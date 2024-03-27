package stf

import (
	"context"
	"errors"
	"fmt"

	appmanager "cosmossdk.io/core/app"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/branch"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	stfgas "cosmossdk.io/server/v2/stf/gas"
)

var _ STF[transaction.Tx] = STF[transaction.Tx]{} // Ensure STF implements STFI.

// STFI defines the state transition handler used by AppManager to execute
// state transitions over some state. STF never writes to state, instead
// returns the state changes caused by the state transitions.
type STFI[T transaction.Tx] interface {
	// DeliverBlock is used to process an entire block, given a state to apply the state transition to.
	// Returns the state changes of the transition.
	DeliverBlock(
		ctx context.Context,
		block *appmanager.BlockRequest[T],
		state store.ReaderMap,
	) (*appmanager.BlockResponse, store.WriterMap, error)
	// Simulate simulates the execution of a transaction over the provided state, with the provided gas limit.
	Simulate(ctx context.Context, state store.ReaderMap, gasLimit uint64, tx T) (appmanager.TxResult, store.WriterMap)
	// Query runs the provided query over the provided readonly state.
	Query(ctx context.Context, state store.ReaderMap, gasLimit uint64, queryRequest transaction.Type) (queryResponse transaction.Type, err error)
	// ValidateTx validates the TX.
	ValidateTx(ctx context.Context, state store.ReaderMap, gasLimit uint64, tx T, hs header.Service) appmanager.TxResult
}

// STF is a struct that manages the state transition component of the app.
type STF[T transaction.Tx] struct {
	handleMsg   func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error)
	handleQuery func(ctx context.Context, req transaction.Type) (resp transaction.Type, err error)

	doPreBlock        func(ctx context.Context, txs []T) error
	doBeginBlock      func(ctx context.Context) error
	doEndBlock        func(ctx context.Context) error
	doValidatorUpdate func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error)

	doTxValidation func(ctx context.Context, tx T) error
	postTxExec     func(ctx context.Context, tx T, success bool) error

	branch           branchdb // branch is a function that given a readonly state it returns a writable version of it.
	getGasMeter      func(gasLimit uint64) gas.Meter
	wrapWithGasMeter func(meter gas.Meter, store store.WriterMap) store.WriterMap
}

// NewSTF returns a new STF instance.
func NewSTF[T transaction.Tx](
	handleMsg func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error),
	handleQuery func(ctx context.Context, req transaction.Type) (resp transaction.Type, err error),
	doPreBlock func(ctx context.Context, txs []T) error,
	doBeginBlock func(ctx context.Context) error,
	doEndBlock func(ctx context.Context) error,
	doTxValidation func(ctx context.Context, tx T) error,
	doValidatorUpdate func(ctx context.Context) ([]appmodulev2.ValidatorUpdate, error),
	postTxExec func(ctx context.Context, tx T, success bool) error,
	branch func(store store.ReaderMap) store.WriterMap,
) *STF[T] {
	return &STF[T]{
		handleMsg:         handleMsg,
		handleQuery:       handleQuery,
		doPreBlock:        doPreBlock,
		doBeginBlock:      doBeginBlock,
		doEndBlock:        doEndBlock,
		doTxValidation:    doTxValidation,
		doValidatorUpdate: doValidatorUpdate,
		postTxExec:        postTxExec, // TODO
		branch:            branch,
		getGasMeter:       stfgas.DefaultGetMeter,
		wrapWithGasMeter:  stfgas.DefaultWrapWithGasMeter,
	}
}

