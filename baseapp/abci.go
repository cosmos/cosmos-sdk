package baseapp

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/codec"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const initialAppVersion = 0
const StoreInfoPath = "info"

type AppVersionError struct {
	Actual  uint64
	Initial uint64
}

func (e *AppVersionError) Error() string {
	return fmt.Sprintf("app version (%d) is not initial (%d)", e.Actual, e.Initial)
}

// InitChain implements the ABCI interface. It runs the initialization logic
// directly on the CommitMultiStore.
func (app *BaseApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	// On a new chain, we consider the init chain block height as 0, even though
	// req.InitialHeight is 1 by default.
	initHeader := tmproto.Header{ChainID: req.ChainId, Time: req.Time}

	// If req.InitialHeight is > 1, then we set the initial version in the
	// stores.
	if req.InitialHeight > 1 {
		app.initialHeight = req.InitialHeight
		initHeader = tmproto.Header{ChainID: req.ChainId, Height: req.InitialHeight, Time: req.Time}
		err := app.cms.SetInitialVersion(req.InitialHeight)
		if err != nil {
			panic(err)
		}
	}

	// initialize the deliver state and check state with a correct header
	app.setDeliverState(initHeader)
	app.setCheckState(initHeader)

	if err := app.SetAppVersion(initialAppVersion); err != nil {
		panic(err)
	}

	// Store the consensus params in the BaseApp's paramstore. Note, this must be
	// done after the deliver state and context have been set as it's persisted
	// to state.
	if req.ConsensusParams != nil {
		// When InitChain is called, the app version should either be absent and determined by the application
		// or set to 0. Panic if it's not.
		if req.ConsensusParams.Version != nil && req.ConsensusParams.Version.AppVersion != initialAppVersion {
			panic(AppVersionError{Actual: req.ConsensusParams.Version.AppVersion, Initial: initialAppVersion})
		}

		app.StoreConsensusParams(app.deliverState.ctx, req.ConsensusParams)
	}

	if app.initChainer == nil {
		return
	}

	// add block gas meter for any genesis transactions (allow infinite gas)
	app.deliverState.ctx = app.deliverState.ctx.WithBlockGasMeter(sdk.NewInfiniteGasMeter())
	app.deliverState.ctx = app.deliverState.ctx.WithIsGenesis(true)

	res = app.initChainer(app.deliverState.ctx, req)

	app.deliverState.ctx = app.deliverState.ctx.WithIsGenesis(false)

	// sanity check
	if len(req.Validators) > 0 {
		if len(req.Validators) != len(res.Validators) {
			panic(
				fmt.Errorf(
					"len(RequestInitChain.Validators) != len(GenesisValidators) (%d != %d)",
					len(req.Validators), len(res.Validators),
				),
			)
		}

		sort.Sort(abci.ValidatorUpdates(req.Validators))
		sort.Sort(abci.ValidatorUpdates(res.Validators))

		for i := range res.Validators {
			if !proto.Equal(&res.Validators[i], &req.Validators[i]) {
				panic(fmt.Errorf("genesisValidators[%d] != req.Validators[%d] ", i, i))
			}
		}
	}

	// In the case of a new chain, AppHash will be the hash of an empty string.
	// During an upgrade, it'll be the hash of the last committed block.
	var appHash []byte
	if !app.LastCommitID().IsZero() {
		appHash = app.LastCommitID().Hash
	} else {
		// $ echo -n '' | sha256sum
		// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		emptyHash := sha256.Sum256([]byte{})
		appHash = emptyHash[:]
	}

	// NOTE: We don't commit, but BeginBlock for block `initial_height` starts from this
	// deliverState.
	return abci.ResponseInitChain{
		ConsensusParams: res.ConsensusParams,
		Validators:      res.Validators,
		AppHash:         appHash,
	}
}

// Info implements the ABCI interface.
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	lastCommitID := app.cms.LastCommitID()

	appVersion, err := app.GetAppVersion()
	if err != nil {
		app.logger.Error("failed to get app version", err)
	}

	return abci.ResponseInfo{
		Data:             app.name,
		Version:          app.version,
		AppVersion:       appVersion,
		LastBlockHeight:  lastCommitID.Version,
		LastBlockAppHash: lastCommitID.Hash,
	}
}

// SetOption implements the ABCI interface.
func (app *BaseApp) SetOption(req abci.RequestSetOption) (res abci.ResponseSetOption) {
	// TODO: Implement!
	return
}

