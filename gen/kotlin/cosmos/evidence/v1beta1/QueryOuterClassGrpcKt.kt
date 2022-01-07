package cosmos.evidence.v1beta1

import cosmos.evidence.v1beta1.QueryGrpc.getServiceDescriptor
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
 * Holder for Kotlin coroutine-based client and server APIs for cosmos.evidence.v1beta1.Query.
 */
object QueryGrpcKt {
  const val SERVICE_NAME: String = QueryGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = QueryGrpc.getServiceDescriptor()

  val evidenceMethod: MethodDescriptor<QueryOuterClass.QueryEvidenceRequest,
      QueryOuterClass.QueryEvidenceResponse>
    @JvmStatic
    get() = QueryGrpc.getEvidenceMethod()

  val allEvidenceMethod: MethodDescriptor<QueryOuterClass.QueryAllEvidenceRequest,
      QueryOuterClass.QueryAllEvidenceResponse>
    @JvmStatic
    get() = QueryGrpc.getAllEvidenceMethod()

  /**
   * A stub for issuing RPCs to a(n) cosmos.evidence.v1beta1.Query service as suspending coroutines.
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
    suspend fun evidence(request: QueryOuterClass.QueryEvidenceRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryEvidenceResponse = unaryRpc(
      channel,
      QueryGrpc.getEvidenceMethod(),
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
    suspend fun allEvidence(request: QueryOuterClass.QueryAllEvidenceRequest, headers: Metadata =
        Metadata()): QueryOuterClass.QueryAllEvidenceResponse = unaryRpc(
      channel,
      QueryGrpc.getAllEvidenceMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the cosmos.evidence.v1beta1.Query service based on Kotlin
   * coroutines.
   */
  abstract class QueryCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for cosmos.evidence.v1beta1.Query.Evidence.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun evidence(request: QueryOuterClass.QueryEvidenceRequest):
        QueryOuterClass.QueryEvidenceResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.evidence.v1beta1.Query.Evidence is unimplemented"))

    /**
     * Returns the response to an RPC for cosmos.evidence.v1beta1.Query.AllEvidence.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun allEvidence(request: QueryOuterClass.QueryAllEvidenceRequest):
        QueryOuterClass.QueryAllEvidenceResponse = throw
        StatusException(UNIMPLEMENTED.withDescription("Method cosmos.evidence.v1beta1.Query.AllEvidence is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getEvidenceMethod(),
      implementation = ::evidence
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = QueryGrpc.getAllEvidenceMethod(),
      implementation = ::allEvidence
    )).build()
  }
}
