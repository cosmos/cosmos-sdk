package cosmos.base.tendermint.v1beta1

import cosmos.base.tendermint.v1beta1.ServiceGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for
 * cosmos.base.tendermint.v1beta1.Service.
 */
object ServiceGrpcKt {
  const val SERVICE_NAME: String = ServiceGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = ServiceGrpc.getServiceDescriptor()

  val getNodeInfoMethod: MethodDescriptor<Query.GetNodeInfoRequest, Query.GetNodeInfoResponse>
    @JvmStatic
    get() = ServiceGrpc.getGetNodeInfoMethod()

  val getSyncingMethod: MethodDescriptor<Query.GetSyncingRequest, Query.GetSyncingResponse>
    @JvmStatic
    get() = ServiceGrpc.getGetSyncingMethod()

  val getLatestBlockMethod: MethodDescriptor<Query.GetLatestBlockRequest,
      Query.GetLatestBlockResponse>
    @JvmStatic
    get() = ServiceGrpc.getGetLatestBlockMethod()

  val getBlockByHeightMethod: MethodDescriptor<Query.GetBlockByHeightRequest,
      Query.GetBlockByHeightResponse>
    @JvmStatic
    get() = ServiceGrpc.getGetBlockByHeightMethod()

  val getLatestValidatorSetMethod: MethodDescriptor<Query.GetLatestValidatorSetRequest,
      Query.GetLatestValidatorSetResponse>
    @JvmStatic
    get() = ServiceGrpc.getGetLatestValidatorSetMethod()

  val getValidatorSetByHeightMethod: MethodDescriptor<Query.GetValidatorSetByHeightRequest,
      Query.GetValidatorSetByHeightResponse>
    @JvmStatic
    get() = ServiceGrpc.getGetValidatorSetByHeightMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.base.tendermint.v1beta1.Service service as suspending
   * coroutines.
   */
  @StubFor(ServiceGrpc::class)
  class ServiceCoroutineStub @JvmOverloads constructor(
    channel: Channel,
    callOptions: CallOptions = DEFAULT
  ) : AbstractCoroutineStub<ServiceCoroutineStub>(channel, callOptions) {
    override fun build(channel: Channel, callOptions: CallOptions): ServiceCoroutineStub =
        ServiceCoroutineStub(channel, callOptions)

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
    suspend fun getNodeInfo(request: Query.GetNodeInfoRequest, headers: Metadata = Metadata()):
        Query.GetNodeInfoResponse = unaryRpc(
      channel,
      ServiceGrpc.getGetNodeInfoMethod(),
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
    suspend fun getSyncing(request: Query.GetSyncingRequest, headers: Metadata = Metadata()):
        Query.GetSyncingResponse = unaryRpc(
      channel,
      ServiceGrpc.getGetSyncingMethod(),
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
    suspend fun getLatestBlock(request: Query.GetLatestBlockRequest, headers: Metadata =
        Metadata()): Query.GetLatestBlockResponse = unaryRpc(
      channel,
      ServiceGrpc.getGetLatestBlockMethod(),
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
    suspend fun getBlockByHeight(request: Query.GetBlockByHeightRequest, headers: Metadata =
        Metadata()): Query.GetBlockByHeightResponse = unaryRpc(
      channel,
      ServiceGrpc.getGetBlockByHeightMethod(),
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
    suspend fun getLatestValidatorSet(request: Query.GetLatestValidatorSetRequest, headers: Metadata
        = Metadata()): Query.GetLatestValidatorSetResponse = unaryRpc(
      channel,
      ServiceGrpc.getGetLatestValidatorSetMethod(),
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
    suspend fun getValidatorSetByHeight(request: Query.GetValidatorSetByHeightRequest,
        headers: Metadata = Metadata()): Query.GetValidatorSetByHeightResponse = unaryRpc(
      channel,
      ServiceGrpc.getGetValidatorSetByHeightMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.base.tendermint.v1beta1.Service service based on Kotlin
   * coroutines.
   */
  abstract class ServiceCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.base.tendermint.v1beta1.Service.GetNodeInfo.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getNodeInfo(request: Query.GetNodeInfoRequest): Query.GetNodeInfoResponse =
        throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.tendermint.v1beta1.Service.GetNodeInfo is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.base.tendermint.v1beta1.Service.GetSyncing.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getSyncing(request: Query.GetSyncingRequest): Query.GetSyncingResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.tendermint.v1beta1.Service.GetSyncing is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.base.tendermint.v1beta1.Service.GetLatestBlock.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getLatestBlock(request: Query.GetLatestBlockRequest):
        Query.GetLatestBlockResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.tendermint.v1beta1.Service.GetLatestBlock is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.base.tendermint.v1beta1.Service.GetBlockByHeight.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getBlockByHeight(request: Query.GetBlockByHeightRequest):
        Query.GetBlockByHeightResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.tendermint.v1beta1.Service.GetBlockByHeight is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.base.tendermint.v1beta1.Service.GetLatestValidatorSet.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getLatestValidatorSet(request: Query.GetLatestValidatorSetRequest):
        Query.GetLatestValidatorSetResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.tendermint.v1beta1.Service.GetLatestValidatorSet is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.base.tendermint.v1beta1.Service.GetValidatorSetByHeight.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getValidatorSetByHeight(request: Query.GetValidatorSetByHeightRequest):
        Query.GetValidatorSetByHeightResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.tendermint.v1beta1.Service.GetValidatorSetByHeight is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getGetNodeInfoMethod(),
      implementation = ::getNodeInfo
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getGetSyncingMethod(),
      implementation = ::getSyncing
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getGetLatestBlockMethod(),
      implementation = ::getLatestBlock
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getGetBlockByHeightMethod(),
      implementation = ::getBlockByHeight
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getGetLatestValidatorSetMethod(),
      implementation = ::getLatestValidatorSet
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getGetValidatorSetByHeightMethod(),
      implementation = ::getValidatorSetByHeight
    )).build()
  }
}
