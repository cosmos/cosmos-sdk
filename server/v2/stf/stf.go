package stf

import (
	"context"
	"errors"
	"fmt"
	"math"

	"google.golang.org/protobuf/proto"

	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/core/transaction"
)

type (
	Type     = proto.Message
	Identity = []byte
)

var runtimeIdentity Identity = []byte("runtime") // TODO: most likely should be moved to core somewhere.

// STF is a struct that manages the state transition component of the app.
type STF[T transaction.Tx] struct {
	handleMsg   func(ctx context.Context, msg Type) (msgResp Type, err error)
	handleQuery func(ctx context.Context, req Type) (resp Type, err error)

	doUpgradeBlock    func(ctx context.Context) (bool, error)
	doBeginBlock      func(ctx context.Context) error
	doEndBlock        func(ctx context.Context) error
	doValidatorUpdate func(ctx context.Context) ([]appmanager.ValidatorUpdate, error)

	doTxValidation func(ctx context.Context, tx T) error
	postTxExec     func(ctx context.Context, tx T, success bool) error
	branch         func(store store.ReadonlyState) store.WritableState // branch is a function that given a readonly store it returns a writable version of it.
}

// DeliverBlock is our state transition function.
// It takes a read only view of the state to apply the block to,
// executes the block and returns the block results and the new state.
func (s STF[T]) DeliverBlock(ctx context.Context, block *appmanager.BlockRequest[T], state store.ReadonlyState) (blockResult *appmanager.BlockResponse, newState store.WritableState, err error) {
	// creates a new branch store, from the readonly view of the state
	// that can be written to.
	newState = s.branch(state)

	// upgrade block is called separate from begin block in order to refresh state updates
	upgradeBlockEvents, err := s.upgradeBlock(ctx, newState)
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
	for i, txBytes := range block.Txs {
		// check if we need to return early or continue delivering txs
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
			txResults[i] = s.deliverTx(ctx, newState, txBytes)
		}
	}
	// end block
	endBlockEvents, valset, err := s.endBlock(ctx, newState)
	if err != nil {
		return nil, nil, err
	}

	return &appmanager.BlockResponse{
		UpgradeBlockEvents: upgradeBlockEvents,
		BeginBlockEvents:   beginBlockEvents,
		TxResults:          txResults,
		EndBlockEvents:     endBlockEvents,
		ValidatorUpdates:   valset,
	}, newState, nil
}

// deliverTx executes a TX and returns the result.
func (s STF[T]) deliverTx(ctx context.Context, state store.WritableState, tx T) appmanager.TxResult {
	// recover in the case of a panic
	// TODO: after discussion with users see if we need middleware
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

	execResp, execGas, execEvents, err := s.execTx(ctx, state, tx.GetGasLimit()-validateGas, tx)
	return appmanager.TxResult{
		Events:  append(validationEvents, execEvents...),
		GasUsed: execGas + validateGas,
		Resp:    execResp,
		Error:   err,
	}
}

// validateTx validates a transaction given the provided WritableState and gas limit.
// If the validation is successful, state is committed
func (s STF[T]) validateTx(ctx context.Context, state store.WritableState, gasLimit uint64, tx T) (gasUsed uint64, events []event.Event, err error) {
	validateState := s.branch(state)
	validateCtx := s.makeContext(ctx, tx.GetSenders(), validateState, gasLimit)
	err = s.doTxValidation(validateCtx, tx)
	if err != nil {
		return 0, nil, err
	}
	return validateCtx.gasUsed, validateCtx.events, applyStateChanges(state, validateState)
}

