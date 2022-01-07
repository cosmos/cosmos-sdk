package cosmos.distribution.v1beta1

import cosmos.distribution.v1beta1.QueryGrpc.getServiceDescriptor
import io.grpc.CallOptions
import io.grpc.CallOptions.DEFAULT
import io.grpc.Channel
import io.grpc.Metadata
import io.grpc.MethodDescriptor
import io.grpc.ServerServiceDefinition
import io.grpc.ServerServiceDefinition.builder
import io.grpc.ServiceDescriptor
import io.grpc.Status
import io.grpc.Status.UNIMPLEMENTED
import io.grpc.StatusException
import io.grpc.kotlin.AbstractCoroutineServerImpl
import io.grpc.kotlin.AbstractCoroutineStub
import io.grpc.kotlin.ClientCalls
import io.grpc.kotlin.ClientCalls.unaryRpc
import io.grpc.kotlin.ServerCalls
import io.grpc.kotlin.ServerCalls.unaryServerMethodDefinition
import io.grpc.kotlin.StubFor
import kotlin.String
import kotlin.coroutines.CoroutineContext
import kotlin.coroutines.EmptyCoroutineContext
import kotlin.jvm.JvmOverloads
import kotlin.jvm.JvmStatic

/**
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.distribution.v1beta1.Query.
 */
object QueryGrpcKt {
  const val SERVICE_NAME: String = QueryGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = QueryGrpc.getServiceDescriptor()

  val paramsMethod: MethodDescriptor<QueryOuterClass.QueryParamsRequest,
      QueryOuterClass.QueryParamsResponse>
    @JvmStatic
    get() = QueryGrpc.getParamsMethod()

  val validatorOutstandingRewardsMethod:
      MethodDescriptor<QueryOuterClass.QueryValidatorOutstandingRewardsRequest,
      QueryOuterClass.QueryValidatorOutstandingRewardsResponse>
    @JvmStatic
    get() = QueryGrpc.getValidatorOutstandingRewardsMethod()

  val validatorCommissionMethod: MethodDescriptor<QueryOuterClass.QueryValidatorCommissionRequest,
      QueryOuterClass.QueryValidatorCommissionResponse>
    @JvmStatic
    get() = QueryGrpc.getValidatorCommissionMethod()

  val validatorSlashesMethod: MethodDescriptor<QueryOuterClass.QueryValidatorSlashesRequest,
      QueryOuterClass.QueryValidatorSlashesResponse>
    @JvmStatic
    get() = QueryGrpc.getValidatorSlashesMethod()

  val delegationRewardsMethod: MethodDescriptor<QueryOuterClass.QueryDelegationRewardsRequest,
      QueryOuterClass.QueryDelegationRewardsResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegationRewardsMethod()

  val delegationTotalRewardsMethod:
      MethodDescriptor<QueryOuterClass.QueryDelegationTotalRewardsRequest,
      QueryOuterClass.QueryDelegationTotalRewardsResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegationTotalRewardsMethod()

  val delegatorValidatorsMethod: MethodDescriptor<QueryOuterClass.QueryDelegatorValidatorsRequest,
      QueryOuterClass.QueryDelegatorValidatorsResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegatorValidatorsMethod()

  val delegatorWithdrawAddressMethod:
      MethodDescriptor<QueryOuterClass.QueryDelegatorWithdrawAddressRequest,
      QueryOuterClass.QueryDelegatorWithdrawAddressResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegatorWithdrawAddressMethod()

