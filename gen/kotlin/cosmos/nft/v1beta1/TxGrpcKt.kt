package cosmos.nft.v1beta1

import cosmos.nft.v1beta1.MsgGrpc.getServiceDescriptor
import io.grpc.CallOptions
import io.grpc.CallOptions.DEFAULT
import io.grpc.Channel
import io.grpc.Metadata
import io.grpc.MethodDescriptor
import io.grpc.ServerServiceDefinition
import io.grpc.ServerServiceDefinition.builder
import io.grpc.ServiceDescriptor
import io.grpc.Status.UNIMPLEMENTED
import io.grpc.StatusException
import io.grpc.kotlin.AbstractCoroutineServerImpl
import io.grpc.kotlin.AbstractCoroutineStub
import io.grpc.kotlin.ClientCalls.unaryRpc
import io.grpc.kotlin.ServerCalls.unaryServerMethodDefinition
import io.grpc.kotlin.StubFor
import kotlin.String
import kotlin.coroutines.CoroutineContext
import kotlin.coroutines.EmptyCoroutineContext
import kotlin.jvm.JvmOverloads
import kotlin.jvm.JvmStatic

/**
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.nft.v1beta1.Msg.
 */
object MsgGrpcKt {
  const val SERVICE_NAME: String = MsgGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = MsgGrpc.getServiceDescriptor()

  val sendMethod: MethodDescriptor<Tx.MsgSend, Tx.MsgSendResponse>
    @JvmStatic
    get() = MsgGrpc.getSendMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.nft.v1beta1.Msg service as suspending coroutines.
   */
  @StubFor(MsgGrpc::class)
  class MsgCoroutineStub @JvmOverloads constructor(
    channel: Channel,
    callOptions: CallOptions = DEFAULT
  ) : AbstractCoroutineStub<MsgCoroutineStub>(channel, callOptions) {
    override fun build(channel: Channel, callOptions: CallOptions): MsgCoroutineStub =
        MsgCoroutineStub(channel, callOptions)

    /**
     * Executes this RPC and returns the response message, suspending until the RPC completes
     * with [`Status.OK`][io.grpc.Status].  If the RPC completes with another status, a
     * corresponding
     * [StatusException] is thrown.  If this coroutine is cancelled, the RPC is also cancelled
     * with the corresponding exception as a cause.
     *
     * @param request The request message to send to the server.
     *
     * @param headers Metadata to attach to the request.  Most users will not need this.
     *
     * @return The single response from the server.
     */
    suspend fun send(request: Tx.MsgSend, headers: Metadata = Metadata()): Tx.MsgSendResponse =
        unaryRpc(
      channel,
      MsgGrpc.getSendMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.nft.v1beta1.Msg service based on Kotlin coroutines.
   */
  abstract class MsgCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.nft.v1beta1.Msg.Send.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [io.grpc.Status].  If this method fails with a [java.util.concurrent.CancellationException],
     * the RPC will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun send(request: Tx.MsgSend): Tx.MsgSendResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.nft.v1beta1.Msg.Send is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getSendMethod(),
      implementation = ::send
    )).build()
  }
}