// FilterPeerByAddrPort filters peers by address/port.
func (app *BaseApp) FilterPeerByAddrPort(info string) abci.ResponseQuery {
	if app.addrPeerFilter != nil {
		return app.addrPeerFilter(info)
	}

	return abci.ResponseQuery{}
}

// FilterPeerByID filters peers by node ID.
func (app *BaseApp) FilterPeerByID(info string) abci.ResponseQuery {
	if app.idPeerFilter != nil {
		return app.idPeerFilter(info)
	}

	return abci.ResponseQuery{}
}

// BeginBlock implements the ABCI application interface.
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	defer telemetry.MeasureSince(time.Now(), "abci", "begin_block")

	if app.cms.TracingEnabled() {
		app.cms.SetTracingContext(sdk.TraceContext(
			map[string]interface{}{"blockHeight": req.Header.Height},
		))
	}

	if err := app.validateHeight(req); err != nil {
		panic(err)
	}

	// Initialize the DeliverTx state. If this is the first block, it should
	// already be initialized in InitChain. Otherwise app.deliverState will be
	// nil, since it is reset on Commit.
	if app.deliverState == nil {
		app.setDeliverState(req.Header)
	} else {
		// In the first block, app.deliverState.ctx will already be initialized
		// by InitChain. Context is now updated with Header information.
		app.deliverState.ctx = app.deliverState.ctx.
			WithBlockHeader(req.Header).
			WithBlockHeight(req.Header.Height)
	}

	// add block gas meter
	var gasMeter sdk.GasMeter
	if maxGas := app.getMaximumBlockGas(app.deliverState.ctx); maxGas > 0 {
		gasMeter = sdk.NewGasMeter(maxGas)
	} else {
		gasMeter = sdk.NewInfiniteGasMeter()
	}

	// NOTE: header hash is not set in NewContext, so we manually set it here

	app.deliverState.ctx = app.deliverState.ctx.
		WithBlockGasMeter(gasMeter).
		WithHeaderHash(req.Hash).
		WithConsensusParams(app.GetConsensusParams(app.deliverState.ctx))

	// we also set block gas meter to checkState in case the application needs to
	// verify gas consumption during (Re)CheckTx
	if app.checkState != nil {
		app.checkState.ctx = app.checkState.ctx.
			WithBlockGasMeter(gasMeter).
			WithHeaderHash(req.Hash)
	}

	if app.beginBlocker != nil {
		res = app.beginBlocker(app.deliverState.ctx, req)
		res.Events = sdk.MarkEventsToIndex(res.Events, app.indexEvents)
	}
	// set the signed validators for addition to context in deliverTx
	app.voteInfos = req.LastCommitInfo.GetVotes()
	return res
}

// EndBlock implements the ABCI interface.
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	defer telemetry.MeasureSince(time.Now(), "abci", "end_block")

	if app.deliverState.ms.TracingEnabled() {
		app.deliverState.ms = app.deliverState.ms.SetTracingContext(nil).(sdk.CacheMultiStore)
	}

	if app.endBlocker != nil {
		res = app.endBlocker(app.deliverState.ctx, req)
		res.Events = sdk.MarkEventsToIndex(res.Events, app.indexEvents)
	}

	if cp := app.GetConsensusParams(app.deliverState.ctx); cp != nil {
		res.ConsensusParamUpdates = cp
	}

	return res
}

// CheckTx implements the ABCI interface and executes a tx in CheckTx mode. In
// CheckTx mode, messages are not executed. This means messages are only validated
// and only the AnteHandler is executed. State is persisted to the BaseApp's
// internal CheckTx state if the AnteHandler passes. Otherwise, the ResponseCheckTx
// will contain releveant error information. Regardless of tx execution outcome,
// the ResponseCheckTx will contain relevant gas execution context.
func (app *BaseApp) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	defer telemetry.MeasureSince(time.Now(), "abci", "check_tx")

	var mode runTxMode

	switch {
	case req.Type == abci.CheckTxType_New:
		mode = runTxModeCheck

	case req.Type == abci.CheckTxType_Recheck:
		mode = runTxModeReCheck

	default:
		panic(fmt.Sprintf("unknown RequestCheckTx type: %s", req.Type))
	}

	gInfo, result, anteEvents, err := app.runTx(mode, req.Tx)
	if err != nil {
		return sdkerrors.ResponseCheckTxWithEvents(err, gInfo.GasWanted, gInfo.GasUsed, anteEvents, app.trace)
	}

	return abci.ResponseCheckTx{
		GasWanted: int64(gInfo.GasWanted), // TODO: Should type accept unsigned ints?
		GasUsed:   int64(gInfo.GasUsed),   // TODO: Should type accept unsigned ints?
		Log:       result.Log,
		Data:      result.Data,
		Events:    sdk.MarkEventsToIndex(result.Events, app.indexEvents),
	}
}

