package tendermint.abci

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
import tendermint.abci.ABCIApplicationGrpc.getServiceDescriptor

/**
 * Holder for Kotlin coroutine-based client and server APIs for tendermint.abci.ABCIApplication.
 */
object ABCIApplicationGrpcKt {
  const val SERVICE_NAME: String = ABCIApplicationGrpc.SERVICE_NAME

  @JvmStatic
  val serviceDescriptor: ServiceDescriptor
    get() = ABCIApplicationGrpc.getServiceDescriptor()

  val echoMethod: MethodDescriptor<Types.RequestEcho, Types.ResponseEcho>
    @JvmStatic
    get() = ABCIApplicationGrpc.getEchoMethod()

  val flushMethod: MethodDescriptor<Types.RequestFlush, Types.ResponseFlush>
    @JvmStatic
    get() = ABCIApplicationGrpc.getFlushMethod()

  val infoMethod: MethodDescriptor<Types.RequestInfo, Types.ResponseInfo>
    @JvmStatic
    get() = ABCIApplicationGrpc.getInfoMethod()

  val setOptionMethod: MethodDescriptor<Types.RequestSetOption, Types.ResponseSetOption>
    @JvmStatic
    get() = ABCIApplicationGrpc.getSetOptionMethod()

  val deliverTxMethod: MethodDescriptor<Types.RequestDeliverTx, Types.ResponseDeliverTx>
    @JvmStatic
    get() = ABCIApplicationGrpc.getDeliverTxMethod()

  val checkTxMethod: MethodDescriptor<Types.RequestCheckTx, Types.ResponseCheckTx>
    @JvmStatic
    get() = ABCIApplicationGrpc.getCheckTxMethod()

  val queryMethod: MethodDescriptor<Types.RequestQuery, Types.ResponseQuery>
    @JvmStatic
    get() = ABCIApplicationGrpc.getQueryMethod()

  val commitMethod: MethodDescriptor<Types.RequestCommit, Types.ResponseCommit>
    @JvmStatic
    get() = ABCIApplicationGrpc.getCommitMethod()

  val initChainMethod: MethodDescriptor<Types.RequestInitChain, Types.ResponseInitChain>
    @JvmStatic
    get() = ABCIApplicationGrpc.getInitChainMethod()

  val beginBlockMethod: MethodDescriptor<Types.RequestBeginBlock, Types.ResponseBeginBlock>
    @JvmStatic
    get() = ABCIApplicationGrpc.getBeginBlockMethod()

  val endBlockMethod: MethodDescriptor<Types.RequestEndBlock, Types.ResponseEndBlock>
    @JvmStatic
    get() = ABCIApplicationGrpc.getEndBlockMethod()

  val listSnapshotsMethod: MethodDescriptor<Types.RequestListSnapshots, Types.ResponseListSnapshots>
    @JvmStatic
    get() = ABCIApplicationGrpc.getListSnapshotsMethod()

  val offerSnapshotMethod: MethodDescriptor<Types.RequestOfferSnapshot, Types.ResponseOfferSnapshot>
    @JvmStatic
    get() = ABCIApplicationGrpc.getOfferSnapshotMethod()

  val loadSnapshotChunkMethod: MethodDescriptor<Types.RequestLoadSnapshotChunk,
      Types.ResponseLoadSnapshotChunk>
    @JvmStatic
    get() = ABCIApplicationGrpc.getLoadSnapshotChunkMethod()

  val applySnapshotChunkMethod: MethodDescriptor<Types.RequestApplySnapshotChunk,
      Types.ResponseApplySnapshotChunk>
    @JvmStatic
    get() = ABCIApplicationGrpc.getApplySnapshotChunkMethod()

