package cosmos.evidence.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Query defines the gRPC querier service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/evidence/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryGrpc {

  private QueryGrpc() {}

  public static final String SERVICE_NAME = "cosmos.evidence.v1beta1.Query";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest,
      cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse> getEvidenceMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Evidence",
      requestType = cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest.class,
      responseType = cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest,
      cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse> getEvidenceMethod() {
    io.grpc.MethodDescriptor<cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest, cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse> getEvidenceMethod;
    if ((getEvidenceMethod = QueryGrpc.getEvidenceMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getEvidenceMethod = QueryGrpc.getEvidenceMethod) == null) {
          QueryGrpc.getEvidenceMethod = getEvidenceMethod =
              io.grpc.MethodDescriptor.<cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest, cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Evidence"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Evidence"))
              .build();
        }
      }
    }
    return getEvidenceMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest,
      cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse> getAllEvidenceMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "AllEvidence",
      requestType = cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest.class,
      responseType = cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest,
      cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse> getAllEvidenceMethod() {
    io.grpc.MethodDescriptor<cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest, cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse> getAllEvidenceMethod;
    if ((getAllEvidenceMethod = QueryGrpc.getAllEvidenceMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getAllEvidenceMethod = QueryGrpc.getAllEvidenceMethod) == null) {
          QueryGrpc.getAllEvidenceMethod = getAllEvidenceMethod =
              io.grpc.MethodDescriptor.<cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest, cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "AllEvidence"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("AllEvidence"))
              .build();
        }
      }
    }
    return getAllEvidenceMethod;
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
     * Evidence queries evidence based on evidence hash.
     * </pre>
     */
    public void evidence(cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest request,
        io.grpc.stub.StreamObserver<cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getEvidenceMethod(), responseObserver);
    }

    /**
     * <pre>
     * AllEvidence queries all evidence.
     * </pre>
     */
    public void allEvidence(cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest request,
        io.grpc.stub.StreamObserver<cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAllEvidenceMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getEvidenceMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest,
                cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse>(
                  this, METHODID_EVIDENCE)))
          .addMethod(
            getAllEvidenceMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest,
                cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse>(
                  this, METHODID_ALL_EVIDENCE)))
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
     * Evidence queries evidence based on evidence hash.
     * </pre>
     */
    public void evidence(cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest request,
        io.grpc.stub.StreamObserver<cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getEvidenceMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * AllEvidence queries all evidence.
     * </pre>
     */
    public void allEvidence(cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest request,
        io.grpc.stub.StreamObserver<cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAllEvidenceMethod(), getCallOptions()), request, responseObserver);
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
     * Evidence queries evidence based on evidence hash.
     * </pre>
     */
    public cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse evidence(cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getEvidenceMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * AllEvidence queries all evidence.
     * </pre>
     */
    public cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse allEvidence(cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAllEvidenceMethod(), getCallOptions(), request);
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
     * Evidence queries evidence based on evidence hash.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse> evidence(
        cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getEvidenceMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * AllEvidence queries all evidence.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse> allEvidence(
        cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAllEvidenceMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_EVIDENCE = 0;
  private static final int METHODID_ALL_EVIDENCE = 1;

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
        case METHODID_EVIDENCE:
          serviceImpl.evidence((cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.evidence.v1beta1.QueryOuterClass.QueryEvidenceResponse>) responseObserver);
          break;
        case METHODID_ALL_EVIDENCE:
          serviceImpl.allEvidence((cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.evidence.v1beta1.QueryOuterClass.QueryAllEvidenceResponse>) responseObserver);
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
      return cosmos.evidence.v1beta1.QueryOuterClass.getDescriptor();
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
              .addMethod(getEvidenceMethod())
              .addMethod(getAllEvidenceMethod())
              .build();
        }
      }
    }
    return result;
  }
}