// DeliverTx implements the ABCI interface and executes a tx in DeliverTx mode.
// State only gets persisted if all messages are valid and get executed successfully.
// Otherwise, the ResponseDeliverTx will contain releveant error information.
// Regardless of tx execution outcome, the ResponseDeliverTx will contain relevant
// gas execution context.
func (app *BaseApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	defer telemetry.MeasureSince(time.Now(), "abci", "deliver_tx")

	gInfo := sdk.GasInfo{}
	resultStr := "successful"

	defer func() {
		telemetry.IncrCounter(1, "tx", "count")
		telemetry.IncrCounter(1, "tx", resultStr)
		telemetry.SetGauge(float32(gInfo.GasUsed), "tx", "gas", "used")
		telemetry.SetGauge(float32(gInfo.GasWanted), "tx", "gas", "wanted")
	}()

	gInfo, result, anteEvents, err := app.runTx(runTxModeDeliver, req.Tx)
	if err != nil {
		resultStr = "failed"
		return sdkerrors.ResponseDeliverTxWithEvents(err, gInfo.GasWanted, gInfo.GasUsed, anteEvents, app.trace)
	}

	return abci.ResponseDeliverTx{
		GasWanted: int64(gInfo.GasWanted), // TODO: Should type accept unsigned ints?
		GasUsed:   int64(gInfo.GasUsed),   // TODO: Should type accept unsigned ints?
		Log:       result.Log,
		Data:      result.Data,
		Events:    sdk.MarkEventsToIndex(result.Events, app.indexEvents),
	}
}

// Commit implements the ABCI interface. It will commit all state that exists in
// the deliver state's multi-store and includes the resulting commit ID in the
// returned abci.ResponseCommit. Commit will set the check state based on the
// latest header and reset the deliver state. Also, if a non-zero halt height is
// defined in config, Commit will execute a deferred function call to check
// against that height and gracefully halt if it matches the latest committed
// height.
func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	defer telemetry.MeasureSince(time.Now(), "abci", "commit")

	header := app.deliverState.ctx.BlockHeader()
	retainHeight := app.GetBlockRetentionHeight(header.Height)

	// Write the DeliverTx state into branched storage and commit the MultiStore.
	// The write to the DeliverTx state writes all state transitions to the root
	// MultiStore (app.cms) so when Commit() is called is persists those values.
	app.deliverState.ms.Write()

	rms, ok := app.cms.(*rootmulti.Store)
	if ok {
		rms.SetCommitHeader(header)
	}

	commitID := app.cms.Commit()

	// Reset the Check state to the latest committed.
	//
	// NOTE: This is safe because Tendermint holds a lock on the mempool for
	// Commit. Use the header from this latest block.
	app.setCheckState(header)

	// empty/reset the deliver state
	app.deliverState = nil

	var halt bool

	switch {
	case app.haltHeight > 0 && uint64(header.Height) >= app.haltHeight:
		halt = true

	case app.haltTime > 0 && header.Time.Unix() >= int64(app.haltTime):
		halt = true
	}

	if halt {
		// Halt the binary and allow Tendermint to receive the ResponseCommit
		// response with the commit ID hash. This will allow the node to successfully
		// restart and process blocks assuming the halt configuration has been
		// reset or moved to a more distant value.
		app.halt()
	}

	go app.snapshotManager.SnapshotIfApplicable(header.Height)

	app.logger.Info("committed ABCI", "height", commitID.Version, "commit_hash", fmt.Sprintf("%X", commitID.Hash), "retain_height", retainHeight)
	return abci.ResponseCommit{
		Data:         commitID.Hash,
		RetainHeight: retainHeight,
	}
}

// halt attempts to gracefully shutdown the node via SIGINT and SIGTERM falling
// back on os.Exit if both fail.
func (app *BaseApp) halt() {
	app.logger.Info("halting node per configuration", "height", app.haltHeight, "time", app.haltTime)

	p, err := os.FindProcess(os.Getpid())
	if err == nil {
		// attempt cascading signals in case SIGINT fails (os dependent)
		sigIntErr := p.Signal(syscall.SIGINT)
		sigTermErr := p.Signal(syscall.SIGTERM)

		if sigIntErr == nil || sigTermErr == nil {
			return
		}
	}

	// Resort to exiting immediately if the process could not be found or killed
	// via SIGINT/SIGTERM signals.
	app.logger.Info("failed to send SIGINT/SIGTERM; exiting...")
	os.Exit(0)
}

