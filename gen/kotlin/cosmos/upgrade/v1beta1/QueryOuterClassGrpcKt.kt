package cosmos.upgrade.v1beta1

import cosmos.upgrade.v1beta1.QueryGrpc.getServiceDescriptor
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
import kotlin.Deprecated
import kotlin.String
import kotlin.coroutines.CoroutineContext
import kotlin.coroutines.EmptyCoroutineContext
import kotlin.jvm.JvmOverloads
import kotlin.jvm.JvmStatic

/**
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.upgrade.v1beta1.Query.
 */
object QueryGrpcKt {
  const val SERVICE_NAME: String = QueryGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = QueryGrpc.getServiceDescriptor()

  val currentPlanMethod: MethodDescriptor<QueryOuterClass.QueryCurrentPlanRequest,
      QueryOuterClass.QueryCurrentPlanResponse>
    @JvmStatic
    get() = QueryGrpc.getCurrentPlanMethod()

  val appliedPlanMethod: MethodDescriptor<QueryOuterClass.QueryAppliedPlanRequest,
      QueryOuterClass.QueryAppliedPlanResponse>
    @JvmStatic
    get() = QueryGrpc.getAppliedPlanMethod()

  val upgradedConsensusStateMethod:
      MethodDescriptor<QueryOuterClass.QueryUpgradedConsensusStateRequest,
      QueryOuterClass.QueryUpgradedConsensusStateResponse>
    @JvmStatic
    get() = QueryGrpc.getUpgradedConsensusStateMethod()

  val moduleVersionsMethod: MethodDescriptor<QueryOuterClass.QueryModuleVersionsRequest,
      QueryOuterClass.QueryModuleVersionsResponse>
    @JvmStatic
    get() = QueryGrpc.getModuleVersionsMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.upgrade.v1beta1.Query service as suspending coroutines.
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
    suspend fun currentPlan(request: QueryOuterClass.QueryCurrentPlanRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryCurrentPlanResponse = unaryRpc(
      channel,
      QueryGrpc.getCurrentPlanMethod(),
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
    suspend fun appliedPlan(request: QueryOuterClass.QueryAppliedPlanRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryAppliedPlanResponse = unaryRpc(
      channel,
      QueryGrpc.getAppliedPlanMethod(),
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
    @Deprecated("The underlying service method is marked deprecated.")
    suspend fun upgradedConsensusState(request: QueryOuterClass.QueryUpgradedConsensusStateRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryUpgradedConsensusStateResponse =
        unaryRpc(
      channel,
      QueryGrpc.getUpgradedConsensusStateMethod(),
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
    suspend fun moduleVersions(request: QueryOuterClass.QueryModuleVersionsRequest,
        headers: Metadata = Metadata()): QueryOuterClass.QueryModuleVersionsResponse = unaryRpc(
      channel,
      QueryGrpc.getModuleVersionsMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.upgrade.v1beta1.Query service based on Kotlin coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.upgrade.v1beta1.Query.CurrentPlan.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun currentPlan(request: QueryOuterClass.QueryCurrentPlanRequest):
        QueryOuterClass.QueryCurrentPlanResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.upgrade.v1beta1.Query.CurrentPlan is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.upgrade.v1beta1.Query.AppliedPlan.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun appliedPlan(request: QueryOuterClass.QueryAppliedPlanRequest):
        QueryOuterClass.QueryAppliedPlanResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.upgrade.v1beta1.Query.AppliedPlan is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.upgrade.v1beta1.Query.UpgradedConsensusState.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    @Deprecated("The underlying service method is marked deprecated.")
    open suspend
        fun upgradedConsensusState(request: QueryOuterClass.QueryUpgradedConsensusStateRequest):
        QueryOuterClass.QueryUpgradedConsensusStateResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.upgrade.v1beta1.Query.UpgradedConsensusState is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.upgrade.v1beta1.Query.ModuleVersions.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun moduleVersions(request: QueryOuterClass.QueryModuleVersionsRequest):
        QueryOuterClass.QueryModuleVersionsResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.upgrade.v1beta1.Query.ModuleVersions is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getCurrentPlanMethod(),
      implementation = ::currentPlan
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getAppliedPlanMethod(),
      implementation = ::appliedPlan
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getUpgradedConsensusStateMethod(),
      implementation = ::upgradedConsensusState
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getModuleVersionsMethod(),
      implementation = ::moduleVersions
    )).build()
  }
}