  /**
   * A stub for issuing RPCs to a(n) tendermint.abci.ABCIApplication service as suspending
   * coroutines.
   */
  @StubFor(ABCIApplicationGrpc::class)
  class ABCIApplicationCoroutineStub @JvmOverloads constructor(
    channel: Channel,
    callOptions: CallOptions = DEFAULT
  ) : AbstractCoroutineStub<ABCIApplicationCoroutineStub>(channel, callOptions) {
    override fun build(channel: Channel, callOptions: CallOptions): ABCIApplicationCoroutineStub =
        ABCIApplicationCoroutineStub(channel, callOptions)

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
    suspend fun echo(request: Types.RequestEcho, headers: Metadata = Metadata()): Types.ResponseEcho
        = unaryRpc(
      channel,
      ABCIApplicationGrpc.getEchoMethod(),
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
    suspend fun flush(request: Types.RequestFlush, headers: Metadata = Metadata()):
        Types.ResponseFlush = unaryRpc(
      channel,
      ABCIApplicationGrpc.getFlushMethod(),
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
    suspend fun info(request: Types.RequestInfo, headers: Metadata = Metadata()): Types.ResponseInfo
        = unaryRpc(
      channel,
      ABCIApplicationGrpc.getInfoMethod(),
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
    suspend fun setOption(request: Types.RequestSetOption, headers: Metadata = Metadata()):
        Types.ResponseSetOption = unaryRpc(
      channel,
      ABCIApplicationGrpc.getSetOptionMethod(),
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
    suspend fun deliverTx(request: Types.RequestDeliverTx, headers: Metadata = Metadata()):
        Types.ResponseDeliverTx = unaryRpc(
      channel,
      ABCIApplicationGrpc.getDeliverTxMethod(),
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
    suspend fun checkTx(request: Types.RequestCheckTx, headers: Metadata = Metadata()):
        Types.ResponseCheckTx = unaryRpc(
      channel,
      ABCIApplicationGrpc.getCheckTxMethod(),
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
    suspend fun query(request: Types.RequestQuery, headers: Metadata = Metadata()):
        Types.ResponseQuery = unaryRpc(
      channel,
      ABCIApplicationGrpc.getQueryMethod(),
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
    suspend fun commit(request: Types.RequestCommit, headers: Metadata = Metadata()):
        Types.ResponseCommit = unaryRpc(
      channel,
      ABCIApplicationGrpc.getCommitMethod(),
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
    suspend fun initChain(request: Types.RequestInitChain, headers: Metadata = Metadata()):
        Types.ResponseInitChain = unaryRpc(
      channel,
      ABCIApplicationGrpc.getInitChainMethod(),
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
    suspend fun beginBlock(request: Types.RequestBeginBlock, headers: Metadata = Metadata()):
        Types.ResponseBeginBlock = unaryRpc(
      channel,
      ABCIApplicationGrpc.getBeginBlockMethod(),
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
    suspend fun endBlock(request: Types.RequestEndBlock, headers: Metadata = Metadata()):
        Types.ResponseEndBlock = unaryRpc(
      channel,
      ABCIApplicationGrpc.getEndBlockMethod(),
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
    suspend fun listSnapshots(request: Types.RequestListSnapshots, headers: Metadata = Metadata()):
        Types.ResponseListSnapshots = unaryRpc(
      channel,
      ABCIApplicationGrpc.getListSnapshotsMethod(),
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
    suspend fun offerSnapshot(request: Types.RequestOfferSnapshot, headers: Metadata = Metadata()):
        Types.ResponseOfferSnapshot = unaryRpc(
      channel,
      ABCIApplicationGrpc.getOfferSnapshotMethod(),
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
    suspend fun loadSnapshotChunk(request: Types.RequestLoadSnapshotChunk, headers: Metadata =
        Metadata()): Types.ResponseLoadSnapshotChunk = unaryRpc(
      channel,
      ABCIApplicationGrpc.getLoadSnapshotChunkMethod(),
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
    suspend fun applySnapshotChunk(request: Types.RequestApplySnapshotChunk, headers: Metadata =
        Metadata()): Types.ResponseApplySnapshotChunk = unaryRpc(
      channel,
      ABCIApplicationGrpc.getApplySnapshotChunkMethod(),
      request,
      callOptions,
      headers
    )}

  /**
   * Skeletal implementation of the tendermint.abci.ABCIApplication service based on Kotlin
   * coroutines.
   */
  abstract class ABCIApplicationCoroutineImplBase(
    coroutineContext: CoroutineContext = EmptyCoroutineContext
  ) : AbstractCoroutineServerImpl(coroutineContext) {
    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.Echo.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun echo(request: Types.RequestEcho): Types.ResponseEcho = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.Echo is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.Flush.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun flush(request: Types.RequestFlush): Types.ResponseFlush = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.Flush is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.Info.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun info(request: Types.RequestInfo): Types.ResponseInfo = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.Info is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.SetOption.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun setOption(request: Types.RequestSetOption): Types.ResponseSetOption = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.SetOption is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.DeliverTx.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun deliverTx(request: Types.RequestDeliverTx): Types.ResponseDeliverTx = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.DeliverTx is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.CheckTx.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun checkTx(request: Types.RequestCheckTx): Types.ResponseCheckTx = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.CheckTx is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.Query.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun query(request: Types.RequestQuery): Types.ResponseQuery = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.Query is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.Commit.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun commit(request: Types.RequestCommit): Types.ResponseCommit = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.Commit is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.InitChain.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun initChain(request: Types.RequestInitChain): Types.ResponseInitChain = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.InitChain is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.BeginBlock.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun beginBlock(request: Types.RequestBeginBlock): Types.ResponseBeginBlock = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.BeginBlock is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.EndBlock.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun endBlock(request: Types.RequestEndBlock): Types.ResponseEndBlock = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.EndBlock is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.ListSnapshots.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun listSnapshots(request: Types.RequestListSnapshots): Types.ResponseListSnapshots
        = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.ListSnapshots is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.OfferSnapshot.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun offerSnapshot(request: Types.RequestOfferSnapshot): Types.ResponseOfferSnapshot
        = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.OfferSnapshot is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.LoadSnapshotChunk.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun loadSnapshotChunk(request: Types.RequestLoadSnapshotChunk):
        Types.ResponseLoadSnapshotChunk = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.LoadSnapshotChunk is unimplemented"))

    /**
     * Returns the response to an RPC for tendermint.abci.ABCIApplication.ApplySnapshotChunk.
     *
     * If this method fails with a [StatusException], the RPC will fail with the corresponding
     * [Status].  If this method fails with a [java.util.concurrent.CancellationException], the RPC
     * will fail
     * with status `Status.CANCELLED`.  If this method fails for any other reason, the RPC will
     * fail with `Status.UNKNOWN` with the exception as a cause.
     *
     * @param request The request from the client.
     */
    open suspend fun applySnapshotChunk(request: Types.RequestApplySnapshotChunk):
        Types.ResponseApplySnapshotChunk = throw
        StatusException(UNIMPLEMENTED.withDescription("Method tendermint.abci.ABCIApplication.ApplySnapshotChunk is unimplemented"))

    final override fun bindService(): ServerServiceDefinition = builder(getServiceDescriptor())
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getEchoMethod(),
      implementation = ::echo
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getFlushMethod(),
      implementation = ::flush
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getInfoMethod(),
      implementation = ::info
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getSetOptionMethod(),
      implementation = ::setOption
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getDeliverTxMethod(),
      implementation = ::deliverTx
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getCheckTxMethod(),
      implementation = ::checkTx
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getQueryMethod(),
      implementation = ::query
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getCommitMethod(),
      implementation = ::commit
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getInitChainMethod(),
      implementation = ::initChain
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getBeginBlockMethod(),
      implementation = ::beginBlock
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getEndBlockMethod(),
      implementation = ::endBlock
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getListSnapshotsMethod(),
      implementation = ::listSnapshots
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getOfferSnapshotMethod(),
      implementation = ::offerSnapshot
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getLoadSnapshotChunkMethod(),
      implementation = ::loadSnapshotChunk
    ))
      .addMethod(unaryServerMethodDefinition(
      context = this.context,
      descriptor = ABCIApplicationGrpc.getApplySnapshotChunkMethod(),
      implementation = ::applySnapshotChunk
    )).build()
  }
}
