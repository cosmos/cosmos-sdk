package cosmos.base.reflection.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * ReflectionService defines a service for interface reflection.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/base/reflection/v1beta1/reflection.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ReflectionServiceGrpc {

  private ReflectionServiceGrpc() {}

  public static final String SERVICE_NAME = "cosmos.base.reflection.v1beta1.ReflectionService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest,
      cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse> getListAllInterfacesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListAllInterfaces",
      requestType = cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest.class,
      responseType = cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest,
      cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse> getListAllInterfacesMethod() {
    io.grpc.MethodDescriptor<cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest, cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse> getListAllInterfacesMethod;
    if ((getListAllInterfacesMethod = ReflectionServiceGrpc.getListAllInterfacesMethod) == null) {
      synchronized (ReflectionServiceGrpc.class) {
        if ((getListAllInterfacesMethod = ReflectionServiceGrpc.getListAllInterfacesMethod) == null) {
          ReflectionServiceGrpc.getListAllInterfacesMethod = getListAllInterfacesMethod =
              io.grpc.MethodDescriptor.<cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest, cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListAllInterfaces"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ReflectionServiceMethodDescriptorSupplier("ListAllInterfaces"))
              .build();
        }
      }
    }
    return getListAllInterfacesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest,
      cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse> getListImplementationsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListImplementations",
      requestType = cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest.class,
      responseType = cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest,
      cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse> getListImplementationsMethod() {
    io.grpc.MethodDescriptor<cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest, cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse> getListImplementationsMethod;
    if ((getListImplementationsMethod = ReflectionServiceGrpc.getListImplementationsMethod) == null) {
      synchronized (ReflectionServiceGrpc.class) {
        if ((getListImplementationsMethod = ReflectionServiceGrpc.getListImplementationsMethod) == null) {
          ReflectionServiceGrpc.getListImplementationsMethod = getListImplementationsMethod =
              io.grpc.MethodDescriptor.<cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest, cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListImplementations"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ReflectionServiceMethodDescriptorSupplier("ListImplementations"))
              .build();
        }
      }
    }
    return getListImplementationsMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ReflectionServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ReflectionServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ReflectionServiceStub>() {
        @java.lang.Override
        public ReflectionServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ReflectionServiceStub(channel, callOptions);
        }
      };
    return ReflectionServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ReflectionServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ReflectionServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ReflectionServiceBlockingStub>() {
        @java.lang.Override
        public ReflectionServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ReflectionServiceBlockingStub(channel, callOptions);
        }
      };
    return ReflectionServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ReflectionServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ReflectionServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ReflectionServiceFutureStub>() {
        @java.lang.Override
        public ReflectionServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ReflectionServiceFutureStub(channel, callOptions);
        }
      };
    return ReflectionServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * ReflectionService defines a service for interface reflection.
   * </pre>
   */
  public static abstract class ReflectionServiceImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * ListAllInterfaces lists all the interfaces registered in the interface
     * registry.
     * </pre>
     */
    public void listAllInterfaces(cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListAllInterfacesMethod(), responseObserver);
    }

    /**
     * <pre>
     * ListImplementations list all the concrete types that implement a given
     * interface.
     * </pre>
     */
    public void listImplementations(cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListImplementationsMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getListAllInterfacesMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest,
                cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse>(
                  this, METHODID_LIST_ALL_INTERFACES)))
          .addMethod(
            getListImplementationsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest,
                cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse>(
                  this, METHODID_LIST_IMPLEMENTATIONS)))
          .build();
    }
  }

  /**
   * <pre>
   * ReflectionService defines a service for interface reflection.
   * </pre>
   */
  public static final class ReflectionServiceStub extends io.grpc.stub.AbstractAsyncStub<ReflectionServiceStub> {
    private ReflectionServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ReflectionServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ReflectionServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * ListAllInterfaces lists all the interfaces registered in the interface
     * registry.
     * </pre>
     */
    public void listAllInterfaces(cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListAllInterfacesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * ListImplementations list all the concrete types that implement a given
     * interface.
     * </pre>
     */
    public void listImplementations(cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest request,
        io.grpc.stub.StreamObserver<cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListImplementationsMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * ReflectionService defines a service for interface reflection.
   * </pre>
   */
  public static final class ReflectionServiceBlockingStub extends io.grpc.stub.AbstractBlockingStub<ReflectionServiceBlockingStub> {
    private ReflectionServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ReflectionServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ReflectionServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * ListAllInterfaces lists all the interfaces registered in the interface
     * registry.
     * </pre>
     */
    public cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse listAllInterfaces(cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListAllInterfacesMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * ListImplementations list all the concrete types that implement a given
     * interface.
     * </pre>
     */
    public cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse listImplementations(cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListImplementationsMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * ReflectionService defines a service for interface reflection.
   * </pre>
   */
  public static final class ReflectionServiceFutureStub extends io.grpc.stub.AbstractFutureStub<ReflectionServiceFutureStub> {
    private ReflectionServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ReflectionServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ReflectionServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * ListAllInterfaces lists all the interfaces registered in the interface
     * registry.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse> listAllInterfaces(
        cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListAllInterfacesMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * ListImplementations list all the concrete types that implement a given
     * interface.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse> listImplementations(
        cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListImplementationsMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_LIST_ALL_INTERFACES = 0;
  private static final int METHODID_LIST_IMPLEMENTATIONS = 1;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final ReflectionServiceImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(ReflectionServiceImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_LIST_ALL_INTERFACES:
          serviceImpl.listAllInterfaces((cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.base.reflection.v1beta1.Reflection.ListAllInterfacesResponse>) responseObserver);
          break;
        case METHODID_LIST_IMPLEMENTATIONS:
          serviceImpl.listImplementations((cosmos.base.reflection.v1beta1.Reflection.ListImplementationsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.base.reflection.v1beta1.Reflection.ListImplementationsResponse>) responseObserver);
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

  private static abstract class ReflectionServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ReflectionServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return cosmos.base.reflection.v1beta1.Reflection.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ReflectionService");
    }
  }

  private static final class ReflectionServiceFileDescriptorSupplier
      extends ReflectionServiceBaseDescriptorSupplier {
    ReflectionServiceFileDescriptorSupplier() {}
  }

  private static final class ReflectionServiceMethodDescriptorSupplier
      extends ReflectionServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    ReflectionServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (ReflectionServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ReflectionServiceFileDescriptorSupplier())
              .addMethod(getListAllInterfacesMethod())
              .addMethod(getListImplementationsMethod())
              .build();
        }
      }
    }
    return result;
  }
}
