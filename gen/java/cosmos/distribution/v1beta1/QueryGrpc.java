package cosmos.distribution.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Query defines the gRPC querier service for distribution module.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/distribution/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryGrpc {

  private QueryGrpc() {}

  public static final String SERVICE_NAME = "cosmos.distribution.v1beta1.Query";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Params",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod;
    if ((getParamsMethod = QueryGrpc.getParamsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getParamsMethod = QueryGrpc.getParamsMethod) == null) {
          QueryGrpc.getParamsMethod = getParamsMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Params"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Params"))
              .build();
        }
      }
    }
    return getParamsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse> getValidatorOutstandingRewardsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ValidatorOutstandingRewards",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse> getValidatorOutstandingRewardsMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse> getValidatorOutstandingRewardsMethod;
    if ((getValidatorOutstandingRewardsMethod = QueryGrpc.getValidatorOutstandingRewardsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getValidatorOutstandingRewardsMethod = QueryGrpc.getValidatorOutstandingRewardsMethod) == null) {
          QueryGrpc.getValidatorOutstandingRewardsMethod = getValidatorOutstandingRewardsMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ValidatorOutstandingRewards"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("ValidatorOutstandingRewards"))
              .build();
        }
      }
    }
    return getValidatorOutstandingRewardsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse> getValidatorCommissionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ValidatorCommission",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse> getValidatorCommissionMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse> getValidatorCommissionMethod;
    if ((getValidatorCommissionMethod = QueryGrpc.getValidatorCommissionMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getValidatorCommissionMethod = QueryGrpc.getValidatorCommissionMethod) == null) {
          QueryGrpc.getValidatorCommissionMethod = getValidatorCommissionMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ValidatorCommission"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("ValidatorCommission"))
              .build();
        }
      }
    }
    return getValidatorCommissionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse> getValidatorSlashesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ValidatorSlashes",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse> getValidatorSlashesMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse> getValidatorSlashesMethod;
    if ((getValidatorSlashesMethod = QueryGrpc.getValidatorSlashesMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getValidatorSlashesMethod = QueryGrpc.getValidatorSlashesMethod) == null) {
          QueryGrpc.getValidatorSlashesMethod = getValidatorSlashesMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ValidatorSlashes"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("ValidatorSlashes"))
              .build();
        }
      }
    }
    return getValidatorSlashesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse> getDelegationRewardsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DelegationRewards",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse> getDelegationRewardsMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse> getDelegationRewardsMethod;
    if ((getDelegationRewardsMethod = QueryGrpc.getDelegationRewardsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getDelegationRewardsMethod = QueryGrpc.getDelegationRewardsMethod) == null) {
          QueryGrpc.getDelegationRewardsMethod = getDelegationRewardsMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DelegationRewards"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("DelegationRewards"))
              .build();
        }
      }
    }
    return getDelegationRewardsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse> getDelegationTotalRewardsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DelegationTotalRewards",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse> getDelegationTotalRewardsMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse> getDelegationTotalRewardsMethod;
    if ((getDelegationTotalRewardsMethod = QueryGrpc.getDelegationTotalRewardsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getDelegationTotalRewardsMethod = QueryGrpc.getDelegationTotalRewardsMethod) == null) {
          QueryGrpc.getDelegationTotalRewardsMethod = getDelegationTotalRewardsMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DelegationTotalRewards"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("DelegationTotalRewards"))
              .build();
        }
      }
    }
    return getDelegationTotalRewardsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse> getDelegatorValidatorsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DelegatorValidators",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse> getDelegatorValidatorsMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse> getDelegatorValidatorsMethod;
    if ((getDelegatorValidatorsMethod = QueryGrpc.getDelegatorValidatorsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getDelegatorValidatorsMethod = QueryGrpc.getDelegatorValidatorsMethod) == null) {
          QueryGrpc.getDelegatorValidatorsMethod = getDelegatorValidatorsMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DelegatorValidators"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("DelegatorValidators"))
              .build();
        }
      }
    }
    return getDelegatorValidatorsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse> getDelegatorWithdrawAddressMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DelegatorWithdrawAddress",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse> getDelegatorWithdrawAddressMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse> getDelegatorWithdrawAddressMethod;
    if ((getDelegatorWithdrawAddressMethod = QueryGrpc.getDelegatorWithdrawAddressMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getDelegatorWithdrawAddressMethod = QueryGrpc.getDelegatorWithdrawAddressMethod) == null) {
          QueryGrpc.getDelegatorWithdrawAddressMethod = getDelegatorWithdrawAddressMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DelegatorWithdrawAddress"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("DelegatorWithdrawAddress"))
              .build();
        }
      }
    }
    return getDelegatorWithdrawAddressMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse> getCommunityPoolMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CommunityPool",
      requestType = cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest.class,
      responseType = cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest,
      cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse> getCommunityPoolMethod() {
    io.grpc.MethodDescriptor<cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse> getCommunityPoolMethod;
    if ((getCommunityPoolMethod = QueryGrpc.getCommunityPoolMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getCommunityPoolMethod = QueryGrpc.getCommunityPoolMethod) == null) {
          QueryGrpc.getCommunityPoolMethod = getCommunityPoolMethod =
              io.grpc.MethodDescriptor.<cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest, cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CommunityPool"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("CommunityPool"))
              .build();
        }
      }
    }
    return getCommunityPoolMethod;
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
   * Query defines the gRPC querier service for distribution module.
   * </pre>
   */
  public static abstract class QueryImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * Params queries params of the distribution module.
     * </pre>
     */
    public void params(cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getParamsMethod(), responseObserver);
    }

    /**
     * <pre>
     * ValidatorOutstandingRewards queries rewards of a validator address.
     * </pre>
     */
    public void validatorOutstandingRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getValidatorOutstandingRewardsMethod(), responseObserver);
    }

    /**
     * <pre>
     * ValidatorCommission queries accumulated commission for a validator.
     * </pre>
     */
    public void validatorCommission(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getValidatorCommissionMethod(), responseObserver);
    }

    /**
     * <pre>
     * ValidatorSlashes queries slash events of a validator.
     * </pre>
     */
    public void validatorSlashes(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getValidatorSlashesMethod(), responseObserver);
    }

    /**
     * <pre>
     * DelegationRewards queries the total rewards accrued by a delegation.
     * </pre>
     */
    public void delegationRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDelegationRewardsMethod(), responseObserver);
    }

    /**
     * <pre>
     * DelegationTotalRewards queries the total rewards accrued by a each
     * validator.
     * </pre>
     */
    public void delegationTotalRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDelegationTotalRewardsMethod(), responseObserver);
    }

    /**
     * <pre>
     * DelegatorValidators queries the validators of a delegator.
     * </pre>
     */
    public void delegatorValidators(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDelegatorValidatorsMethod(), responseObserver);
    }

    /**
     * <pre>
     * DelegatorWithdrawAddress queries withdraw address of a delegator.
     * </pre>
     */
    public void delegatorWithdrawAddress(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDelegatorWithdrawAddressMethod(), responseObserver);
    }

    /**
     * <pre>
     * CommunityPool queries the community pool coins.
     * </pre>
     */
    public void communityPool(cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCommunityPoolMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getParamsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse>(
                  this, METHODID_PARAMS)))
          .addMethod(
            getValidatorOutstandingRewardsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse>(
                  this, METHODID_VALIDATOR_OUTSTANDING_REWARDS)))
          .addMethod(
            getValidatorCommissionMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse>(
                  this, METHODID_VALIDATOR_COMMISSION)))
          .addMethod(
            getValidatorSlashesMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse>(
                  this, METHODID_VALIDATOR_SLASHES)))
          .addMethod(
            getDelegationRewardsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse>(
                  this, METHODID_DELEGATION_REWARDS)))
          .addMethod(
            getDelegationTotalRewardsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse>(
                  this, METHODID_DELEGATION_TOTAL_REWARDS)))
          .addMethod(
            getDelegatorValidatorsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse>(
                  this, METHODID_DELEGATOR_VALIDATORS)))
          .addMethod(
            getDelegatorWithdrawAddressMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse>(
                  this, METHODID_DELEGATOR_WITHDRAW_ADDRESS)))
          .addMethod(
            getCommunityPoolMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest,
                cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse>(
                  this, METHODID_COMMUNITY_POOL)))
          .build();
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service for distribution module.
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
     * Params queries params of the distribution module.
     * </pre>
     */
    public void params(cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getParamsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * ValidatorOutstandingRewards queries rewards of a validator address.
     * </pre>
     */
    public void validatorOutstandingRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getValidatorOutstandingRewardsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * ValidatorCommission queries accumulated commission for a validator.
     * </pre>
     */
    public void validatorCommission(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getValidatorCommissionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * ValidatorSlashes queries slash events of a validator.
     * </pre>
     */
    public void validatorSlashes(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getValidatorSlashesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * DelegationRewards queries the total rewards accrued by a delegation.
     * </pre>
     */
    public void delegationRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDelegationRewardsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * DelegationTotalRewards queries the total rewards accrued by a each
     * validator.
     * </pre>
     */
    public void delegationTotalRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDelegationTotalRewardsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * DelegatorValidators queries the validators of a delegator.
     * </pre>
     */
    public void delegatorValidators(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDelegatorValidatorsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * DelegatorWithdrawAddress queries withdraw address of a delegator.
     * </pre>
     */
    public void delegatorWithdrawAddress(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDelegatorWithdrawAddressMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * CommunityPool queries the community pool coins.
     * </pre>
     */
    public void communityPool(cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest request,
        io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCommunityPoolMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service for distribution module.
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
     * Params queries params of the distribution module.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse params(cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getParamsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * ValidatorOutstandingRewards queries rewards of a validator address.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse validatorOutstandingRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getValidatorOutstandingRewardsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * ValidatorCommission queries accumulated commission for a validator.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse validatorCommission(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getValidatorCommissionMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * ValidatorSlashes queries slash events of a validator.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse validatorSlashes(cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getValidatorSlashesMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * DelegationRewards queries the total rewards accrued by a delegation.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse delegationRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDelegationRewardsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * DelegationTotalRewards queries the total rewards accrued by a each
     * validator.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse delegationTotalRewards(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDelegationTotalRewardsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * DelegatorValidators queries the validators of a delegator.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse delegatorValidators(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDelegatorValidatorsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * DelegatorWithdrawAddress queries withdraw address of a delegator.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse delegatorWithdrawAddress(cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDelegatorWithdrawAddressMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * CommunityPool queries the community pool coins.
     * </pre>
     */
    public cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse communityPool(cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCommunityPoolMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service for distribution module.
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
     * Params queries params of the distribution module.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse> params(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getParamsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * ValidatorOutstandingRewards queries rewards of a validator address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse> validatorOutstandingRewards(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getValidatorOutstandingRewardsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * ValidatorCommission queries accumulated commission for a validator.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse> validatorCommission(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getValidatorCommissionMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * ValidatorSlashes queries slash events of a validator.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse> validatorSlashes(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getValidatorSlashesMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * DelegationRewards queries the total rewards accrued by a delegation.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse> delegationRewards(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDelegationRewardsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * DelegationTotalRewards queries the total rewards accrued by a each
     * validator.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse> delegationTotalRewards(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDelegationTotalRewardsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * DelegatorValidators queries the validators of a delegator.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse> delegatorValidators(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDelegatorValidatorsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * DelegatorWithdrawAddress queries withdraw address of a delegator.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse> delegatorWithdrawAddress(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDelegatorWithdrawAddressMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * CommunityPool queries the community pool coins.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse> communityPool(
        cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCommunityPoolMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PARAMS = 0;
  private static final int METHODID_VALIDATOR_OUTSTANDING_REWARDS = 1;
  private static final int METHODID_VALIDATOR_COMMISSION = 2;
  private static final int METHODID_VALIDATOR_SLASHES = 3;
  private static final int METHODID_DELEGATION_REWARDS = 4;
  private static final int METHODID_DELEGATION_TOTAL_REWARDS = 5;
  private static final int METHODID_DELEGATOR_VALIDATORS = 6;
  private static final int METHODID_DELEGATOR_WITHDRAW_ADDRESS = 7;
  private static final int METHODID_COMMUNITY_POOL = 8;

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
          serviceImpl.params((cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryParamsResponse>) responseObserver);
          break;
        case METHODID_VALIDATOR_OUTSTANDING_REWARDS:
          serviceImpl.validatorOutstandingRewards((cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorOutstandingRewardsResponse>) responseObserver);
          break;
        case METHODID_VALIDATOR_COMMISSION:
          serviceImpl.validatorCommission((cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorCommissionResponse>) responseObserver);
          break;
        case METHODID_VALIDATOR_SLASHES:
          serviceImpl.validatorSlashes((cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryValidatorSlashesResponse>) responseObserver);
          break;
        case METHODID_DELEGATION_REWARDS:
          serviceImpl.delegationRewards((cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationRewardsResponse>) responseObserver);
          break;
        case METHODID_DELEGATION_TOTAL_REWARDS:
          serviceImpl.delegationTotalRewards((cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegationTotalRewardsResponse>) responseObserver);
          break;
        case METHODID_DELEGATOR_VALIDATORS:
          serviceImpl.delegatorValidators((cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorValidatorsResponse>) responseObserver);
          break;
        case METHODID_DELEGATOR_WITHDRAW_ADDRESS:
          serviceImpl.delegatorWithdrawAddress((cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryDelegatorWithdrawAddressResponse>) responseObserver);
          break;
        case METHODID_COMMUNITY_POOL:
          serviceImpl.communityPool((cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.distribution.v1beta1.QueryOuterClass.QueryCommunityPoolResponse>) responseObserver);
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
      return cosmos.distribution.v1beta1.QueryOuterClass.getDescriptor();
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
              .addMethod(getValidatorOutstandingRewardsMethod())
              .addMethod(getValidatorCommissionMethod())
              .addMethod(getValidatorSlashesMethod())
              .addMethod(getDelegationRewardsMethod())
              .addMethod(getDelegationTotalRewardsMethod())
              .addMethod(getDelegatorValidatorsMethod())
              .addMethod(getDelegatorWithdrawAddressMethod())
              .addMethod(getCommunityPoolMethod())
              .build();
        }
      }
    }
    return result;
  }
}
