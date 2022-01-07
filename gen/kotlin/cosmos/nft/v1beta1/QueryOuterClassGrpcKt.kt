package cosmos.nft.v1beta1

import cosmos.nft.v1beta1.QueryGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.nft.v1beta1.Query.
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

  val ownerMethod: MethodDescriptor<QueryOuterClass.QueryOwnerRequest,
      QueryOuterClass.QueryOwnerResponse>
    @JvmStatic
    get() = QueryGrpc.getOwnerMethod()

  val supplyMethod: MethodDescriptor<QueryOuterClass.QuerySupplyRequest,
      QueryOuterClass.QuerySupplyResponse>
    @JvmStatic
    get() = QueryGrpc.getSupplyMethod()

  val nFTsOfClassMethod: MethodDescriptor<QueryOuterClass.QueryNFTsOfClassRequest,
      QueryOuterClass.QueryNFTsOfClassResponse>
    @JvmStatic
    get() = QueryGrpc.getNFTsOfClassMethod()

  val nFTMethod: MethodDescriptor<QueryOuterClass.QueryNFTRequest, QueryOuterClass.QueryNFTResponse>
    @JvmStatic
    get() = QueryGrpc.getNFTMethod()

  val classMethod: MethodDescriptor<QueryOuterClass.QueryClassRequest,
      QueryOuterClass.QueryClassResponse>
    @JvmStatic
    get() = QueryGrpc.getClassMethod()

  val classesMethod: MethodDescriptor<QueryOuterClass.QueryClassesRequest,
      QueryOuterClass.QueryClassesResponse>
    @JvmStatic
    get() = QueryGrpc.getClassesMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.nft.v1beta1.Query service as suspending coroutines.
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
    suspend fun owner(request: QueryOuterClass.QueryOwnerRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QueryOwnerResponse = unaryRpc(
      channel,
      QueryGrpc.getOwnerMethod(),
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
    suspend fun supply(request: QueryOuterClass.QuerySupplyRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QuerySupplyResponse = unaryRpc(
      channel,
      QueryGrpc.getSupplyMethod(),
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
    suspend fun nFTsOfClass(request: QueryOuterClass.QueryNFTsOfClassRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryNFTsOfClassResponse = unaryRpc(
      channel,
      QueryGrpc.getNFTsOfClassMethod(),
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
    suspend fun nFT(request: QueryOuterClass.QueryNFTRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QueryNFTResponse = unaryRpc(
      channel,
      QueryGrpc.getNFTMethod(),
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
    suspend fun `class`(request: QueryOuterClass.QueryClassRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QueryClassResponse = unaryRpc(
      channel,
      QueryGrpc.getClassMethod(),
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
    suspend fun classes(request: QueryOuterClass.QueryClassesRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryClassesResponse = unaryRpc(
      channel,
      QueryGrpc.getClassesMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.nft.v1beta1.Query service based on Kotlin coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.nft.v1beta1.Query.Balance.
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
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.nft.v1beta1.Query.Balance is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.nft.v1beta1.Query.Owner.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun owner(request: QueryOuterClass.QueryOwnerRequest):
        QueryOuterClass.QueryOwnerResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.nft.v1beta1.Query.Owner is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.nft.v1beta1.Query.Supply.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun supply(request: QueryOuterClass.QuerySupplyRequest):
        QueryOuterClass.QuerySupplyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.nft.v1beta1.Query.Supply is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.nft.v1beta1.Query.NFTsOfClass.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun nFTsOfClass(request: QueryOuterClass.QueryNFTsOfClassRequest):
        QueryOuterClass.QueryNFTsOfClassResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.nft.v1beta1.Query.NFTsOfClass is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.nft.v1beta1.Query.NFT.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun nFT(request: QueryOuterClass.QueryNFTRequest): QueryOuterClass.QueryNFTResponse
        = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.nft.v1beta1.Query.NFT is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.nft.v1beta1.Query.Class.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun `class`(request: QueryOuterClass.QueryClassRequest):
        QueryOuterClass.QueryClassResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.nft.v1beta1.Query.Class is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.nft.v1beta1.Query.Classes.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun classes(request: QueryOuterClass.QueryClassesRequest):
        QueryOuterClass.QueryClassesResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.nft.v1beta1.Query.Classes is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getBalanceMethod(),
      implementation = ::balance
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getOwnerMethod(),
      implementation = ::owner
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getSupplyMethod(),
      implementation = ::supply
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getNFTsOfClassMethod(),
      implementation = ::nFTsOfClass
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getNFTMethod(),
      implementation = ::nFT
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getClassMethod(),
      implementation = ::`class`
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getClassesMethod(),
      implementation = ::classes
    )).build()
  }
}
