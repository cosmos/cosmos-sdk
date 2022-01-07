package cosmos.group.v1beta1

import cosmos.group.v1beta1.QueryGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.group.v1beta1.Query.
 */
object QueryGrpcKt {
  const val SERVICE_NAME: String = QueryGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = QueryGrpc.getServiceDescriptor()

  val groupInfoMethod: MethodDescriptor<QueryOuterClass.QueryGroupInfoRequest,
      QueryOuterClass.QueryGroupInfoResponse>
    @JvmStatic
    get() = QueryGrpc.getGroupInfoMethod()

  val groupPolicyInfoMethod: MethodDescriptor<QueryOuterClass.QueryGroupPolicyInfoRequest,
      QueryOuterClass.QueryGroupPolicyInfoResponse>
    @JvmStatic
    get() = QueryGrpc.getGroupPolicyInfoMethod()

  val groupMembersMethod: MethodDescriptor<QueryOuterClass.QueryGroupMembersRequest,
      QueryOuterClass.QueryGroupMembersResponse>
    @JvmStatic
    get() = QueryGrpc.getGroupMembersMethod()

  val groupsByAdminMethod: MethodDescriptor<QueryOuterClass.QueryGroupsByAdminRequest,
      QueryOuterClass.QueryGroupsByAdminResponse>
    @JvmStatic
    get() = QueryGrpc.getGroupsByAdminMethod()

  val groupPoliciesByGroupMethod: MethodDescriptor<QueryOuterClass.QueryGroupPoliciesByGroupRequest,
      QueryOuterClass.QueryGroupPoliciesByGroupResponse>
    @JvmStatic
    get() = QueryGrpc.getGroupPoliciesByGroupMethod()

  val groupPoliciesByAdminMethod: MethodDescriptor<QueryOuterClass.QueryGroupPoliciesByAdminRequest,
      QueryOuterClass.QueryGroupPoliciesByAdminResponse>
    @JvmStatic
    get() = QueryGrpc.getGroupPoliciesByAdminMethod()

  val proposalMethod: MethodDescriptor<QueryOuterClass.QueryProposalRequest,
      QueryOuterClass.QueryProposalResponse>
    @JvmStatic
    get() = QueryGrpc.getProposalMethod()

  val proposalsByGroupPolicyMethod:
      MethodDescriptor<QueryOuterClass.QueryProposalsByGroupPolicyRequest,
      QueryOuterClass.QueryProposalsByGroupPolicyResponse>
    @JvmStatic
    get() = QueryGrpc.getProposalsByGroupPolicyMethod()

  val voteByProposalVoterMethod: MethodDescriptor<QueryOuterClass.QueryVoteByProposalVoterRequest,
      QueryOuterClass.QueryVoteByProposalVoterResponse>
    @JvmStatic
    get() = QueryGrpc.getVoteByProposalVoterMethod()

  val votesByProposalMethod: MethodDescriptor<QueryOuterClass.QueryVotesByProposalRequest,
      QueryOuterClass.QueryVotesByProposalResponse>
    @JvmStatic
    get() = QueryGrpc.getVotesByProposalMethod()

  val votesByVoterMethod: MethodDescriptor<QueryOuterClass.QueryVotesByVoterRequest,
      QueryOuterClass.QueryVotesByVoterResponse>
    @JvmStatic
    get() = QueryGrpc.getVotesByVoterMethod()

