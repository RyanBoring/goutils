// package: proto.rpc.webrtc.v1
// file: proto/rpc/webrtc/v1/signaling.proto

import * as jspb from "google-protobuf";
import * as google_api_annotations_pb from "../../../../google/api/annotations_pb";
import * as google_rpc_status_pb from "../../../../google/rpc/status_pb";

export class CallRequest extends jspb.Message {
  getSdp(): string;
  setSdp(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CallRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CallRequest): CallRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CallRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CallRequest;
  static deserializeBinaryFromReader(message: CallRequest, reader: jspb.BinaryReader): CallRequest;
}

export namespace CallRequest {
  export type AsObject = {
    sdp: string,
  }
}

export class CallResponse extends jspb.Message {
  getSdp(): string;
  setSdp(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CallResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CallResponse): CallResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CallResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CallResponse;
  static deserializeBinaryFromReader(message: CallResponse, reader: jspb.BinaryReader): CallResponse;
}

export namespace CallResponse {
  export type AsObject = {
    sdp: string,
  }
}

export class ICEServer extends jspb.Message {
  clearUrlsList(): void;
  getUrlsList(): Array<string>;
  setUrlsList(value: Array<string>): void;
  addUrls(value: string, index?: number): string;

  getUsername(): string;
  setUsername(value: string): void;

  getCredential(): string;
  setCredential(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ICEServer.AsObject;
  static toObject(includeInstance: boolean, msg: ICEServer): ICEServer.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ICEServer, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ICEServer;
  static deserializeBinaryFromReader(message: ICEServer, reader: jspb.BinaryReader): ICEServer;
}

export namespace ICEServer {
  export type AsObject = {
    urlsList: Array<string>,
    username: string,
    credential: string,
  }
}

export class WebRTCConfig extends jspb.Message {
  clearAdditionalIceServersList(): void;
  getAdditionalIceServersList(): Array<ICEServer>;
  setAdditionalIceServersList(value: Array<ICEServer>): void;
  addAdditionalIceServers(value?: ICEServer, index?: number): ICEServer;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WebRTCConfig.AsObject;
  static toObject(includeInstance: boolean, msg: WebRTCConfig): WebRTCConfig.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WebRTCConfig, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WebRTCConfig;
  static deserializeBinaryFromReader(message: WebRTCConfig, reader: jspb.BinaryReader): WebRTCConfig;
}

export namespace WebRTCConfig {
  export type AsObject = {
    additionalIceServersList: Array<ICEServer.AsObject>,
  }
}

export class AnswerRequest extends jspb.Message {
  getSdp(): string;
  setSdp(value: string): void;

  hasOptionalConfig(): boolean;
  clearOptionalConfig(): void;
  getOptionalConfig(): WebRTCConfig | undefined;
  setOptionalConfig(value?: WebRTCConfig): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AnswerRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AnswerRequest): AnswerRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AnswerRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AnswerRequest;
  static deserializeBinaryFromReader(message: AnswerRequest, reader: jspb.BinaryReader): AnswerRequest;
}

export namespace AnswerRequest {
  export type AsObject = {
    sdp: string,
    optionalConfig?: WebRTCConfig.AsObject,
  }
}

export class AnswerResponse extends jspb.Message {
  hasStatus(): boolean;
  clearStatus(): void;
  getStatus(): google_rpc_status_pb.Status | undefined;
  setStatus(value?: google_rpc_status_pb.Status): void;

  getSdp(): string;
  setSdp(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AnswerResponse.AsObject;
  static toObject(includeInstance: boolean, msg: AnswerResponse): AnswerResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AnswerResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AnswerResponse;
  static deserializeBinaryFromReader(message: AnswerResponse, reader: jspb.BinaryReader): AnswerResponse;
}

export namespace AnswerResponse {
  export type AsObject = {
    status?: google_rpc_status_pb.Status.AsObject,
    sdp: string,
  }
}

export class OptionalWebRTCConfigRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OptionalWebRTCConfigRequest.AsObject;
  static toObject(includeInstance: boolean, msg: OptionalWebRTCConfigRequest): OptionalWebRTCConfigRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: OptionalWebRTCConfigRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OptionalWebRTCConfigRequest;
  static deserializeBinaryFromReader(message: OptionalWebRTCConfigRequest, reader: jspb.BinaryReader): OptionalWebRTCConfigRequest;
}

export namespace OptionalWebRTCConfigRequest {
  export type AsObject = {
  }
}

export class OptionalWebRTCConfigResponse extends jspb.Message {
  hasConfig(): boolean;
  clearConfig(): void;
  getConfig(): WebRTCConfig | undefined;
  setConfig(value?: WebRTCConfig): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OptionalWebRTCConfigResponse.AsObject;
  static toObject(includeInstance: boolean, msg: OptionalWebRTCConfigResponse): OptionalWebRTCConfigResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: OptionalWebRTCConfigResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OptionalWebRTCConfigResponse;
  static deserializeBinaryFromReader(message: OptionalWebRTCConfigResponse, reader: jspb.BinaryReader): OptionalWebRTCConfigResponse;
}

export namespace OptionalWebRTCConfigResponse {
  export type AsObject = {
    config?: WebRTCConfig.AsObject,
  }
}