  val communityPoolMethod: MethodDescriptor<QueryOuterClass.QueryCommunityPoolRequest,
      QueryOuterClass.QueryCommunityPoolResponse>
    @JvmStatic
    get() = QueryGrpc.getCommunityPoolMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.distribution.v1beta1.Query service as suspending
   * coroutines.
   */
  @StubFor(QueryGrpc::class)
  class QueryCoroutineStub @JvmOverloads constructor(
    channel: Channel,
    callOptions: CallOptions = DEFAULT
  ) : AbstractCoroutineStub<QueryCoroutineStub>(channel, callOptions) {
    override fun build(channel: Channel, callOptions: CallOptions): QueryCoroutineStub =
        QueryCoroutineStub(channel, callOptions)

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun params(request: QueryOuterClass.QueryParamsRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QueryParamsResponse = unaryRpc(
      channel,
      QueryGrpc.getParamsMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend
        fun validatorOutstandingRewards(request: QueryOuterClass.QueryValidatorOutstandingRewardsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryValidatorOutstandingRewardsResponse =
        unaryRpc(
      channel,
      QueryGrpc.getValidatorOutstandingRewardsMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun validatorCommission(request: QueryOuterClass.QueryValidatorCommissionRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryValidatorCommissionResponse =
        unaryRpc(
      channel,
      QueryGrpc.getValidatorCommissionMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun validatorSlashes(request: QueryOuterClass.QueryValidatorSlashesRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryValidatorSlashesResponse = unaryRpc(
      channel,
      QueryGrpc.getValidatorSlashesMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun delegationRewards(request: QueryOuterClass.QueryDelegationRewardsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryDelegationRewardsResponse = unaryRpc(
      channel,
      QueryGrpc.getDelegationRewardsMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun delegationTotalRewards(request: QueryOuterClass.QueryDelegationTotalRewardsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryDelegationTotalRewardsResponse =
        unaryRpc(
      channel,
      QueryGrpc.getDelegationTotalRewardsMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun delegatorValidators(request: QueryOuterClass.QueryDelegatorValidatorsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryDelegatorValidatorsResponse =
        unaryRpc(
      channel,
      QueryGrpc.getDelegatorValidatorsMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend
        fun delegatorWithdrawAddress(request: QueryOuterClass.QueryDelegatorWithdrawAddressRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryDelegatorWithdrawAddressResponse =
        unaryRpc(
      channel,
      QueryGrpc.getDelegatorWithdrawAddressMethod(),
      request,
      callOptions,
      headers
    )
    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][Status].  If the RPC completes with another status, a corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun communityPool(request: QueryOuterClass.QueryCommunityPoolRequest, headers: Metadata
        = Metadata()): QueryOuterClass.QueryCommunityPoolResponse = unaryRpc(
      channel,
      QueryGrpc.getCommunityPoolMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.distribution.v1beta1.Query service based on Kotlin
   * coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Query.Params.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun params(request: QueryOuterClass.QueryParamsRequest):
        QueryOuterClass.QueryParamsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.Params is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.distribution.v1beta1.Query.ValidatorOutstandingRewards.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend
        fun validatorOutstandingRewards(request: QueryOuterClass.QueryValidatorOutstandingRewardsRequest):
        QueryOuterClass.QueryValidatorOutstandingRewardsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.ValidatorOutstandingRewards is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Query.ValidatorCommission.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun validatorCommission(request: QueryOuterClass.QueryValidatorCommissionRequest):
        QueryOuterClass.QueryValidatorCommissionResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.ValidatorCommission is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Query.ValidatorSlashes.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun validatorSlashes(request: QueryOuterClass.QueryValidatorSlashesRequest):
        QueryOuterClass.QueryValidatorSlashesResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.ValidatorSlashes is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Query.DelegationRewards.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun delegationRewards(request: QueryOuterClass.QueryDelegationRewardsRequest):
        QueryOuterClass.QueryDelegationRewardsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.DelegationRewards is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Query.DelegationTotalRewards.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend
        fun delegationTotalRewards(request: QueryOuterClass.QueryDelegationTotalRewardsRequest):
        QueryOuterClass.QueryDelegationTotalRewardsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.DelegationTotalRewards is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Query.DelegatorValidators.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun delegatorValidators(request: QueryOuterClass.QueryDelegatorValidatorsRequest):
        QueryOuterClass.QueryDelegatorValidatorsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.DelegatorValidators is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.distribution.v1beta1.Query.DelegatorWithdrawAddress.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend
        fun delegatorWithdrawAddress(request: QueryOuterClass.QueryDelegatorWithdrawAddressRequest):
        QueryOuterClass.QueryDelegatorWithdrawAddressResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.DelegatorWithdrawAddress is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Query.CommunityPool.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun communityPool(request: QueryOuterClass.QueryCommunityPoolRequest):
        QueryOuterClass.QueryCommunityPoolResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Query.CommunityPool is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getParamsMethod(),
      implementation = ::params
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getValidatorOutstandingRewardsMethod(),
      implementation = ::validatorOutstandingRewards
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getValidatorCommissionMethod(),
      implementation = ::validatorCommission
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getValidatorSlashesMethod(),
      implementation = ::validatorSlashes
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegationRewardsMethod(),
      implementation = ::delegationRewards
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegationTotalRewardsMethod(),
      implementation = ::delegationTotalRewards
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegatorValidatorsMethod(),
      implementation = ::delegatorValidators
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegatorWithdrawAddressMethod(),
      implementation = ::delegatorWithdrawAddress
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getCommunityPoolMethod(),
      implementation = ::communityPool
    )).build()
  }
}