// DeliverBlock is our state transition function.
// It takes a read only view of the state to apply the block to,
// executes the block and returns the block results and the new state.
func (s STF[T]) DeliverBlock(ctx context.Context, block *appmanager.BlockRequest[T], state store.ReaderMap) (blockResult *appmanager.BlockResponse, newState store.WriterMap, err error) {
	// creates a new branch state, from the readonly view of the state
	// that can be written to.
	newState = s.branch(state)
	ctx = s.contextInitCache(ctx)

	for _, msg := range block.ConsensusMessages {
		_, err := s.handleMsg(ctx, msg)
		if err != nil {
			return nil, nil, err
		}
	}

	// pre block is called separate from begin block in order to prepopulate state
	preBlockEvents, err := s.preBlock(ctx, newState, block.Txs)
	if err != nil {
		return nil, nil, err
	}

	if err = isCtxCancelled(ctx); err != nil {
		return nil, nil, err
	}

	// begin block
	beginBlockEvents, err := s.beginBlock(ctx, newState)
	if err != nil {
		return nil, nil, err
	}

	// check if we need to return early
	if err = isCtxCancelled(ctx); err != nil {
		return nil, nil, err
	}

	hs := HeaderService{Info: header.Info{
		Hash:    block.Hash,
		AppHash: block.AppHash,
		ChainID: block.ChainId,
		Time:    block.Time,
		Height:  int64(block.Height),
	}}

	// execute txs
	txResults := make([]appmanager.TxResult, len(block.Txs))
	// TODO: skip first tx if vote extensions are enabled (marko)
	for i, txBytes := range block.Txs {
		// check if we need to return early or continue delivering txs
		if err = isCtxCancelled(ctx); err != nil {
			return nil, nil, err
		}
		txResults[i] = s.deliverTx(ctx, newState, txBytes, corecontext.ExecModeFinalize, hs)
	}
	// end block
	endBlockEvents, valset, err := s.endBlock(ctx, newState, hs)
	if err != nil {
		return nil, nil, err
	}

	return &appmanager.BlockResponse{
		PreBlockEvents:   preBlockEvents,
		BeginBlockEvents: beginBlockEvents,
		TxResults:        txResults,
		EndBlockEvents:   endBlockEvents,
		ValidatorUpdates: valset,
	}, newState, nil
}

