package rpcwebrtc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/edaniels/golog"
	"github.com/pion/webrtc/v3"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.viam.com/utils"
	webrtcpb "go.viam.com/utils/proto/rpc/webrtc/v1"
	"go.viam.com/utils/rpc/dialer"
)

// A SignalingServer implements a signaling service for WebRTC by exchanging
// SDPs (https://webrtcforthecurious.com/docs/02-signaling/#what-is-the-session-description-protocol-sdp)
// via gRPC. The service consists of a many-to-many interaction where there are many callers
// and many answerers. The callers provide an SDP to the service which asks a corresponding
// waiting answerer to provide an SDP in exchange in order to establish a P2P connection between
// the two parties.
type SignalingServer struct {
	webrtcpb.UnimplementedSignalingServiceServer
	mu                   sync.RWMutex
	callQueue            CallQueue
	hostICEServers       map[string]hostICEServers
	webrtcConfigProvider ConfigProvider
}

// NewSignalingServer makes a new signaling server that uses the given
// call queue and looks routes based on a given robot host.
func NewSignalingServer(callQueue CallQueue, webrtcConfigProvider ConfigProvider) *SignalingServer {
	return &SignalingServer{
		callQueue:            callQueue,
		hostICEServers:       map[string]hostICEServers{},
		webrtcConfigProvider: webrtcConfigProvider,
	}
}

// RPCHostMetadataField is the identifier of a host.
const RPCHostMetadataField = "rpc-host"

func hostFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md[RPCHostMetadataField]) == 0 {
		return "", fmt.Errorf("expected %s to be set in metadata", RPCHostMetadataField)
	}
	host := md[RPCHostMetadataField][0]
	if host == "" {
		return "", fmt.Errorf("expected non-empty %s", RPCHostMetadataField)
	}
	return host, nil
}

// Call is a request/offer to start a caller with the connected answerer.
func (srv *SignalingServer) Call(ctx context.Context, req *webrtcpb.CallRequest) (*webrtcpb.CallResponse, error) {
	host, err := hostFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	respSDP, err := srv.callQueue.SendOffer(ctx, host, req.Sdp)
	if err != nil {
		return nil, err
	}
	return &webrtcpb.CallResponse{Sdp: respSDP}, nil
}

type hostICEServers struct {
	Servers []*webrtcpb.ICEServer
	Expires time.Time
}

func (srv *SignalingServer) additionalICEServers(ctx context.Context, host string, cache bool) ([]*webrtcpb.ICEServer, error) {
	if srv.webrtcConfigProvider == nil {
		return nil, nil
	}
	srv.mu.RLock()
	hostServers := srv.hostICEServers[host]
	srv.mu.RUnlock()
	if time.Now().Before(hostServers.Expires) {
		return hostServers.Servers, nil
	}
	config, err := srv.webrtcConfigProvider.Config(ctx)
	if err != nil {
		return nil, err
	}
	if cache {
		srv.mu.Lock()
		srv.hostICEServers[host] = hostICEServers{
			Servers: config.ICEServers,
			Expires: config.Expires,
		}
		srv.mu.Unlock()
	}
	return config.ICEServers, nil
}

// Note: We expect but do not enforce one host for one answer. If this is not true, a race
// can happen where we may double fetch additional ICE servers.
func (srv *SignalingServer) clearAdditionalICEServers(host string) {
	srv.mu.Lock()
	delete(srv.hostICEServers, host)
	srv.mu.Unlock()
}

// Answer listens on call/offer queue forever responding with SDPs to agreed to calls.
// TODO(https://github.com/viamrobotics/core/issues/104): This should be authorized for robots only.
// Note: We expect but do not enforce one host for one answer.
func (srv *SignalingServer) Answer(server webrtcpb.SignalingService_AnswerServer) error {
	ctx := server.Context()
	host, err := hostFromCtx(ctx)
	if err != nil {
		return err
	}
	defer srv.clearAdditionalICEServers(host)

	for {
		offer, err := srv.callQueue.RecvOffer(ctx, host)
		if err != nil {
			return err
		}
		iceServers, err := srv.additionalICEServers(ctx, host, true)
		if err != nil {
			return err
		}
		if err := server.Send(&webrtcpb.AnswerRequest{
			Sdp: offer.SDP(),
			OptionalConfig: &webrtcpb.WebRTCConfig{
				AdditionalIceServers: iceServers,
			},
		}); err != nil {
			return err
		}
		answer, err := server.Recv()
		if err != nil {
			return err
		}
		respStatus := status.FromProto(answer.Status)
		var ans CallAnswer
		if respStatus.Code() == codes.OK {
			ans = CallAnswer{SDP: answer.Sdp}
		} else {
			ans = CallAnswer{Err: respStatus.Err()}
		}
		if err := offer.Respond(ctx, ans); err != nil {
			return err
		}
	}
}

