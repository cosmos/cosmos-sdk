package stf

import (
	"context"
	"errors"
	"fmt"

	appmanager "cosmossdk.io/core/app"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	stfgas "cosmossdk.io/server/v2/stf/gas"
)

// STF is a struct that manages the state transition component of the app.
type STF[T transaction.Tx] struct {
	logger      log.Logger
	handleMsg   func(ctx context.Context, msg transaction.Type) (transaction.Type, error)
	handleQuery func(ctx context.Context, req transaction.Type) (transaction.Type, error)

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

// NewSTF returns a new STF instance.
func NewSTF[T transaction.Tx](
	handleMsg func(ctx context.Context, msg transaction.Type) (transaction.Type, error),
	handleQuery func(ctx context.Context, req transaction.Type) (transaction.Type, error),
	doPreBlock func(ctx context.Context, txs []T) error,
	doBeginBlock func(ctx context.Context) error,
	doEndBlock func(ctx context.Context) error,
	doTxValidation func(ctx context.Context, tx T) error,
	doValidatorUpdate func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error),
	postTxExec func(ctx context.Context, tx T, success bool) error,
	branch func(store store.ReaderMap) store.WriterMap,
) *STF[T] {
	return &STF[T]{
		handleMsg:           handleMsg,
		handleQuery:         handleQuery,
		doPreBlock:          doPreBlock,
		doBeginBlock:        doBeginBlock,
		doEndBlock:          doEndBlock,
		doTxValidation:      doTxValidation,
		doValidatorUpdate:   doValidatorUpdate,
		postTxExec:          postTxExec, // TODO
		branchFn:            branch,
		makeGasMeter:        stfgas.DefaultGasMeter,
		makeGasMeteredState: stfgas.DefaultWrapWithGasMeter,
	}
}

// DeliverBlock is our state transition function.
// It takes a read only view of the state to apply the block to,
// executes the block and returns the block results and the new state.
func (s STF[T]) DeliverBlock(
	ctx context.Context,
	block *appmanager.BlockRequest[T],
	state store.ReaderMap,
) (blockResult *appmanager.BlockResponse, newState store.WriterMap, err error) {
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

	exCtx := s.makeContext(ctx, appmanager.ConsensusIdentity, newState, corecontext.ExecModeFinalize)
	exCtx.setHeaderInfo(hi)
	consMessagesResponses, err := s.runConsensusMessages(exCtx, block.ConsensusMessages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute consensus messages: %w", err)
	}

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
	beginBlockEvents, err := s.beginBlock(exCtx)
	if err != nil {
		return nil, nil, err
	}

	// check if we need to return early
	if err = isCtxCancelled(ctx); err != nil {
		return nil, nil, err
	}

	// execute txs
	txResults := make([]appmanager.TxResult, len(block.Txs))
	// TODO: skip first tx if vote extensions are enabled (marko)
	for i, txBytes := range block.Txs {
		// check if we need to return early or continue delivering txs
		if err = isCtxCancelled(ctx); err != nil {
			return nil, nil, err
		}
		txResults[i] = s.deliverTx(ctx, newState, txBytes, corecontext.ExecModeFinalize, hi)
	}
	// reset events
	exCtx.events = make([]event.Event, 0)
	// end block
	endBlockEvents, valset, err := s.endBlock(exCtx)
	if err != nil {
		return nil, nil, err
	}

	return &appmanager.BlockResponse{
		Apphash:                   nil,
		ConsensusMessagesResponse: consMessagesResponses,
		ValidatorUpdates:          valset,
		PreBlockEvents:            preBlockEvents,
		BeginBlockEvents:          beginBlockEvents,
		TxResults:                 txResults,
		EndBlockEvents:            endBlockEvents,
	}, newState, nil
}

