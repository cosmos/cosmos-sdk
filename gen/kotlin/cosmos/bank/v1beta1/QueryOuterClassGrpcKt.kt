package cosmos.bank.v1beta1

import cosmos.bank.v1beta1.QueryGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.bank.v1beta1.Query.
 */
object QueryGrpcKt {
  const val SERVICE_NAME: String = QueryGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = QueryGrpc.getServiceDescriptor()

  val balanceMethod: MethodDescriptor<QueryOuterClass.QueryBalanceRequest,
      QueryOuterClass.QueryBalanceResponse>
    @JvmStatic
    get() = QueryGrpc.getBalanceMethod()

  val allBalancesMethod: MethodDescriptor<QueryOuterClass.QueryAllBalancesRequest,
      QueryOuterClass.QueryAllBalancesResponse>
    @JvmStatic
    get() = QueryGrpc.getAllBalancesMethod()

  val totalSupplyMethod: MethodDescriptor<QueryOuterClass.QueryTotalSupplyRequest,
      QueryOuterClass.QueryTotalSupplyResponse>
    @JvmStatic
    get() = QueryGrpc.getTotalSupplyMethod()

  val supplyOfMethod: MethodDescriptor<QueryOuterClass.QuerySupplyOfRequest,
      QueryOuterClass.QuerySupplyOfResponse>
    @JvmStatic
    get() = QueryGrpc.getSupplyOfMethod()

  val paramsMethod: MethodDescriptor<QueryOuterClass.QueryParamsRequest,
      QueryOuterClass.QueryParamsResponse>
    @JvmStatic
    get() = QueryGrpc.getParamsMethod()

  val denomMetadataMethod: MethodDescriptor<QueryOuterClass.QueryDenomMetadataRequest,
      QueryOuterClass.QueryDenomMetadataResponse>
    @JvmStatic
    get() = QueryGrpc.getDenomMetadataMethod()

  val denomsMetadataMethod: MethodDescriptor<QueryOuterClass.QueryDenomsMetadataRequest,
      QueryOuterClass.QueryDenomsMetadataResponse>
    @JvmStatic
    get() = QueryGrpc.getDenomsMetadataMethod()

  val denomOwnersMethod: MethodDescriptor<QueryOuterClass.QueryDenomOwnersRequest,
      QueryOuterClass.QueryDenomOwnersResponse>
    @JvmStatic
    get() = QueryGrpc.getDenomOwnersMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.bank.v1beta1.Query service as suspending coroutines.
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
    suspend fun balance(request: QueryOuterClass.QueryBalanceRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryBalanceResponse = unaryRpc(
      channel,
      QueryGrpc.getBalanceMethod(),
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
    suspend fun allBalances(request: QueryOuterClass.QueryAllBalancesRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryAllBalancesResponse = unaryRpc(
      channel,
      QueryGrpc.getAllBalancesMethod(),
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
    suspend fun totalSupply(request: QueryOuterClass.QueryTotalSupplyRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryTotalSupplyResponse = unaryRpc(
      channel,
      QueryGrpc.getTotalSupplyMethod(),
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
    suspend fun supplyOf(request: QueryOuterClass.QuerySupplyOfRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QuerySupplyOfResponse = unaryRpc(
      channel,
      QueryGrpc.getSupplyOfMethod(),
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
    suspend fun denomMetadata(request: QueryOuterClass.QueryDenomMetadataRequest, headers: Metadata
        = Metadata()): QueryOuterClass.QueryDenomMetadataResponse = unaryRpc(
      channel,
      QueryGrpc.getDenomMetadataMethod(),
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
    suspend fun denomsMetadata(request: QueryOuterClass.QueryDenomsMetadataRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryDenomsMetadataResponse = unaryRpc(
      channel,
      QueryGrpc.getDenomsMetadataMethod(),
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
    suspend fun denomOwners(request: QueryOuterClass.QueryDenomOwnersRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryDenomOwnersResponse = unaryRpc(
      channel,
      QueryGrpc.getDenomOwnersMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.bank.v1beta1.Query service based on Kotlin coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.bank.v1beta1.Query.Balance.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun balance(request: QueryOuterClass.QueryBalanceRequest):
        QueryOuterClass.QueryBalanceResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.bank.v1beta1.Query.Balance is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.bank.v1beta1.Query.AllBalances.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun allBalances(request: QueryOuterClass.QueryAllBalancesRequest):
        QueryOuterClass.QueryAllBalancesResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.bank.v1beta1.Query.AllBalances is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.bank.v1beta1.Query.TotalSupply.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun totalSupply(request: QueryOuterClass.QueryTotalSupplyRequest):
        QueryOuterClass.QueryTotalSupplyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.bank.v1beta1.Query.TotalSupply is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.bank.v1beta1.Query.SupplyOf.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun supplyOf(request: QueryOuterClass.QuerySupplyOfRequest):
        QueryOuterClass.QuerySupplyOfResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.bank.v1beta1.Query.SupplyOf is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.bank.v1beta1.Query.Params.
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
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.bank.v1beta1.Query.Params is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.bank.v1beta1.Query.DenomMetadata.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun denomMetadata(request: QueryOuterClass.QueryDenomMetadataRequest):
        QueryOuterClass.QueryDenomMetadataResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.bank.v1beta1.Query.DenomMetadata is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.bank.v1beta1.Query.DenomsMetadata.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun denomsMetadata(request: QueryOuterClass.QueryDenomsMetadataRequest):
        QueryOuterClass.QueryDenomsMetadataResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.bank.v1beta1.Query.DenomsMetadata is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.bank.v1beta1.Query.DenomOwners.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun denomOwners(request: QueryOuterClass.QueryDenomOwnersRequest):
        QueryOuterClass.QueryDenomOwnersResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.bank.v1beta1.Query.DenomOwners is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getBalanceMethod(),
      implementation = ::balance
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getAllBalancesMethod(),
      implementation = ::allBalances
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getTotalSupplyMethod(),
      implementation = ::totalSupply
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getSupplyOfMethod(),
      implementation = ::supplyOf
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getParamsMethod(),
      implementation = ::params
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDenomMetadataMethod(),
      implementation = ::denomMetadata
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDenomsMetadataMethod(),
      implementation = ::denomsMetadata
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDenomOwnersMethod(),
      implementation = ::denomOwners
    )).build()
  }
}
