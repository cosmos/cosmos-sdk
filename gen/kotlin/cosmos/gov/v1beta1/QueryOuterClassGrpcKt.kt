package cosmos.gov.v1beta1

import cosmos.gov.v1beta1.QueryGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.gov.v1beta1.Query.
 */
object QueryGrpcKt {
  const val SERVICE_NAME: String = QueryGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = QueryGrpc.getServiceDescriptor()

  val proposalMethod: MethodDescriptor<QueryOuterClass.QueryProposalRequest,
      QueryOuterClass.QueryProposalResponse>
    @JvmStatic
    get() = QueryGrpc.getProposalMethod()

  val proposalsMethod: MethodDescriptor<QueryOuterClass.QueryProposalsRequest,
      QueryOuterClass.QueryProposalsResponse>
    @JvmStatic
    get() = QueryGrpc.getProposalsMethod()

  val voteMethod: MethodDescriptor<QueryOuterClass.QueryVoteRequest,
      QueryOuterClass.QueryVoteResponse>
    @JvmStatic
    get() = QueryGrpc.getVoteMethod()

  val votesMethod: MethodDescriptor<QueryOuterClass.QueryVotesRequest,
      QueryOuterClass.QueryVotesResponse>
    @JvmStatic
    get() = QueryGrpc.getVotesMethod()

  val paramsMethod: MethodDescriptor<QueryOuterClass.QueryParamsRequest,
      QueryOuterClass.QueryParamsResponse>
    @JvmStatic
    get() = QueryGrpc.getParamsMethod()

  val depositMethod: MethodDescriptor<QueryOuterClass.QueryDepositRequest,
      QueryOuterClass.QueryDepositResponse>
    @JvmStatic
    get() = QueryGrpc.getDepositMethod()

  val depositsMethod: MethodDescriptor<QueryOuterClass.QueryDepositsRequest,
      QueryOuterClass.QueryDepositsResponse>
    @JvmStatic
    get() = QueryGrpc.getDepositsMethod()

  val tallyResultMethod: MethodDescriptor<QueryOuterClass.QueryTallyResultRequest,
      QueryOuterClass.QueryTallyResultResponse>
    @JvmStatic
    get() = QueryGrpc.getTallyResultMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.gov.v1beta1.Query service as suspending coroutines.
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
    suspend fun proposal(request: QueryOuterClass.QueryProposalRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryProposalResponse = unaryRpc(
      channel,
      QueryGrpc.getProposalMethod(),
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
    suspend fun proposals(request: QueryOuterClass.QueryProposalsRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryProposalsResponse = unaryRpc(
      channel,
      QueryGrpc.getProposalsMethod(),
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
    suspend fun vote(request: QueryOuterClass.QueryVoteRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QueryVoteResponse = unaryRpc(
      channel,
      QueryGrpc.getVoteMethod(),
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
    suspend fun votes(request: QueryOuterClass.QueryVotesRequest, headers: Metadata = Metadata()):
        QueryOuterClass.QueryVotesResponse = unaryRpc(
      channel,
      QueryGrpc.getVotesMethod(),
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
    suspend fun deposit(request: QueryOuterClass.QueryDepositRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryDepositResponse = unaryRpc(
      channel,
      QueryGrpc.getDepositMethod(),
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
    suspend fun deposits(request: QueryOuterClass.QueryDepositsRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryDepositsResponse = unaryRpc(
      channel,
      QueryGrpc.getDepositsMethod(),
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
    suspend fun tallyResult(request: QueryOuterClass.QueryTallyResultRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryTallyResultResponse = unaryRpc(
      channel,
      QueryGrpc.getTallyResultMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.gov.v1beta1.Query service based on Kotlin coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.gov.v1beta1.Query.Proposal.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun proposal(request: QueryOuterClass.QueryProposalRequest):
        QueryOuterClass.QueryProposalResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta1.Query.Proposal is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta1.Query.Proposals.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun proposals(request: QueryOuterClass.QueryProposalsRequest):
        QueryOuterClass.QueryProposalsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta1.Query.Proposals is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta1.Query.Vote.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun vote(request: QueryOuterClass.QueryVoteRequest):
        QueryOuterClass.QueryVoteResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta1.Query.Vote is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta1.Query.Votes.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun votes(request: QueryOuterClass.QueryVotesRequest):
        QueryOuterClass.QueryVotesResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta1.Query.Votes is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta1.Query.Params.
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
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta1.Query.Params is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta1.Query.Deposit.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun deposit(request: QueryOuterClass.QueryDepositRequest):
        QueryOuterClass.QueryDepositResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta1.Query.Deposit is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta1.Query.Deposits.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun deposits(request: QueryOuterClass.QueryDepositsRequest):
        QueryOuterClass.QueryDepositsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta1.Query.Deposits is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.gov.v1beta1.Query.TallyResult.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun tallyResult(request: QueryOuterClass.QueryTallyResultRequest):
        QueryOuterClass.QueryTallyResultResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.gov.v1beta1.Query.TallyResult is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getProposalMethod(),
      implementation = ::proposal
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getProposalsMethod(),
      implementation = ::proposals
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getVoteMethod(),
      implementation = ::vote
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getVotesMethod(),
      implementation = ::votes
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getParamsMethod(),
      implementation = ::params
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDepositMethod(),
      implementation = ::deposit
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getDepositsMethod(),
      implementation = ::deposits
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getTallyResultMethod(),
      implementation = ::tallyResult
    )).build()
  }
}
