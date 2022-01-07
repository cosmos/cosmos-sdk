package cosmos.auth.v1beta1

import cosmos.auth.v1beta1.QueryGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.auth.v1beta1.Query.
 */
object QueryGrpcKt {
  const val SERVICE_NAME: String = QueryGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = QueryGrpc.getServiceDescriptor()

  val accountsMethod: MethodDescriptor<QueryOuterClass.QueryAccountsRequest,
      QueryOuterClass.QueryAccountsResponse>
    @JvmStatic
    get() = QueryGrpc.getAccountsMethod()

  val accountMethod: MethodDescriptor<QueryOuterClass.QueryAccountRequest,
      QueryOuterClass.QueryAccountResponse>
    @JvmStatic
    get() = QueryGrpc.getAccountMethod()

  val paramsMethod: MethodDescriptor<QueryOuterClass.QueryParamsRequest,
      QueryOuterClass.QueryParamsResponse>
    @JvmStatic
    get() = QueryGrpc.getParamsMethod()

  val moduleAccountsMethod: MethodDescriptor<QueryOuterClass.QueryModuleAccountsRequest,
      QueryOuterClass.QueryModuleAccountsResponse>
    @JvmStatic
    get() = QueryGrpc.getModuleAccountsMethod()

  val bech32PrefixMethod: MethodDescriptor<QueryOuterClass.Bech32PrefixRequest,
      QueryOuterClass.Bech32PrefixResponse>
    @JvmStatic
    get() = QueryGrpc.getBech32PrefixMethod()

  val addressBytesToStringMethod: MethodDescriptor<QueryOuterClass.AddressBytesToStringRequest,
      QueryOuterClass.AddressBytesToStringResponse>
    @JvmStatic
    get() = QueryGrpc.getAddressBytesToStringMethod()

  val addressStringToBytesMethod: MethodDescriptor<QueryOuterClass.AddressStringToBytesRequest,
      QueryOuterClass.AddressStringToBytesResponse>
    @JvmStatic
    get() = QueryGrpc.getAddressStringToBytesMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.auth.v1beta1.Query service as suspending coroutines.
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
    suspend fun accounts(request: QueryOuterClass.QueryAccountsRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryAccountsResponse = unaryRpc(
      channel,
      QueryGrpc.getAccountsMethod(),
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
    suspend fun account(request: QueryOuterClass.QueryAccountRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryAccountResponse = unaryRpc(
      channel,
      QueryGrpc.getAccountMethod(),
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
    suspend fun moduleAccounts(request: QueryOuterClass.QueryModuleAccountsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryModuleAccountsResponse = unaryRpc(
      channel,
      QueryGrpc.getModuleAccountsMethod(),
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
    suspend fun bech32Prefix(request: QueryOuterClass.Bech32PrefixRequest, headers: Metadata =
        Metadata()): QueryOuterClass.Bech32PrefixResponse = unaryRpc(
      channel,
      QueryGrpc.getBech32PrefixMethod(),
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
    suspend fun addressBytesToString(request: QueryOuterClass.AddressBytesToStringRequest,
        headers: Metadata = Metadata()): QueryOuterClass.AddressBytesToStringResponse = unaryRpc(
      channel,
      QueryGrpc.getAddressBytesToStringMethod(),
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
    suspend fun addressStringToBytes(request: QueryOuterClass.AddressStringToBytesRequest,
        headers: Metadata = Metadata()): QueryOuterClass.AddressStringToBytesResponse = unaryRpc(
      channel,
      QueryGrpc.getAddressStringToBytesMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.auth.v1beta1.Query service based on Kotlin coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.auth.v1beta1.Query.Accounts.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun accounts(request: QueryOuterClass.QueryAccountsRequest):
        QueryOuterClass.QueryAccountsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.auth.v1beta1.Query.Accounts is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.auth.v1beta1.Query.Account.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun account(request: QueryOuterClass.QueryAccountRequest):
        QueryOuterClass.QueryAccountResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.auth.v1beta1.Query.Account is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.auth.v1beta1.Query.Params.
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
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.auth.v1beta1.Query.Params is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.auth.v1beta1.Query.ModuleAccounts.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun moduleAccounts(request: QueryOuterClass.QueryModuleAccountsRequest):
        QueryOuterClass.QueryModuleAccountsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.auth.v1beta1.Query.ModuleAccounts is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.auth.v1beta1.Query.Bech32Prefix.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun bech32Prefix(request: QueryOuterClass.Bech32PrefixRequest):
        QueryOuterClass.Bech32PrefixResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.auth.v1beta1.Query.Bech32Prefix is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.auth.v1beta1.Query.AddressBytesToString.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun addressBytesToString(request: QueryOuterClass.AddressBytesToStringRequest):
        QueryOuterClass.AddressBytesToStringResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.auth.v1beta1.Query.AddressBytesToString is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.auth.v1beta1.Query.AddressStringToBytes.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun addressStringToBytes(request: QueryOuterClass.AddressStringToBytesRequest):
        QueryOuterClass.AddressStringToBytesResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.auth.v1beta1.Query.AddressStringToBytes is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getAccountsMethod(),
      implementation = ::accounts
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getAccountMethod(),
      implementation = ::account
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getParamsMethod(),
      implementation = ::params
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getModuleAccountsMethod(),
      implementation = ::moduleAccounts
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getBech32PrefixMethod(),
      implementation = ::bech32Prefix
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getAddressBytesToStringMethod(),
      implementation = ::addressBytesToString
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getAddressStringToBytesMethod(),
      implementation = ::addressStringToBytes
    )).build()
  }
}
