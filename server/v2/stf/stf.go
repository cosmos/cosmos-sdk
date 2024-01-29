package stf

import (
	"context"
	"errors"
	"fmt"
	"math"

	"cosmossdk.io/core/appmodule"
	corecontext "cosmossdk.io/core/context"
	coreevent "cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/store"
)

var runtimeIdentity transaction.Identity = []byte("runtime") // TODO: most likely should be moved to core somewhere.

// STF is a struct that manages the state transition component of the app.
type STF[T transaction.Tx] struct {
	handleMsg   func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error)
	handleQuery func(ctx context.Context, req transaction.Type) (resp transaction.Type, err error)

	doPreBlock        func(ctx context.Context, txs []T) error
	doBeginBlock      func(ctx context.Context) error
	doEndBlock        func(ctx context.Context) error
	doValidatorUpdate func(ctx context.Context) ([]appmodule.ValidatorUpdate, error)

	doTxValidation func(ctx context.Context, tx T) error // TODO: rewrite antehandlers remove simulate
	postTxExec     func(ctx context.Context, tx T, success bool) error
	branch         func(state store.GetReader) store.GetWriter // branch is a function that given a readonly store it returns a writable version of it.
	// TODO: add gas store
}

// NewSTF returns a new STF instance.
func NewSTF[T transaction.Tx](
	handleMsg func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error),
	handleQuery func(ctx context.Context, req transaction.Type) (resp transaction.Type, err error),
	doPreBlock func(ctx context.Context, txs []T) error,
	doBeginBlock func(ctx context.Context) error,
	doEndBlock func(ctx context.Context) error,
	doTxValidation func(ctx context.Context, tx T) error,
	doValidatorUpdate func(ctx context.Context) ([]appmodule.ValidatorUpdate, error),
	branch func(store store.GetReader) store.GetWriter,
) *STF[T] {
	return &STF[T]{
		handleMsg:         handleMsg,
		handleQuery:       handleQuery,
		doPreBlock:        doPreBlock,
		doBeginBlock:      doBeginBlock,
		doEndBlock:        doEndBlock,
		doTxValidation:    doTxValidation,
		doValidatorUpdate: doValidatorUpdate,
		postTxExec:        nil, // TODO
		branch:            branch,
	}
}

// DeliverBlock is our state transition function.
// It takes a read only view of the state to apply the block to,
// executes the block and returns the block results and the new state.
func (s STF[T]) DeliverBlock(ctx context.Context, block *appmanager.BlockRequest[T], state store.GetReader) (blockResult *appmanager.BlockResponse, newState store.GetWriter, err error) {
	// creates a new branch store, from the readonly view of the state
	// that can be written to.
	newState = s.branch(state)

	// TODO: handle consensus messages

	// pre block is called separate from begin block in order to prepopulate state
	preBlockEvents, err := s.preBlock(ctx, newState, block.Txs)
	if err != nil {
		return nil, nil, err
	}

	// check if we need to return early
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		// continue
	}

	// begin block
	beginBlockEvents, err := s.beginBlock(ctx, newState)
	if err != nil {
		return nil, nil, err
	}

	// check if we need to return early
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		// continue
	}

	// execute txs
	txResults := make([]appmanager.TxResult, len(block.Txs))
	// TODO: skip first tx if vote extensions are enabled (marko)
	for i, txBytes := range block.Txs {
		// check if we need to return early or continue delivering txs
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
			txResults[i] = s.deliverTx(ctx, newState, txBytes, corecontext.ExecModeFinalize)
		}
	}
	// end block
	endBlockEvents, valset, err := s.endBlock(ctx, newState)
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
func (s STF[T]) deliverTx(ctx context.Context, state store.GetWriter, tx T, execMode corecontext.ExecMode) appmanager.TxResult {
	// recover in the case of a panic
	var recoveryError error
	defer func() {
		if r := recover(); r != nil {
			recoveryError = fmt.Errorf("panic during transaction execution: %s", r)
		}
	}()
	if recoveryError != nil {
		return appmanager.TxResult{
			Error: recoveryError,
		}
	}

	validateGas, validationEvents, err := s.validateTx(ctx, state, tx.GetGasLimit(), tx)
	if err != nil {
		return appmanager.TxResult{
			Error: err,
		}
	}

	execResp, execGas, execEvents, err := s.execTx(ctx, state, tx.GetGasLimit()-validateGas, tx, execMode)
	return appmanager.TxResult{
		Events:  append(validationEvents, execEvents...),
		GasUsed: execGas + validateGas,
		Resp:    execResp,
		Error:   err,
	}
}