// execTx executes the tx messages on the provided state. If the tx fails then the state is discarded.
func (s STF[T]) execTx(ctx context.Context, state store.WritableState, gasLimit uint64, tx T) ([]Type, uint64, []event.Event, error) {
	execState := s.branch(state)
	execCtx := s.makeContext(ctx, tx.GetSenders(), execState, gasLimit)

	// atomic execution of the all messages in a transaction, TODO: we should allow messages to fail in a specific mode
	msgsResp, txErr := s.runTxMsgs(ctx, execState, gasLimit, tx)
	if txErr != nil {
		// in case of error during message execution, we do not apply the exec state.
		// instead we run the post exec handler in a new branch from the initial state.
		postTxState := s.branch(state)
		postTxCtx := s.makeContext(ctx, []Identity{runtimeIdentity}, postTxState, math.MaxUint64) // NO gas limit.
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
	postTxCtx := s.makeContext(ctx, []Identity{runtimeIdentity}, execState, math.MaxUint64) // NO gas limit.
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
func (s STF[T]) runTxMsgs(ctx context.Context, state store.WritableState, gasLimit uint64, tx T) ([]Type, error) {
	execCtx := s.makeContext(ctx, tx.GetSenders(), state, gasLimit)
	msgs := tx.GetMessages()
	msgResps := make([]Type, len(msgs))
	for i, msg := range msgs {
		resp, err := s.handleMsg(execCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("message execution at index %d failed: %w", i, err)
		}
		msgResps[i] = resp
	}
	return msgResps, nil
}

func (s STF[T]) upgradeBlock(ctx context.Context, state store.WritableState) ([]event.Event, error) {
	pbCtx := s.makeContext(ctx, []Identity{runtimeIdentity}, state, 0) // TODO: gas limit
	refresh, err := s.doUpgradeBlock(pbCtx)
	if err != nil {
		return nil, err
	}
	if refresh {
		// TODO: update context with updated params
	}

	return pbCtx.events, nil
}

func (s STF[T]) beginBlock(ctx context.Context, state store.WritableState) (beginBlockEvents []event.Event, err error) {
	bbCtx := s.makeContext(ctx, []Identity{runtimeIdentity}, state, 0) // TODO: gas limit
	err = s.doBeginBlock(bbCtx)
	if err != nil {
		return nil, err
	}
	return bbCtx.events, nil
}

func (s STF[T]) endBlock(ctx context.Context, state store.WritableState) ([]event.Event, []appmanager.ValidatorUpdate, error) {
	ebCtx := s.makeContext(ctx, []Identity{runtimeIdentity}, state, 0) // TODO: gas limit
	err := s.doEndBlock(ebCtx)
	if err != nil {
		return nil, nil, err
	}

	events, valsetUpdates, err := s.validatorUpdates(ctx, state)
	if err != nil {
		return nil, nil, err
	}

	ebCtx.events = append(ebCtx.events, events...)

	return ebCtx.events, valsetUpdates, nil
}

// validatorUpdates returns the validator updates for the current block. It is called by endBlock after the endblock execution has concluded
func (s STF[T]) validatorUpdates(ctx context.Context, state store.WritableState) ([]event.Event, []appmanager.ValidatorUpdate, error) {
	ebCtx := s.makeContext(ctx, []Identity{runtimeIdentity}, state, 0) // TODO: gas limit
	valSetUpdates, err := s.doValidatorUpdate(ebCtx)
	if err != nil {
		return nil, nil, err
	}
	return ebCtx.events, valSetUpdates, nil
}

// Simulate simulates the execution of a tx on the provided state.
func (s STF[T]) Simulate(ctx context.Context, state store.ReadonlyState, gasLimit uint64, tx T) appmanager.TxResult {
	simulationState := s.branch(state)
	return s.deliverTx(ctx, simulationState, tx)
}

// ValidateTx will run only the validation steps required for a transaction.
// Validations are run over the provided state, with the provided gas limit.
func (s STF[T]) ValidateTx(ctx context.Context, state store.ReadonlyState, gasLimit uint64, tx T) appmanager.TxResult {
	validationState := s.branch(state)
	gasUsed, events, err := s.validateTx(ctx, validationState, gasLimit, tx)
	return appmanager.TxResult{
		Events:  events,
		GasUsed: gasUsed,
		Error:   err,
	}
}

// Query executes the query on the provided state with the provided gas limits.
func (s STF[T]) Query(ctx context.Context, state store.ReadonlyState, gasLimit uint64, req Type) (Type, error) {
	queryState := s.branch(state)
	queryCtx := s.makeContext(ctx, nil, queryState, gasLimit)
	return s.handleQuery(queryCtx, req)
}

type executionContext struct {
	context.Context
	store    store.WritableState
	gasUsed  uint64
	gasLimit uint64
	events   []event.Event
	sender   []Identity
}

func (s STF[T]) makeContext(
	ctx context.Context,
	sender []Identity,
	store store.WritableState,
	gasLimit uint64,
	// TODO add exec mode
) *executionContext {
	return &executionContext{
		Context:  ctx,
		store:    store,
		gasUsed:  0,
		gasLimit: gasLimit,
		events:   make([]event.Event, 0),
		sender:   sender,
	}
}

// applyStateChanges writes the changes in state from src to dst.
func applyStateChanges(dst, src store.WritableState) error {
	changes, err := src.ChangeSets()
	if err != nil {
		return err
	}

	return dst.ApplyChangeSets(changes)
}
