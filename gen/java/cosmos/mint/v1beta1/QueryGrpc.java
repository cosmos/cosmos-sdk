package cosmos.mint.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Query provides defines the gRPC querier service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/mint/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryGrpc {

  private QueryGrpc() {}

  public static final String SERVICE_NAME = "cosmos.mint.v1beta1.Query";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest,
      cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Params",
      requestType = cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest.class,
      responseType = cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest,
      cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod() {
    io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest, cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod;
    if ((getParamsMethod = QueryGrpc.getParamsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getParamsMethod = QueryGrpc.getParamsMethod) == null) {
          QueryGrpc.getParamsMethod = getParamsMethod =
              io.grpc.MethodDescriptor.<cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest, cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Params"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Params"))
              .build();
        }
      }
    }
    return getParamsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest,
      cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse> getInflationMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Inflation",
      requestType = cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest.class,
      responseType = cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest,
      cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse> getInflationMethod() {
    io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest, cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse> getInflationMethod;
    if ((getInflationMethod = QueryGrpc.getInflationMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getInflationMethod = QueryGrpc.getInflationMethod) == null) {
          QueryGrpc.getInflationMethod = getInflationMethod =
              io.grpc.MethodDescriptor.<cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest, cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Inflation"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Inflation"))
              .build();
        }
      }
    }
    return getInflationMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest,
      cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse> getAnnualProvisionsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "AnnualProvisions",
      requestType = cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest.class,
      responseType = cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest,
      cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse> getAnnualProvisionsMethod() {
    io.grpc.MethodDescriptor<cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest, cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse> getAnnualProvisionsMethod;
    if ((getAnnualProvisionsMethod = QueryGrpc.getAnnualProvisionsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getAnnualProvisionsMethod = QueryGrpc.getAnnualProvisionsMethod) == null) {
          QueryGrpc.getAnnualProvisionsMethod = getAnnualProvisionsMethod =
              io.grpc.MethodDescriptor.<cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest, cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "AnnualProvisions"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("AnnualProvisions"))
              .build();
        }
      }
    }
    return getAnnualProvisionsMethod;
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
   * Query provides defines the gRPC querier service.
   * </pre>
   */
  public static abstract class QueryImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * Params returns the total set of minting parameters.
     * </pre>
     */
    public void params(cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest request,
        io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getParamsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Inflation returns the current minting inflation value.
     * </pre>
     */
    public void inflation(cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest request,
        io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getInflationMethod(), responseObserver);
    }

    /**
     * <pre>
     * AnnualProvisions current minting annual provisions value.
     * </pre>
     */
    public void annualProvisions(cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest request,
        io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAnnualProvisionsMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getParamsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest,
                cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse>(
                  this, METHODID_PARAMS)))
          .addMethod(
            getInflationMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest,
                cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse>(
                  this, METHODID_INFLATION)))
          .addMethod(
            getAnnualProvisionsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest,
                cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse>(
                  this, METHODID_ANNUAL_PROVISIONS)))
          .build();
    }
  }

  /**
   * <pre>
   * Query provides defines the gRPC querier service.
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
     * Params returns the total set of minting parameters.
     * </pre>
     */
    public void params(cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest request,
        io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getParamsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Inflation returns the current minting inflation value.
     * </pre>
     */
    public void inflation(cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest request,
        io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getInflationMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * AnnualProvisions current minting annual provisions value.
     * </pre>
     */
    public void annualProvisions(cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest request,
        io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAnnualProvisionsMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Query provides defines the gRPC querier service.
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
     * Params returns the total set of minting parameters.
     * </pre>
     */
    public cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse params(cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getParamsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Inflation returns the current minting inflation value.
     * </pre>
     */
    public cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse inflation(cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getInflationMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * AnnualProvisions current minting annual provisions value.
     * </pre>
     */
    public cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse annualProvisions(cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAnnualProvisionsMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Query provides defines the gRPC querier service.
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
     * Params returns the total set of minting parameters.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse> params(
        cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getParamsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Inflation returns the current minting inflation value.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse> inflation(
        cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getInflationMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * AnnualProvisions current minting annual provisions value.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse> annualProvisions(
        cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAnnualProvisionsMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PARAMS = 0;
  private static final int METHODID_INFLATION = 1;
  private static final int METHODID_ANNUAL_PROVISIONS = 2;

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
          serviceImpl.params((cosmos.mint.v1beta1.QueryOuterClass.QueryParamsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryParamsResponse>) responseObserver);
          break;
        case METHODID_INFLATION:
          serviceImpl.inflation((cosmos.mint.v1beta1.QueryOuterClass.QueryInflationRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryInflationResponse>) responseObserver);
          break;
        case METHODID_ANNUAL_PROVISIONS:
          serviceImpl.annualProvisions((cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.mint.v1beta1.QueryOuterClass.QueryAnnualProvisionsResponse>) responseObserver);
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
      return cosmos.mint.v1beta1.QueryOuterClass.getDescriptor();
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
              .addMethod(getInflationMethod())
              .addMethod(getAnnualProvisionsMethod())
              .build();
        }
      }
    }
    return result;
  }
}
