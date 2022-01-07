package cosmos.base.reflection.v2alpha1

import cosmos.base.reflection.v2alpha1.ReflectionServiceGrpc.getServiceDescriptor
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
 * cosmos.base.reflection.v2alpha1.ReflectionService.
 */
object ReflectionServiceGrpcKt {
  const val SERVICE_NAME: String = ReflectionServiceGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = ReflectionServiceGrpc.getServiceDescriptor()

  val getAuthnDescriptorMethod: MethodDescriptor<Reflection.GetAuthnDescriptorRequest,
      Reflection.GetAuthnDescriptorResponse>
    @JvmStatic
    get() = ReflectionServiceGrpc.getGetAuthnDescriptorMethod()

  val getChainDescriptorMethod: MethodDescriptor<Reflection.GetChainDescriptorRequest,
      Reflection.GetChainDescriptorResponse>
    @JvmStatic
    get() = ReflectionServiceGrpc.getGetChainDescriptorMethod()

  val getCodecDescriptorMethod: MethodDescriptor<Reflection.GetCodecDescriptorRequest,
      Reflection.GetCodecDescriptorResponse>
    @JvmStatic
    get() = ReflectionServiceGrpc.getGetCodecDescriptorMethod()

  val getConfigurationDescriptorMethod:
      MethodDescriptor<Reflection.GetConfigurationDescriptorRequest,
      Reflection.GetConfigurationDescriptorResponse>
    @JvmStatic
    get() = ReflectionServiceGrpc.getGetConfigurationDescriptorMethod()

  val getQueryServicesDescriptorMethod:
      MethodDescriptor<Reflection.GetQueryServicesDescriptorRequest,
      Reflection.GetQueryServicesDescriptorResponse>
    @JvmStatic
    get() = ReflectionServiceGrpc.getGetQueryServicesDescriptorMethod()

  val getTxDescriptorMethod: MethodDescriptor<Reflection.GetTxDescriptorRequest,
      Reflection.GetTxDescriptorResponse>
    @JvmStatic
    get() = ReflectionServiceGrpc.getGetTxDescriptorMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.base.reflection.v2alpha1.ReflectionService service as
   * suspending coroutines.
   */
  @StubFor(ReflectionServiceGrpc::class)
  class ReflectionServiceCoroutineStub @JvmOverloads constructor(
    channel: Channel,
    callOptions: CallOptions = DEFAULT
  ) : AbstractCoroutineStub<ReflectionServiceCoroutineStub>(channel, callOptions) {
    override fun build(channel: Channel, callOptions: CallOptions): ReflectionServiceCoroutineStub =
        ReflectionServiceCoroutineStub(channel, callOptions)

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
    suspend fun getAuthnDescriptor(request: Reflection.GetAuthnDescriptorRequest, headers: Metadata
        = Metadata()): Reflection.GetAuthnDescriptorResponse = unaryRpc(
      channel,
      ReflectionServiceGrpc.getGetAuthnDescriptorMethod(),
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
    suspend fun getChainDescriptor(request: Reflection.GetChainDescriptorRequest, headers: Metadata
        = Metadata()): Reflection.GetChainDescriptorResponse = unaryRpc(
      channel,
      ReflectionServiceGrpc.getGetChainDescriptorMethod(),
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
    suspend fun getCodecDescriptor(request: Reflection.GetCodecDescriptorRequest, headers: Metadata
        = Metadata()): Reflection.GetCodecDescriptorResponse = unaryRpc(
      channel,
      ReflectionServiceGrpc.getGetCodecDescriptorMethod(),
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
    suspend fun getConfigurationDescriptor(request: Reflection.GetConfigurationDescriptorRequest,
        headers: Metadata = Metadata()): Reflection.GetConfigurationDescriptorResponse = unaryRpc(
      channel,
      ReflectionServiceGrpc.getGetConfigurationDescriptorMethod(),
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
    suspend fun getQueryServicesDescriptor(request: Reflection.GetQueryServicesDescriptorRequest,
        headers: Metadata = Metadata()): Reflection.GetQueryServicesDescriptorResponse = unaryRpc(
      channel,
      ReflectionServiceGrpc.getGetQueryServicesDescriptorMethod(),
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
    suspend fun getTxDescriptor(request: Reflection.GetTxDescriptorRequest, headers: Metadata =
        Metadata()): Reflection.GetTxDescriptorResponse = unaryRpc(
      channel,
      ReflectionServiceGrpc.getGetTxDescriptorMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.base.reflection.v2alpha1.ReflectionService service based
   * on Kotlin coroutines.
   */
  abstract class ReflectionServiceCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for
     * cosmos.base.reflection.v2alpha1.ReflectionService.GetAuthnDescriptor.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getAuthnDescriptor(request: Reflection.GetAuthnDescriptorRequest):
        Reflection.GetAuthnDescriptorResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.reflection.v2alpha1.ReflectionService.GetAuthnDescriptor is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.base.reflection.v2alpha1.ReflectionService.GetChainDescriptor.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getChainDescriptor(request: Reflection.GetChainDescriptorRequest):
        Reflection.GetChainDescriptorResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.reflection.v2alpha1.ReflectionService.GetChainDescriptor is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.base.reflection.v2alpha1.ReflectionService.GetCodecDescriptor.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getCodecDescriptor(request: Reflection.GetCodecDescriptorRequest):
        Reflection.GetCodecDescriptorResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.reflection.v2alpha1.ReflectionService.GetCodecDescriptor is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.base.reflection.v2alpha1.ReflectionService.GetConfigurationDescriptor.
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
        fun getConfigurationDescriptor(request: Reflection.GetConfigurationDescriptorRequest):
        Reflection.GetConfigurationDescriptorResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.reflection.v2alpha1.ReflectionService.GetConfigurationDescriptor is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.base.reflection.v2alpha1.ReflectionService.GetQueryServicesDescriptor.
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
        fun getQueryServicesDescriptor(request: Reflection.GetQueryServicesDescriptorRequest):
        Reflection.GetQueryServicesDescriptorResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.reflection.v2alpha1.ReflectionService.GetQueryServicesDescriptor is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.base.reflection.v2alpha1.ReflectionService.GetTxDescriptor.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun getTxDescriptor(request: Reflection.GetTxDescriptorRequest):
        Reflection.GetTxDescriptorResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.reflection.v2alpha1.ReflectionService.GetTxDescriptor is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ReflectionServiceGrpc.getGetAuthnDescriptorMethod(),
      implementation = ::getAuthnDescriptor
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ReflectionServiceGrpc.getGetChainDescriptorMethod(),
      implementation = ::getChainDescriptor
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ReflectionServiceGrpc.getGetCodecDescriptorMethod(),
      implementation = ::getCodecDescriptor
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ReflectionServiceGrpc.getGetConfigurationDescriptorMethod(),
      implementation = ::getConfigurationDescriptor
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ReflectionServiceGrpc.getGetQueryServicesDescriptorMethod(),
      implementation = ::getQueryServicesDescriptor
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ReflectionServiceGrpc.getGetTxDescriptorMethod(),
      implementation = ::getTxDescriptor
    )).build()
  }
}