// OptionalWebRTCConfig returns any WebRTC configuration the caller may want to use.
func (srv *SignalingServer) OptionalWebRTCConfig(ctx context.Context, req *webrtcpb.OptionalWebRTCConfigRequest) (*webrtcpb.OptionalWebRTCConfigResponse, error) {
	host, err := hostFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	iceServers, err := srv.additionalICEServers(ctx, host, false)
	if err != nil {
		return nil, err
	}
	return &webrtcpb.OptionalWebRTCConfigResponse{Config: &webrtcpb.WebRTCConfig{
		AdditionalIceServers: iceServers,
	}}, nil
}

// A SignalingAnswerer listens for and answers calls with a given signaling service. It is
// directly connected to a Server that will handle the actual calls/connections over WebRTC
// data channels.
type SignalingAnswerer struct {
	address                 string
	host                    string
	client                  webrtcpb.SignalingService_AnswerClient
	server                  *Server
	insecure                bool
	webrtcConfig            webrtc.Configuration
	activeBackgroundWorkers sync.WaitGroup
	cancelBackgroundWorkers func()
	closeCtx                context.Context
	logger                  golog.Logger
}

// NewSignalingAnswerer makes an answerer that will connect to and listen for calls at the given
// address. Note that using this assumes that the connection at the given address is secure and
// assumed that all calls are authenticated. Random ports will be opened on this host to establish
// connections as a means to service ICE (https://webrtcforthecurious.com/docs/03-connecting/#how-does-it-work).
func NewSignalingAnswerer(address, host string, server *Server, insecure bool, webrtcConfig webrtc.Configuration, logger golog.Logger) *SignalingAnswerer {
	closeCtx, cancel := context.WithCancel(context.Background())
	return &SignalingAnswerer{
		address:                 address,
		host:                    host,
		server:                  server,
		insecure:                insecure,
		webrtcConfig:            webrtcConfig,
		cancelBackgroundWorkers: cancel,
		closeCtx:                closeCtx,
		logger:                  logger,
	}
}

const answererReconnectWait = time.Second

// Start connects to the signaling service and listens forever until instructed to stop
// via Stop.
func (ans *SignalingAnswerer) Start() error {
	var connInUse dialer.ClientConn
	var connMu sync.Mutex
	connect := func() error {
		connMu.Lock()
		conn := connInUse
		connMu.Unlock()
		if conn != nil {
			if err := conn.Close(); err != nil {
				ans.logger.Errorw("error closing existing signaling connection", "error", err)
			}
		}
		setupCtx, timeoutCancel := context.WithTimeout(ans.closeCtx, 5*time.Second)
		defer timeoutCancel()
		conn, err := dialer.DialDirectGRPC(setupCtx, ans.address, ans.insecure)
		if err != nil {
			return err
		}
		connMu.Lock()
		connInUse = conn
		connMu.Unlock()

		client := webrtcpb.NewSignalingServiceClient(conn)
		md := metadata.New(map[string]string{RPCHostMetadataField: ans.host})
		answerCtx := metadata.NewOutgoingContext(ans.closeCtx, md)
		answerClient, err := client.Answer(answerCtx)
		if err != nil {
			return multierr.Combine(err, conn.Close())
		}
		ans.client = answerClient
		return nil
	}

	ans.activeBackgroundWorkers.Add(1)
	utils.ManagedGo(func() {
		for {
			select {
			case <-ans.closeCtx.Done():
				return
			default:
			}
			if err := ans.answer(); err != nil && utils.FilterOutError(err, context.Canceled) != nil {
				shouldLogError := false
				if _, isGRPCErr := status.FromError(err); !(isGRPCErr || errors.Is(err, errSignalingAnswererDisconnected)) {
					shouldLogError = true
				}
				if shouldLogError {
					ans.logger.Errorw("error answering", "error", err)
				}
				for {
					if shouldLogError {
						ans.logger.Debugw("reconnecting answer client", "in", answererReconnectWait.String())
					}
					if !utils.SelectContextOrWait(ans.closeCtx, answererReconnectWait) {
						return
					}
					if connectErr := connect(); connectErr != nil {
						ans.logger.Errorw("error reconnecting answer client", "error", err, "reconnect_err", connectErr)
						continue
					}
					if shouldLogError {
						ans.logger.Debug("reconnected answer client")
					}
					break
				}
			}
		}
	}, func() {
		defer ans.activeBackgroundWorkers.Done()
		defer func() {
			connMu.Lock()
			conn := connInUse
			connMu.Unlock()
			if conn == nil {
				return
			}
			if err := conn.Close(); err != nil {
				ans.logger.Errorw("error closing signaling connection", "error", err)
			}
		}()
		defer func() {
			if ans.client == nil {
				return
			}
			if err := ans.client.CloseSend(); err != nil {
				ans.logger.Errorw("error closing send side of answering client", "error", err)
			}
		}()
	})

	return nil
}

