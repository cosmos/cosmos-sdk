package stf

import (
	"context"
	"errors"
	"fmt"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/log"
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/schema/appdata"
	stfgas "cosmossdk.io/server/v2/stf/gas"
	"cosmossdk.io/server/v2/stf/internal"
)

type eContextKey struct{}

var executionContextKey = eContextKey{}

// STF is a struct that manages the state transition component of the app.
type STF[T transaction.Tx] struct {
	logger log.Logger

	msgRouter   coreRouterImpl
	queryRouter coreRouterImpl

	doPreBlock        func(ctx context.Context, txs []T) error
	doBeginBlock      func(ctx context.Context) error
	doEndBlock        func(ctx context.Context) error
	doValidatorUpdate func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error)

	doTxValidation func(ctx context.Context, tx T) error
	postTxExec     func(ctx context.Context, tx T, success bool) error

	branchFn            branchFn // branchFn is a function that given a readonly state it returns a writable version of it.
	makeGasMeter        makeGasMeterFn
	makeGasMeteredState makeGasMeteredStateFn
}

// New returns a new STF instance.
func New[T transaction.Tx](
	logger log.Logger,
	msgRouterBuilder *MsgRouterBuilder,
	queryRouterBuilder *MsgRouterBuilder,
	doPreBlock func(ctx context.Context, txs []T) error,
	doBeginBlock func(ctx context.Context) error,
	doEndBlock func(ctx context.Context) error,
	doTxValidation func(ctx context.Context, tx T) error,
	doValidatorUpdate func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error),
	postTxExec func(ctx context.Context, tx T, success bool) error,
	branch func(store store.ReaderMap) store.WriterMap,
) (*STF[T], error) {
	msgRouter, err := msgRouterBuilder.build()
	if err != nil {
		return nil, fmt.Errorf("build msg router: %w", err)
	}
	queryRouter, err := queryRouterBuilder.build()
	if err != nil {
		return nil, fmt.Errorf("build query router: %w", err)
	}

	return &STF[T]{
		logger:              logger,
		msgRouter:           msgRouter,
		queryRouter:         queryRouter,
		doPreBlock:          doPreBlock,
		doBeginBlock:        doBeginBlock,
		doEndBlock:          doEndBlock,
		doValidatorUpdate:   doValidatorUpdate,
		doTxValidation:      doTxValidation,
		postTxExec:          postTxExec, // TODO
		branchFn:            branch,
		makeGasMeter:        stfgas.DefaultGasMeter,
		makeGasMeteredState: stfgas.DefaultWrapWithGasMeter,
	}, nil
}

// DeliverBlock is our state transition function.
// It takes a read only view of the state to apply the block to,
// executes the block and returns the block results and the new state.
func (s STF[T]) DeliverBlock(
	ctx context.Context,
	block *server.BlockRequest[T],
	state store.ReaderMap,
) (blockResult *server.BlockResponse, newState store.WriterMap, err error) {
	// creates a new branchFn state, from the readonly view of the state
	// that can be written to.
	newState = s.branchFn(state)
	hi := header.Info{
		Hash:    block.Hash,
		AppHash: block.AppHash,
		ChainID: block.ChainId,
		Time:    block.Time,
		Height:  int64(block.Height),
	}
	// set header info
	err = s.setHeaderInfo(newState, hi)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to set initial header info, %w", err)
	}

	exCtx := s.makeContext(ctx, ConsensusIdentity, newState, internal.ExecModeFinalize)
	exCtx.setHeaderInfo(hi)

	// reset events
	exCtx.events = make([]event.Event, 0)
	// pre block is called separate from begin block in order to prepopulate state
	preBlockEvents, err := s.preBlock(exCtx, block.Txs)
	if err != nil {
		return nil, nil, err
	}

	if err = isCtxCancelled(ctx); err != nil {
		return nil, nil, err
	}

	// reset events
	exCtx.events = make([]event.Event, 0)
	// begin block
	var beginBlockEvents []event.Event
	if !block.IsGenesis {
		// begin block
		beginBlockEvents, err = s.beginBlock(exCtx)
		if err != nil {
			return nil, nil, err
		}
	}

	// check if we need to return early
	if err = isCtxCancelled(ctx); err != nil {
		return nil, nil, err
	}

	// execute txs
	txResults := make([]server.TxResult, len(block.Txs))
	// TODO: skip first tx if vote extensions are enabled (marko)
	for i, txBytes := range block.Txs {
		// check if we need to return early or continue delivering txs
		if err = isCtxCancelled(ctx); err != nil {
			return nil, nil, err
		}
		txResults[i] = s.deliverTx(exCtx, newState, txBytes, transaction.ExecModeFinalize, hi, int32(i+1))
	}
	// reset events
	exCtx.events = make([]event.Event, 0)
	// end block
	endBlockEvents, valset, err := s.endBlock(exCtx)
	if err != nil {
		return nil, nil, err
	}

	return &server.BlockResponse{
		ValidatorUpdates: valset,
		PreBlockEvents:   preBlockEvents,
		BeginBlockEvents: beginBlockEvents,
		TxResults:        txResults,
		EndBlockEvents:   endBlockEvents,
	}, newState, nil
}