// Query implements the ABCI interface. It delegates to CommitMultiStore if it
// implements Queryable.
func (app *BaseApp) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	defer telemetry.MeasureSince(time.Now(), "abci", "query")

	// Add panic recovery for all queries.
	// ref: https://github.com/cosmos/cosmos-sdk/pull/8039
	defer func() {
		if r := recover(); r != nil {
			res = sdkerrors.QueryResultWithDebug(sdkerrors.Wrapf(sdkerrors.ErrPanic, "%v", r), app.trace)
		}
	}()

	// when a client did not provide a query height, manually inject the latest
	if req.Height == 0 {
		req.Height = app.LastBlockHeight()
	}

	if req.Path == "/cosmos.tx.v1beta1.Service/BroadcastTx" {
		return sdkerrors.QueryResultWithDebug(sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "can't route a broadcast tx message"), app.trace)
	}

	// handle gRPC routes first rather than calling splitPath because '/' characters
	// are used as part of gRPC paths
	if grpcHandler := app.grpcQueryRouter.Route(req.Path); grpcHandler != nil {
		return app.handleQueryGRPC(grpcHandler, req)
	}

	path := splitPath(req.Path)
	if len(path) == 0 {
		sdkerrors.QueryResultWithDebug(sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "no query path provided"), app.trace)
	}

	switch path[0] {
	// "/app" prefix for special application queries
	case "app":
		return handleQueryApp(app, path, req)

	case "store":
		return handleQueryStore(app, path, req)

	case "p2p":
		return handleQueryP2P(app, path)

	case "custom":
		return handleQueryCustom(app, path, req)
	}

	return sdkerrors.QueryResultWithDebug(sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown query path"), app.trace)
}

// ListSnapshots implements the ABCI interface. It delegates to app.snapshotManager if set.
func (app *BaseApp) ListSnapshots(req abci.RequestListSnapshots) abci.ResponseListSnapshots {
	resp := abci.ResponseListSnapshots{Snapshots: []*abci.Snapshot{}}
	if app.snapshotManager == nil {
		return resp
	}

	snapshots, err := app.snapshotManager.List()
	if err != nil {
		app.logger.Error("failed to list snapshots", "err", err)
		return resp
	}

	for _, snapshot := range snapshots {
		abciSnapshot, err := snapshot.ToABCI()
		if err != nil {
			app.logger.Error("failed to list snapshots", "err", err)
			return resp
		}
		resp.Snapshots = append(resp.Snapshots, &abciSnapshot)
	}

	return resp
}

// LoadSnapshotChunk implements the ABCI interface. It delegates to app.snapshotManager if set.
func (app *BaseApp) LoadSnapshotChunk(req abci.RequestLoadSnapshotChunk) abci.ResponseLoadSnapshotChunk {
	if app.snapshotManager == nil {
		return abci.ResponseLoadSnapshotChunk{}
	}
	chunk, err := app.snapshotManager.LoadChunk(req.Height, req.Format, req.Chunk)
	if err != nil {
		app.logger.Error(
			"failed to load snapshot chunk",
			"height", req.Height,
			"format", req.Format,
			"chunk", req.Chunk,
			"err", err,
		)
		return abci.ResponseLoadSnapshotChunk{}
	}
	return abci.ResponseLoadSnapshotChunk{Chunk: chunk}
}

// OfferSnapshot implements the ABCI interface. It delegates to app.snapshotManager if set.
func (app *BaseApp) OfferSnapshot(req abci.RequestOfferSnapshot) abci.ResponseOfferSnapshot {
	if app.snapshotManager == nil {
		app.logger.Error("snapshot manager not configured")
		return abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}
	}

	if req.Snapshot == nil {
		app.logger.Error("received nil snapshot")
		return abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}
	}

	snapshot, err := snapshottypes.SnapshotFromABCI(req.Snapshot)
	if err != nil {
		app.logger.Error("failed to decode snapshot metadata", "err", err)
		return abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}
	}

	err = app.snapshotManager.Restore(snapshot)
	switch {
	case err == nil:
		return abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}

	case errors.Is(err, snapshottypes.ErrUnknownFormat):
		return abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT_FORMAT}

	case errors.Is(err, snapshottypes.ErrInvalidMetadata):
		app.logger.Error(
			"rejecting invalid snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)
		return abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_REJECT}

	default:
		app.logger.Error(
			"failed to restore snapshot",
			"height", req.Snapshot.Height,
			"format", req.Snapshot.Format,
			"err", err,
		)

		// We currently don't support resetting the IAVL stores and retrying a different snapshot,
		// so we ask Tendermint to abort all snapshot restoration.
		return abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}
	}
}

