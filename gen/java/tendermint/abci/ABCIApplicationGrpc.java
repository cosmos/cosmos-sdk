package tendermint.abci;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: tendermint/abci/types.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ABCIApplicationGrpc {

  private ABCIApplicationGrpc() {}

  public static final String SERVICE_NAME = "tendermint.abci.ABCIApplication";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestEcho,
      tendermint.abci.Types.ResponseEcho> getEchoMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Echo",
      requestType = tendermint.abci.Types.RequestEcho.class,
      responseType = tendermint.abci.Types.ResponseEcho.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestEcho,
      tendermint.abci.Types.ResponseEcho> getEchoMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestEcho, tendermint.abci.Types.ResponseEcho> getEchoMethod;
    if ((getEchoMethod = ABCIApplicationGrpc.getEchoMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getEchoMethod = ABCIApplicationGrpc.getEchoMethod) == null) {
          ABCIApplicationGrpc.getEchoMethod = getEchoMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestEcho, tendermint.abci.Types.ResponseEcho>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Echo"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestEcho.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseEcho.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("Echo"))
              .build();
        }
      }
    }
    return getEchoMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestFlush,
      tendermint.abci.Types.ResponseFlush> getFlushMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Flush",
      requestType = tendermint.abci.Types.RequestFlush.class,
      responseType = tendermint.abci.Types.ResponseFlush.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestFlush,
      tendermint.abci.Types.ResponseFlush> getFlushMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestFlush, tendermint.abci.Types.ResponseFlush> getFlushMethod;
    if ((getFlushMethod = ABCIApplicationGrpc.getFlushMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getFlushMethod = ABCIApplicationGrpc.getFlushMethod) == null) {
          ABCIApplicationGrpc.getFlushMethod = getFlushMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestFlush, tendermint.abci.Types.ResponseFlush>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Flush"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestFlush.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseFlush.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("Flush"))
              .build();
        }
      }
    }
    return getFlushMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestInfo,
      tendermint.abci.Types.ResponseInfo> getInfoMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Info",
      requestType = tendermint.abci.Types.RequestInfo.class,
      responseType = tendermint.abci.Types.ResponseInfo.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestInfo,
      tendermint.abci.Types.ResponseInfo> getInfoMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestInfo, tendermint.abci.Types.ResponseInfo> getInfoMethod;
    if ((getInfoMethod = ABCIApplicationGrpc.getInfoMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getInfoMethod = ABCIApplicationGrpc.getInfoMethod) == null) {
          ABCIApplicationGrpc.getInfoMethod = getInfoMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestInfo, tendermint.abci.Types.ResponseInfo>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Info"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestInfo.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseInfo.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("Info"))
              .build();
        }
      }
    }
    return getInfoMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestSetOption,
      tendermint.abci.Types.ResponseSetOption> getSetOptionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SetOption",
      requestType = tendermint.abci.Types.RequestSetOption.class,
      responseType = tendermint.abci.Types.ResponseSetOption.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestSetOption,
      tendermint.abci.Types.ResponseSetOption> getSetOptionMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestSetOption, tendermint.abci.Types.ResponseSetOption> getSetOptionMethod;
    if ((getSetOptionMethod = ABCIApplicationGrpc.getSetOptionMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getSetOptionMethod = ABCIApplicationGrpc.getSetOptionMethod) == null) {
          ABCIApplicationGrpc.getSetOptionMethod = getSetOptionMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestSetOption, tendermint.abci.Types.ResponseSetOption>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SetOption"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestSetOption.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseSetOption.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("SetOption"))
              .build();
        }
      }
    }
    return getSetOptionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestDeliverTx,
      tendermint.abci.Types.ResponseDeliverTx> getDeliverTxMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DeliverTx",
      requestType = tendermint.abci.Types.RequestDeliverTx.class,
      responseType = tendermint.abci.Types.ResponseDeliverTx.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestDeliverTx,
      tendermint.abci.Types.ResponseDeliverTx> getDeliverTxMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestDeliverTx, tendermint.abci.Types.ResponseDeliverTx> getDeliverTxMethod;
    if ((getDeliverTxMethod = ABCIApplicationGrpc.getDeliverTxMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getDeliverTxMethod = ABCIApplicationGrpc.getDeliverTxMethod) == null) {
          ABCIApplicationGrpc.getDeliverTxMethod = getDeliverTxMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestDeliverTx, tendermint.abci.Types.ResponseDeliverTx>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DeliverTx"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestDeliverTx.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseDeliverTx.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("DeliverTx"))
              .build();
        }
      }
    }
    return getDeliverTxMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestCheckTx,
      tendermint.abci.Types.ResponseCheckTx> getCheckTxMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CheckTx",
      requestType = tendermint.abci.Types.RequestCheckTx.class,
      responseType = tendermint.abci.Types.ResponseCheckTx.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestCheckTx,
      tendermint.abci.Types.ResponseCheckTx> getCheckTxMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestCheckTx, tendermint.abci.Types.ResponseCheckTx> getCheckTxMethod;
    if ((getCheckTxMethod = ABCIApplicationGrpc.getCheckTxMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getCheckTxMethod = ABCIApplicationGrpc.getCheckTxMethod) == null) {
          ABCIApplicationGrpc.getCheckTxMethod = getCheckTxMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestCheckTx, tendermint.abci.Types.ResponseCheckTx>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CheckTx"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestCheckTx.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseCheckTx.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("CheckTx"))
              .build();
        }
      }
    }
    return getCheckTxMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestQuery,
      tendermint.abci.Types.ResponseQuery> getQueryMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Query",
      requestType = tendermint.abci.Types.RequestQuery.class,
      responseType = tendermint.abci.Types.ResponseQuery.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestQuery,
      tendermint.abci.Types.ResponseQuery> getQueryMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestQuery, tendermint.abci.Types.ResponseQuery> getQueryMethod;
    if ((getQueryMethod = ABCIApplicationGrpc.getQueryMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getQueryMethod = ABCIApplicationGrpc.getQueryMethod) == null) {
          ABCIApplicationGrpc.getQueryMethod = getQueryMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestQuery, tendermint.abci.Types.ResponseQuery>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Query"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestQuery.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseQuery.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("Query"))
              .build();
        }
      }
    }
    return getQueryMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestCommit,
      tendermint.abci.Types.ResponseCommit> getCommitMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Commit",
      requestType = tendermint.abci.Types.RequestCommit.class,
      responseType = tendermint.abci.Types.ResponseCommit.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestCommit,
      tendermint.abci.Types.ResponseCommit> getCommitMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestCommit, tendermint.abci.Types.ResponseCommit> getCommitMethod;
    if ((getCommitMethod = ABCIApplicationGrpc.getCommitMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getCommitMethod = ABCIApplicationGrpc.getCommitMethod) == null) {
          ABCIApplicationGrpc.getCommitMethod = getCommitMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestCommit, tendermint.abci.Types.ResponseCommit>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Commit"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestCommit.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseCommit.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("Commit"))
              .build();
        }
      }
    }
    return getCommitMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestInitChain,
      tendermint.abci.Types.ResponseInitChain> getInitChainMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "InitChain",
      requestType = tendermint.abci.Types.RequestInitChain.class,
      responseType = tendermint.abci.Types.ResponseInitChain.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestInitChain,
      tendermint.abci.Types.ResponseInitChain> getInitChainMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestInitChain, tendermint.abci.Types.ResponseInitChain> getInitChainMethod;
    if ((getInitChainMethod = ABCIApplicationGrpc.getInitChainMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getInitChainMethod = ABCIApplicationGrpc.getInitChainMethod) == null) {
          ABCIApplicationGrpc.getInitChainMethod = getInitChainMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestInitChain, tendermint.abci.Types.ResponseInitChain>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "InitChain"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestInitChain.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseInitChain.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("InitChain"))
              .build();
        }
      }
    }
    return getInitChainMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestBeginBlock,
      tendermint.abci.Types.ResponseBeginBlock> getBeginBlockMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "BeginBlock",
      requestType = tendermint.abci.Types.RequestBeginBlock.class,
      responseType = tendermint.abci.Types.ResponseBeginBlock.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestBeginBlock,
      tendermint.abci.Types.ResponseBeginBlock> getBeginBlockMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestBeginBlock, tendermint.abci.Types.ResponseBeginBlock> getBeginBlockMethod;
    if ((getBeginBlockMethod = ABCIApplicationGrpc.getBeginBlockMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getBeginBlockMethod = ABCIApplicationGrpc.getBeginBlockMethod) == null) {
          ABCIApplicationGrpc.getBeginBlockMethod = getBeginBlockMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestBeginBlock, tendermint.abci.Types.ResponseBeginBlock>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "BeginBlock"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestBeginBlock.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseBeginBlock.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("BeginBlock"))
              .build();
        }
      }
    }
    return getBeginBlockMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestEndBlock,
      tendermint.abci.Types.ResponseEndBlock> getEndBlockMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "EndBlock",
      requestType = tendermint.abci.Types.RequestEndBlock.class,
      responseType = tendermint.abci.Types.ResponseEndBlock.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestEndBlock,
      tendermint.abci.Types.ResponseEndBlock> getEndBlockMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestEndBlock, tendermint.abci.Types.ResponseEndBlock> getEndBlockMethod;
    if ((getEndBlockMethod = ABCIApplicationGrpc.getEndBlockMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getEndBlockMethod = ABCIApplicationGrpc.getEndBlockMethod) == null) {
          ABCIApplicationGrpc.getEndBlockMethod = getEndBlockMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestEndBlock, tendermint.abci.Types.ResponseEndBlock>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "EndBlock"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestEndBlock.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseEndBlock.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("EndBlock"))
              .build();
        }
      }
    }
    return getEndBlockMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestListSnapshots,
      tendermint.abci.Types.ResponseListSnapshots> getListSnapshotsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListSnapshots",
      requestType = tendermint.abci.Types.RequestListSnapshots.class,
      responseType = tendermint.abci.Types.ResponseListSnapshots.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestListSnapshots,
      tendermint.abci.Types.ResponseListSnapshots> getListSnapshotsMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestListSnapshots, tendermint.abci.Types.ResponseListSnapshots> getListSnapshotsMethod;
    if ((getListSnapshotsMethod = ABCIApplicationGrpc.getListSnapshotsMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getListSnapshotsMethod = ABCIApplicationGrpc.getListSnapshotsMethod) == null) {
          ABCIApplicationGrpc.getListSnapshotsMethod = getListSnapshotsMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestListSnapshots, tendermint.abci.Types.ResponseListSnapshots>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListSnapshots"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestListSnapshots.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseListSnapshots.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("ListSnapshots"))
              .build();
        }
      }
    }
    return getListSnapshotsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestOfferSnapshot,
      tendermint.abci.Types.ResponseOfferSnapshot> getOfferSnapshotMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "OfferSnapshot",
      requestType = tendermint.abci.Types.RequestOfferSnapshot.class,
      responseType = tendermint.abci.Types.ResponseOfferSnapshot.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestOfferSnapshot,
      tendermint.abci.Types.ResponseOfferSnapshot> getOfferSnapshotMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestOfferSnapshot, tendermint.abci.Types.ResponseOfferSnapshot> getOfferSnapshotMethod;
    if ((getOfferSnapshotMethod = ABCIApplicationGrpc.getOfferSnapshotMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getOfferSnapshotMethod = ABCIApplicationGrpc.getOfferSnapshotMethod) == null) {
          ABCIApplicationGrpc.getOfferSnapshotMethod = getOfferSnapshotMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestOfferSnapshot, tendermint.abci.Types.ResponseOfferSnapshot>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "OfferSnapshot"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestOfferSnapshot.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseOfferSnapshot.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("OfferSnapshot"))
              .build();
        }
      }
    }
    return getOfferSnapshotMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestLoadSnapshotChunk,
      tendermint.abci.Types.ResponseLoadSnapshotChunk> getLoadSnapshotChunkMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "LoadSnapshotChunk",
      requestType = tendermint.abci.Types.RequestLoadSnapshotChunk.class,
      responseType = tendermint.abci.Types.ResponseLoadSnapshotChunk.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestLoadSnapshotChunk,
      tendermint.abci.Types.ResponseLoadSnapshotChunk> getLoadSnapshotChunkMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestLoadSnapshotChunk, tendermint.abci.Types.ResponseLoadSnapshotChunk> getLoadSnapshotChunkMethod;
    if ((getLoadSnapshotChunkMethod = ABCIApplicationGrpc.getLoadSnapshotChunkMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getLoadSnapshotChunkMethod = ABCIApplicationGrpc.getLoadSnapshotChunkMethod) == null) {
          ABCIApplicationGrpc.getLoadSnapshotChunkMethod = getLoadSnapshotChunkMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestLoadSnapshotChunk, tendermint.abci.Types.ResponseLoadSnapshotChunk>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "LoadSnapshotChunk"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestLoadSnapshotChunk.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseLoadSnapshotChunk.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("LoadSnapshotChunk"))
              .build();
        }
      }
    }
    return getLoadSnapshotChunkMethod;
  }

  private static volatile io.grpc.MethodDescriptor<tendermint.abci.Types.RequestApplySnapshotChunk,
      tendermint.abci.Types.ResponseApplySnapshotChunk> getApplySnapshotChunkMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ApplySnapshotChunk",
      requestType = tendermint.abci.Types.RequestApplySnapshotChunk.class,
      responseType = tendermint.abci.Types.ResponseApplySnapshotChunk.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<tendermint.abci.Types.RequestApplySnapshotChunk,
      tendermint.abci.Types.ResponseApplySnapshotChunk> getApplySnapshotChunkMethod() {
    io.grpc.MethodDescriptor<tendermint.abci.Types.RequestApplySnapshotChunk, tendermint.abci.Types.ResponseApplySnapshotChunk> getApplySnapshotChunkMethod;
    if ((getApplySnapshotChunkMethod = ABCIApplicationGrpc.getApplySnapshotChunkMethod) == null) {
      synchronized (ABCIApplicationGrpc.class) {
        if ((getApplySnapshotChunkMethod = ABCIApplicationGrpc.getApplySnapshotChunkMethod) == null) {
          ABCIApplicationGrpc.getApplySnapshotChunkMethod = getApplySnapshotChunkMethod =
              io.grpc.MethodDescriptor.<tendermint.abci.Types.RequestApplySnapshotChunk, tendermint.abci.Types.ResponseApplySnapshotChunk>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ApplySnapshotChunk"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.RequestApplySnapshotChunk.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  tendermint.abci.Types.ResponseApplySnapshotChunk.getDefaultInstance()))
              .setSchemaDescriptor(new ABCIApplicationMethodDescriptorSupplier("ApplySnapshotChunk"))
              .build();
        }
      }
    }
    return getApplySnapshotChunkMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ABCIApplicationStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ABCIApplicationStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ABCIApplicationStub>() {
        @java.lang.Override
        public ABCIApplicationStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ABCIApplicationStub(channel, callOptions);
        }
      };
    return ABCIApplicationStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ABCIApplicationBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ABCIApplicationBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ABCIApplicationBlockingStub>() {
        @java.lang.Override
        public ABCIApplicationBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ABCIApplicationBlockingStub(channel, callOptions);
        }
      };
    return ABCIApplicationBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ABCIApplicationFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ABCIApplicationFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ABCIApplicationFutureStub>() {
        @java.lang.Override
        public ABCIApplicationFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ABCIApplicationFutureStub(channel, callOptions);
        }
      };
    return ABCIApplicationFutureStub.newStub(factory, channel);
  }

  /**
   */
  public static abstract class ABCIApplicationImplBase implements io.grpc.BindableService {

    /**
     */
    public void echo(tendermint.abci.Types.RequestEcho request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseEcho> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getEchoMethod(), responseObserver);
    }

    /**
     */
    public void flush(tendermint.abci.Types.RequestFlush request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseFlush> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getFlushMethod(), responseObserver);
    }

    /**
     */
    public void info(tendermint.abci.Types.RequestInfo request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseInfo> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getInfoMethod(), responseObserver);
    }

    /**
     */
    public void setOption(tendermint.abci.Types.RequestSetOption request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseSetOption> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSetOptionMethod(), responseObserver);
    }

    /**
     */
    public void deliverTx(tendermint.abci.Types.RequestDeliverTx request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseDeliverTx> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDeliverTxMethod(), responseObserver);
    }

    /**
     */
    public void checkTx(tendermint.abci.Types.RequestCheckTx request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseCheckTx> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCheckTxMethod(), responseObserver);
    }

    /**
     */
    public void query(tendermint.abci.Types.RequestQuery request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseQuery> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getQueryMethod(), responseObserver);
    }

    /**
     */
    public void commit(tendermint.abci.Types.RequestCommit request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseCommit> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCommitMethod(), responseObserver);
    }

    /**
     */
    public void initChain(tendermint.abci.Types.RequestInitChain request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseInitChain> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getInitChainMethod(), responseObserver);
    }

    /**
     */
    public void beginBlock(tendermint.abci.Types.RequestBeginBlock request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseBeginBlock> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getBeginBlockMethod(), responseObserver);
    }

    /**
     */
    public void endBlock(tendermint.abci.Types.RequestEndBlock request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseEndBlock> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getEndBlockMethod(), responseObserver);
    }

    /**
     */
    public void listSnapshots(tendermint.abci.Types.RequestListSnapshots request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseListSnapshots> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListSnapshotsMethod(), responseObserver);
    }

    /**
     */
    public void offerSnapshot(tendermint.abci.Types.RequestOfferSnapshot request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseOfferSnapshot> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getOfferSnapshotMethod(), responseObserver);
    }

    /**
     */
    public void loadSnapshotChunk(tendermint.abci.Types.RequestLoadSnapshotChunk request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseLoadSnapshotChunk> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getLoadSnapshotChunkMethod(), responseObserver);
    }

    /**
     */
    public void applySnapshotChunk(tendermint.abci.Types.RequestApplySnapshotChunk request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseApplySnapshotChunk> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getApplySnapshotChunkMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getEchoMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestEcho,
                tendermint.abci.Types.ResponseEcho>(
                  this, METHODID_ECHO)))
          .addMethod(
            getFlushMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestFlush,
                tendermint.abci.Types.ResponseFlush>(
                  this, METHODID_FLUSH)))
          .addMethod(
            getInfoMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestInfo,
                tendermint.abci.Types.ResponseInfo>(
                  this, METHODID_INFO)))
          .addMethod(
            getSetOptionMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestSetOption,
                tendermint.abci.Types.ResponseSetOption>(
                  this, METHODID_SET_OPTION)))
          .addMethod(
            getDeliverTxMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestDeliverTx,
                tendermint.abci.Types.ResponseDeliverTx>(
                  this, METHODID_DELIVER_TX)))
          .addMethod(
            getCheckTxMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestCheckTx,
                tendermint.abci.Types.ResponseCheckTx>(
                  this, METHODID_CHECK_TX)))
          .addMethod(
            getQueryMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestQuery,
                tendermint.abci.Types.ResponseQuery>(
                  this, METHODID_QUERY)))
          .addMethod(
            getCommitMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestCommit,
                tendermint.abci.Types.ResponseCommit>(
                  this, METHODID_COMMIT)))
          .addMethod(
            getInitChainMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestInitChain,
                tendermint.abci.Types.ResponseInitChain>(
                  this, METHODID_INIT_CHAIN)))
          .addMethod(
            getBeginBlockMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestBeginBlock,
                tendermint.abci.Types.ResponseBeginBlock>(
                  this, METHODID_BEGIN_BLOCK)))
          .addMethod(
            getEndBlockMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestEndBlock,
                tendermint.abci.Types.ResponseEndBlock>(
                  this, METHODID_END_BLOCK)))
          .addMethod(
            getListSnapshotsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestListSnapshots,
                tendermint.abci.Types.ResponseListSnapshots>(
                  this, METHODID_LIST_SNAPSHOTS)))
          .addMethod(
            getOfferSnapshotMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestOfferSnapshot,
                tendermint.abci.Types.ResponseOfferSnapshot>(
                  this, METHODID_OFFER_SNAPSHOT)))
          .addMethod(
            getLoadSnapshotChunkMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestLoadSnapshotChunk,
                tendermint.abci.Types.ResponseLoadSnapshotChunk>(
                  this, METHODID_LOAD_SNAPSHOT_CHUNK)))
          .addMethod(
            getApplySnapshotChunkMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                tendermint.abci.Types.RequestApplySnapshotChunk,
                tendermint.abci.Types.ResponseApplySnapshotChunk>(
                  this, METHODID_APPLY_SNAPSHOT_CHUNK)))
          .build();
    }
  }

  /**
   */
  public static final class ABCIApplicationStub extends io.grpc.stub.AbstractAsyncStub<ABCIApplicationStub> {
    private ABCIApplicationStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ABCIApplicationStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ABCIApplicationStub(channel, callOptions);
    }

    /**
     */
    public void echo(tendermint.abci.Types.RequestEcho request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseEcho> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getEchoMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void flush(tendermint.abci.Types.RequestFlush request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseFlush> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getFlushMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void info(tendermint.abci.Types.RequestInfo request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseInfo> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getInfoMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void setOption(tendermint.abci.Types.RequestSetOption request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseSetOption> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSetOptionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void deliverTx(tendermint.abci.Types.RequestDeliverTx request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseDeliverTx> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDeliverTxMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void checkTx(tendermint.abci.Types.RequestCheckTx request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseCheckTx> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCheckTxMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void query(tendermint.abci.Types.RequestQuery request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseQuery> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getQueryMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void commit(tendermint.abci.Types.RequestCommit request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseCommit> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCommitMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void initChain(tendermint.abci.Types.RequestInitChain request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseInitChain> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getInitChainMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void beginBlock(tendermint.abci.Types.RequestBeginBlock request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseBeginBlock> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getBeginBlockMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void endBlock(tendermint.abci.Types.RequestEndBlock request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseEndBlock> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getEndBlockMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void listSnapshots(tendermint.abci.Types.RequestListSnapshots request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseListSnapshots> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListSnapshotsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void offerSnapshot(tendermint.abci.Types.RequestOfferSnapshot request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseOfferSnapshot> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getOfferSnapshotMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void loadSnapshotChunk(tendermint.abci.Types.RequestLoadSnapshotChunk request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseLoadSnapshotChunk> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getLoadSnapshotChunkMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void applySnapshotChunk(tendermint.abci.Types.RequestApplySnapshotChunk request,
        io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseApplySnapshotChunk> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getApplySnapshotChunkMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   */
  public static final class ABCIApplicationBlockingStub extends io.grpc.stub.AbstractBlockingStub<ABCIApplicationBlockingStub> {
    private ABCIApplicationBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ABCIApplicationBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ABCIApplicationBlockingStub(channel, callOptions);
    }

    /**
     */
    public tendermint.abci.Types.ResponseEcho echo(tendermint.abci.Types.RequestEcho request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getEchoMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseFlush flush(tendermint.abci.Types.RequestFlush request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getFlushMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseInfo info(tendermint.abci.Types.RequestInfo request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getInfoMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseSetOption setOption(tendermint.abci.Types.RequestSetOption request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSetOptionMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseDeliverTx deliverTx(tendermint.abci.Types.RequestDeliverTx request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDeliverTxMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseCheckTx checkTx(tendermint.abci.Types.RequestCheckTx request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCheckTxMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseQuery query(tendermint.abci.Types.RequestQuery request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getQueryMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseCommit commit(tendermint.abci.Types.RequestCommit request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCommitMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseInitChain initChain(tendermint.abci.Types.RequestInitChain request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getInitChainMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseBeginBlock beginBlock(tendermint.abci.Types.RequestBeginBlock request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getBeginBlockMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseEndBlock endBlock(tendermint.abci.Types.RequestEndBlock request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getEndBlockMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseListSnapshots listSnapshots(tendermint.abci.Types.RequestListSnapshots request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListSnapshotsMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseOfferSnapshot offerSnapshot(tendermint.abci.Types.RequestOfferSnapshot request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getOfferSnapshotMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseLoadSnapshotChunk loadSnapshotChunk(tendermint.abci.Types.RequestLoadSnapshotChunk request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getLoadSnapshotChunkMethod(), getCallOptions(), request);
    }

    /**
     */
    public tendermint.abci.Types.ResponseApplySnapshotChunk applySnapshotChunk(tendermint.abci.Types.RequestApplySnapshotChunk request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getApplySnapshotChunkMethod(), getCallOptions(), request);
    }
  }

  /**
   */
  public static final class ABCIApplicationFutureStub extends io.grpc.stub.AbstractFutureStub<ABCIApplicationFutureStub> {
    private ABCIApplicationFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ABCIApplicationFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ABCIApplicationFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseEcho> echo(
        tendermint.abci.Types.RequestEcho request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getEchoMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseFlush> flush(
        tendermint.abci.Types.RequestFlush request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getFlushMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseInfo> info(
        tendermint.abci.Types.RequestInfo request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getInfoMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseSetOption> setOption(
        tendermint.abci.Types.RequestSetOption request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSetOptionMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseDeliverTx> deliverTx(
        tendermint.abci.Types.RequestDeliverTx request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDeliverTxMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseCheckTx> checkTx(
        tendermint.abci.Types.RequestCheckTx request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCheckTxMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseQuery> query(
        tendermint.abci.Types.RequestQuery request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getQueryMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseCommit> commit(
        tendermint.abci.Types.RequestCommit request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCommitMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseInitChain> initChain(
        tendermint.abci.Types.RequestInitChain request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getInitChainMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseBeginBlock> beginBlock(
        tendermint.abci.Types.RequestBeginBlock request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getBeginBlockMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseEndBlock> endBlock(
        tendermint.abci.Types.RequestEndBlock request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getEndBlockMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseListSnapshots> listSnapshots(
        tendermint.abci.Types.RequestListSnapshots request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListSnapshotsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseOfferSnapshot> offerSnapshot(
        tendermint.abci.Types.RequestOfferSnapshot request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getOfferSnapshotMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseLoadSnapshotChunk> loadSnapshotChunk(
        tendermint.abci.Types.RequestLoadSnapshotChunk request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getLoadSnapshotChunkMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<tendermint.abci.Types.ResponseApplySnapshotChunk> applySnapshotChunk(
        tendermint.abci.Types.RequestApplySnapshotChunk request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getApplySnapshotChunkMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_ECHO = 0;
  private static final int METHODID_FLUSH = 1;
  private static final int METHODID_INFO = 2;
  private static final int METHODID_SET_OPTION = 3;
  private static final int METHODID_DELIVER_TX = 4;
  private static final int METHODID_CHECK_TX = 5;
  private static final int METHODID_QUERY = 6;
  private static final int METHODID_COMMIT = 7;
  private static final int METHODID_INIT_CHAIN = 8;
  private static final int METHODID_BEGIN_BLOCK = 9;
  private static final int METHODID_END_BLOCK = 10;
  private static final int METHODID_LIST_SNAPSHOTS = 11;
  private static final int METHODID_OFFER_SNAPSHOT = 12;
  private static final int METHODID_LOAD_SNAPSHOT_CHUNK = 13;
  private static final int METHODID_APPLY_SNAPSHOT_CHUNK = 14;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final ABCIApplicationImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(ABCIApplicationImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_ECHO:
          serviceImpl.echo((tendermint.abci.Types.RequestEcho) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseEcho>) responseObserver);
          break;
        case METHODID_FLUSH:
          serviceImpl.flush((tendermint.abci.Types.RequestFlush) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseFlush>) responseObserver);
          break;
        case METHODID_INFO:
          serviceImpl.info((tendermint.abci.Types.RequestInfo) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseInfo>) responseObserver);
          break;
        case METHODID_SET_OPTION:
          serviceImpl.setOption((tendermint.abci.Types.RequestSetOption) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseSetOption>) responseObserver);
          break;
        case METHODID_DELIVER_TX:
          serviceImpl.deliverTx((tendermint.abci.Types.RequestDeliverTx) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseDeliverTx>) responseObserver);
          break;
        case METHODID_CHECK_TX:
          serviceImpl.checkTx((tendermint.abci.Types.RequestCheckTx) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseCheckTx>) responseObserver);
          break;
        case METHODID_QUERY:
          serviceImpl.query((tendermint.abci.Types.RequestQuery) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseQuery>) responseObserver);
          break;
        case METHODID_COMMIT:
          serviceImpl.commit((tendermint.abci.Types.RequestCommit) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseCommit>) responseObserver);
          break;
        case METHODID_INIT_CHAIN:
          serviceImpl.initChain((tendermint.abci.Types.RequestInitChain) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseInitChain>) responseObserver);
          break;
        case METHODID_BEGIN_BLOCK:
          serviceImpl.beginBlock((tendermint.abci.Types.RequestBeginBlock) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseBeginBlock>) responseObserver);
          break;
        case METHODID_END_BLOCK:
          serviceImpl.endBlock((tendermint.abci.Types.RequestEndBlock) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseEndBlock>) responseObserver);
          break;
        case METHODID_LIST_SNAPSHOTS:
          serviceImpl.listSnapshots((tendermint.abci.Types.RequestListSnapshots) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseListSnapshots>) responseObserver);
          break;
        case METHODID_OFFER_SNAPSHOT:
          serviceImpl.offerSnapshot((tendermint.abci.Types.RequestOfferSnapshot) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseOfferSnapshot>) responseObserver);
          break;
        case METHODID_LOAD_SNAPSHOT_CHUNK:
          serviceImpl.loadSnapshotChunk((tendermint.abci.Types.RequestLoadSnapshotChunk) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseLoadSnapshotChunk>) responseObserver);
          break;
        case METHODID_APPLY_SNAPSHOT_CHUNK:
          serviceImpl.applySnapshotChunk((tendermint.abci.Types.RequestApplySnapshotChunk) request,
              (io.grpc.stub.StreamObserver<tendermint.abci.Types.ResponseApplySnapshotChunk>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  private static abstract class ABCIApplicationBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ABCIApplicationBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return tendermint.abci.Types.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ABCIApplication");
    }
  }

  private static final class ABCIApplicationFileDescriptorSupplier
      extends ABCIApplicationBaseDescriptorSupplier {
    ABCIApplicationFileDescriptorSupplier() {}
  }

  private static final class ABCIApplicationMethodDescriptorSupplier
      extends ABCIApplicationBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    ABCIApplicationMethodDescriptorSupplier(String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (ABCIApplicationGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ABCIApplicationFileDescriptorSupplier())
              .addMethod(getEchoMethod())
              .addMethod(getFlushMethod())
              .addMethod(getInfoMethod())
              .addMethod(getSetOptionMethod())
              .addMethod(getDeliverTxMethod())
              .addMethod(getCheckTxMethod())
              .addMethod(getQueryMethod())
              .addMethod(getCommitMethod())
              .addMethod(getInitChainMethod())
              .addMethod(getBeginBlockMethod())
              .addMethod(getEndBlockMethod())
              .addMethod(getListSnapshotsMethod())
              .addMethod(getOfferSnapshotMethod())
              .addMethod(getLoadSnapshotChunkMethod())
              .addMethod(getApplySnapshotChunkMethod())
              .build();
        }
      }
    }
    return result;
  }
}