// deliverTx executes a TX and returns the result.
func (s STF[T]) deliverTx(
	ctx context.Context,
	state store.WriterMap,
	tx T,
	execMode transaction.ExecMode,
	hi header.Info,
	txIndex int32,
) server.TxResult {
	// recover in the case of a panic
	var recoveryError error
	defer func() {
		if r := recover(); r != nil {
			recoveryError = fmt.Errorf("panic during transaction execution: %s", r)
			s.logger.Error("panic during transaction execution", "error", recoveryError)
		}
	}()
	// handle error from GetGasLimit
	gasLimit, gasLimitErr := tx.GetGasLimit()
	if gasLimitErr != nil {
		return server.TxResult{
			Error: gasLimitErr,
		}
	}

	if recoveryError != nil {
		return server.TxResult{
			Error: recoveryError,
		}
	}
	validateGas, validationEvents, err := s.validateTx(ctx, state, gasLimit, tx, execMode)
	if err != nil {
		return server.TxResult{
			Error: err,
		}
	}
	events := make([]event.Event, 0)
	// set the event indexes, set MsgIndex to 0 in validation events
	for i, e := range validationEvents {
		e.BlockStage = appdata.TxProcessingStage
		e.TxIndex = txIndex
		e.MsgIndex = 0
		e.EventIndex = int32(i + 1)
		events = append(events, e)
	}

	execResp, execGas, execEvents, err := s.execTx(ctx, state, gasLimit-validateGas, tx, execMode, hi)
	// set the TxIndex in the exec events
	for _, e := range execEvents {
		e.BlockStage = appdata.TxProcessingStage
		e.TxIndex = txIndex
		events = append(events, e)
	}

	return server.TxResult{
		Events:    events,
		GasUsed:   execGas + validateGas,
		GasWanted: gasLimit,
		Resp:      execResp,
		Error:     err,
	}
}

// validateTx validates a transaction given the provided WritableState and gas limit.
// If the validation is successful, state is committed
func (s STF[T]) validateTx(
	ctx context.Context,
	state store.WriterMap,
	gasLimit uint64,
	tx T,
	execMode transaction.ExecMode,
) (gasUsed uint64, events []event.Event, err error) {
	validateState := s.branchFn(state)
	hi, err := s.getHeaderInfo(validateState)
	if err != nil {
		return 0, nil, err
	}
	validateCtx := s.makeContext(ctx, RuntimeIdentity, validateState, execMode)
	validateCtx.setHeaderInfo(hi)
	validateCtx.setGasLimit(gasLimit)
	err = s.doTxValidation(validateCtx, tx)
	if err != nil {
		return 0, nil, err
	}

	consumed := validateCtx.meter.Limit() - validateCtx.meter.Remaining()

	return consumed, validateCtx.events, applyStateChanges(state, validateState)
}