// ApplySnapshotChunk implements the ABCI interface. It delegates to app.snapshotManager if set.
func (app *BaseApp) ApplySnapshotChunk(req abci.RequestApplySnapshotChunk) abci.ResponseApplySnapshotChunk {
	if app.snapshotManager == nil {
		app.logger.Error("snapshot manager not configured")
		return abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}
	}

	_, err := app.snapshotManager.RestoreChunk(req.Chunk)
	switch {
	case err == nil:
		return abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ACCEPT}

	case errors.Is(err, snapshottypes.ErrChunkHashMismatch):
		app.logger.Error(
			"chunk checksum mismatch; rejecting sender and requesting refetch",
			"chunk", req.Index,
			"sender", req.Sender,
			"err", err,
		)
		return abci.ResponseApplySnapshotChunk{
			Result:        abci.ResponseApplySnapshotChunk_RETRY,
			RefetchChunks: []uint32{req.Index},
			RejectSenders: []string{req.Sender},
		}

	default:
		app.logger.Error("failed to restore snapshot", "err", err)
		return abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}
	}
}

// This comes from ICA, and is made out of an abundance of caution.
// Should be deleted during the next Osmosis major release, as we don't need this extra safety check.
var stateMachineQueryList = []string{"/ibc.applications.transfer.v1.Query/DenomTrace", "/cosmos.auth.v1beta1.Query/Account", "/cosmos.auth.v1beta1.Query/Params", "/cosmos.bank.v1beta1.Query/Balance", "/cosmos.bank.v1beta1.Query/DenomMetadata", "/cosmos.bank.v1beta1.Query/Params", "/cosmos.bank.v1beta1.Query/SupplyOf", "/cosmos.distribution.v1beta1.Query/Params", "/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress", "/cosmos.distribution.v1beta1.Query/ValidatorCommission", "/cosmos.gov.v1beta1.Query/Deposit", "/cosmos.gov.v1beta1.Query/Params", "/cosmos.gov.v1beta1.Query/Vote", "/cosmos.slashing.v1beta1.Query/Params", "/cosmos.slashing.v1beta1.Query/SigningInfo", "/cosmos.staking.v1beta1.Query/Delegation", "/cosmos.staking.v1beta1.Query/Params", "/cosmos.staking.v1beta1.Query/Validator", "/osmosis.epochs.v1beta1.Query/EpochInfos", "/osmosis.epochs.v1beta1.Query/CurrentEpoch", "/osmosis.gamm.v1beta1.Query/NumPools", "/osmosis.gamm.v1beta1.Query/TotalLiquidity", "/osmosis.gamm.v1beta1.Query/Pool", "/osmosis.gamm.v1beta1.Query/PoolParams", "/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity", "/osmosis.gamm.v1beta1.Query/TotalShares", "/osmosis.gamm.v1beta1.Query/CalcJoinPoolShares", "/osmosis.gamm.v1beta1.Query/CalcExitPoolCoinsFromShares", "/osmosis.gamm.v1beta1.Query/CalcJoinPoolNoSwapShares", "/osmosis.gamm.v1beta1.Query/PoolType", "/osmosis.gamm.v2.Query/SpotPrice", "/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn", "/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut", "/osmosis.incentives.Query/ModuleToDistributeCoins", "/osmosis.incentives.Query/LockableDurations", "/osmosis.lockup.Query/ModuleBalance", "/osmosis.lockup.Query/ModuleLockedAmount", "/osmosis.lockup.Query/AccountUnlockableCoins", "/osmosis.lockup.Query/AccountUnlockingCoins", "/osmosis.lockup.Query/LockedDenom", "/osmosis.lockup.Query/LockedByID", "/osmosis.lockup.Query/NextLockID", "/osmosis.mint.v1beta1.Query/EpochProvisions", "/osmosis.mint.v1beta1.Query/Params", "/osmosis.poolincentives.v1beta1.Query/GaugeIds", "/osmosis.superfluid.Query/Params", "/osmosis.superfluid.Query/AssetType", "/osmosis.superfluid.Query/AllAssets", "/osmosis.superfluid.Query/AssetMultiplier", "/osmosis.poolmanager.v1beta1.Query/NumPools", "/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountIn", "/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountOut", "/osmosis.txfees.v1beta1.Query/FeeTokens", "/osmosis.txfees.v1beta1.Query/DenomSpotPrice", "/osmosis.txfees.v1beta1.Query/DenomPoolId", "/osmosis.txfees.v1beta1.Query/BaseDenom", "/osmosis.tokenfactory.v1beta1.Query/Params", "/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata", "/osmosis.twap.v1beta1.Query/ArithmeticTwap", "/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow", "/osmosis.twap.v1beta1.Query/GeometricTwap", "/osmosis.twap.v1beta1.Query/GeometricTwapToNow", "/osmosis.twap.v1beta1.Query/Params", "/osmosis.downtimedetector.v1beta1.Query/RecoveredSinceDowntimeOfLength"}

