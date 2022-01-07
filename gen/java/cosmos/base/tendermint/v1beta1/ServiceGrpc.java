package cosmos.base.tendermint.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Service defines the gRPC querier service for tendermint queries.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/base/tendermint/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ServiceGrpc {

  private ServiceGrpc() {}

  public static final String SERVICE_NAME = "cosmos.base.tendermint.v1beta1.Service";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest,
      cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse> getGetNodeInfoMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetNodeInfo",
      requestType = cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest.class,
      responseType = cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest,
      cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse> getGetNodeInfoMethod() {
    io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest, cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse> getGetNodeInfoMethod;
    if ((getGetNodeInfoMethod = ServiceGrpc.getGetNodeInfoMethod) == null) {
      synchronized (ServiceGrpc.class) {
        if ((getGetNodeInfoMethod = ServiceGrpc.getGetNodeInfoMethod) == null) {
          ServiceGrpc.getGetNodeInfoMethod = getGetNodeInfoMethod =
              io.grpc.MethodDescriptor.<cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest, cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetNodeInfo"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ServiceMethodDescriptorSupplier("GetNodeInfo"))
              .build();
        }
      }
    }
    return getGetNodeInfoMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest,
      cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse> getGetSyncingMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetSyncing",
      requestType = cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest.class,
      responseType = cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest,
      cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse> getGetSyncingMethod() {
    io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest, cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse> getGetSyncingMethod;
    if ((getGetSyncingMethod = ServiceGrpc.getGetSyncingMethod) == null) {
      synchronized (ServiceGrpc.class) {
        if ((getGetSyncingMethod = ServiceGrpc.getGetSyncingMethod) == null) {
          ServiceGrpc.getGetSyncingMethod = getGetSyncingMethod =
              io.grpc.MethodDescriptor.<cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest, cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetSyncing"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ServiceMethodDescriptorSupplier("GetSyncing"))
              .build();
        }
      }
    }
    return getGetSyncingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest,
      cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse> getGetLatestBlockMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetLatestBlock",
      requestType = cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest.class,
      responseType = cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest,
      cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse> getGetLatestBlockMethod() {
    io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest, cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse> getGetLatestBlockMethod;
    if ((getGetLatestBlockMethod = ServiceGrpc.getGetLatestBlockMethod) == null) {
      synchronized (ServiceGrpc.class) {
        if ((getGetLatestBlockMethod = ServiceGrpc.getGetLatestBlockMethod) == null) {
          ServiceGrpc.getGetLatestBlockMethod = getGetLatestBlockMethod =
              io.grpc.MethodDescriptor.<cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest, cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetLatestBlock"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ServiceMethodDescriptorSupplier("GetLatestBlock"))
              .build();
        }
      }
    }
    return getGetLatestBlockMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest,
      cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse> getGetBlockByHeightMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetBlockByHeight",
      requestType = cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest.class,
      responseType = cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest,
      cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse> getGetBlockByHeightMethod() {
    io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest, cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse> getGetBlockByHeightMethod;
    if ((getGetBlockByHeightMethod = ServiceGrpc.getGetBlockByHeightMethod) == null) {
      synchronized (ServiceGrpc.class) {
        if ((getGetBlockByHeightMethod = ServiceGrpc.getGetBlockByHeightMethod) == null) {
          ServiceGrpc.getGetBlockByHeightMethod = getGetBlockByHeightMethod =
              io.grpc.MethodDescriptor.<cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest, cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetBlockByHeight"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ServiceMethodDescriptorSupplier("GetBlockByHeight"))
              .build();
        }
      }
    }
    return getGetBlockByHeightMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest,
      cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse> getGetLatestValidatorSetMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetLatestValidatorSet",
      requestType = cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest.class,
      responseType = cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest,
      cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse> getGetLatestValidatorSetMethod() {
    io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest, cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse> getGetLatestValidatorSetMethod;
    if ((getGetLatestValidatorSetMethod = ServiceGrpc.getGetLatestValidatorSetMethod) == null) {
      synchronized (ServiceGrpc.class) {
        if ((getGetLatestValidatorSetMethod = ServiceGrpc.getGetLatestValidatorSetMethod) == null) {
          ServiceGrpc.getGetLatestValidatorSetMethod = getGetLatestValidatorSetMethod =
              io.grpc.MethodDescriptor.<cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest, cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetLatestValidatorSet"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ServiceMethodDescriptorSupplier("GetLatestValidatorSet"))
              .build();
        }
      }
    }
    return getGetLatestValidatorSetMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest,
      cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse> getGetValidatorSetByHeightMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetValidatorSetByHeight",
      requestType = cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest.class,
      responseType = cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest,
      cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse> getGetValidatorSetByHeightMethod() {
    io.grpc.MethodDescriptor<cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest, cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse> getGetValidatorSetByHeightMethod;
    if ((getGetValidatorSetByHeightMethod = ServiceGrpc.getGetValidatorSetByHeightMethod) == null) {
      synchronized (ServiceGrpc.class) {
        if ((getGetValidatorSetByHeightMethod = ServiceGrpc.getGetValidatorSetByHeightMethod) == null) {
          ServiceGrpc.getGetValidatorSetByHeightMethod = getGetValidatorSetByHeightMethod =
              io.grpc.MethodDescriptor.<cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest, cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetValidatorSetByHeight"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ServiceMethodDescriptorSupplier("GetValidatorSetByHeight"))
              .build();
        }
      }
    }
    return getGetValidatorSetByHeightMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ServiceStub>() {
        @java.lang.Override
        public ServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ServiceStub(channel, callOptions);
        }
      };
    return ServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ServiceBlockingStub>() {
        @java.lang.Override
        public ServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ServiceBlockingStub(channel, callOptions);
        }
      };
    return ServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ServiceFutureStub>() {
        @java.lang.Override
        public ServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ServiceFutureStub(channel, callOptions);
        }
      };
    return ServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Service defines the gRPC querier service for tendermint queries.
   * </pre>
   */
  public static abstract class ServiceImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * GetNodeInfo queries the current node info.
     * </pre>
     */
    public void getNodeInfo(cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetNodeInfoMethod(), responseObserver);
    }

    /**
     * <pre>
     * GetSyncing queries node syncing.
     * </pre>
     */
    public void getSyncing(cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetSyncingMethod(), responseObserver);
    }

    /**
     * <pre>
     * GetLatestBlock returns the latest block.
     * </pre>
     */
    public void getLatestBlock(cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetLatestBlockMethod(), responseObserver);
    }

    /**
     * <pre>
     * GetBlockByHeight queries block for given height.
     * </pre>
     */
    public void getBlockByHeight(cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetBlockByHeightMethod(), responseObserver);
    }

    /**
     * <pre>
     * GetLatestValidatorSet queries latest validator-set.
     * </pre>
     */
    public void getLatestValidatorSet(cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetLatestValidatorSetMethod(), responseObserver);
    }

    /**
     * <pre>
     * GetValidatorSetByHeight queries validator-set at a given height.
     * </pre>
     */
    public void getValidatorSetByHeight(cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetValidatorSetByHeightMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getGetNodeInfoMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest,
                cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse>(
                  this, METHODID_GET_NODE_INFO)))
          .addMethod(
            getGetSyncingMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest,
                cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse>(
                  this, METHODID_GET_SYNCING)))
          .addMethod(
            getGetLatestBlockMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest,
                cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse>(
                  this, METHODID_GET_LATEST_BLOCK)))
          .addMethod(
            getGetBlockByHeightMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest,
                cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse>(
                  this, METHODID_GET_BLOCK_BY_HEIGHT)))
          .addMethod(
            getGetLatestValidatorSetMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest,
                cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse>(
                  this, METHODID_GET_LATEST_VALIDATOR_SET)))
          .addMethod(
            getGetValidatorSetByHeightMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest,
                cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse>(
                  this, METHODID_GET_VALIDATOR_SET_BY_HEIGHT)))
          .build();
    }
  }

  /**
   * <pre>
   * Service defines the gRPC querier service for tendermint queries.
   * </pre>
   */
  public static final class ServiceStub extends io.grpc.stub.AbstractAsyncStub<ServiceStub> {
    private ServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * GetNodeInfo queries the current node info.
     * </pre>
     */
    public void getNodeInfo(cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetNodeInfoMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GetSyncing queries node syncing.
     * </pre>
     */
    public void getSyncing(cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetSyncingMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GetLatestBlock returns the latest block.
     * </pre>
     */
    public void getLatestBlock(cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetLatestBlockMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GetBlockByHeight queries block for given height.
     * </pre>
     */
    public void getBlockByHeight(cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetBlockByHeightMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GetLatestValidatorSet queries latest validator-set.
     * </pre>
     */
    public void getLatestValidatorSet(cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetLatestValidatorSetMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GetValidatorSetByHeight queries validator-set at a given height.
     * </pre>
     */
    public void getValidatorSetByHeight(cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetValidatorSetByHeightMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Service defines the gRPC querier service for tendermint queries.
   * </pre>
   */
  public static final class ServiceBlockingStub extends io.grpc.stub.AbstractBlockingStub<ServiceBlockingStub> {
    private ServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * GetNodeInfo queries the current node info.
     * </pre>
     */
    public cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse getNodeInfo(cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetNodeInfoMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GetSyncing queries node syncing.
     * </pre>
     */
    public cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse getSyncing(cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetSyncingMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GetLatestBlock returns the latest block.
     * </pre>
     */
    public cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse getLatestBlock(cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetLatestBlockMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GetBlockByHeight queries block for given height.
     * </pre>
     */
    public cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse getBlockByHeight(cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetBlockByHeightMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GetLatestValidatorSet queries latest validator-set.
     * </pre>
     */
    public cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse getLatestValidatorSet(cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetLatestValidatorSetMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GetValidatorSetByHeight queries validator-set at a given height.
     * </pre>
     */
    public cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse getValidatorSetByHeight(cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetValidatorSetByHeightMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Service defines the gRPC querier service for tendermint queries.
   * </pre>
   */
  public static final class ServiceFutureStub extends io.grpc.stub.AbstractFutureStub<ServiceFutureStub> {
    private ServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * GetNodeInfo queries the current node info.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse> getNodeInfo(
        cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetNodeInfoMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GetSyncing queries node syncing.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse> getSyncing(
        cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetSyncingMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GetLatestBlock returns the latest block.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse> getLatestBlock(
        cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetLatestBlockMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GetBlockByHeight queries block for given height.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse> getBlockByHeight(
        cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetBlockByHeightMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GetLatestValidatorSet queries latest validator-set.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse> getLatestValidatorSet(
        cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetLatestValidatorSetMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GetValidatorSetByHeight queries validator-set at a given height.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse> getValidatorSetByHeight(
        cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetValidatorSetByHeightMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GET_NODE_INFO = 0;
  private static final int METHODID_GET_SYNCING = 1;
  private static final int METHODID_GET_LATEST_BLOCK = 2;
  private static final int METHODID_GET_BLOCK_BY_HEIGHT = 3;
  private static final int METHODID_GET_LATEST_VALIDATOR_SET = 4;
  private static final int METHODID_GET_VALIDATOR_SET_BY_HEIGHT = 5;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final ServiceImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(ServiceImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_GET_NODE_INFO:
          serviceImpl.getNodeInfo((cosmos.base.tendermint.v1beta1.Query.GetNodeInfoRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetNodeInfoResponse>) responseObserver);
          break;
        case METHODID_GET_SYNCING:
          serviceImpl.getSyncing((cosmos.base.tendermint.v1beta1.Query.GetSyncingRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetSyncingResponse>) responseObserver);
          break;
        case METHODID_GET_LATEST_BLOCK:
          serviceImpl.getLatestBlock((cosmos.base.tendermint.v1beta1.Query.GetLatestBlockRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetLatestBlockResponse>) responseObserver);
          break;
        case METHODID_GET_BLOCK_BY_HEIGHT:
          serviceImpl.getBlockByHeight((cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetBlockByHeightResponse>) responseObserver);
          break;
        case METHODID_GET_LATEST_VALIDATOR_SET:
          serviceImpl.getLatestValidatorSet((cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetLatestValidatorSetResponse>) responseObserver);
          break;
        case METHODID_GET_VALIDATOR_SET_BY_HEIGHT:
          serviceImpl.getValidatorSetByHeight((cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.base.tendermint.v1beta1.Query.GetValidatorSetByHeightResponse>) responseObserver);
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

  private static abstract class ServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return cosmos.base.tendermint.v1beta1.Query.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("Service");
    }
  }

  private static final class ServiceFileDescriptorSupplier
      extends ServiceBaseDescriptorSupplier {
    ServiceFileDescriptorSupplier() {}
  }

  private static final class ServiceMethodDescriptorSupplier
      extends ServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    ServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (ServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ServiceFileDescriptorSupplier())
              .addMethod(getGetNodeInfoMethod())
              .addMethod(getGetSyncingMethod())
              .addMethod(getGetLatestBlockMethod())
              .addMethod(getGetBlockByHeightMethod())
              .addMethod(getGetLatestValidatorSetMethod())
              .addMethod(getGetValidatorSetByHeightMethod())
              .build();
        }
      }
    }
    return result;
  }
}