// execTx executes the tx messages on the provided state. If the tx fails then the state is discarded.
func (s STF[T]) execTx(
	ctx context.Context,
	state store.WriterMap,
	gasLimit uint64,
	tx T,
	execMode transaction.ExecMode,
	hi header.Info,
) ([]transaction.Msg, uint64, []event.Event, error) {
	execState := s.branchFn(state)

	msgsResp, gasUsed, runTxMsgsEvents, txErr := s.runTxMsgs(ctx, execState, gasLimit, tx, execMode, hi)
	if txErr != nil {
		// in case of error during message execution, we do not apply the exec state.
		// instead we run the post exec handler in a new branchFn from the initial state.
		postTxState := s.branchFn(state)
		postTxCtx := s.makeContext(ctx, RuntimeIdentity, postTxState, execMode)
		postTxCtx.setHeaderInfo(hi)

		postTxErr := s.postTxExec(postTxCtx, tx, false)
		if postTxErr != nil {
			// if the post tx handler fails, then we do not apply any state change to the initial state.
			// we just return the exec gas used and a joined error from TX error and post TX error.
			return nil, gasUsed, nil, errors.Join(txErr, postTxErr)
		}
		// in case post tx is successful, then we commit the post tx state to the initial state,
		// and we return post tx events alongside exec gas used and the error of the tx.
		applyErr := applyStateChanges(state, postTxState)
		if applyErr != nil {
			return nil, 0, nil, applyErr
		}
		// set the event indexes, set MsgIndex to -1 in post tx events
		for i := range postTxCtx.events {
			postTxCtx.events[i].EventIndex = int32(i + 1)
			postTxCtx.events[i].MsgIndex = -1
		}

		return nil, gasUsed, postTxCtx.events, txErr
	}
	// tx execution went fine, now we use the same state to run the post tx exec handler,
	// in case the execution of the post tx fails, then no state change is applied and the
	// whole execution step is rolled back.
	postTxCtx := s.makeContext(ctx, RuntimeIdentity, execState, execMode) // NO gas limit.
	postTxCtx.setHeaderInfo(hi)
	postTxErr := s.postTxExec(postTxCtx, tx, true)
	if postTxErr != nil {
		// if post tx fails, then we do not apply any state change, we return the post tx error,
		// alongside the gas used.
		return nil, gasUsed, nil, postTxErr
	}
	// both the execution and post tx execution step were successful, so we apply the state changes
	// to the provided state, and we return responses, and events from exec tx and post tx exec.
	applyErr := applyStateChanges(state, execState)
	if applyErr != nil {
		return nil, 0, nil, applyErr
	}
	// set the event indexes, set MsgIndex to -1 in post tx events
	for i := range postTxCtx.events {
		postTxCtx.events[i].EventIndex = int32(i + 1)
		postTxCtx.events[i].MsgIndex = -1
	}

	return msgsResp, gasUsed, append(runTxMsgsEvents, postTxCtx.events...), nil
}

// runTxMsgs will execute the messages contained in the TX with the provided state.
func (s STF[T]) runTxMsgs(
	ctx context.Context,
	state store.WriterMap,
	gasLimit uint64,
	tx T,
	execMode transaction.ExecMode,
	hi header.Info,
) ([]transaction.Msg, uint64, []event.Event, error) {
	txSenders, err := tx.GetSenders()
	if err != nil {
		return nil, 0, nil, err
	}
	msgs, err := tx.GetMessages()
	if err != nil {
		return nil, 0, nil, err
	}
	msgResps := make([]transaction.Msg, len(msgs))

	execCtx := s.makeContext(ctx, RuntimeIdentity, state, execMode)
	execCtx.setHeaderInfo(hi)
	execCtx.setGasLimit(gasLimit)
	events := make([]event.Event, 0)
	for i, msg := range msgs {
		execCtx.sender = txSenders[i]
		execCtx.events = make([]event.Event, 0) // reset events
		resp, err := s.msgRouter.Invoke(execCtx, msg)
		if err != nil {
			return nil, 0, nil, err // do not wrap the error or we lose the original error type
		}
		msgResps[i] = resp
		for j, e := range execCtx.events {
			e.MsgIndex = int32(i + 1)
			e.EventIndex = int32(j + 1)
			events = append(events, e)
		}
	}

	consumed := execCtx.meter.Limit() - execCtx.meter.Remaining()
	return msgResps, consumed, events, nil
}

