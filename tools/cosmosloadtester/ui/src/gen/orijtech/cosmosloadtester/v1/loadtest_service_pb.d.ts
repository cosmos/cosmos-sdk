import * as jspb from 'google-protobuf'

import * as google_api_annotations_pb from '../../../google/api/annotations_pb';
import * as google_protobuf_duration_pb from 'google-protobuf/google/protobuf/duration_pb';


export class RunLoadtestRequest extends jspb.Message {
  getClientFactory(): string;
  setClientFactory(value: string): RunLoadtestRequest;

  getConnectionCount(): number;
  setConnectionCount(value: number): RunLoadtestRequest;

  getDuration(): google_protobuf_duration_pb.Duration | undefined;
  setDuration(value?: google_protobuf_duration_pb.Duration): RunLoadtestRequest;
  hasDuration(): boolean;
  clearDuration(): RunLoadtestRequest;

  getSendPeriod(): google_protobuf_duration_pb.Duration | undefined;
  setSendPeriod(value?: google_protobuf_duration_pb.Duration): RunLoadtestRequest;
  hasSendPeriod(): boolean;
  clearSendPeriod(): RunLoadtestRequest;

  getTransactionsPerSecond(): number;
  setTransactionsPerSecond(value: number): RunLoadtestRequest;

  getTransactionSizeBytes(): number;
  setTransactionSizeBytes(value: number): RunLoadtestRequest;

  getTransactionCount(): number;
  setTransactionCount(value: number): RunLoadtestRequest;

  getBroadcastTxMethod(): RunLoadtestRequest.BroadcastTxMethod;
  setBroadcastTxMethod(value: RunLoadtestRequest.BroadcastTxMethod): RunLoadtestRequest;

  getEndpointsList(): Array<string>;
  setEndpointsList(value: Array<string>): RunLoadtestRequest;
  clearEndpointsList(): RunLoadtestRequest;
  addEndpoints(value: string, index?: number): RunLoadtestRequest;

  getEndpointSelectMethod(): RunLoadtestRequest.EndpointSelectMethod;
  setEndpointSelectMethod(value: RunLoadtestRequest.EndpointSelectMethod): RunLoadtestRequest;

  getExpectPeersCount(): number;
  setExpectPeersCount(value: number): RunLoadtestRequest;

  getMaxEndpointCount(): number;
  setMaxEndpointCount(value: number): RunLoadtestRequest;

  getPeerConnectTimeout(): google_protobuf_duration_pb.Duration | undefined;
  setPeerConnectTimeout(value?: google_protobuf_duration_pb.Duration): RunLoadtestRequest;
  hasPeerConnectTimeout(): boolean;
  clearPeerConnectTimeout(): RunLoadtestRequest;

  getMinPeerConnectivityCount(): number;
  setMinPeerConnectivityCount(value: number): RunLoadtestRequest;

  getStatsOutputFilePath(): string;
  setStatsOutputFilePath(value: string): RunLoadtestRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RunLoadtestRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RunLoadtestRequest): RunLoadtestRequest.AsObject;
  static serializeBinaryToWriter(message: RunLoadtestRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RunLoadtestRequest;
  static deserializeBinaryFromReader(message: RunLoadtestRequest, reader: jspb.BinaryReader): RunLoadtestRequest;
}

export namespace RunLoadtestRequest {
  export type AsObject = {
    clientFactory: string,
    connectionCount: number,
    duration?: google_protobuf_duration_pb.Duration.AsObject,
    sendPeriod?: google_protobuf_duration_pb.Duration.AsObject,
    transactionsPerSecond: number,
    transactionSizeBytes: number,
    transactionCount: number,
    broadcastTxMethod: RunLoadtestRequest.BroadcastTxMethod,
    endpointsList: Array<string>,
    endpointSelectMethod: RunLoadtestRequest.EndpointSelectMethod,
    expectPeersCount: number,
    maxEndpointCount: number,
    peerConnectTimeout?: google_protobuf_duration_pb.Duration.AsObject,
    minPeerConnectivityCount: number,
    statsOutputFilePath: string,
  }

  export enum BroadcastTxMethod { 
    BROADCAST_TX_METHOD_UNSPECIFIED = 0,
    BROADCAST_TX_METHOD_SYNC = 1,
    BROADCAST_TX_METHOD_ASYNC = 2,
    BROADCAST_TX_METHOD_COMMIT = 3,
  }

  export enum EndpointSelectMethod { 
    ENDPOINT_SELECT_METHOD_UNSPECIFIED = 0,
    ENDPOINT_SELECT_METHOD_SUPPLIED = 1,
    ENDPOINT_SELECT_METHOD_DISCOVERED = 2,
    ENDPOINT_SELECT_METHOD_ANY = 3,
  }
}

export class RunLoadtestResponse extends jspb.Message {
  getTotalTxs(): number;
  setTotalTxs(value: number): RunLoadtestResponse;

  getTotalTime(): google_protobuf_duration_pb.Duration | undefined;
  setTotalTime(value?: google_protobuf_duration_pb.Duration): RunLoadtestResponse;
  hasTotalTime(): boolean;
  clearTotalTime(): RunLoadtestResponse;

