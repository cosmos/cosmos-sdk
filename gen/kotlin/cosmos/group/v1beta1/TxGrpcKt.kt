package cosmos.group.v1beta1

import cosmos.group.v1beta1.MsgGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.group.v1beta1.Msg.
 */
object MsgGrpcKt {
  const val SERVICE_NAME: String = MsgGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = MsgGrpc.getServiceDescriptor()

  val createGroupMethod: MethodDescriptor<Tx.MsgCreateGroup, Tx.MsgCreateGroupResponse>
    @JvmStatic
    get() = MsgGrpc.getCreateGroupMethod()

  val updateGroupMembersMethod: MethodDescriptor<Tx.MsgUpdateGroupMembers,
      Tx.MsgUpdateGroupMembersResponse>
    @JvmStatic
    get() = MsgGrpc.getUpdateGroupMembersMethod()

  val updateGroupAdminMethod: MethodDescriptor<Tx.MsgUpdateGroupAdmin,
      Tx.MsgUpdateGroupAdminResponse>
    @JvmStatic
    get() = MsgGrpc.getUpdateGroupAdminMethod()

  val updateGroupMetadataMethod: MethodDescriptor<Tx.MsgUpdateGroupMetadata,
      Tx.MsgUpdateGroupMetadataResponse>
    @JvmStatic
    get() = MsgGrpc.getUpdateGroupMetadataMethod()

  val createGroupPolicyMethod: MethodDescriptor<Tx.MsgCreateGroupPolicy,
      Tx.MsgCreateGroupPolicyResponse>
    @JvmStatic
    get() = MsgGrpc.getCreateGroupPolicyMethod()

  val updateGroupPolicyAdminMethod: MethodDescriptor<Tx.MsgUpdateGroupPolicyAdmin,
      Tx.MsgUpdateGroupPolicyAdminResponse>
    @JvmStatic
    get() = MsgGrpc.getUpdateGroupPolicyAdminMethod()

  val updateGroupPolicyDecisionPolicyMethod: MethodDescriptor<Tx.MsgUpdateGroupPolicyDecisionPolicy,
      Tx.MsgUpdateGroupPolicyDecisionPolicyResponse>
    @JvmStatic
    get() = MsgGrpc.getUpdateGroupPolicyDecisionPolicyMethod()

  val updateGroupPolicyMetadataMethod: MethodDescriptor<Tx.MsgUpdateGroupPolicyMetadata,
      Tx.MsgUpdateGroupPolicyMetadataResponse>
    @JvmStatic
    get() = MsgGrpc.getUpdateGroupPolicyMetadataMethod()

  val createProposalMethod: MethodDescriptor<Tx.MsgCreateProposal, Tx.MsgCreateProposalResponse>
    @JvmStatic
    get() = MsgGrpc.getCreateProposalMethod()

  val voteMethod: MethodDescriptor<Tx.MsgVote, Tx.MsgVoteResponse>
    @JvmStatic
    get() = MsgGrpc.getVoteMethod()