func queryListContainsReq(req abci.RequestQuery) bool {
	for _, query := range stateMachineQueryList {
		if req.Path == query {
			return true
		}
	}
	return false
}

func (app *BaseApp) handleQueryGRPC(handler GRPCQueryHandler, req abci.RequestQuery) abci.ResponseQuery {
	ctx, err := app.createQueryContext(req.Height, req.Prove)
	if err != nil {
		return sdkerrors.QueryResultWithDebug(err, app.trace)
	}
	// use infinite gas metering for potential state machine queries
	// delete in next major release
	if queryListContainsReq(req) {
		ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	}

	res, err := handler(ctx, req)
	if err != nil {
		res = sdkerrors.QueryResultWithDebug(gRPCErrorToSDKError(err), app.trace)
		res.Height = req.Height
		return res
	}

	return res
}

func gRPCErrorToSDKError(err error) error {
	status, ok := grpcstatus.FromError(err)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	switch status.Code() {
	case codes.NotFound:
		return sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, err.Error())
	case codes.InvalidArgument:
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	case codes.FailedPrecondition:
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	case codes.Unauthenticated:
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, err.Error())
	default:
		return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, err.Error())
	}
}

func checkNegativeHeight(height int64) error {
	if height < 0 {
		// Reject invalid heights.
		return sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"cannot query with height < 0; please provide a valid height",
		)
	}
	return nil
}

// createQueryContext creates a new sdk.Context for a query, taking as args
// the block height and whether the query needs a proof or not.
func (app *BaseApp) createQueryContext(height int64, prove bool) (sdk.Context, error) {
	if err := checkNegativeHeight(height); err != nil {
		return sdk.Context{}, err
	}

	// when a client did not provide a query height, manually inject the latest
	if height == 0 {
		height = app.LastBlockHeight()
	}

	if height <= 1 && prove {
		return sdk.Context{},
			sdkerrors.Wrap(
				sdkerrors.ErrInvalidRequest,
				"cannot query with proof when height <= 1; please provide a valid height",
			)
	}

	lastBlockHeight := app.LastBlockHeight()
	if height > lastBlockHeight {
		return sdk.Context{},
			sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"failed to load state at height %d; (latest height: %d)", height, lastBlockHeight,
			)
	}

	cacheMS, err := app.cms.CacheMultiStoreWithVersion(height)
	if err != nil {
		return sdk.Context{}, fmt.Errorf("failed to load cache multi store for height %d: %w", height, err)
	}

	gasLimit := app.queryGasLimit
	// arbitrarily declare that historical queries cause 2x gas load.
	if height != lastBlockHeight {
		gasLimit /= 2
	}

	// branch the commit-multistore for safety
	ctx := sdk.NewContext(
		cacheMS, app.checkState.ctx.BlockHeader(), true, app.logger,
	).WithMinGasPrices(app.minGasPrices).WithBlockHeight(height).WithGasMeter(sdk.NewGasMeter(gasLimit))

	if height != lastBlockHeight {
		rms, ok := app.cms.(*rootmulti.Store)
		if ok {
			cInfo, err := rms.GetCommitInfoFromDB(height)
			if cInfo != nil && err == nil {
				ctx = ctx.WithBlockTime(cInfo.Timestamp)
			}
		}
	}

	return ctx, nil
}