// validateTx validates a transaction given the provided WritableState and gas limit.
// If the validation is successful, state is committed
func (s STF[T]) validateTx(ctx context.Context, state store.GetWriter, gasLimit uint64, tx T) (gasUsed uint64, events []event.Event, err error) {
	validateState := s.branch(state)
	validateCtx := s.makeContext(ctx, tx.GetSenders(), validateState, gasLimit, corecontext.ExecModeCheck)
	err = s.doTxValidation(validateCtx, tx)
	if err != nil {
		return 0, nil, err
	}

	return validateCtx.gasUsed, validateCtx.events, applyStateChanges(state, validateState)
}

// execTx executes the tx messages on the provided state. If the tx fails then the state is discarded.
func (s STF[T]) execTx(ctx context.Context, state store.GetWriter, gasLimit uint64, tx T, execMode corecontext.ExecMode) ([]transaction.Type, uint64, []event.Event, error) {
	execState := s.branch(state)
	execCtx := s.makeContext(ctx, tx.GetSenders(), execState, gasLimit, execMode)

	// atomic execution of the all messages in a transaction, TODO: we should allow messages to fail in a specific mode
	msgsResp, txErr := s.runTxMsgs(ctx, execState, gasLimit, tx, execMode)
	if txErr != nil {
		// in case of error during message execution, we do not apply the exec state.
		// instead we run the post exec handler in a new branch from the initial state.
		postTxState := s.branch(state)
		postTxCtx := s.makeContext(ctx, []transaction.Identity{runtimeIdentity}, postTxState, math.MaxUint64, execMode) // NO gas limit.

		// TODO: runtime sets a noop posttxexec if the app doesnt set anything (julien)

		postTxErr := s.postTxExec(postTxCtx, tx, false)
		if postTxErr != nil {
			// if the post tx handler fails, then we do not apply any state change to the initial state.
			// we just return the exec gas used and a joined error from TX error and post TX error.
			return nil, execCtx.gasUsed, nil, errors.Join(txErr, postTxErr)
		}
		// in case post tx is successful, then we commit the post tx state to the initial state,
		// and we return post tx events alongside exec gas used and the error of the tx.
		applyErr := applyStateChanges(state, postTxState)
		if applyErr != nil {
			return nil, 0, nil, applyErr
		}
		return nil, execCtx.gasUsed, postTxCtx.events, txErr
	}
	// tx execution went fine, now we use the same state to run the post tx exec handler,
	// in case the execution of the post tx fails, then no state change is applied and the
	// whole execution step is rolled back.
	postTxCtx := s.makeContext(ctx, []transaction.Identity{runtimeIdentity}, execState, math.MaxUint64, execMode) // NO gas limit.
	postTxErr := s.postTxExec(postTxCtx, tx, true)
	if postTxErr != nil {
		// if post tx fails, then we do not apply any state change, we return the post tx error,
		// alongside the gas used.
		return nil, execCtx.gasUsed, nil, postTxErr
	}
	// both the execution and post tx execution step were successful, so we apply the state changes
	// to the provided state, and we return responses, and events from exec tx and post tx exec.
	applyErr := applyStateChanges(state, execState)
	if applyErr != nil {
		return nil, 0, nil, applyErr
	}

	return msgsResp, execCtx.gasUsed, append(execCtx.events, postTxCtx.events...), nil
}

// runTxMsgs will execute the messages contained in the TX with the provided state.
// TODO: multimessage both atomic and non atomic
func (s STF[T]) runTxMsgs(ctx context.Context, state store.GetWriter, gasLimit uint64, tx T, execMode corecontext.ExecMode) ([]transaction.Type, error) {
	execCtx := s.makeContext(ctx, tx.GetSenders(), state, gasLimit, execMode)
	msgs := tx.GetMessages()
	msgResps := make([]transaction.Type, len(msgs))
	for i, msg := range msgs {
		resp, err := s.handleMsg(execCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("message execution at index %d failed: %w", i, err)
		}
		msgResps[i] = resp
	}
	return msgResps, nil
}

func (s STF[T]) preBlock(ctx context.Context, state store.GetWriter, txs []T) ([]event.Event, error) {
	pbCtx := s.makeContext(ctx, []transaction.Identity{runtimeIdentity}, state, math.MaxUint64, corecontext.ExecModeFinalize)
	err := s.doPreBlock(pbCtx, txs)
	if err != nil {
		return nil, err
	}

	for i, e := range pbCtx.events {
		pbCtx.events[i].Attributes = append(
			e.Attributes,
			coreevent.Attribute{Key: "mode", Value: "PreBlock"},
		)
	}
	// TODO: update consensus module to accept consensus messages (facu)

	return pbCtx.events, nil
}