// Stop waits for the answer to stop listening and return.
func (ans *SignalingAnswerer) Stop() {
	ans.cancelBackgroundWorkers()
	ans.activeBackgroundWorkers.Wait()
}

var errSignalingAnswererDisconnected = errors.New("signaling answerer disconnected")

// answer accepts a single call offer, responds with a corresponding SDP, and
// attempts to establish a WebRTC connection with the caller via ICE. Once established,
// the designated WebRTC data channel is passed off to the underlying Server which
// is then used as the server end of a gRPC connection.
func (ans *SignalingAnswerer) answer() (err error) {
	if ans.client == nil {
		return errSignalingAnswererDisconnected
	}
	resp, err := ans.client.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	extendedConfig := extendWebRTCConfig(&ans.webrtcConfig, resp.OptionalConfig)
	pc, dc, err := newPeerConnectionForServer(ans.closeCtx, resp.Sdp, extendedConfig, ans.logger)
	if err != nil {
		return ans.client.Send(&webrtcpb.AnswerResponse{
			Status: ErrorToStatus(err).Proto(),
		})
	}
	var successful bool
	defer func() {
		if !(successful && err == nil) {
			err = multierr.Combine(err, pc.Close())
		}
	}()

	encodedSDP, err := EncodeSDP(pc.LocalDescription())
	if err != nil {
		return ans.client.Send(&webrtcpb.AnswerResponse{
			Status: ErrorToStatus(err).Proto(),
		})
	}

	ans.server.NewChannel(pc, dc)

	successful = true
	return ans.client.Send(&webrtcpb.AnswerResponse{
		Status: ErrorToStatus(nil).Proto(),
		Sdp:    encodedSDP,
	})
}

// Adapted from https://github.com/pion/webrtc/blob/master/examples/internal/signal/signal.go

// EncodeSDP encodes the given SDP in base64.
func EncodeSDP(sdp *webrtc.SessionDescription) (string, error) {
	b, err := json.Marshal(sdp)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// DecodeSDP decodes the input from base64 into the given SDP.
func DecodeSDP(in string, sdp *webrtc.SessionDescription) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, sdp)
}

func extendWebRTCConfig(original *webrtc.Configuration, optional *webrtcpb.WebRTCConfig) webrtc.Configuration {
	configCopy := *original
	if len(optional.AdditionalIceServers) > 0 {
		iceServers := make([]webrtc.ICEServer, len(original.ICEServers)+len(optional.AdditionalIceServers))
		copy(iceServers, original.ICEServers)
		for _, server := range optional.AdditionalIceServers {
			iceServers = append(iceServers, webrtc.ICEServer{
				URLs:       server.Urls,
				Username:   server.Username,
				Credential: server.Credential,
			})
		}
		configCopy.ICEServers = iceServers
	}
	return configCopy
}