// deliverTx executes a TX and returns the result.
func (s STF[T]) deliverTx(
	ctx context.Context,
	state store.WriterMap,
	tx T,
	execMode corecontext.ExecMode,
	hi header.Info,
) appmanager.TxResult {
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
		return appmanager.TxResult{
			Error: gasLimitErr,
		}
	}

	if recoveryError != nil {
		return appmanager.TxResult{
			Error: recoveryError,
		}
	}

	validateGas, validationEvents, err := s.validateTx(ctx, state, gasLimit, tx)
	if err != nil {
		return appmanager.TxResult{
			Error: err,
		}
	}

	execResp, execGas, execEvents, err := s.execTx(ctx, state, gasLimit-validateGas, tx, execMode, hi)
	return appmanager.TxResult{
		Events:    append(validationEvents, execEvents...),
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
) (gasUsed uint64, events []event.Event, err error) {
	validateState := s.branchFn(state)
	hi, err := s.getHeaderInfo(validateState)
	if err != nil {
		return 0, nil, err
	}
	validateCtx := s.makeContext(ctx, appmanager.RuntimeIdentity, validateState, corecontext.ExecModeCheck)
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
	execMode corecontext.ExecMode,
	hi header.Info,
) ([]transaction.Type, uint64, []event.Event, error) {
	execState := s.branchFn(state)

	msgsResp, gasUsed, runTxMsgsEvents, txErr := s.runTxMsgs(ctx, execState, gasLimit, tx, execMode, hi)
	if txErr != nil {
		// in case of error during message execution, we do not apply the exec state.
		// instead we run the post exec handler in a new branchFn from the initial state.
		postTxState := s.branchFn(state)
		postTxCtx := s.makeContext(ctx, appmanager.RuntimeIdentity, postTxState, execMode)
		postTxCtx.setHeaderInfo(hi)

		// TODO: runtime sets a noop posttxexec if the app doesnt set anything (julien)

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
		return nil, gasUsed, postTxCtx.events, txErr
	}
	// tx execution went fine, now we use the same state to run the post tx exec handler,
	// in case the execution of the post tx fails, then no state change is applied and the
	// whole execution step is rolled back.
	postTxCtx := s.makeContext(ctx, appmanager.RuntimeIdentity, execState, execMode) // NO gas limit.
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

	return msgsResp, gasUsed, append(runTxMsgsEvents, postTxCtx.events...), nil
}

// runTxMsgs will execute the messages contained in the TX with the provided state.
func (s STF[T]) runTxMsgs(
	ctx context.Context,
	state store.WriterMap,
	gasLimit uint64,
	tx T,
	execMode corecontext.ExecMode,
	hi header.Info,
) ([]transaction.Type, uint64, []event.Event, error) {
	txSenders, err := tx.GetSenders()
	if err != nil {
		return nil, 0, nil, err
	}
	msgs, err := tx.GetMessages()
	if err != nil {
		return nil, 0, nil, err
	}
	msgResps := make([]transaction.Type, len(msgs))

	execCtx := s.makeContext(ctx, nil, state, execMode)
	execCtx.setHeaderInfo(hi)
	execCtx.setGasLimit(gasLimit)
	for i, msg := range msgs {
		execCtx.sender = txSenders[i]
		resp, err := s.handleMsg(execCtx, msg)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("message execution at index %d failed: %w", i, err)
		}
		msgResps[i] = resp
	}

	consumed := execCtx.meter.Limit() - execCtx.meter.Remaining()
	return msgResps, consumed, execCtx.events, nil
}

func (s STF[T]) preBlock(
	ctx *executionContext,
	txs []T,
) ([]event.Event, error) {
	err := s.doPreBlock(ctx, txs)
	if err != nil {
		return nil, err
	}

	for i, e := range ctx.events {
		ctx.events[i].Attributes = append(
			e.Attributes,
			event.Attribute{Key: "mode", Value: "PreBlock"},
		)
	}

	return ctx.events, nil
}

func (s STF[T]) runConsensusMessages(
	ctx *executionContext,
	messages []transaction.Type,
) ([]transaction.Type, error) {
	responses := make([]transaction.Type, len(messages))
	for i := range messages {
		resp, err := s.handleMsg(ctx, messages[i])
		if err != nil {
			return nil, err
		}
		responses[i] = resp
	}

	return responses, nil
}

func (s STF[T]) beginBlock(
	ctx *executionContext,
) (beginBlockEvents []event.Event, err error) {
	err = s.doBeginBlock(ctx)
	if err != nil {
		return nil, err
	}

	for i, e := range ctx.events {
		ctx.events[i].Attributes = append(
			e.Attributes,
			event.Attribute{Key: "mode", Value: "BeginBlock"},
		)
	}

	return ctx.events, nil
}

func (s STF[T]) endBlock(
	ctx *executionContext,
) ([]event.Event, []appmodulev2.ValidatorUpdate, error) {
	err := s.doEndBlock(ctx)
	if err != nil {
		return nil, nil, err
	}

	events, valsetUpdates, err := s.validatorUpdates(ctx)
	if err != nil {
		return nil, nil, err
	}

	ctx.events = append(ctx.events, events...)

	for i, e := range ctx.events {
		ctx.events[i].Attributes = append(
			e.Attributes,
			event.Attribute{Key: "mode", Value: "BeginBlock"},
		)
	}

	return ctx.events, valsetUpdates, nil
}

// validatorUpdates returns the validator updates for the current block. It is called by endBlock after the endblock execution has concluded
func (s STF[T]) validatorUpdates(
	ctx *executionContext,
) ([]event.Event, []appmodulev2.ValidatorUpdate, error) {
	valSetUpdates, err := s.doValidatorUpdate(ctx)
	if err != nil {
		return nil, nil, err
	}
	return ctx.events, valSetUpdates, nil
}