// GetBlockRetentionHeight returns the height for which all blocks below this height
// are pruned from Tendermint. Given a commitment height and a non-zero local
// minRetainBlocks configuration, the retentionHeight is the smallest height that
// satisfies:
//
// - Unbonding (safety threshold) time: The block interval in which validators
// can be economically punished for misbehavior. Blocks in this interval must be
// auditable e.g. by the light client.
//
// - Logical store snapshot interval: The block interval at which the underlying
// logical store database is persisted to disk, e.g. every 10000 heights. Blocks
// since the last IAVL snapshot must be available for replay on application restart.
//
// - State sync snapshots: Blocks since the oldest available snapshot must be
// available for state sync nodes to catch up (oldest because a node may be
// restoring an old snapshot while a new snapshot was taken).
//
// - Local (minRetainBlocks) config: Archive nodes may want to retain more or
// all blocks, e.g. via a local config option min-retain-blocks. There may also
// be a need to vary retention for other nodes, e.g. sentry nodes which do not
// need historical blocks.
func (app *BaseApp) GetBlockRetentionHeight(commitHeight int64) int64 {
	// pruning is disabled if minRetainBlocks is zero
	if app.minRetainBlocks == 0 {
		return 0
	}

	minNonZero := func(x, y int64) int64 {
		switch {
		case x == 0:
			return y
		case y == 0:
			return x
		case x < y:
			return x
		default:
			return y
		}
	}

	// Define retentionHeight as the minimum value that satisfies all non-zero
	// constraints. All blocks below (commitHeight-retentionHeight) are pruned
	// from Tendermint.
	var retentionHeight int64

	// Define the number of blocks needed to protect against misbehaving validators
	// which allows light clients to operate safely. Note, we piggy back of the
	// evidence parameters instead of computing an estimated nubmer of blocks based
	// on the unbonding period and block commitment time as the two should be
	// equivalent.
	cp := app.GetConsensusParams(app.deliverState.ctx)
	if cp != nil && cp.Evidence != nil && cp.Evidence.MaxAgeNumBlocks > 0 {
		retentionHeight = commitHeight - cp.Evidence.MaxAgeNumBlocks
	}

	if app.snapshotManager != nil {
		snapshotInterval := int64(app.snapshotManager.GetInterval())
		// Define the state pruning offset, i.e. the block offset at which the
		// underlying logical database is persisted to disk.
		if snapshotInterval > 0 {
			if commitHeight > snapshotInterval {
				v := commitHeight - (commitHeight % snapshotInterval)
				retentionHeight = minNonZero(retentionHeight, v)
			} else {
				// Hitting this case means we have persisting enabled but have yet to reach
				// a height in which we persist state, so we return zero regardless of other
				// conditions. Otherwise, we could end up pruning blocks without having
				// any state committed to disk.
				return 0
			}
			snapshotRetentionHeights := app.snapshotManager.GetSnapshotBlockRetentionHeights()
			if snapshotRetentionHeights > 0 {
				retentionHeight = minNonZero(retentionHeight, commitHeight-snapshotRetentionHeights)
			}
		}
	}

	v := commitHeight - int64(app.minRetainBlocks)
	retentionHeight = minNonZero(retentionHeight, v)

	if retentionHeight <= 0 {
		// prune nothing in the case of a non-positive height
		return 0
	}

	return retentionHeight
}

