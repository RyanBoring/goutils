package rpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/edaniels/golog"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerTracingInterceptor(logger golog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		remoteSpanContext, err := remoteSpanContextFromContext(ctx)
		var span *trace.Span
		if err == nil {
			ctx, span = trace.StartSpanWithRemoteParent(ctx, "server_root", remoteSpanContext)
			defer span.End()
		} else {
			logger.Warnf("client did not send valid Span metadata in headers, local Spans will not be linked to client. reason: %w", err)
		}

		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		if _, ok := status.FromError(err); ok {
			return nil, err
		}
		if s := status.FromContextError(err); s != nil {
			return nil, s.Err()
		}
		return nil, err
	}
}

func StreamServerTracingInterceptor(logger golog.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		remoteSpanContext, err := remoteSpanContextFromContext(stream.Context())
		if err == nil {
			newCtx, span := trace.StartSpanWithRemoteParent(stream.Context(), "server_root", remoteSpanContext)
			defer span.End()
			stream = WrapServerStream(stream, newCtx)
		} else {
			logger.Warnf("client did not send valid Span metadata in headers, local Spans will not be linked to client. reason: %w", err)
		}

		err = handler(srv, stream)
		if err == nil {
			return nil
		}
		if _, ok := status.FromError(err); ok {
			return err
		}
		if s := status.FromContextError(err); s != nil {
			return s.Err()
		}
		return err
	}
}

type ServerStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *ServerStreamWrapper) Context() context.Context {
	return s.ctx
}

func WrapServerStream(stream grpc.ServerStream, ctx context.Context) *ServerStreamWrapper {
	s := ServerStreamWrapper{ServerStream: stream, ctx: ctx}
	return &s
}

func remoteSpanContextFromContext(ctx context.Context) (trace.SpanContext, error) {
	var err error

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return trace.SpanContext{}, errors.New("no metadata in context")
	}

	// Extract trace-id
	traceIDMetadata := md.Get("trace-id")
	if len(traceIDMetadata) == 0 {
		return trace.SpanContext{}, errors.New("trace-id is missing from metadata")
	}

	traceIDBytes, err := hex.DecodeString(traceIDMetadata[0])
	if err != nil {
		return trace.SpanContext{}, fmt.Errorf("trace-id could not be decoded: %w", err)
	}
	var traceID trace.TraceID
	copy(traceID[:], traceIDBytes)

	// Extract span-id
	spanIDMetadata := md.Get("span-id")
	spanIDBytes, err := hex.DecodeString(spanIDMetadata[0])
	if err != nil {
		return trace.SpanContext{}, fmt.Errorf("span-id could not be decoded: %w", err)
	}
	var spanID trace.SpanID
	copy(spanID[:], spanIDBytes)

	// Extract trace-options
	traceOptionsMetadata := md.Get("trace-options")
	if len(traceOptionsMetadata) == 0 {
		return trace.SpanContext{}, errors.New("trace-options is missing from metadata")
	}

	traceOptionsUint, err := strconv.ParseUint(traceOptionsMetadata[0], 10, 32)
	if err != nil {
		return trace.SpanContext{}, fmt.Errorf("trace-options could not be parsed as uint: %w", err)
	}
	traceOptions := trace.TraceOptions(traceOptionsUint)

	return trace.SpanContext{TraceID: traceID, SpanID: spanID, TraceOptions: traceOptions, Tracestate: nil}, nil
}