const headerInfoPrefix = 0x0

// setHeaderInfo sets the header info in the state to be used by queries in the future.
func (s STF[T]) setHeaderInfo(state store.WriterMap, headerInfo header.Info) error {
	runtimeStore, err := state.GetWriter(appmanager.RuntimeIdentity)
	if err != nil {
		return err
	}
	bz, err := headerInfo.Bytes()
	if err != nil {
		return err
	}
	err = runtimeStore.Set([]byte{headerInfoPrefix}, bz)
	if err != nil {
		return err
	}
	return nil
}

// getHeaderInfo gets the header info from the state. It should only be used for queries
func (s STF[T]) getHeaderInfo(state store.WriterMap) (i header.Info, err error) {
	runtimeStore, err := state.GetWriter(appmanager.RuntimeIdentity)
	if err != nil {
		return header.Info{}, err
	}
	v, err := runtimeStore.Get([]byte{headerInfoPrefix})
	if err != nil {
		return header.Info{}, err
	}
	if v == nil {
		return header.Info{}, nil
	}

	err = i.FromBytes(v)
	return i, err
}

// Simulate simulates the execution of a tx on the provided state.
func (s STF[T]) Simulate(
	ctx context.Context,
	state store.ReaderMap,
	gasLimit uint64,
	tx T,
) (appmanager.TxResult, store.WriterMap) {
	simulationState := s.branchFn(state)
	hi, err := s.getHeaderInfo(simulationState)
	if err != nil {
		return appmanager.TxResult{}, nil
	}
	txr := s.deliverTx(ctx, simulationState, tx, corecontext.ExecModeSimulate, hi)

	return txr, simulationState
}

// ValidateTx will run only the validation steps required for a transaction.
// Validations are run over the provided state, with the provided gas limit.
func (s STF[T]) ValidateTx(
	ctx context.Context,
	state store.ReaderMap,
	gasLimit uint64,
	tx T,
) appmanager.TxResult {
	validationState := s.branchFn(state)
	gasUsed, events, err := s.validateTx(ctx, validationState, gasLimit, tx)
	return appmanager.TxResult{
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
	req transaction.Type,
) (transaction.Type, error) {
	queryState := s.branchFn(state)
	hi, err := s.getHeaderInfo(queryState)
	if err != nil {
		return nil, err
	}
	queryCtx := s.makeContext(ctx, nil, queryState, corecontext.ExecModeSimulate)
	queryCtx.setHeaderInfo(hi)
	queryCtx.setGasLimit(gasLimit)
	return s.handleQuery(queryCtx, req)
}

func (s STF[T]) Message(ctx context.Context, msg transaction.Type) (response transaction.Type, err error) {
	return s.handleMsg(ctx, msg)
}

// RunWithCtx is made to support genesis, if genesis was just the execution of messages instead
// of being something custom then we would not need this. PLEASE DO NOT USE.
// TODO: Remove
func (s STF[T]) RunWithCtx(
	ctx context.Context,
	state store.ReaderMap,
	closure func(ctx context.Context) error,
) (store.WriterMap, error) {
	branchedState := s.branchFn(state)
	// TODO  do we need headerinfo for genesis?
	stfCtx := s.makeContext(ctx, nil, branchedState, corecontext.ExecModeFinalize)
	return branchedState, closure(stfCtx)
}

// clone clones STF.
func (s STF[T]) clone() STF[T] {
	return STF[T]{
		handleMsg:           s.handleMsg,
		handleQuery:         s.handleQuery,
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
	execMode corecontext.ExecMode

	branchFn            branchFn
	makeGasMeter        makeGasMeterFn
	makeGasMeteredStore makeGasMeteredStateFn
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
	execMode corecontext.ExecMode,
) *executionContext {
	return newExecutionContext(
		s.makeGasMeter,
		s.makeGasMeteredState,
		s.branchFn,
		ctx,
		sender,
		store,
		execMode,
	)
}

func newExecutionContext(
	makeGasMeterFn makeGasMeterFn,
	makeGasMeteredStoreFn makeGasMeteredStateFn,
	branchFn branchFn,
	ctx context.Context,
	sender transaction.Identity,
	state store.WriterMap,
	execMode corecontext.ExecMode,
) *executionContext {
	meter := makeGasMeterFn(gas.NoGasLimit)
	meteredState := makeGasMeteredStoreFn(meter, state)

	return &executionContext{
		Context:             ctx,
		unmeteredState:      state,
		state:               meteredState,
		meter:               meter,
		events:              make([]event.Event, 0),
		sender:              sender,
		headerInfo:          header.Info{},
		execMode:            execMode,
		branchFn:            branchFn,
		makeGasMeter:        makeGasMeterFn,
		makeGasMeteredStore: makeGasMeteredStoreFn,
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