  getTotalBytes(): number;
  setTotalBytes(value: number): RunLoadtestResponse;

  getAvgTxsPerSecond(): number;
  setAvgTxsPerSecond(value: number): RunLoadtestResponse;

  getAvgBytesPerSecond(): number;
  setAvgBytesPerSecond(value: number): RunLoadtestResponse;

  getPerSecList(): Array<PerSecond>;
  setPerSecList(value: Array<PerSecond>): RunLoadtestResponse;
  clearPerSecList(): RunLoadtestResponse;
  addPerSec(value?: PerSecond, index?: number): PerSecond;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RunLoadtestResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RunLoadtestResponse): RunLoadtestResponse.AsObject;
  static serializeBinaryToWriter(message: RunLoadtestResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RunLoadtestResponse;
  static deserializeBinaryFromReader(message: RunLoadtestResponse, reader: jspb.BinaryReader): RunLoadtestResponse;
}

export namespace RunLoadtestResponse {
  export type AsObject = {
    totalTxs: number,
    totalTime?: google_protobuf_duration_pb.Duration.AsObject,
    totalBytes: number,
    avgTxsPerSecond: number,
    avgBytesPerSecond: number,
    perSecList: Array<PerSecond.AsObject>,
  }
}

export class PerSecond extends jspb.Message {
  getSec(): number;
  setSec(value: number): PerSecond;

  getQps(): number;
  setQps(value: number): PerSecond;

  getBytesSent(): number;
  setBytesSent(value: number): PerSecond;

  getBytesRankings(): Ranking | undefined;
  setBytesRankings(value?: Ranking): PerSecond;
  hasBytesRankings(): boolean;
  clearBytesRankings(): PerSecond;

  getLatencyRankings(): Ranking | undefined;
  setLatencyRankings(value?: Ranking): PerSecond;
  hasLatencyRankings(): boolean;
  clearLatencyRankings(): PerSecond;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PerSecond.AsObject;
  static toObject(includeInstance: boolean, msg: PerSecond): PerSecond.AsObject;
  static serializeBinaryToWriter(message: PerSecond, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PerSecond;
  static deserializeBinaryFromReader(message: PerSecond, reader: jspb.BinaryReader): PerSecond;
}

export namespace PerSecond {
  export type AsObject = {
    sec: number,
    qps: number,
    bytesSent: number,
    bytesRankings?: Ranking.AsObject,
    latencyRankings?: Ranking.AsObject,
  }
}

export class Percentile extends jspb.Message {
  getStartOffset(): google_protobuf_duration_pb.Duration | undefined;
  setStartOffset(value?: google_protobuf_duration_pb.Duration): Percentile;
  hasStartOffset(): boolean;
  clearStartOffset(): Percentile;

  getLatency(): google_protobuf_duration_pb.Duration | undefined;
  setLatency(value?: google_protobuf_duration_pb.Duration): Percentile;
  hasLatency(): boolean;
  clearLatency(): Percentile;

  getBytesSent(): number;
  setBytesSent(value: number): Percentile;

  getAtStr(): string;
  setAtStr(value: string): Percentile;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Percentile.AsObject;
  static toObject(includeInstance: boolean, msg: Percentile): Percentile.AsObject;
  static serializeBinaryToWriter(message: Percentile, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Percentile;
  static deserializeBinaryFromReader(message: Percentile, reader: jspb.BinaryReader): Percentile;
}

export namespace Percentile {
  export type AsObject = {
    startOffset?: google_protobuf_duration_pb.Duration.AsObject,
    latency?: google_protobuf_duration_pb.Duration.AsObject,
    bytesSent: number,
    atStr: string,
  }
}

export class Ranking extends jspb.Message {
  getP50(): Percentile | undefined;
  setP50(value?: Percentile): Ranking;
  hasP50(): boolean;
  clearP50(): Ranking;

  getP75(): Percentile | undefined;
  setP75(value?: Percentile): Ranking;
  hasP75(): boolean;
  clearP75(): Ranking;

  getP90(): Percentile | undefined;
  setP90(value?: Percentile): Ranking;
  hasP90(): boolean;
  clearP90(): Ranking;

  getP95(): Percentile | undefined;
  setP95(value?: Percentile): Ranking;
  hasP95(): boolean;
  clearP95(): Ranking;

  getP99(): Percentile | undefined;
  setP99(value?: Percentile): Ranking;
  hasP99(): boolean;
  clearP99(): Ranking;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Ranking.AsObject;
  static toObject(includeInstance: boolean, msg: Ranking): Ranking.AsObject;
  static serializeBinaryToWriter(message: Ranking, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Ranking;
  static deserializeBinaryFromReader(message: Ranking, reader: jspb.BinaryReader): Ranking;
}

export namespace Ranking {
  export type AsObject = {
    p50?: Percentile.AsObject,
    p75?: Percentile.AsObject,
    p90?: Percentile.AsObject,
    p95?: Percentile.AsObject,
    p99?: Percentile.AsObject,
  }
}