// beginBlock executes the begin block logic and returns the events.
// it is called by deliver block after the upgrade block logic has concluded. It is assumed simulations will not call begin block.
func (s STF[T]) beginBlock(ctx context.Context, state store.GetWriter) (beginBlockEvents []event.Event, err error) {
	bbCtx := s.makeContext(ctx, []transaction.Identity{runtimeIdentity}, state, math.MaxUint64, corecontext.ExecModeFinalize) // unlimited gas
	err = s.doBeginBlock(bbCtx)
	if err != nil {
		return nil, err
	}

	for i, e := range bbCtx.events {
		bbCtx.events[i].Attributes = append(
			e.Attributes,
			coreevent.Attribute{Key: "mode", Value: "BeginBlock"},
		)
	}

	return bbCtx.events, nil
}

// endBlock executes the end block logic and returns the events and validator updates.
// it is called by deliver block after the execution of the txs has concluded. It is assumed simulations will not call end block.
func (s STF[T]) endBlock(ctx context.Context, state store.GetWriter) ([]event.Event, []appmodule.ValidatorUpdate, error) {
	ebCtx := s.makeContext(ctx, []transaction.Identity{runtimeIdentity}, state, math.MaxUint64, corecontext.ExecModeFinalize) // unlimited gas
	err := s.doEndBlock(ebCtx)
	if err != nil {
		return nil, nil, err
	}

	events, valsetUpdates, err := s.validatorUpdates(ctx, state)
	if err != nil {
		return nil, nil, err
	}

	ebCtx.events = append(ebCtx.events, events...)

	for i, e := range ebCtx.events {
		ebCtx.events[i].Attributes = append(
			e.Attributes,
			coreevent.Attribute{Key: "mode", Value: "BeginBlock"},
		)
	}

	return ebCtx.events, valsetUpdates, nil
}

// validatorUpdates returns the validator updates for the current block. It is called by endBlock after the endblock execution has concluded
// and before the state is committed. It is assumed simulations will not call validator updates.
func (s STF[T]) validatorUpdates(ctx context.Context, state store.GetWriter) ([]event.Event, []appmodule.ValidatorUpdate, error) {
	ebCtx := s.makeContext(ctx, []transaction.Identity{runtimeIdentity}, state, math.MaxUint64, corecontext.ExecModeFinalize) // unlimited gas
	valSetUpdates, err := s.doValidatorUpdate(ebCtx)
	if err != nil {
		return nil, nil, err
	}
	return ebCtx.events, valSetUpdates, nil
}

// Simulate simulates the execution of a tx on the provided state.
func (s STF[T]) Simulate(ctx context.Context, state store.GetReader, gasLimit uint64, tx T) (appmanager.TxResult, []store.StateChanges) {
	simulationState := s.branch(state)
	cs, err := simulationState.GetStateChanges()
	if err != nil {
		return appmanager.TxResult{}, nil
	}
	return s.deliverTx(ctx, simulationState, tx, corecontext.ExecModeSimulate), cs
}

// ValidateTx will run only the validation steps required for a transaction.
// Validations are run over the provided state, with the provided gas limit.
func (s STF[T]) ValidateTx(ctx context.Context, state store.GetReader, gasLimit uint64, tx T) appmanager.TxResult {
	validationState := s.branch(state)
	gasUsed, events, err := s.validateTx(ctx, validationState, gasLimit, tx)
	return appmanager.TxResult{
		Events:  events,
		GasUsed: gasUsed,
		Error:   err,
	}
}

// Query executes the query on the provided state with the provided gas limits.
func (s STF[T]) Query(ctx context.Context, state store.GetReader, gasLimit uint64, req transaction.Type) (transaction.Type, error) {
	queryState := s.branch(state)
	queryCtx := s.makeContext(ctx, nil, queryState, gasLimit, corecontext.ExecModeSimulate)
	return s.handleQuery(queryCtx, req)
}

// executionContext is a struct that holds the context for the execution of a tx.
// TODO: look if we are missing anything here
type executionContext struct {
	context.Context
	store    store.GetWriter
	gasUsed  uint64
	gasLimit uint64
	events   []event.Event
	sender   []transaction.Identity
	execMode corecontext.ExecMode
	// TODO: add gas meter/kv
	// TODO: add services
}

func (s STF[T]) makeContext(
	ctx context.Context,
	sender []transaction.Identity,
	store store.GetWriter,
	gasLimit uint64,
	execMode corecontext.ExecMode,
) *executionContext {
	return &executionContext{
		Context:  ctx,
		store:    store,
		gasUsed:  0,
		gasLimit: gasLimit,
		events:   make([]event.Event, 0),
		sender:   sender,
		execMode: execMode,
	}
}

func applyStateChanges(dst, src store.GetWriter) error {
	changes, err := src.GetStateChanges()
	if err != nil {
		return err
	}
	return dst.ApplyStateChanges(changes)
}
