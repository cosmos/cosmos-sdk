package cosmos.gov.v1beta2

import cosmos.gov.v1beta2.MsgGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.gov.v1beta2.Msg.
 */
object MsgGrpcKt {
  const val SERVICE_NAME: String = MsgGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = MsgGrpc.getServiceDescriptor()

  val submitProposalMethod: MethodDescriptor<Tx.MsgSubmitProposal, Tx.MsgSubmitProposalResponse>
    @JvmStatic
    get() = MsgGrpc.getSubmitProposalMethod()

  val voteMethod: MethodDescriptor<Tx.MsgVote, Tx.MsgVoteResponse>
    @JvmStatic
    get() = MsgGrpc.getVoteMethod()

  val voteWeightedMethod: MethodDescriptor<Tx.MsgVoteWeighted, Tx.MsgVoteWeightedResponse>
    @JvmStatic
    get() = MsgGrpc.getVoteWeightedMethod()

  val depositMethod: MethodDescriptor<Tx.MsgDeposit, Tx.MsgDepositResponse>
    @JvmStatic
    get() = MsgGrpc.getDepositMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.gov.v1beta2.Msg service as suspending coroutines.
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
    suspend fun submitProposal(request: Tx.MsgSubmitProposal, headers: Metadata = Metadata()):
        Tx.MsgSubmitProposalResponse = unaryRpc(
      channel,
      MsgGrpc.getSubmitProposalMethod(),
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
    suspend fun vote(request: Tx.MsgVote, headers: Metadata = Metadata()): Tx.MsgVoteResponse =
        unaryRpc(
      channel,
      MsgGrpc.getVoteMethod(),
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
    suspend fun voteWeighted(request: Tx.MsgVoteWeighted, headers: Metadata = Metadata()):
        Tx.MsgVoteWeightedResponse = unaryRpc(
      channel,
      MsgGrpc.getVoteWeightedMethod(),
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
    suspend fun deposit(request: Tx.MsgDeposit, headers: Metadata = Metadata()):
        Tx.MsgDepositResponse = unaryRpc(
      channel,
      MsgGrpc.getDepositMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.gov.v1beta2.Msg service based on Kotlin coroutines.
   */
  abstract class MsgCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.gov.v1beta2.Msg.SubmitProposal.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun submitProposal(request: Tx.MsgSubmitProposal): Tx.MsgSubmitProposalResponse =
        throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta2.Msg.SubmitProposal is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta2.Msg.Vote.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun vote(request: Tx.MsgVote): Tx.MsgVoteResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta2.Msg.Vote is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta2.Msg.VoteWeighted.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun voteWeighted(request: Tx.MsgVoteWeighted): Tx.MsgVoteWeightedResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta2.Msg.VoteWeighted is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta2.Msg.Deposit.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun deposit(request: Tx.MsgDeposit): Tx.MsgDepositResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta2.Msg.Deposit is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getSubmitProposalMethod(),
      implementation = ::submitProposal
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getVoteMethod(),
      implementation = ::vote
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getVoteWeightedMethod(),
      implementation = ::voteWeighted
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getDepositMethod(),
      implementation = ::deposit
    )).build()
  }
}
