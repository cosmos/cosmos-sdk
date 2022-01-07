package cosmos.tx.v1beta1

import cosmos.tx.v1beta1.ServiceGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.tx.v1beta1.Service.
 */
object ServiceGrpcKt {
  const val SERVICE_NAME: String = ServiceGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = ServiceGrpc.getServiceDescriptor()

  val simulateMethod: MethodDescriptor<ServiceOuterClass.SimulateRequest,
      ServiceOuterClass.SimulateResponse>
    @JvmStatic
    get() = ServiceGrpc.getSimulateMethod()

  val getTxMethod: MethodDescriptor<ServiceOuterClass.GetTxRequest, ServiceOuterClass.GetTxResponse>
    @JvmStatic
    get() = ServiceGrpc.getGetTxMethod()

  val broadcastTxMethod: MethodDescriptor<ServiceOuterClass.BroadcastTxRequest,
      ServiceOuterClass.BroadcastTxResponse>
    @JvmStatic
    get() = ServiceGrpc.getBroadcastTxMethod()

  val getTxsEventMethod: MethodDescriptor<ServiceOuterClass.GetTxsEventRequest,
      ServiceOuterClass.GetTxsEventResponse>
    @JvmStatic
    get() = ServiceGrpc.getGetTxsEventMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.tx.v1beta1.Service service as suspending coroutines.
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
    suspend fun simulate(request: ServiceOuterClass.SimulateRequest, headers: Metadata =
        Metadata()): ServiceOuterClass.SimulateResponse = unaryRpc(
      channel,
      ServiceGrpc.getSimulateMethod(),
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
    suspend fun getTx(request: ServiceOuterClass.GetTxRequest, headers: Metadata = Metadata()):
        ServiceOuterClass.GetTxResponse = unaryRpc(
      channel,
      ServiceGrpc.getGetTxMethod(),
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
    suspend fun broadcastTx(request: ServiceOuterClass.BroadcastTxRequest, headers: Metadata =
        Metadata()): ServiceOuterClass.BroadcastTxResponse = unaryRpc(
      channel,
      ServiceGrpc.getBroadcastTxMethod(),
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
    suspend fun getTxsEvent(request: ServiceOuterClass.GetTxsEventRequest, headers: Metadata =
        Metadata()): ServiceOuterClass.GetTxsEventResponse = unaryRpc(
      channel,
      ServiceGrpc.getGetTxsEventMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.tx.v1beta1.Service service based on Kotlin coroutines.
   */
  abstract class ServiceCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.tx.v1beta1.Service.Simulate.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun simulate(request: ServiceOuterClass.SimulateRequest):
        ServiceOuterClass.SimulateResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.tx.v1beta1.Service.Simulate is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.tx.v1beta1.Service.GetTx.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getTx(request: ServiceOuterClass.GetTxRequest): ServiceOuterClass.GetTxResponse
        = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.tx.v1beta1.Service.GetTx is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.tx.v1beta1.Service.BroadcastTx.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun broadcastTx(request: ServiceOuterClass.BroadcastTxRequest):
        ServiceOuterClass.BroadcastTxResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.tx.v1beta1.Service.BroadcastTx is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.tx.v1beta1.Service.GetTxsEvent.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getTxsEvent(request: ServiceOuterClass.GetTxsEventRequest):
        ServiceOuterClass.GetTxsEventResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.tx.v1beta1.Service.GetTxsEvent is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getSimulateMethod(),
      implementation = ::simulate
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getGetTxMethod(),
      implementation = ::getTx
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getBroadcastTxMethod(),
      implementation = ::broadcastTx
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ServiceGrpc.getGetTxsEventMethod(),
      implementation = ::getTxsEvent
    )).build()
  }
}
