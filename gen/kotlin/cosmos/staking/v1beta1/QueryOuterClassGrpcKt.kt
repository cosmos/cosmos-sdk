package cosmos.staking.v1beta1

import cosmos.staking.v1beta1.QueryGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.staking.v1beta1.Query.
 */
object QueryGrpcKt {
  const val SERVICE_NAME: String = QueryGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = QueryGrpc.getServiceDescriptor()

  val validatorsMethod: MethodDescriptor<QueryOuterClass.QueryValidatorsRequest,
      QueryOuterClass.QueryValidatorsResponse>
    @JvmStatic
    get() = QueryGrpc.getValidatorsMethod()

  val validatorMethod: MethodDescriptor<QueryOuterClass.QueryValidatorRequest,
      QueryOuterClass.QueryValidatorResponse>
    @JvmStatic
    get() = QueryGrpc.getValidatorMethod()

  val validatorDelegationsMethod: MethodDescriptor<QueryOuterClass.QueryValidatorDelegationsRequest,
      QueryOuterClass.QueryValidatorDelegationsResponse>
    @JvmStatic
    get() = QueryGrpc.getValidatorDelegationsMethod()

  val validatorUnbondingDelegationsMethod:
      MethodDescriptor<QueryOuterClass.QueryValidatorUnbondingDelegationsRequest,
      QueryOuterClass.QueryValidatorUnbondingDelegationsResponse>
    @JvmStatic
    get() = QueryGrpc.getValidatorUnbondingDelegationsMethod()

  val delegationMethod: MethodDescriptor<QueryOuterClass.QueryDelegationRequest,
      QueryOuterClass.QueryDelegationResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegationMethod()

  val unbondingDelegationMethod: MethodDescriptor<QueryOuterClass.QueryUnbondingDelegationRequest,
      QueryOuterClass.QueryUnbondingDelegationResponse>
    @JvmStatic
    get() = QueryGrpc.getUnbondingDelegationMethod()

  val delegatorDelegationsMethod: MethodDescriptor<QueryOuterClass.QueryDelegatorDelegationsRequest,
      QueryOuterClass.QueryDelegatorDelegationsResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegatorDelegationsMethod()

  val delegatorUnbondingDelegationsMethod:
      MethodDescriptor<QueryOuterClass.QueryDelegatorUnbondingDelegationsRequest,
      QueryOuterClass.QueryDelegatorUnbondingDelegationsResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegatorUnbondingDelegationsMethod()

  val redelegationsMethod: MethodDescriptor<QueryOuterClass.QueryRedelegationsRequest,
      QueryOuterClass.QueryRedelegationsResponse>
    @JvmStatic
    get() = QueryGrpc.getRedelegationsMethod()

  val delegatorValidatorsMethod: MethodDescriptor<QueryOuterClass.QueryDelegatorValidatorsRequest,
      QueryOuterClass.QueryDelegatorValidatorsResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegatorValidatorsMethod()

  val delegatorValidatorMethod: MethodDescriptor<QueryOuterClass.QueryDelegatorValidatorRequest,
      QueryOuterClass.QueryDelegatorValidatorResponse>
    @JvmStatic
    get() = QueryGrpc.getDelegatorValidatorMethod()

  val historicalInfoMethod: MethodDescriptor<QueryOuterClass.QueryHistoricalInfoRequest,
      QueryOuterClass.QueryHistoricalInfoResponse>
    @JvmStatic
    get() = QueryGrpc.getHistoricalInfoMethod()

  val poolMethod: MethodDescriptor<QueryOuterClass.QueryPoolRequest,
      QueryOuterClass.QueryPoolResponse>
    @JvmStatic
    get() = QueryGrpc.getPoolMethod()