// deliverTx executes a TX and returns the result.
func (s STF[T]) deliverTx(ctx context.Context, state store.WriterMap, tx T, execMode corecontext.ExecMode, hs header.Service) appmanager.TxResult {
	// recover in the case of a panic
	var recoveryError error
	defer func() {
		if r := recover(); r != nil {
			recoveryError = fmt.Errorf("panic during transaction execution: %s", r)
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

	validateGas, validationEvents, err := s.validateTx(ctx, state, gasLimit, tx, hs)
	if err != nil {
		return appmanager.TxResult{
			Error: err,
		}
	}

	execResp, execGas, execEvents, err := s.execTx(ctx, state, gasLimit-validateGas, tx, execMode, hs)
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
func (s STF[T]) validateTx(ctx context.Context, state store.WriterMap, gasLimit uint64, tx T, hs header.Service) (gasUsed uint64, events []event.Event, err error) {
	validateState := s.branch(state)
	txSenders, err := tx.GetSenders()
	if err != nil {
		return 0, nil, err
	}
	validateCtx := s.makeContext(ctx, txSenders, validateState, gasLimit, corecontext.ExecModeCheck, hs)
	err = s.doTxValidation(validateCtx, tx)
	if err != nil {
		return 0, nil, err
	}
	applyContextCache(ctx, validateCtx)

	return validateCtx.meter.Consumed(), validateCtx.events, applyStateChanges(state, validateState)
}

// execTx executes the tx messages on the provided state. If the tx fails then the state is discarded.
func (s STF[T]) execTx(ctx context.Context, state store.WriterMap, gasLimit uint64, tx T, execMode corecontext.ExecMode, hs header.Service) ([]transaction.Type, uint64, []event.Event, error) {
	execState := s.branch(state)
	txSenders, err := tx.GetSenders()
	if err != nil {
		return nil, 0, nil, err
	}
	execCtx := s.makeContext(ctx, txSenders, execState, gasLimit, execMode, nil)

	// atomic execution of the all messages in a transaction,
	msgsResp, txErr := s.runTxMsgs(ctx, execState, gasLimit, tx, execMode, hs)
	if txErr != nil {
		// in case of error during message execution, we do not apply the exec state.
		// instead we run the post exec handler in a new branch from the initial state.
		postTxState := s.branch(state)
		postTxCtx := s.makeContext(ctx, []transaction.Identity{appmanager.RuntimeIdentity}, postTxState, gas.NoGasLimit, execMode, hs)

		// TODO: runtime sets a noop posttxexec if the app doesnt set anything (julien)

		postTxErr := s.postTxExec(postTxCtx, tx, false)
		if postTxErr != nil {
			// if the post tx handler fails, then we do not apply any state change to the initial state.
			// we just return the exec gas used and a joined error from TX error and post TX error.
			return nil, execCtx.meter.Consumed(), nil, errors.Join(txErr, postTxErr)
		}
		// in case post tx is successful, then we commit the post tx state to the initial state,
		// and we return post tx events alongside exec gas used and the error of the tx.
		applyErr := applyStateChanges(state, postTxState)
		if applyErr != nil {
			return nil, 0, nil, applyErr
		}
		applyContextCache(ctx, postTxCtx)
		return nil, execCtx.meter.Consumed(), postTxCtx.events, txErr
	}
	// tx execution went fine, now we use the same state to run the post tx exec handler,
	// in case the execution of the post tx fails, then no state change is applied and the
	// whole execution step is rolled back.
	postTxCtx := s.makeContext(ctx, []transaction.Identity{appmanager.RuntimeIdentity}, execState, gas.NoGasLimit, execMode, hs) // NO gas limit.
	postTxErr := s.postTxExec(postTxCtx, tx, true)
	if postTxErr != nil {
		// if post tx fails, then we do not apply any state change, we return the post tx error,
		// alongside the gas used.
		return nil, execCtx.meter.Consumed(), nil, postTxErr
	}
	// both the execution and post tx execution step were successful, so we apply the state changes
	// to the provided state, and we return responses, and events from exec tx and post tx exec.
	applyErr := applyStateChanges(state, execState)
	if applyErr != nil {
		return nil, 0, nil, applyErr
	}
	applyContextCache(ctx, postTxCtx)

	return msgsResp, execCtx.meter.Consumed(), append(execCtx.events, postTxCtx.events...), nil
}

// runTxMsgs will execute the messages contained in the TX with the provided state.
func (s STF[T]) runTxMsgs(ctx context.Context, state store.WriterMap, gasLimit uint64, tx T, execMode corecontext.ExecMode, hs header.Service) ([]transaction.Type, error) {
	txSenders, err := tx.GetSenders()
	if err != nil {
		return nil, err
	}
	execCtx := s.makeContext(ctx, txSenders, state, gasLimit, execMode, hs)
	msgs, err := tx.GetMessages()
	if err != nil {
		return nil, err
	}
	msgResps := make([]transaction.Type, len(msgs))
	for i, msg := range msgs {
		resp, err := s.handleMsg(execCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("message execution at index %d failed: %w", i, err)
		}
		msgResps[i] = resp
	}
	applyContextCache(ctx, execCtx)
	return msgResps, nil
}

func (s STF[T]) preBlock(ctx context.Context, state store.WriterMap, txs []T) ([]event.Event, error) {
	pbCtx := s.makeContext(ctx, []transaction.Identity{appmanager.RuntimeIdentity}, state, gas.NoGasLimit, corecontext.ExecModeFinalize, nil)
	err := s.doPreBlock(pbCtx, txs)
	if err != nil {
		return nil, err
	}

	for i, e := range pbCtx.events {
		pbCtx.events[i].Attributes = append(
			e.Attributes,
			event.Attribute{Key: "mode", Value: "PreBlock"},
		)
	}
	applyContextCache(ctx, pbCtx)

	return pbCtx.events, nil
}

func (s STF[T]) beginBlock(ctx context.Context, state store.WriterMap) (beginBlockEvents []event.Event, err error) {
	bbCtx := s.makeContext(ctx, []transaction.Identity{appmanager.RuntimeIdentity}, state, gas.NoGasLimit, corecontext.ExecModeFinalize, nil)
	err = s.doBeginBlock(bbCtx)
	if err != nil {
		return nil, err
	}

	for i, e := range bbCtx.events {
		bbCtx.events[i].Attributes = append(
			e.Attributes,
			event.Attribute{Key: "mode", Value: "BeginBlock"},
		)
	}
	applyContextCache(ctx, bbCtx)

	return bbCtx.events, nil
}

func (s STF[T]) endBlock(ctx context.Context, state store.WriterMap, hs header.Service) ([]event.Event, []appmodulev2.ValidatorUpdate, error) {
	ebCtx := s.makeContext(ctx, []transaction.Identity{appmanager.RuntimeIdentity}, state, gas.NoGasLimit, corecontext.ExecModeFinalize, hs)
	err := s.doEndBlock(ebCtx)
	if err != nil {
		return nil, nil, err
	}

	events, valsetUpdates, err := s.validatorUpdates(ctx, state, hs)
	if err != nil {
		return nil, nil, err
	}

	ebCtx.events = append(ebCtx.events, events...)

	for i, e := range ebCtx.events {
		ebCtx.events[i].Attributes = append(
			e.Attributes,
			event.Attribute{Key: "mode", Value: "BeginBlock"},
		)
	}
	applyContextCache(ctx, ebCtx)

	return ebCtx.events, valsetUpdates, nil
}

// validatorUpdates returns the validator updates for the current block. It is called by endBlock after the endblock execution has concluded
func (s STF[T]) validatorUpdates(ctx context.Context, state store.WriterMap, hs header.Service) ([]event.Event, []appmodulev2.ValidatorUpdate, error) {
	ebCtx := s.makeContext(ctx, []transaction.Identity{appmanager.RuntimeIdentity}, state, gas.NoGasLimit, corecontext.ExecModeFinalize, hs)
	valSetUpdates, err := s.doValidatorUpdate(ebCtx)
	if err != nil {
		return nil, nil, err
	}
	applyContextCache(ctx, ebCtx)
	return ebCtx.events, valSetUpdates, nil
}

// Simulate simulates the execution of a tx on the provided state.
func (s STF[T]) Simulate(ctx context.Context, state store.ReaderMap, gasLimit uint64, tx T) (appmanager.TxResult, store.WriterMap) {
	simulationState := s.branch(state)
	txr := s.deliverTx(ctx, simulationState, tx, corecontext.ExecModeSimulate, nil)

	return txr, simulationState
}

// ValidateTx will run only the validation steps required for a transaction.
// Validations are run over the provided state, with the provided gas limit.
func (s STF[T]) ValidateTx(ctx context.Context, state store.ReaderMap, gasLimit uint64, tx T, hs header.Service) appmanager.TxResult {
	validationState := s.branch(state)
	gasUsed, events, err := s.validateTx(ctx, validationState, gasLimit, tx, hs)
	return appmanager.TxResult{
		Events:  events,
		GasUsed: gasUsed,
		Error:   err,
	}
}

// Query executes the query on the provided state with the provided gas limits.
func (s STF[T]) Query(ctx context.Context, state store.ReaderMap, gasLimit uint64, req transaction.Type) (transaction.Type, error) {
	queryState := s.branch(state)
	queryCtx := s.makeContext(ctx, nil, queryState, gasLimit, corecontext.ExecModeSimulate, nil)
	return s.handleQuery(queryCtx, req)
}

// clone clones STF.
func (s STF[T]) clone() STF[T] {
	return STF[T]{
		handleMsg:         s.handleMsg,
		handleQuery:       s.handleQuery,
		doPreBlock:        s.doPreBlock,
		doBeginBlock:      s.doBeginBlock,
		doEndBlock:        s.doEndBlock,
		doValidatorUpdate: s.doValidatorUpdate,
		doTxValidation:    s.doTxValidation,
		postTxExec:        s.postTxExec,
		branch:            s.branch,
		getGasMeter:       s.getGasMeter,
		wrapWithGasMeter:  s.wrapWithGasMeter,
	}
}

// executionContext is a struct that holds the context for the execution of a tx.
type executionContext struct {
	context.Context

	state         store.WriterMap
	meter         gas.Meter
	events        []event.Event
	sender        []transaction.Identity
	branchdb      branch.Service
	headerService header.Service
	execMode      corecontext.ExecMode
	State         store.WriterMap
	Cache         ModuleContainer
}

func (s STF[T]) makeContext(
	ctx context.Context,
	sender []transaction.Identity,
	store store.WriterMap,
	gasLimit uint64,
	execMode corecontext.ExecMode,
	headerservice header.Service,
) *executionContext {
	meter := s.getGasMeter(gasLimit)
	store = s.wrapWithGasMeter(meter, store)
	var cache ModuleContainer
	executionCtx, ok := ctx.(*executionContext)
	if ok {
		cache = executionCtx.Cache
	} else {
		cache = NewModuleContainer()
	}
	return &executionContext{
		Context:       ctx,
		state:         store,
		meter:         meter,
		events:        make([]event.Event, 0),
		sender:        sender,
		execMode:      execMode,
		branchdb:      BrachService{s.branch},
		headerService: headerservice, // todo need a wrapper
		State:         store,
		Cache:         cache,
	}
}

func (s STF[T]) contextInitCache(
	ctx context.Context,
) *executionContext {
	return &executionContext{
		Context: ctx,
		Cache:   NewModuleContainer(),
	}
}

func applyStateChanges(dst, src store.WriterMap) error {
	changes, err := src.GetStateChanges()
	if err != nil {
		return err
	}
	return dst.ApplyStateChanges(changes)
}

func applyContextCache(dst, src context.Context) error {
	srcExecutionCtx, ok := src.(*executionContext)
	if !ok {
		return fmt.Errorf("Can not convert ctx to executionContext")
	}
	_, ok = dst.(*executionContext)
	if !ok {
		return fmt.Errorf("Can not convert ctx to executionContext")
	}
	dst.(*executionContext).Cache = srcExecutionCtx.Cache
	return nil
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
