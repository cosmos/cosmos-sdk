package cosmos.base.reflection.v1beta1

import cosmos.base.reflection.v1beta1.ReflectionServiceGrpc.getServiceDescriptor
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
 * cosmos.base.reflection.v1beta1.ReflectionService.
 */
object ReflectionServiceGrpcKt {
  const val SERVICE_NAME: String = ReflectionServiceGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = ReflectionServiceGrpc.getServiceDescriptor()

  val listAllInterfacesMethod: MethodDescriptor<Reflection.ListAllInterfacesRequest,
      Reflection.ListAllInterfacesResponse>
    @JvmStatic
    get() = ReflectionServiceGrpc.getListAllInterfacesMethod()

  val listImplementationsMethod: MethodDescriptor<Reflection.ListImplementationsRequest,
      Reflection.ListImplementationsResponse>
    @JvmStatic
    get() = ReflectionServiceGrpc.getListImplementationsMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.base.reflection.v1beta1.ReflectionService service as
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
    suspend fun listAllInterfaces(request: Reflection.ListAllInterfacesRequest, headers: Metadata =
        Metadata()): Reflection.ListAllInterfacesResponse = unaryRpc(
      channel,
      ReflectionServiceGrpc.getListAllInterfacesMethod(),
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
    suspend fun listImplementations(request: Reflection.ListImplementationsRequest,
        headers: Metadata = Metadata()): Reflection.ListImplementationsResponse = unaryRpc(
      channel,
      ReflectionServiceGrpc.getListImplementationsMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.base.reflection.v1beta1.ReflectionService service based
   * on Kotlin coroutines.
   */
  abstract class ReflectionServiceCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for
     * cosmos.base.reflection.v1beta1.ReflectionService.ListAllInterfaces.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun listAllInterfaces(request: Reflection.ListAllInterfacesRequest):
        Reflection.ListAllInterfacesResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.reflection.v1beta1.ReflectionService.ListAllInterfaces is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.base.reflection.v1beta1.ReflectionService.ListImplementations.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun listImplementations(request: Reflection.ListImplementationsRequest):
        Reflection.ListImplementationsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.base.reflection.v1beta1.ReflectionService.ListImplementations is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ReflectionServiceGrpc.getListAllInterfacesMethod(),
      implementation = ::listAllInterfaces
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ReflectionServiceGrpc.getListImplementationsMethod(),
      implementation = ::listImplementations
    )).build()
  }
}