  val paramsMethod: MethodDescriptor<QueryOuterClass.QueryParamsRequest,
      QueryOuterClass.QueryParamsResponse>
    @JvmStatic
    get() = QueryGrpc.getParamsMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.staking.v1beta1.Query service as suspending coroutines.
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
    suspend fun validators(request: QueryOuterClass.QueryValidatorsRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryValidatorsResponse = unaryRpc(
      channel,
      QueryGrpc.getValidatorsMethod(),
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
    suspend fun validator(request: QueryOuterClass.QueryValidatorRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryValidatorResponse = unaryRpc(
      channel,
      QueryGrpc.getValidatorMethod(),
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
    suspend fun validatorDelegations(request: QueryOuterClass.QueryValidatorDelegationsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryValidatorDelegationsResponse =
        unaryRpc(
      channel,
      QueryGrpc.getValidatorDelegationsMethod(),
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
        fun validatorUnbondingDelegations(request: QueryOuterClass.QueryValidatorUnbondingDelegationsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryValidatorUnbondingDelegationsResponse
        = unaryRpc(
      channel,
      QueryGrpc.getValidatorUnbondingDelegationsMethod(),
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
    suspend fun delegation(request: QueryOuterClass.QueryDelegationRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryDelegationResponse = unaryRpc(
      channel,
      QueryGrpc.getDelegationMethod(),
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
    suspend fun unbondingDelegation(request: QueryOuterClass.QueryUnbondingDelegationRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryUnbondingDelegationResponse =
        unaryRpc(
      channel,
      QueryGrpc.getUnbondingDelegationMethod(),
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
    suspend fun delegatorDelegations(request: QueryOuterClass.QueryDelegatorDelegationsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryDelegatorDelegationsResponse =
        unaryRpc(
      channel,
      QueryGrpc.getDelegatorDelegationsMethod(),
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
        fun delegatorUnbondingDelegations(request: QueryOuterClass.QueryDelegatorUnbondingDelegationsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryDelegatorUnbondingDelegationsResponse
        = unaryRpc(
      channel,
      QueryGrpc.getDelegatorUnbondingDelegationsMethod(),
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
    suspend fun redelegations(request: QueryOuterClass.QueryRedelegationsRequest, headers: Metadata
        = Metadata()): QueryOuterClass.QueryRedelegationsResponse = unaryRpc(
      channel,
      QueryGrpc.getRedelegationsMethod(),
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
    suspend fun delegatorValidator(request: QueryOuterClass.QueryDelegatorValidatorRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryDelegatorValidatorResponse = unaryRpc(
      channel,
      QueryGrpc.getDelegatorValidatorMethod(),
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
    suspend fun historicalInfo(request: QueryOuterClass.QueryHistoricalInfoRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryHistoricalInfoResponse = unaryRpc(
      channel,
      QueryGrpc.getHistoricalInfoMethod(),
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
    suspend fun pool(request: QueryOuterClass.QueryPoolRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QueryPoolResponse = unaryRpc(
      channel,
      QueryGrpc.getPoolMethod(),
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
    suspend fun params(request: QueryOuterClass.QueryParamsRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QueryParamsResponse = unaryRpc(
      channel,
      QueryGrpc.getParamsMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.staking.v1beta1.Query service based on Kotlin coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.Validators.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun validators(request: QueryOuterClass.QueryValidatorsRequest):
        QueryOuterClass.QueryValidatorsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.Validators is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.Validator.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun validator(request: QueryOuterClass.QueryValidatorRequest):
        QueryOuterClass.QueryValidatorResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.Validator is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.ValidatorDelegations.
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
        fun validatorDelegations(request: QueryOuterClass.QueryValidatorDelegationsRequest):
        QueryOuterClass.QueryValidatorDelegationsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.ValidatorDelegations is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.staking.v1beta1.Query.ValidatorUnbondingDelegations.
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
        fun validatorUnbondingDelegations(request: QueryOuterClass.QueryValidatorUnbondingDelegationsRequest):
        QueryOuterClass.QueryValidatorUnbondingDelegationsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.ValidatorUnbondingDelegations is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.Delegation.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun delegation(request: QueryOuterClass.QueryDelegationRequest):
        QueryOuterClass.QueryDelegationResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.Delegation is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.UnbondingDelegation.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun unbondingDelegation(request: QueryOuterClass.QueryUnbondingDelegationRequest):
        QueryOuterClass.QueryUnbondingDelegationResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.UnbondingDelegation is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.DelegatorDelegations.
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
        fun delegatorDelegations(request: QueryOuterClass.QueryDelegatorDelegationsRequest):
        QueryOuterClass.QueryDelegatorDelegationsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.DelegatorDelegations is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.staking.v1beta1.Query.DelegatorUnbondingDelegations.
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
        fun delegatorUnbondingDelegations(request: QueryOuterClass.QueryDelegatorUnbondingDelegationsRequest):
        QueryOuterClass.QueryDelegatorUnbondingDelegationsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.DelegatorUnbondingDelegations is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.Redelegations.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun redelegations(request: QueryOuterClass.QueryRedelegationsRequest):
        QueryOuterClass.QueryRedelegationsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.Redelegations is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.DelegatorValidators.
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
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.DelegatorValidators is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.DelegatorValidator.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun delegatorValidator(request: QueryOuterClass.QueryDelegatorValidatorRequest):
        QueryOuterClass.QueryDelegatorValidatorResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.DelegatorValidator is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.HistoricalInfo.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun historicalInfo(request: QueryOuterClass.QueryHistoricalInfoRequest):
        QueryOuterClass.QueryHistoricalInfoResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.HistoricalInfo is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.Pool.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun pool(request: QueryOuterClass.QueryPoolRequest):
        QueryOuterClass.QueryPoolResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.Pool is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Query.Params.
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
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Query.Params is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getValidatorsMethod(),
      implementation = ::validators
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getValidatorMethod(),
      implementation = ::validator
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getValidatorDelegationsMethod(),
      implementation = ::validatorDelegations
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getValidatorUnbondingDelegationsMethod(),
      implementation = ::validatorUnbondingDelegations
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegationMethod(),
      implementation = ::delegation
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getUnbondingDelegationMethod(),
      implementation = ::unbondingDelegation
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegatorDelegationsMethod(),
      implementation = ::delegatorDelegations
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegatorUnbondingDelegationsMethod(),
      implementation = ::delegatorUnbondingDelegations
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getRedelegationsMethod(),
      implementation = ::redelegations
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegatorValidatorsMethod(),
      implementation = ::delegatorValidators
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDelegatorValidatorMethod(),
      implementation = ::delegatorValidator
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getHistoricalInfoMethod(),
      implementation = ::historicalInfo
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getPoolMethod(),
      implementation = ::pool
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getParamsMethod(),
      implementation = ::params
    )).build()
  }
}