  val groupsByMemberMethod: MethodDescriptor<QueryOuterClass.QueryGroupsByMemberRequest,
      QueryOuterClass.QueryGroupsByMemberResponse>
    @JvmStatic
    get() = QueryGrpc.getGroupsByMemberMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.group.v1beta1.Query service as suspending coroutines.
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
    suspend fun groupInfo(request: QueryOuterClass.QueryGroupInfoRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryGroupInfoResponse = unaryRpc(
      channel,
      QueryGrpc.getGroupInfoMethod(),
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
    suspend fun groupPolicyInfo(request: QueryOuterClass.QueryGroupPolicyInfoRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryGroupPolicyInfoResponse = unaryRpc(
      channel,
      QueryGrpc.getGroupPolicyInfoMethod(),
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
    suspend fun groupMembers(request: QueryOuterClass.QueryGroupMembersRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryGroupMembersResponse = unaryRpc(
      channel,
      QueryGrpc.getGroupMembersMethod(),
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
    suspend fun groupsByAdmin(request: QueryOuterClass.QueryGroupsByAdminRequest, headers: Metadata
        = Metadata()): QueryOuterClass.QueryGroupsByAdminResponse = unaryRpc(
      channel,
      QueryGrpc.getGroupsByAdminMethod(),
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
    suspend fun groupPoliciesByGroup(request: QueryOuterClass.QueryGroupPoliciesByGroupRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryGroupPoliciesByGroupResponse =
        unaryRpc(
      channel,
      QueryGrpc.getGroupPoliciesByGroupMethod(),
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
    suspend fun groupPoliciesByAdmin(request: QueryOuterClass.QueryGroupPoliciesByAdminRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryGroupPoliciesByAdminResponse =
        unaryRpc(
      channel,
      QueryGrpc.getGroupPoliciesByAdminMethod(),
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
    suspend fun proposalsByGroupPolicy(request: QueryOuterClass.QueryProposalsByGroupPolicyRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryProposalsByGroupPolicyResponse =
        unaryRpc(
      channel,
      QueryGrpc.getProposalsByGroupPolicyMethod(),
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
    suspend fun voteByProposalVoter(request: QueryOuterClass.QueryVoteByProposalVoterRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryVoteByProposalVoterResponse =
        unaryRpc(
      channel,
      QueryGrpc.getVoteByProposalVoterMethod(),
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
    suspend fun votesByProposal(request: QueryOuterClass.QueryVotesByProposalRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryVotesByProposalResponse = unaryRpc(
      channel,
      QueryGrpc.getVotesByProposalMethod(),
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
    suspend fun votesByVoter(request: QueryOuterClass.QueryVotesByVoterRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryVotesByVoterResponse = unaryRpc(
      channel,
      QueryGrpc.getVotesByVoterMethod(),
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
    suspend fun groupsByMember(request: QueryOuterClass.QueryGroupsByMemberRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryGroupsByMemberResponse = unaryRpc(
      channel,
      QueryGrpc.getGroupsByMemberMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.group.v1beta1.Query service based on Kotlin coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.GroupInfo.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun groupInfo(request: QueryOuterClass.QueryGroupInfoRequest):
        QueryOuterClass.QueryGroupInfoResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.GroupInfo is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.GroupPolicyInfo.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun groupPolicyInfo(request: QueryOuterClass.QueryGroupPolicyInfoRequest):
        QueryOuterClass.QueryGroupPolicyInfoResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.GroupPolicyInfo is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.GroupMembers.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun groupMembers(request: QueryOuterClass.QueryGroupMembersRequest):
        QueryOuterClass.QueryGroupMembersResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.GroupMembers is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.GroupsByAdmin.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun groupsByAdmin(request: QueryOuterClass.QueryGroupsByAdminRequest):
        QueryOuterClass.QueryGroupsByAdminResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.GroupsByAdmin is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.GroupPoliciesByGroup.
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
        fun groupPoliciesByGroup(request: QueryOuterClass.QueryGroupPoliciesByGroupRequest):
        QueryOuterClass.QueryGroupPoliciesByGroupResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.GroupPoliciesByGroup is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.GroupPoliciesByAdmin.
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
        fun groupPoliciesByAdmin(request: QueryOuterClass.QueryGroupPoliciesByAdminRequest):
        QueryOuterClass.QueryGroupPoliciesByAdminResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.GroupPoliciesByAdmin is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.Proposal.
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
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.Proposal is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.ProposalsByGroupPolicy.
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
        fun proposalsByGroupPolicy(request: QueryOuterClass.QueryProposalsByGroupPolicyRequest):
        QueryOuterClass.QueryProposalsByGroupPolicyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.ProposalsByGroupPolicy is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.VoteByProposalVoter.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun voteByProposalVoter(request: QueryOuterClass.QueryVoteByProposalVoterRequest):
        QueryOuterClass.QueryVoteByProposalVoterResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.VoteByProposalVoter is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.VotesByProposal.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun votesByProposal(request: QueryOuterClass.QueryVotesByProposalRequest):
        QueryOuterClass.QueryVotesByProposalResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.VotesByProposal is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.VotesByVoter.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun votesByVoter(request: QueryOuterClass.QueryVotesByVoterRequest):
        QueryOuterClass.QueryVotesByVoterResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.VotesByVoter is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Query.GroupsByMember.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun groupsByMember(request: QueryOuterClass.QueryGroupsByMemberRequest):
        QueryOuterClass.QueryGroupsByMemberResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Query.GroupsByMember is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getGroupInfoMethod(),
      implementation = ::groupInfo
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getGroupPolicyInfoMethod(),
      implementation = ::groupPolicyInfo
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getGroupMembersMethod(),
      implementation = ::groupMembers
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getGroupsByAdminMethod(),
      implementation = ::groupsByAdmin
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getGroupPoliciesByGroupMethod(),
      implementation = ::groupPoliciesByGroup
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getGroupPoliciesByAdminMethod(),
      implementation = ::groupPoliciesByAdmin
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getProposalMethod(),
      implementation = ::proposal
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getProposalsByGroupPolicyMethod(),
      implementation = ::proposalsByGroupPolicy
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getVoteByProposalVoterMethod(),
      implementation = ::voteByProposalVoter
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getVotesByProposalMethod(),
      implementation = ::votesByProposal
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getVotesByVoterMethod(),
      implementation = ::votesByVoter
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getGroupsByMemberMethod(),
      implementation = ::groupsByMember
    )).build()
  }
}