// preBlock executes the pre block logic.
func (s STF[T]) preBlock(
	ctx *executionContext,
	txs []T,
) ([]event.Event, error) {
	err := s.doPreBlock(ctx, txs)
	if err != nil {
		return nil, err
	}

	for i := range ctx.events {
		ctx.events[i].BlockStage = appdata.PreBlockStage
		ctx.events[i].EventIndex = int32(i + 1)
	}

	return ctx.events, nil
}

// beginBlock executes the begin block logic.
func (s STF[T]) beginBlock(
	ctx *executionContext,
) (beginBlockEvents []event.Event, err error) {
	err = s.doBeginBlock(ctx)
	if err != nil {
		return nil, err
	}

	for i := range ctx.events {
		ctx.events[i].BlockStage = appdata.BeginBlockStage
		ctx.events[i].EventIndex = int32(i + 1)
	}

	return ctx.events, nil
}

// endBlock executes the end block logic.
func (s STF[T]) endBlock(
	ctx *executionContext,
) ([]event.Event, []appmodulev2.ValidatorUpdate, error) {
	err := s.doEndBlock(ctx)
	if err != nil {
		return nil, nil, err
	}
	events := ctx.events
	ctx.events = make([]event.Event, 0) // reset events
	valsetUpdates, err := s.validatorUpdates(ctx)
	if err != nil {
		return nil, nil, err
	}
	events = append(events, ctx.events...)
	for i := range events {
		events[i].BlockStage = appdata.EndBlockStage
		events[i].EventIndex = int32(i + 1)
	}

	return events, valsetUpdates, nil
}

// validatorUpdates returns the validator updates for the current block. It is called by endBlock after the endblock execution has concluded
func (s STF[T]) validatorUpdates(
	ctx *executionContext,
) ([]appmodulev2.ValidatorUpdate, error) {
	valSetUpdates, err := s.doValidatorUpdate(ctx)
	if err != nil {
		return nil, err
	}
	return valSetUpdates, nil
}

// Simulate simulates the execution of a tx on the provided state.
func (s STF[T]) Simulate(
	ctx context.Context,
	state store.ReaderMap,
	gasLimit uint64,
	tx T,
) (server.TxResult, store.WriterMap) {
	simulationState := s.branchFn(state)
	hi, err := s.getHeaderInfo(simulationState)
	if err != nil {
		return server.TxResult{}, nil
	}
	txr := s.deliverTx(ctx, simulationState, tx, internal.ExecModeSimulate, hi, 0)

	return txr, simulationState
}

// ValidateTx will run only the validation steps required for a transaction.
// Validations are run over the provided state, with the provided gas limit.
func (s STF[T]) ValidateTx(
	ctx context.Context,
	state store.ReaderMap,
	gasLimit uint64,
	tx T,
) server.TxResult {
	validationState := s.branchFn(state)
	gasUsed, events, err := s.validateTx(ctx, validationState, gasLimit, tx, transaction.ExecModeCheck)
	return server.TxResult{
		Events:  events,
		GasUsed: gasUsed,
		Error:   err,
	}
}

// Query executes the query on the provided state with the provided gas limits.
func (s STF[T]) Query(
	ctx context.Context,
	state store.ReaderMap,
	gasLimit uint64,
	req transaction.Msg,
) (transaction.Msg, error) {
	queryState := s.branchFn(state)
	hi, err := s.getHeaderInfo(queryState)
	if err != nil {
		return nil, err
	}
	queryCtx := s.makeContext(ctx, nil, queryState, internal.ExecModeSimulate)
	queryCtx.setHeaderInfo(hi)
	queryCtx.setGasLimit(gasLimit)
	return s.queryRouter.Invoke(queryCtx, req)
}

// clone clones STF.
func (s STF[T]) clone() STF[T] {
	return STF[T]{
		logger:              s.logger,
		msgRouter:           s.msgRouter,
		queryRouter:         s.queryRouter,
		doPreBlock:          s.doPreBlock,
		doBeginBlock:        s.doBeginBlock,
		doEndBlock:          s.doEndBlock,
		doValidatorUpdate:   s.doValidatorUpdate,
		doTxValidation:      s.doTxValidation,
		postTxExec:          s.postTxExec,
		branchFn:            s.branchFn,
		makeGasMeter:        s.makeGasMeter,
		makeGasMeteredState: s.makeGasMeteredState,
	}
}

