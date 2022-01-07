package cosmos.feegrant.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Query defines the gRPC querier service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/feegrant/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryGrpc {

  private QueryGrpc() {}

  public static final String SERVICE_NAME = "cosmos.feegrant.v1beta1.Query";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest,
      cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse> getAllowanceMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Allowance",
      requestType = cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest.class,
      responseType = cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest,
      cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse> getAllowanceMethod() {
    io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest, cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse> getAllowanceMethod;
    if ((getAllowanceMethod = QueryGrpc.getAllowanceMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getAllowanceMethod = QueryGrpc.getAllowanceMethod) == null) {
          QueryGrpc.getAllowanceMethod = getAllowanceMethod =
              io.grpc.MethodDescriptor.<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest, cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Allowance"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Allowance"))
              .build();
        }
      }
    }
    return getAllowanceMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest,
      cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse> getAllowancesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Allowances",
      requestType = cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest.class,
      responseType = cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest,
      cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse> getAllowancesMethod() {
    io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest, cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse> getAllowancesMethod;
    if ((getAllowancesMethod = QueryGrpc.getAllowancesMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getAllowancesMethod = QueryGrpc.getAllowancesMethod) == null) {
          QueryGrpc.getAllowancesMethod = getAllowancesMethod =
              io.grpc.MethodDescriptor.<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest, cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Allowances"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Allowances"))
              .build();
        }
      }
    }
    return getAllowancesMethod;
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
   * Query defines the gRPC querier service.
   * </pre>
   */
  public static abstract class QueryImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * Allowance returns fee granted to the grantee by the granter.
     * </pre>
     */
    public void allowance(cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest request,
        io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAllowanceMethod(), responseObserver);
    }

    /**
     * <pre>
     * Allowances returns all the grants for address.
     * </pre>
     */
    public void allowances(cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest request,
        io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAllowancesMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getAllowanceMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest,
                cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse>(
                  this, METHODID_ALLOWANCE)))
          .addMethod(
            getAllowancesMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest,
                cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse>(
                  this, METHODID_ALLOWANCES)))
          .build();
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service.
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
     * Allowance returns fee granted to the grantee by the granter.
     * </pre>
     */
    public void allowance(cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest request,
        io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAllowanceMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Allowances returns all the grants for address.
     * </pre>
     */
    public void allowances(cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest request,
        io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAllowancesMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service.
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
     * Allowance returns fee granted to the grantee by the granter.
     * </pre>
     */
    public cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse allowance(cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAllowanceMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Allowances returns all the grants for address.
     * </pre>
     */
    public cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse allowances(cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAllowancesMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service.
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
     * Allowance returns fee granted to the grantee by the granter.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse> allowance(
        cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAllowanceMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Allowances returns all the grants for address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse> allowances(
        cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAllowancesMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_ALLOWANCE = 0;
  private static final int METHODID_ALLOWANCES = 1;

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
        case METHODID_ALLOWANCE:
          serviceImpl.allowance((cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowanceResponse>) responseObserver);
          break;
        case METHODID_ALLOWANCES:
          serviceImpl.allowances((cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.QueryOuterClass.QueryAllowancesResponse>) responseObserver);
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
      return cosmos.feegrant.v1beta1.QueryOuterClass.getDescriptor();
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
              .addMethod(getAllowanceMethod())
              .addMethod(getAllowancesMethod())
              .build();
        }
      }
    }
    return result;
  }
}
