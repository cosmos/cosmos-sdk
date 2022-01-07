package cosmos.distribution.v1beta1

import cosmos.distribution.v1beta1.MsgGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.distribution.v1beta1.Msg.
 */
object MsgGrpcKt {
  const val SERVICE_NAME: String = MsgGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = MsgGrpc.getServiceDescriptor()

  val setWithdrawAddressMethod: MethodDescriptor<Tx.MsgSetWithdrawAddress,
      Tx.MsgSetWithdrawAddressResponse>
    @JvmStatic
    get() = MsgGrpc.getSetWithdrawAddressMethod()

  val withdrawDelegatorRewardMethod: MethodDescriptor<Tx.MsgWithdrawDelegatorReward,
      Tx.MsgWithdrawDelegatorRewardResponse>
    @JvmStatic
    get() = MsgGrpc.getWithdrawDelegatorRewardMethod()

  val withdrawValidatorCommissionMethod: MethodDescriptor<Tx.MsgWithdrawValidatorCommission,
      Tx.MsgWithdrawValidatorCommissionResponse>
    @JvmStatic
    get() = MsgGrpc.getWithdrawValidatorCommissionMethod()

  val fundCommunityPoolMethod: MethodDescriptor<Tx.MsgFundCommunityPool,
      Tx.MsgFundCommunityPoolResponse>
    @JvmStatic
    get() = MsgGrpc.getFundCommunityPoolMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.distribution.v1beta1.Msg service as suspending
   * coroutines.
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
    suspend fun setWithdrawAddress(request: Tx.MsgSetWithdrawAddress, headers: Metadata =
        Metadata()): Tx.MsgSetWithdrawAddressResponse = unaryRpc(
      channel,
      MsgGrpc.getSetWithdrawAddressMethod(),
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
    suspend fun withdrawDelegatorReward(request: Tx.MsgWithdrawDelegatorReward, headers: Metadata =
        Metadata()): Tx.MsgWithdrawDelegatorRewardResponse = unaryRpc(
      channel,
      MsgGrpc.getWithdrawDelegatorRewardMethod(),
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
    suspend fun withdrawValidatorCommission(request: Tx.MsgWithdrawValidatorCommission,
        headers: Metadata = Metadata()): Tx.MsgWithdrawValidatorCommissionResponse = unaryRpc(
      channel,
      MsgGrpc.getWithdrawValidatorCommissionMethod(),
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
    suspend fun fundCommunityPool(request: Tx.MsgFundCommunityPool, headers: Metadata = Metadata()):
        Tx.MsgFundCommunityPoolResponse = unaryRpc(
      channel,
      MsgGrpc.getFundCommunityPoolMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.distribution.v1beta1.Msg service based on Kotlin
   * coroutines.
   */
  abstract class MsgCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Msg.SetWithdrawAddress.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun setWithdrawAddress(request: Tx.MsgSetWithdrawAddress):
        Tx.MsgSetWithdrawAddressResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Msg.SetWithdrawAddress is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Msg.WithdrawDelegatorReward.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun withdrawDelegatorReward(request: Tx.MsgWithdrawDelegatorReward):
        Tx.MsgWithdrawDelegatorRewardResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Msg.WithdrawDelegatorReward is unimplemented"))

    /**
     * Returns the response to an RPC for
     * cosmos.distribution.v1beta1.Msg.WithdrawValidatorCommission.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun withdrawValidatorCommission(request: Tx.MsgWithdrawValidatorCommission):
        Tx.MsgWithdrawValidatorCommissionResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Msg.WithdrawValidatorCommission is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.distribution.v1beta1.Msg.FundCommunityPool.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun fundCommunityPool(request: Tx.MsgFundCommunityPool):
        Tx.MsgFundCommunityPoolResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.distribution.v1beta1.Msg.FundCommunityPool is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getSetWithdrawAddressMethod(),
      implementation = ::setWithdrawAddress
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getWithdrawDelegatorRewardMethod(),
      implementation = ::withdrawDelegatorReward
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getWithdrawValidatorCommissionMethod(),
      implementation = ::withdrawValidatorCommission
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getFundCommunityPoolMethod(),
      implementation = ::fundCommunityPool
    )).build()
  }
}
