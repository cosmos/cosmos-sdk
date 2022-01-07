package cosmos.authz.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Msg defines the authz Msg service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/authz/v1beta1/tx.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class MsgGrpc {

  private MsgGrpc() {}

  public static final String SERVICE_NAME = "cosmos.authz.v1beta1.Msg";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgGrant,
      cosmos.authz.v1beta1.Tx.MsgGrantResponse> getGrantMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Grant",
      requestType = cosmos.authz.v1beta1.Tx.MsgGrant.class,
      responseType = cosmos.authz.v1beta1.Tx.MsgGrantResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgGrant,
      cosmos.authz.v1beta1.Tx.MsgGrantResponse> getGrantMethod() {
    io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgGrant, cosmos.authz.v1beta1.Tx.MsgGrantResponse> getGrantMethod;
    if ((getGrantMethod = MsgGrpc.getGrantMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getGrantMethod = MsgGrpc.getGrantMethod) == null) {
          MsgGrpc.getGrantMethod = getGrantMethod =
              io.grpc.MethodDescriptor.<cosmos.authz.v1beta1.Tx.MsgGrant, cosmos.authz.v1beta1.Tx.MsgGrantResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Grant"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.authz.v1beta1.Tx.MsgGrant.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.authz.v1beta1.Tx.MsgGrantResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("Grant"))
              .build();
        }
      }
    }
    return getGrantMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgExec,
      cosmos.authz.v1beta1.Tx.MsgExecResponse> getExecMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Exec",
      requestType = cosmos.authz.v1beta1.Tx.MsgExec.class,
      responseType = cosmos.authz.v1beta1.Tx.MsgExecResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgExec,
      cosmos.authz.v1beta1.Tx.MsgExecResponse> getExecMethod() {
    io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgExec, cosmos.authz.v1beta1.Tx.MsgExecResponse> getExecMethod;
    if ((getExecMethod = MsgGrpc.getExecMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getExecMethod = MsgGrpc.getExecMethod) == null) {
          MsgGrpc.getExecMethod = getExecMethod =
              io.grpc.MethodDescriptor.<cosmos.authz.v1beta1.Tx.MsgExec, cosmos.authz.v1beta1.Tx.MsgExecResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Exec"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.authz.v1beta1.Tx.MsgExec.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.authz.v1beta1.Tx.MsgExecResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("Exec"))
              .build();
        }
      }
    }
    return getExecMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgRevoke,
      cosmos.authz.v1beta1.Tx.MsgRevokeResponse> getRevokeMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Revoke",
      requestType = cosmos.authz.v1beta1.Tx.MsgRevoke.class,
      responseType = cosmos.authz.v1beta1.Tx.MsgRevokeResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgRevoke,
      cosmos.authz.v1beta1.Tx.MsgRevokeResponse> getRevokeMethod() {
    io.grpc.MethodDescriptor<cosmos.authz.v1beta1.Tx.MsgRevoke, cosmos.authz.v1beta1.Tx.MsgRevokeResponse> getRevokeMethod;
    if ((getRevokeMethod = MsgGrpc.getRevokeMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getRevokeMethod = MsgGrpc.getRevokeMethod) == null) {
          MsgGrpc.getRevokeMethod = getRevokeMethod =
              io.grpc.MethodDescriptor.<cosmos.authz.v1beta1.Tx.MsgRevoke, cosmos.authz.v1beta1.Tx.MsgRevokeResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Revoke"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.authz.v1beta1.Tx.MsgRevoke.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.authz.v1beta1.Tx.MsgRevokeResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("Revoke"))
              .build();
        }
      }
    }
    return getRevokeMethod;
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
   * Msg defines the authz Msg service.
   * </pre>
   */
  public static abstract class MsgImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * Grant grants the provided authorization to the grantee on the granter's
     * account with the provided expiration time. If there is already a grant
     * for the given (granter, grantee, Authorization) triple, then the grant
     * will be overwritten.
     * </pre>
     */
    public void grant(cosmos.authz.v1beta1.Tx.MsgGrant request,
        io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgGrantResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGrantMethod(), responseObserver);
    }

    /**
     * <pre>
     * Exec attempts to execute the provided messages using
     * authorizations granted to the grantee. Each message should have only
     * one signer corresponding to the granter of the authorization.
     * </pre>
     */
    public void exec(cosmos.authz.v1beta1.Tx.MsgExec request,
        io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgExecResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getExecMethod(), responseObserver);
    }

    /**
     * <pre>
     * Revoke revokes any authorization corresponding to the provided method name on the
     * granter's account that has been granted to the grantee.
     * </pre>
     */
    public void revoke(cosmos.authz.v1beta1.Tx.MsgRevoke request,
        io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgRevokeResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRevokeMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getGrantMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.authz.v1beta1.Tx.MsgGrant,
                cosmos.authz.v1beta1.Tx.MsgGrantResponse>(
                  this, METHODID_GRANT)))
          .addMethod(
            getExecMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.authz.v1beta1.Tx.MsgExec,
                cosmos.authz.v1beta1.Tx.MsgExecResponse>(
                  this, METHODID_EXEC)))
          .addMethod(
            getRevokeMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.authz.v1beta1.Tx.MsgRevoke,
                cosmos.authz.v1beta1.Tx.MsgRevokeResponse>(
                  this, METHODID_REVOKE)))
          .build();
    }
  }

  /**
   * <pre>
   * Msg defines the authz Msg service.
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
     * Grant grants the provided authorization to the grantee on the granter's
     * account with the provided expiration time. If there is already a grant
     * for the given (granter, grantee, Authorization) triple, then the grant
     * will be overwritten.
     * </pre>
     */
    public void grant(cosmos.authz.v1beta1.Tx.MsgGrant request,
        io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgGrantResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGrantMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Exec attempts to execute the provided messages using
     * authorizations granted to the grantee. Each message should have only
     * one signer corresponding to the granter of the authorization.
     * </pre>
     */
    public void exec(cosmos.authz.v1beta1.Tx.MsgExec request,
        io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgExecResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getExecMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Revoke revokes any authorization corresponding to the provided method name on the
     * granter's account that has been granted to the grantee.
     * </pre>
     */
    public void revoke(cosmos.authz.v1beta1.Tx.MsgRevoke request,
        io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgRevokeResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRevokeMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Msg defines the authz Msg service.
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
     * Grant grants the provided authorization to the grantee on the granter's
     * account with the provided expiration time. If there is already a grant
     * for the given (granter, grantee, Authorization) triple, then the grant
     * will be overwritten.
     * </pre>
     */
    public cosmos.authz.v1beta1.Tx.MsgGrantResponse grant(cosmos.authz.v1beta1.Tx.MsgGrant request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGrantMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Exec attempts to execute the provided messages using
     * authorizations granted to the grantee. Each message should have only
     * one signer corresponding to the granter of the authorization.
     * </pre>
     */
    public cosmos.authz.v1beta1.Tx.MsgExecResponse exec(cosmos.authz.v1beta1.Tx.MsgExec request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getExecMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Revoke revokes any authorization corresponding to the provided method name on the
     * granter's account that has been granted to the grantee.
     * </pre>
     */
    public cosmos.authz.v1beta1.Tx.MsgRevokeResponse revoke(cosmos.authz.v1beta1.Tx.MsgRevoke request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRevokeMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Msg defines the authz Msg service.
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
     * Grant grants the provided authorization to the grantee on the granter's
     * account with the provided expiration time. If there is already a grant
     * for the given (granter, grantee, Authorization) triple, then the grant
     * will be overwritten.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.authz.v1beta1.Tx.MsgGrantResponse> grant(
        cosmos.authz.v1beta1.Tx.MsgGrant request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGrantMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Exec attempts to execute the provided messages using
     * authorizations granted to the grantee. Each message should have only
     * one signer corresponding to the granter of the authorization.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.authz.v1beta1.Tx.MsgExecResponse> exec(
        cosmos.authz.v1beta1.Tx.MsgExec request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getExecMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Revoke revokes any authorization corresponding to the provided method name on the
     * granter's account that has been granted to the grantee.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.authz.v1beta1.Tx.MsgRevokeResponse> revoke(
        cosmos.authz.v1beta1.Tx.MsgRevoke request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRevokeMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GRANT = 0;
  private static final int METHODID_EXEC = 1;
  private static final int METHODID_REVOKE = 2;

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
        case METHODID_GRANT:
          serviceImpl.grant((cosmos.authz.v1beta1.Tx.MsgGrant) request,
              (io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgGrantResponse>) responseObserver);
          break;
        case METHODID_EXEC:
          serviceImpl.exec((cosmos.authz.v1beta1.Tx.MsgExec) request,
              (io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgExecResponse>) responseObserver);
          break;
        case METHODID_REVOKE:
          serviceImpl.revoke((cosmos.authz.v1beta1.Tx.MsgRevoke) request,
              (io.grpc.stub.StreamObserver<cosmos.authz.v1beta1.Tx.MsgRevokeResponse>) responseObserver);
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
      return cosmos.authz.v1beta1.Tx.getDescriptor();
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
              .addMethod(getGrantMethod())
              .addMethod(getExecMethod())
              .addMethod(getRevokeMethod())
              .build();
        }
      }
    }
    return result;
  }
}
