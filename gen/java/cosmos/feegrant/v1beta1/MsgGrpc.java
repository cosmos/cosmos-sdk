package cosmos.feegrant.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Msg defines the feegrant msg service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/feegrant/v1beta1/tx.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class MsgGrpc {

  private MsgGrpc() {}

  public static final String SERVICE_NAME = "cosmos.feegrant.v1beta1.Msg";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance,
      cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse> getGrantAllowanceMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GrantAllowance",
      requestType = cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance.class,
      responseType = cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance,
      cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse> getGrantAllowanceMethod() {
    io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance, cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse> getGrantAllowanceMethod;
    if ((getGrantAllowanceMethod = MsgGrpc.getGrantAllowanceMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getGrantAllowanceMethod = MsgGrpc.getGrantAllowanceMethod) == null) {
          MsgGrpc.getGrantAllowanceMethod = getGrantAllowanceMethod =
              io.grpc.MethodDescriptor.<cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance, cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GrantAllowance"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("GrantAllowance"))
              .build();
        }
      }
    }
    return getGrantAllowanceMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance,
      cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse> getRevokeAllowanceMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RevokeAllowance",
      requestType = cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance.class,
      responseType = cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance,
      cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse> getRevokeAllowanceMethod() {
    io.grpc.MethodDescriptor<cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance, cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse> getRevokeAllowanceMethod;
    if ((getRevokeAllowanceMethod = MsgGrpc.getRevokeAllowanceMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getRevokeAllowanceMethod = MsgGrpc.getRevokeAllowanceMethod) == null) {
          MsgGrpc.getRevokeAllowanceMethod = getRevokeAllowanceMethod =
              io.grpc.MethodDescriptor.<cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance, cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RevokeAllowance"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("RevokeAllowance"))
              .build();
        }
      }
    }
    return getRevokeAllowanceMethod;
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
   * Msg defines the feegrant msg service.
   * </pre>
   */
  public static abstract class MsgImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * GrantAllowance grants fee allowance to the grantee on the granter's
     * account with the provided expiration time.
     * </pre>
     */
    public void grantAllowance(cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance request,
        io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGrantAllowanceMethod(), responseObserver);
    }

    /**
     * <pre>
     * RevokeAllowance revokes any fee allowance of granter's account that
     * has been granted to the grantee.
     * </pre>
     */
    public void revokeAllowance(cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance request,
        io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRevokeAllowanceMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getGrantAllowanceMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance,
                cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse>(
                  this, METHODID_GRANT_ALLOWANCE)))
          .addMethod(
            getRevokeAllowanceMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance,
                cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse>(
                  this, METHODID_REVOKE_ALLOWANCE)))
          .build();
    }
  }

  /**
   * <pre>
   * Msg defines the feegrant msg service.
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
     * GrantAllowance grants fee allowance to the grantee on the granter's
     * account with the provided expiration time.
     * </pre>
     */
    public void grantAllowance(cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance request,
        io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGrantAllowanceMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * RevokeAllowance revokes any fee allowance of granter's account that
     * has been granted to the grantee.
     * </pre>
     */
    public void revokeAllowance(cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance request,
        io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRevokeAllowanceMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Msg defines the feegrant msg service.
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
     * GrantAllowance grants fee allowance to the grantee on the granter's
     * account with the provided expiration time.
     * </pre>
     */
    public cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse grantAllowance(cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGrantAllowanceMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * RevokeAllowance revokes any fee allowance of granter's account that
     * has been granted to the grantee.
     * </pre>
     */
    public cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse revokeAllowance(cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRevokeAllowanceMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Msg defines the feegrant msg service.
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
     * GrantAllowance grants fee allowance to the grantee on the granter's
     * account with the provided expiration time.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse> grantAllowance(
        cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGrantAllowanceMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * RevokeAllowance revokes any fee allowance of granter's account that
     * has been granted to the grantee.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse> revokeAllowance(
        cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRevokeAllowanceMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GRANT_ALLOWANCE = 0;
  private static final int METHODID_REVOKE_ALLOWANCE = 1;

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
        case METHODID_GRANT_ALLOWANCE:
          serviceImpl.grantAllowance((cosmos.feegrant.v1beta1.Tx.MsgGrantAllowance) request,
              (io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.Tx.MsgGrantAllowanceResponse>) responseObserver);
          break;
        case METHODID_REVOKE_ALLOWANCE:
          serviceImpl.revokeAllowance((cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowance) request,
              (io.grpc.stub.StreamObserver<cosmos.feegrant.v1beta1.Tx.MsgRevokeAllowanceResponse>) responseObserver);
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
      return cosmos.feegrant.v1beta1.Tx.getDescriptor();
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
              .addMethod(getGrantAllowanceMethod())
              .addMethod(getRevokeAllowanceMethod())
              .build();
        }
      }
    }
    return result;
  }
}
