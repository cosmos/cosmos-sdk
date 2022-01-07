package cosmos.vesting.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Msg defines the bank Msg service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/vesting/v1beta1/tx.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class MsgGrpc {

  private MsgGrpc() {}

  public static final String SERVICE_NAME = "cosmos.vesting.v1beta1.Msg";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount,
      cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse> getCreateVestingAccountMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateVestingAccount",
      requestType = cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount.class,
      responseType = cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount,
      cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse> getCreateVestingAccountMethod() {
    io.grpc.MethodDescriptor<cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount, cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse> getCreateVestingAccountMethod;
    if ((getCreateVestingAccountMethod = MsgGrpc.getCreateVestingAccountMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getCreateVestingAccountMethod = MsgGrpc.getCreateVestingAccountMethod) == null) {
          MsgGrpc.getCreateVestingAccountMethod = getCreateVestingAccountMethod =
              io.grpc.MethodDescriptor.<cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount, cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateVestingAccount"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("CreateVestingAccount"))
              .build();
        }
      }
    }
    return getCreateVestingAccountMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount,
      cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse> getCreatePeriodicVestingAccountMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreatePeriodicVestingAccount",
      requestType = cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount.class,
      responseType = cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount,
      cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse> getCreatePeriodicVestingAccountMethod() {
    io.grpc.MethodDescriptor<cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount, cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse> getCreatePeriodicVestingAccountMethod;
    if ((getCreatePeriodicVestingAccountMethod = MsgGrpc.getCreatePeriodicVestingAccountMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getCreatePeriodicVestingAccountMethod = MsgGrpc.getCreatePeriodicVestingAccountMethod) == null) {
          MsgGrpc.getCreatePeriodicVestingAccountMethod = getCreatePeriodicVestingAccountMethod =
              io.grpc.MethodDescriptor.<cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount, cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreatePeriodicVestingAccount"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("CreatePeriodicVestingAccount"))
              .build();
        }
      }
    }
    return getCreatePeriodicVestingAccountMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static MsgStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<MsgStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<MsgStub>() {
        @java.lang.Override
        public MsgStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new MsgStub(channel, callOptions);
        }
      };
    return MsgStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static MsgBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<MsgBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<MsgBlockingStub>() {
        @java.lang.Override
        public MsgBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new MsgBlockingStub(channel, callOptions);
        }
      };
    return MsgBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static MsgFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<MsgFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<MsgFutureStub>() {
        @java.lang.Override
        public MsgFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new MsgFutureStub(channel, callOptions);
        }
      };
    return MsgFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Msg defines the bank Msg service.
   * </pre>
   */
  public static abstract class MsgImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * CreateVestingAccount defines a method that enables creating a vesting
     * account.
     * </pre>
     */
    public void createVestingAccount(cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount request,
        io.grpc.stub.StreamObserver<cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateVestingAccountMethod(), responseObserver);
    }

    /**
     * <pre>
     * CreatePeriodicVestingAccount defines a method that enables creating a
     * periodic vesting account.
     * </pre>
     */
    public void createPeriodicVestingAccount(cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount request,
        io.grpc.stub.StreamObserver<cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreatePeriodicVestingAccountMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getCreateVestingAccountMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount,
                cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse>(
                  this, METHODID_CREATE_VESTING_ACCOUNT)))
          .addMethod(
            getCreatePeriodicVestingAccountMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount,
                cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse>(
                  this, METHODID_CREATE_PERIODIC_VESTING_ACCOUNT)))
          .build();
    }
  }

  /**
   * <pre>
   * Msg defines the bank Msg service.
   * </pre>
   */
  public static final class MsgStub extends io.grpc.stub.AbstractAsyncStub<MsgStub> {
    private MsgStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected MsgStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new MsgStub(channel, callOptions);
    }

    /**
     * <pre>
     * CreateVestingAccount defines a method that enables creating a vesting
     * account.
     * </pre>
     */
    public void createVestingAccount(cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount request,
        io.grpc.stub.StreamObserver<cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateVestingAccountMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * CreatePeriodicVestingAccount defines a method that enables creating a
     * periodic vesting account.
     * </pre>
     */
    public void createPeriodicVestingAccount(cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount request,
        io.grpc.stub.StreamObserver<cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreatePeriodicVestingAccountMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Msg defines the bank Msg service.
   * </pre>
   */
  public static final class MsgBlockingStub extends io.grpc.stub.AbstractBlockingStub<MsgBlockingStub> {
    private MsgBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected MsgBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new MsgBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * CreateVestingAccount defines a method that enables creating a vesting
     * account.
     * </pre>
     */
    public cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse createVestingAccount(cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateVestingAccountMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * CreatePeriodicVestingAccount defines a method that enables creating a
     * periodic vesting account.
     * </pre>
     */
    public cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse createPeriodicVestingAccount(cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreatePeriodicVestingAccountMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Msg defines the bank Msg service.
   * </pre>
   */
  public static final class MsgFutureStub extends io.grpc.stub.AbstractFutureStub<MsgFutureStub> {
    private MsgFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected MsgFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new MsgFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * CreateVestingAccount defines a method that enables creating a vesting
     * account.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse> createVestingAccount(
        cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateVestingAccountMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * CreatePeriodicVestingAccount defines a method that enables creating a
     * periodic vesting account.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse> createPeriodicVestingAccount(
        cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreatePeriodicVestingAccountMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CREATE_VESTING_ACCOUNT = 0;
  private static final int METHODID_CREATE_PERIODIC_VESTING_ACCOUNT = 1;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final MsgImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(MsgImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_CREATE_VESTING_ACCOUNT:
          serviceImpl.createVestingAccount((cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccount) request,
              (io.grpc.stub.StreamObserver<cosmos.vesting.v1beta1.Tx.MsgCreateVestingAccountResponse>) responseObserver);
          break;
        case METHODID_CREATE_PERIODIC_VESTING_ACCOUNT:
          serviceImpl.createPeriodicVestingAccount((cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccount) request,
              (io.grpc.stub.StreamObserver<cosmos.vesting.v1beta1.Tx.MsgCreatePeriodicVestingAccountResponse>) responseObserver);
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

  private static abstract class MsgBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    MsgBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return cosmos.vesting.v1beta1.Tx.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("Msg");
    }
  }

  private static final class MsgFileDescriptorSupplier
      extends MsgBaseDescriptorSupplier {
    MsgFileDescriptorSupplier() {}
  }

  private static final class MsgMethodDescriptorSupplier
      extends MsgBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    MsgMethodDescriptorSupplier(String methodName) {
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
      synchronized (MsgGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new MsgFileDescriptorSupplier())
              .addMethod(getCreateVestingAccountMethod())
              .addMethod(getCreatePeriodicVestingAccountMethod())
              .build();
        }
      }
    }
    return result;
  }
}
