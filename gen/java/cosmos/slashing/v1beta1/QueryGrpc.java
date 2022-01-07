package cosmos.slashing.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Query provides defines the gRPC querier service
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/slashing/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryGrpc {

  private QueryGrpc() {}

  public static final String SERVICE_NAME = "cosmos.slashing.v1beta1.Query";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest,
      cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Params",
      requestType = cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest.class,
      responseType = cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest,
      cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod() {
    io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest, cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod;
    if ((getParamsMethod = QueryGrpc.getParamsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getParamsMethod = QueryGrpc.getParamsMethod) == null) {
          QueryGrpc.getParamsMethod = getParamsMethod =
              io.grpc.MethodDescriptor.<cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest, cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Params"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Params"))
              .build();
        }
      }
    }
    return getParamsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest,
      cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse> getSigningInfoMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SigningInfo",
      requestType = cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest.class,
      responseType = cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest,
      cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse> getSigningInfoMethod() {
    io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest, cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse> getSigningInfoMethod;
    if ((getSigningInfoMethod = QueryGrpc.getSigningInfoMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getSigningInfoMethod = QueryGrpc.getSigningInfoMethod) == null) {
          QueryGrpc.getSigningInfoMethod = getSigningInfoMethod =
              io.grpc.MethodDescriptor.<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest, cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SigningInfo"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("SigningInfo"))
              .build();
        }
      }
    }
    return getSigningInfoMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest,
      cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse> getSigningInfosMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SigningInfos",
      requestType = cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest.class,
      responseType = cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest,
      cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse> getSigningInfosMethod() {
    io.grpc.MethodDescriptor<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest, cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse> getSigningInfosMethod;
    if ((getSigningInfosMethod = QueryGrpc.getSigningInfosMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getSigningInfosMethod = QueryGrpc.getSigningInfosMethod) == null) {
          QueryGrpc.getSigningInfosMethod = getSigningInfosMethod =
              io.grpc.MethodDescriptor.<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest, cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SigningInfos"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("SigningInfos"))
              .build();
        }
      }
    }
    return getSigningInfosMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static QueryStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryStub>() {
        @java.lang.Override
        public QueryStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryStub(channel, callOptions);
        }
      };
    return QueryStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static QueryBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryBlockingStub>() {
        @java.lang.Override
        public QueryBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryBlockingStub(channel, callOptions);
        }
      };
    return QueryBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static QueryFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryFutureStub>() {
        @java.lang.Override
        public QueryFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryFutureStub(channel, callOptions);
        }
      };
    return QueryFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Query provides defines the gRPC querier service
   * </pre>
   */
  public static abstract class QueryImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * Params queries the parameters of slashing module
     * </pre>
     */
    public void params(cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest request,
        io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getParamsMethod(), responseObserver);
    }

    /**
     * <pre>
     * SigningInfo queries the signing info of given cons address
     * </pre>
     */
    public void signingInfo(cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest request,
        io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSigningInfoMethod(), responseObserver);
    }

    /**
     * <pre>
     * SigningInfos queries signing info of all validators
     * </pre>
     */
    public void signingInfos(cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest request,
        io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSigningInfosMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getParamsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest,
                cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse>(
                  this, METHODID_PARAMS)))
          .addMethod(
            getSigningInfoMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest,
                cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse>(
                  this, METHODID_SIGNING_INFO)))
          .addMethod(
            getSigningInfosMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest,
                cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse>(
                  this, METHODID_SIGNING_INFOS)))
          .build();
    }
  }

  /**
   * <pre>
   * Query provides defines the gRPC querier service
   * </pre>
   */
  public static final class QueryStub extends io.grpc.stub.AbstractAsyncStub<QueryStub> {
    private QueryStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryStub(channel, callOptions);
    }

    /**
     * <pre>
     * Params queries the parameters of slashing module
     * </pre>
     */
    public void params(cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest request,
        io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getParamsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * SigningInfo queries the signing info of given cons address
     * </pre>
     */
    public void signingInfo(cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest request,
        io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSigningInfoMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * SigningInfos queries signing info of all validators
     * </pre>
     */
    public void signingInfos(cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest request,
        io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSigningInfosMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Query provides defines the gRPC querier service
   * </pre>
   */
  public static final class QueryBlockingStub extends io.grpc.stub.AbstractBlockingStub<QueryBlockingStub> {
    private QueryBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Params queries the parameters of slashing module
     * </pre>
     */
    public cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse params(cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getParamsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * SigningInfo queries the signing info of given cons address
     * </pre>
     */
    public cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse signingInfo(cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSigningInfoMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * SigningInfos queries signing info of all validators
     * </pre>
     */
    public cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse signingInfos(cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSigningInfosMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Query provides defines the gRPC querier service
   * </pre>
   */
  public static final class QueryFutureStub extends io.grpc.stub.AbstractFutureStub<QueryFutureStub> {
    private QueryFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Params queries the parameters of slashing module
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse> params(
        cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getParamsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * SigningInfo queries the signing info of given cons address
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse> signingInfo(
        cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSigningInfoMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * SigningInfos queries signing info of all validators
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse> signingInfos(
        cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSigningInfosMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PARAMS = 0;
  private static final int METHODID_SIGNING_INFO = 1;
  private static final int METHODID_SIGNING_INFOS = 2;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final QueryImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(QueryImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_PARAMS:
          serviceImpl.params((cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QueryParamsResponse>) responseObserver);
          break;
        case METHODID_SIGNING_INFO:
          serviceImpl.signingInfo((cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfoResponse>) responseObserver);
          break;
        case METHODID_SIGNING_INFOS:
          serviceImpl.signingInfos((cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.slashing.v1beta1.QueryOuterClass.QuerySigningInfosResponse>) responseObserver);
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

  private static abstract class QueryBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    QueryBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return cosmos.slashing.v1beta1.QueryOuterClass.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("Query");
    }
  }

  private static final class QueryFileDescriptorSupplier
      extends QueryBaseDescriptorSupplier {
    QueryFileDescriptorSupplier() {}
  }

  private static final class QueryMethodDescriptorSupplier
      extends QueryBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    QueryMethodDescriptorSupplier(String methodName) {
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
      synchronized (QueryGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new QueryFileDescriptorSupplier())
              .addMethod(getParamsMethod())
              .addMethod(getSigningInfoMethod())
              .addMethod(getSigningInfosMethod())
              .build();
        }
      }
    }
    return result;
  }
}