func handleQueryApp(app *BaseApp, path []string, req abci.RequestQuery) abci.ResponseQuery {
	if len(path) >= 2 {
		switch path[1] {
		case "simulate":
			txBytes := req.Data

			gInfo, res, err := app.Simulate(txBytes)
			if err != nil {
				return sdkerrors.QueryResultWithDebug(sdkerrors.Wrap(err, "failed to simulate tx"), app.trace)
			}

			simRes := &sdk.SimulationResponse{
				GasInfo: gInfo,
				Result:  res,
			}

			bz, err := codec.ProtoMarshalJSON(simRes, app.interfaceRegistry)
			if err != nil {
				return sdkerrors.QueryResultWithDebug(sdkerrors.Wrap(err, "failed to JSON encode simulation response"), app.trace)
			}

			return abci.ResponseQuery{
				Codespace: sdkerrors.RootCodespace,
				Height:    req.Height,
				Value:     bz,
			}

		case "version":

			return abci.ResponseQuery{
				Codespace: sdkerrors.RootCodespace,
				Height:    req.Height,
				Value:     []byte(app.version),
			}

		case "snapshots":
			var responseValue []byte

			response := app.ListSnapshots(abci.RequestListSnapshots{})

			responseValue, err := json.Marshal(response)
			if err != nil {
				sdkerrors.QueryResult(sdkerrors.Wrap(err, fmt.Sprintf("failed to marshal list snapshots response %v", response)))
			}

			return abci.ResponseQuery{
				Codespace: sdkerrors.RootCodespace,
				Height:    req.Height,
				Value:     responseValue,
			}

		default:
			return sdkerrors.QueryResultWithDebug(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query: %s", path), app.trace)
		}
	}

	return sdkerrors.QueryResultWithDebug(
		sdkerrors.Wrap(
			sdkerrors.ErrUnknownRequest,
			"expected second parameter to be either 'simulate' or 'version', neither was present",
		), app.trace)
}

func handleQueryStore(app *BaseApp, path []string, req abci.RequestQuery) abci.ResponseQuery {
	// "/store" prefix for store queries
	queryable, ok := app.cms.(sdk.Queryable)
	if !ok {
		return sdkerrors.QueryResultWithDebug(sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "multistore doesn't support queries"), app.trace)
	}

	req.Path = "/" + strings.Join(path[1:], "/")

	if req.Height <= 1 && req.Prove {
		return sdkerrors.QueryResultWithDebug(
			sdkerrors.Wrap(
				sdkerrors.ErrInvalidRequest,
				"cannot query with proof when height <= 1; please provide a valid height",
			), app.trace)
	}

	resp := queryable.Query(req)
	resp.Height = req.Height

	if path[1] == StoreInfoPath {
		commitInfo, err := app.cms.GetCommitInfoFromDB(resp.Height)
		if err != nil {
			return sdkerrors.QueryResult(err)
		}

		// sort the commitInfo, otherwise it returns in a different order every height
		sort.Slice(commitInfo.StoreInfos, func(i, j int) bool {
			return commitInfo.StoreInfos[i].Name < commitInfo.StoreInfos[j].Name
		})

		bz, err := codec.ProtoMarshalJSON(commitInfo, app.interfaceRegistry)
		if err != nil {
			return sdkerrors.QueryResult(sdkerrors.Wrap(err, "failed to JSON encode simulation response"))
		}

		return abci.ResponseQuery{
			Codespace: sdkerrors.RootCodespace,
			Height:    req.Height,
			Value:     bz,
		}
	}
	return resp

}

func handleQueryP2P(app *BaseApp, path []string) abci.ResponseQuery {
	// "/p2p" prefix for p2p queries
	if len(path) < 4 {
		return sdkerrors.QueryResultWithDebug(
			sdkerrors.Wrap(
				sdkerrors.ErrUnknownRequest, "path should be p2p filter <addr|id> <parameter>",
			), app.trace)
	}

	var resp abci.ResponseQuery

	cmd, typ, arg := path[1], path[2], path[3]
	switch cmd {
	case "filter":
		switch typ {
		case "addr":
			resp = app.FilterPeerByAddrPort(arg)

		case "id":
			resp = app.FilterPeerByID(arg)
		}

	default:
		resp = sdkerrors.QueryResultWithDebug(sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "expected second parameter to be 'filter'"), app.trace)
	}

	return resp
}

func handleQueryCustom(app *BaseApp, path []string, req abci.RequestQuery) abci.ResponseQuery {
	// path[0] should be "custom" because "/custom" prefix is required for keeper
	// queries.
	//
	// The QueryRouter routes using path[1]. For example, in the path
	// "custom/gov/proposal", QueryRouter routes using "gov".
	if len(path) < 2 || path[1] == "" {
		return sdkerrors.QueryResultWithDebug(sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "no route for custom query specified"), app.trace)
	}

	querier := app.queryRouter.Route(path[1])
	if querier == nil {
		return sdkerrors.QueryResultWithDebug(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "no custom querier found for route %s", path[1]), app.trace)
	}

	ctx, err := app.createQueryContext(req.Height, req.Prove)
	if err != nil {
		return sdkerrors.QueryResultWithDebug(err, app.trace)
	}

	// Passes the rest of the path as an argument to the querier.
	//
	// For example, in the path "custom/gov/proposal/test", the gov querier gets
	// []string{"proposal", "test"} as the path.
	resBytes, err := querier(ctx, path[2:], req)
	if err != nil {
		res := sdkerrors.QueryResultWithDebug(err, app.trace)
		res.Height = req.Height
		return res
	}

	return abci.ResponseQuery{
		Height: req.Height,
		Value:  resBytes,
	}
}

// splitPath splits a string path using the delimiter '/'.
//
// e.g. "this/is/funny" becomes []string{"this", "is", "funny"}
func splitPath(requestPath string) (path []string) {
	path = strings.Split(requestPath, "/")

	// first element is empty string
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}

	return path
}