  val execMethod: MethodDescriptor<Tx.MsgExec, Tx.MsgExecResponse>
    @JvmStatic
    get() = MsgGrpc.getExecMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.group.v1beta1.Msg service as suspending coroutines.
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
    suspend fun createGroup(request: Tx.MsgCreateGroup, headers: Metadata = Metadata()):
        Tx.MsgCreateGroupResponse = unaryRpc(
      channel,
      MsgGrpc.getCreateGroupMethod(),
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
    suspend fun updateGroupMembers(request: Tx.MsgUpdateGroupMembers, headers: Metadata =
        Metadata()): Tx.MsgUpdateGroupMembersResponse = unaryRpc(
      channel,
      MsgGrpc.getUpdateGroupMembersMethod(),
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
    suspend fun updateGroupAdmin(request: Tx.MsgUpdateGroupAdmin, headers: Metadata = Metadata()):
        Tx.MsgUpdateGroupAdminResponse = unaryRpc(
      channel,
      MsgGrpc.getUpdateGroupAdminMethod(),
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
    suspend fun updateGroupMetadata(request: Tx.MsgUpdateGroupMetadata, headers: Metadata =
        Metadata()): Tx.MsgUpdateGroupMetadataResponse = unaryRpc(
      channel,
      MsgGrpc.getUpdateGroupMetadataMethod(),
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
    suspend fun createGroupPolicy(request: Tx.MsgCreateGroupPolicy, headers: Metadata = Metadata()):
        Tx.MsgCreateGroupPolicyResponse = unaryRpc(
      channel,
      MsgGrpc.getCreateGroupPolicyMethod(),
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
    suspend fun updateGroupPolicyAdmin(request: Tx.MsgUpdateGroupPolicyAdmin, headers: Metadata =
        Metadata()): Tx.MsgUpdateGroupPolicyAdminResponse = unaryRpc(
      channel,
      MsgGrpc.getUpdateGroupPolicyAdminMethod(),
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
    suspend fun updateGroupPolicyDecisionPolicy(request: Tx.MsgUpdateGroupPolicyDecisionPolicy,
        headers: Metadata = Metadata()): Tx.MsgUpdateGroupPolicyDecisionPolicyResponse = unaryRpc(
      channel,
      MsgGrpc.getUpdateGroupPolicyDecisionPolicyMethod(),
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
    suspend fun updateGroupPolicyMetadata(request: Tx.MsgUpdateGroupPolicyMetadata,
        headers: Metadata = Metadata()): Tx.MsgUpdateGroupPolicyMetadataResponse = unaryRpc(
      channel,
      MsgGrpc.getUpdateGroupPolicyMetadataMethod(),
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
    suspend fun createProposal(request: Tx.MsgCreateProposal, headers: Metadata = Metadata()):
        Tx.MsgCreateProposalResponse = unaryRpc(
      channel,
      MsgGrpc.getCreateProposalMethod(),
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
    suspend fun exec(request: Tx.MsgExec, headers: Metadata = Metadata()): Tx.MsgExecResponse =
        unaryRpc(
      channel,
      MsgGrpc.getExecMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.group.v1beta1.Msg service based on Kotlin coroutines.
   */
  abstract class MsgCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.CreateGroup.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun createGroup(request: Tx.MsgCreateGroup): Tx.MsgCreateGroupResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.CreateGroup is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.UpdateGroupMembers.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun updateGroupMembers(request: Tx.MsgUpdateGroupMembers):
        Tx.MsgUpdateGroupMembersResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.UpdateGroupMembers is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.UpdateGroupAdmin.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun updateGroupAdmin(request: Tx.MsgUpdateGroupAdmin):
        Tx.MsgUpdateGroupAdminResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.UpdateGroupAdmin is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.UpdateGroupMetadata.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun updateGroupMetadata(request: Tx.MsgUpdateGroupMetadata):
        Tx.MsgUpdateGroupMetadataResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.UpdateGroupMetadata is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.CreateGroupPolicy.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun createGroupPolicy(request: Tx.MsgCreateGroupPolicy):
        Tx.MsgCreateGroupPolicyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.CreateGroupPolicy is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.UpdateGroupPolicyAdmin.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun updateGroupPolicyAdmin(request: Tx.MsgUpdateGroupPolicyAdmin):
        Tx.MsgUpdateGroupPolicyAdminResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.UpdateGroupPolicyAdmin is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.UpdateGroupPolicyDecisionPolicy.
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
        fun updateGroupPolicyDecisionPolicy(request: Tx.MsgUpdateGroupPolicyDecisionPolicy):
        Tx.MsgUpdateGroupPolicyDecisionPolicyResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.UpdateGroupPolicyDecisionPolicy is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.UpdateGroupPolicyMetadata.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun updateGroupPolicyMetadata(request: Tx.MsgUpdateGroupPolicyMetadata):
        Tx.MsgUpdateGroupPolicyMetadataResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.UpdateGroupPolicyMetadata is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.CreateProposal.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun createProposal(request: Tx.MsgCreateProposal): Tx.MsgCreateProposalResponse =
        throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.CreateProposal is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.Vote.
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
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.Vote is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.group.v1beta1.Msg.Exec.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun exec(request: Tx.MsgExec): Tx.MsgExecResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.group.v1beta1.Msg.Exec is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getCreateGroupMethod(),
      implementation = ::createGroup
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getUpdateGroupMembersMethod(),
      implementation = ::updateGroupMembers
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getUpdateGroupAdminMethod(),
      implementation = ::updateGroupAdmin
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getUpdateGroupMetadataMethod(),
      implementation = ::updateGroupMetadata
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getCreateGroupPolicyMethod(),
      implementation = ::createGroupPolicy
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getUpdateGroupPolicyAdminMethod(),
      implementation = ::updateGroupPolicyAdmin
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getUpdateGroupPolicyDecisionPolicyMethod(),
      implementation = ::updateGroupPolicyDecisionPolicy
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getUpdateGroupPolicyMetadataMethod(),
      implementation = ::updateGroupPolicyMetadata
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getCreateProposalMethod(),
      implementation = ::createProposal
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getVoteMethod(),
      implementation = ::vote
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = MsgGrpc.getExecMethod(),
      implementation = ::exec
    )).build()
  }
}
