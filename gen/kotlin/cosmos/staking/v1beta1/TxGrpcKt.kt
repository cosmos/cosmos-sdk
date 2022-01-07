package cosmos.staking.v1beta1

import cosmos.staking.v1beta1.MsgGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.staking.v1beta1.Msg.
 */
object MsgGrpcKt {
  const val SERVICE_NAME: String = MsgGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = MsgGrpc.getServiceDescriptor()

  val createValidatorMethod: MethodDescriptor<Tx.MsgCreateValidator, Tx.MsgCreateValidatorResponse>
    @JvmStatic
    get() = MsgGrpc.getCreateValidatorMethod()

  val editValidatorMethod: MethodDescriptor<Tx.MsgEditValidator, Tx.MsgEditValidatorResponse>
    @JvmStatic
    get() = MsgGrpc.getEditValidatorMethod()

  val delegateMethod: MethodDescriptor<Tx.MsgDelegate, Tx.MsgDelegateResponse>
    @JvmStatic
    get() = MsgGrpc.getDelegateMethod()

  val beginRedelegateMethod: MethodDescriptor<Tx.MsgBeginRedelegate, Tx.MsgBeginRedelegateResponse>
    @JvmStatic
    get() = MsgGrpc.getBeginRedelegateMethod()

  val undelegateMethod: MethodDescriptor<Tx.MsgUndelegate, Tx.MsgUndelegateResponse>
    @JvmStatic
    get() = MsgGrpc.getUndelegateMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.staking.v1beta1.Msg service as suspending coroutines.
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
    suspend fun createValidator(request: Tx.MsgCreateValidator, headers: Metadata = Metadata()):
        Tx.MsgCreateValidatorResponse = unaryRpc(
      channel,
      MsgGrpc.getCreateValidatorMethod(),
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
    suspend fun editValidator(request: Tx.MsgEditValidator, headers: Metadata = Metadata()):
        Tx.MsgEditValidatorResponse = unaryRpc(
      channel,
      MsgGrpc.getEditValidatorMethod(),
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
    suspend fun delegate(request: Tx.MsgDelegate, headers: Metadata = Metadata()):
        Tx.MsgDelegateResponse = unaryRpc(
      channel,
      MsgGrpc.getDelegateMethod(),
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
    suspend fun beginRedelegate(request: Tx.MsgBeginRedelegate, headers: Metadata = Metadata()):
        Tx.MsgBeginRedelegateResponse = unaryRpc(
      channel,
      MsgGrpc.getBeginRedelegateMethod(),
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
    suspend fun undelegate(request: Tx.MsgUndelegate, headers: Metadata = Metadata()):
        Tx.MsgUndelegateResponse = unaryRpc(
      channel,
      MsgGrpc.getUndelegateMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.staking.v1beta1.Msg service based on Kotlin coroutines.
   */
  abstract class MsgCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Msg.CreateValidator.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun createValidator(request: Tx.MsgCreateValidator): Tx.MsgCreateValidatorResponse
        = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Msg.CreateValidator is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Msg.EditValidator.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun editValidator(request: Tx.MsgEditValidator): Tx.MsgEditValidatorResponse =
        throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Msg.EditValidator is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Msg.Delegate.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun delegate(request: Tx.MsgDelegate): Tx.MsgDelegateResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Msg.Delegate is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Msg.BeginRedelegate.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun beginRedelegate(request: Tx.MsgBeginRedelegate): Tx.MsgBeginRedelegateResponse
        = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Msg.BeginRedelegate is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.staking.v1beta1.Msg.Undelegate.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun undelegate(request: Tx.MsgUndelegate): Tx.MsgUndelegateResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.staking.v1beta1.Msg.Undelegate is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getCreateValidatorMethod(),
      implementation = ::createValidator
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getEditValidatorMethod(),
      implementation = ::editValidator
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getDelegateMethod(),
      implementation = ::delegate
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getBeginRedelegateMethod(),
      implementation = ::beginRedelegate
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getUndelegateMethod(),
      implementation = ::undelegate
    )).build()
  }
}
