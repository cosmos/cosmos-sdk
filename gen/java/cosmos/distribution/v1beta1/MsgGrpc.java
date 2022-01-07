package cosmos.distribution.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Msg defines the distribution Msg service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/distribution/v1beta1/tx.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class MsgGrpc {

  private MsgGrpc() {}

  public static final String SERVICE_NAME = "cosmos.distribution.v1beta1.Msg";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress,
      cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse> getSetWithdrawAddressMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SetWithdrawAddress",
      requestType = cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress.class,
      responseType = cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress,
      cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse> getSetWithdrawAddressMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress, cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse> getSetWithdrawAddressMethod;
    if ((getSetWithdrawAddressMethod = MsgGrpc.getSetWithdrawAddressMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getSetWithdrawAddressMethod = MsgGrpc.getSetWithdrawAddressMethod) == null) {
          MsgGrpc.getSetWithdrawAddressMethod = getSetWithdrawAddressMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress, cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SetWithdrawAddress"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("SetWithdrawAddress"))
              .build();
        }
      }
    }
    return getSetWithdrawAddressMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward,
      cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse> getWithdrawDelegatorRewardMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "WithdrawDelegatorReward",
      requestType = cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward.class,
      responseType = cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward,
      cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse> getWithdrawDelegatorRewardMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward, cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse> getWithdrawDelegatorRewardMethod;
    if ((getWithdrawDelegatorRewardMethod = MsgGrpc.getWithdrawDelegatorRewardMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getWithdrawDelegatorRewardMethod = MsgGrpc.getWithdrawDelegatorRewardMethod) == null) {
          MsgGrpc.getWithdrawDelegatorRewardMethod = getWithdrawDelegatorRewardMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward, cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "WithdrawDelegatorReward"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("WithdrawDelegatorReward"))
              .build();
        }
      }
    }
    return getWithdrawDelegatorRewardMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission,
      cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse> getWithdrawValidatorCommissionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "WithdrawValidatorCommission",
      requestType = cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission.class,
      responseType = cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission,
      cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse> getWithdrawValidatorCommissionMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission, cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse> getWithdrawValidatorCommissionMethod;
    if ((getWithdrawValidatorCommissionMethod = MsgGrpc.getWithdrawValidatorCommissionMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getWithdrawValidatorCommissionMethod = MsgGrpc.getWithdrawValidatorCommissionMethod) == null) {
          MsgGrpc.getWithdrawValidatorCommissionMethod = getWithdrawValidatorCommissionMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission, cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "WithdrawValidatorCommission"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("WithdrawValidatorCommission"))
              .build();
        }
      }
    }
    return getWithdrawValidatorCommissionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool,
      cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse> getFundCommunityPoolMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "FundCommunityPool",
      requestType = cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool.class,
      responseType = cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool,
      cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse> getFundCommunityPoolMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool, cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse> getFundCommunityPoolMethod;
    if ((getFundCommunityPoolMethod = MsgGrpc.getFundCommunityPoolMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getFundCommunityPoolMethod = MsgGrpc.getFundCommunityPoolMethod) == null) {
          MsgGrpc.getFundCommunityPoolMethod = getFundCommunityPoolMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool, cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "FundCommunityPool"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("FundCommunityPool"))
              .build();
        }
      }
    }
    return getFundCommunityPoolMethod;
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
   * Msg defines the distribution Msg service.
   * </pre>
   */
  public static abstract class MsgImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * SetWithdrawAddress defines a method to change the withdraw address
     * for a delegator (or validator self-delegation).
     * </pre>
     */
    public void setWithdrawAddress(cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSetWithdrawAddressMethod(), responseObserver);
    }

    /**
     * <pre>
     * WithdrawDelegatorReward defines a method to withdraw rewards of delegator
     * from a single validator.
     * </pre>
     */
    public void withdrawDelegatorReward(cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getWithdrawDelegatorRewardMethod(), responseObserver);
    }

    /**
     * <pre>
     * WithdrawValidatorCommission defines a method to withdraw the
     * full commission to the validator address.
     * </pre>
     */
    public void withdrawValidatorCommission(cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getWithdrawValidatorCommissionMethod(), responseObserver);
    }

    /**
     * <pre>
     * FundCommunityPool defines a method to allow an account to directly
     * fund the community pool.
     * </pre>
     */
    public void fundCommunityPool(cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getFundCommunityPoolMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getSetWithdrawAddressMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress,
                cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse>(
                  this, METHODID_SET_WITHDRAW_ADDRESS)))
          .addMethod(
            getWithdrawDelegatorRewardMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward,
                cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse>(
                  this, METHODID_WITHDRAW_DELEGATOR_REWARD)))
          .addMethod(
            getWithdrawValidatorCommissionMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission,
                cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse>(
                  this, METHODID_WITHDRAW_VALIDATOR_COMMISSION)))
          .addMethod(
            getFundCommunityPoolMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool,
                cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse>(
                  this, METHODID_FUND_COMMUNITY_POOL)))
          .build();
    }
  }

  /**
   * <pre>
   * Msg defines the distribution Msg service.
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
     * SetWithdrawAddress defines a method to change the withdraw address
     * for a delegator (or validator self-delegation).
     * </pre>
     */
    public void setWithdrawAddress(cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSetWithdrawAddressMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * WithdrawDelegatorReward defines a method to withdraw rewards of delegator
     * from a single validator.
     * </pre>
     */
    public void withdrawDelegatorReward(cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getWithdrawDelegatorRewardMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * WithdrawValidatorCommission defines a method to withdraw the
     * full commission to the validator address.
     * </pre>
     */
    public void withdrawValidatorCommission(cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getWithdrawValidatorCommissionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * FundCommunityPool defines a method to allow an account to directly
     * fund the community pool.
     * </pre>
     */
    public void fundCommunityPool(cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getFundCommunityPoolMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Msg defines the distribution Msg service.
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
     * SetWithdrawAddress defines a method to change the withdraw address
     * for a delegator (or validator self-delegation).
     * </pre>
     */
    public cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse setWithdrawAddress(cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSetWithdrawAddressMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * WithdrawDelegatorReward defines a method to withdraw rewards of delegator
     * from a single validator.
     * </pre>
     */
    public cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse withdrawDelegatorReward(cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getWithdrawDelegatorRewardMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * WithdrawValidatorCommission defines a method to withdraw the
     * full commission to the validator address.
     * </pre>
     */
    public cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse withdrawValidatorCommission(cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getWithdrawValidatorCommissionMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * FundCommunityPool defines a method to allow an account to directly
     * fund the community pool.
     * </pre>
     */
    public cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse fundCommunityPool(cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getFundCommunityPoolMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Msg defines the distribution Msg service.
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
     * SetWithdrawAddress defines a method to change the withdraw address
     * for a delegator (or validator self-delegation).
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse> setWithdrawAddress(
        cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSetWithdrawAddressMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * WithdrawDelegatorReward defines a method to withdraw rewards of delegator
     * from a single validator.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse> withdrawDelegatorReward(
        cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getWithdrawDelegatorRewardMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * WithdrawValidatorCommission defines a method to withdraw the
     * full commission to the validator address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse> withdrawValidatorCommission(
        cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getWithdrawValidatorCommissionMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * FundCommunityPool defines a method to allow an account to directly
     * fund the community pool.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse> fundCommunityPool(
        cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getFundCommunityPoolMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_SET_WITHDRAW_ADDRESS = 0;
  private static final int METHODID_WITHDRAW_DELEGATOR_REWARD = 1;
  private static final int METHODID_WITHDRAW_VALIDATOR_COMMISSION = 2;
  private static final int METHODID_FUND_COMMUNITY_POOL = 3;

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
        case METHODID_SET_WITHDRAW_ADDRESS:
          serviceImpl.setWithdrawAddress((cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddress) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgSetWithdrawAddressResponse>) responseObserver);
          break;
        case METHODID_WITHDRAW_DELEGATOR_REWARD:
          serviceImpl.withdrawDelegatorReward((cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorReward) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgWithdrawDelegatorRewardResponse>) responseObserver);
          break;
        case METHODID_WITHDRAW_VALIDATOR_COMMISSION:
          serviceImpl.withdrawValidatorCommission((cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommission) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgWithdrawValidatorCommissionResponse>) responseObserver);
          break;
        case METHODID_FUND_COMMUNITY_POOL:
          serviceImpl.fundCommunityPool((cosmos.distribution.v1beta1.Tx.MsgFundCommunityPool) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.Tx.MsgFundCommunityPoolResponse>) responseObserver);
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
      return cosmos.distribution.v1beta1.Tx.getDescriptor();
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
              .addMethod(getSetWithdrawAddressMethod())
              .addMethod(getWithdrawDelegatorRewardMethod())
              .addMethod(getWithdrawValidatorCommissionMethod())
              .addMethod(getFundCommunityPoolMethod())
              .build();
        }
      }
    }
    return result;
  }
}