// executionContext is a struct that holds the context for the execution of a tx.
type executionContext struct {
	context.Context

	// unmeteredState is storage without metering. Changes here are propagated to state which is the metered
	// version.
	unmeteredState store.WriterMap
	// state is the gas metered state.
	state store.WriterMap
	// meter is the gas meter.
	meter gas.Meter
	// events are the current events.
	events []event.Event
	// sender is the causer of the state transition.
	sender transaction.Identity
	// headerInfo contains the block info.
	headerInfo header.Info
	// execMode retains information about the exec mode.
	execMode transaction.ExecMode

	branchFn            branchFn
	makeGasMeter        makeGasMeterFn
	makeGasMeteredStore makeGasMeteredStateFn

	msgRouter   router.Service
	queryRouter router.Service
}

// setHeaderInfo sets the header info in the state to be used by queries in the future.
func (e *executionContext) setHeaderInfo(hi header.Info) {
	e.headerInfo = hi
}

// setGasLimit will update the gas limit of the *executionContext
func (e *executionContext) setGasLimit(limit uint64) {
	meter := e.makeGasMeter(limit)
	meteredState := e.makeGasMeteredStore(meter, e.unmeteredState)

	e.meter = meter
	e.state = meteredState
}

func (e *executionContext) Value(key any) any {
	if key == executionContextKey {
		return e
	}

	return e.Context.Value(key)
}

// TODO: too many calls to makeContext can be expensive
// makeContext creates and returns a new execution context for the STF[T] type.
// It takes in the following parameters:
// - ctx: The context.Context object for the execution.
// - sender: The transaction.Identity object representing the sender of the transaction.
// - state: The store.WriterMap object for accessing and modifying the state.
// - gasLimit: The maximum amount of gas allowed for the execution.
// - execMode: The corecontext.ExecMode object representing the execution mode.
//
// It returns a pointer to the executionContext struct
func (s STF[T]) makeContext(
	ctx context.Context,
	sender transaction.Identity,
	store store.WriterMap,
	execMode transaction.ExecMode,
) *executionContext {
	valuedCtx := context.WithValue(ctx, corecontext.ExecModeKey, execMode)
	return newExecutionContext(
		valuedCtx,
		s.makeGasMeter,
		s.makeGasMeteredState,
		s.branchFn,
		sender,
		store,
		execMode,
		s.msgRouter,
		s.queryRouter,
	)
}

func newExecutionContext(
	ctx context.Context,
	makeGasMeterFn makeGasMeterFn,
	makeGasMeteredStoreFn makeGasMeteredStateFn,
	branchFn branchFn,
	sender transaction.Identity,
	state store.WriterMap,
	execMode transaction.ExecMode,
	msgRouter coreRouterImpl,
	queryRouter coreRouterImpl,
) *executionContext {
	meter := makeGasMeterFn(gas.NoGasLimit)
	meteredState := makeGasMeteredStoreFn(meter, state)

	return &executionContext{
		Context:             ctx,
		unmeteredState:      state,
		state:               meteredState,
		meter:               meter,
		events:              make([]event.Event, 0),
		headerInfo:          header.Info{},
		execMode:            execMode,
		sender:              sender,
		branchFn:            branchFn,
		makeGasMeter:        makeGasMeterFn,
		makeGasMeteredStore: makeGasMeteredStoreFn,
		msgRouter:           msgRouter,
		queryRouter:         queryRouter,
	}
}

// applyStateChanges applies the state changes from the source store to the destination store.
// It retrieves the state changes from the source store using GetStateChanges method,
// and then applies those changes to the destination store using ApplyStateChanges method.
// If an error occurs during the retrieval or application of state changes, it is returned.
func applyStateChanges(dst, src store.WriterMap) error {
	changes, err := src.GetStateChanges()
	if err != nil {
		return err
	}
	return dst.ApplyStateChanges(changes)
}

// isCtxCancelled reports if the context was canceled.
func isCtxCancelled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
