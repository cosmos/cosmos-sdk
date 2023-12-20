package stf

import (
	"context"

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

	doBeginBlock func(ctx context.Context) error
	doEndBlock   func(ctx context.Context) error

	doTxValidation func(ctx context.Context, tx T) error

	decodeTx func(txBytes []byte) (T, error)

	branch func(store store.ReadonlyState) store.WritableState // branch is a function that given a readonly store it returns a writable version of it.
}

// DeliverBlock is our state transition function.
// It takes a read only view of the state to apply the block to,
// executes the block and returns the block results and the new state.
func (s STF[T]) DeliverBlock(ctx context.Context, block appmanager.BlockRequest, state store.ReadonlyState) (blockResult *appmanager.BlockResponse, newState store.WritableState, err error) {
	// creates a new branch store, from the readonly view of the state
	// that can be written to.
	newState = s.branch(state)
	// begin block
	beginBlockEvents, err := s.beginBlock(ctx, newState)
	if err != nil {
		return nil, nil, err
	}
	// execute txs
	txResults := make([]appmanager.TxResult, len(block.Txs))
	for i, txBytes := range block.Txs {
		txResults[i] = s.deliverTx(ctx, newState, txBytes)
	}
	// end block
	endBlockEvents, err := s.endBlock(ctx, newState, block)
	if err != nil {
		return nil, nil, err
	}

	return &appmanager.BlockResponse{
		BeginBlockEvents: beginBlockEvents,
		TxResults:        txResults,
		EndBlockEvents:   endBlockEvents,
	}, newState, nil
}

// DeliverTx executes a TX and returns the result.
func (s STF[T]) deliverTx(ctx context.Context, state store.WritableState, txBytes []byte) appmanager.TxResult {
	tx, err := s.decodeTx(txBytes)
	if err != nil {
		return appmanager.TxResult{
			Error: err,
		}
	}

	validateGas, validationEvents, err := s.validateTx(ctx, state, tx.GetGasLimit(), tx)
	if err != nil {
		return appmanager.TxResult{
			Error: err,
		}
	}

	execResp, execGas, execEvents, err := s.execTx(ctx, state, tx.GetGasLimit()-validateGas, tx)
	if err != nil {
		return appmanager.TxResult{
			Events:  validationEvents,
			GasUsed: validateGas + execGas,
			Error:   err,
		}
	}

	return appmanager.TxResult{
		Events:  append(validationEvents, execEvents...),
		GasUsed: execGas + validateGas,
		Resp:    execResp,
		Error:   nil,
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

func (s STF[T]) execTx(ctx context.Context, state store.WritableState, gasLimit uint64, tx T) (msgResp Type, gasUsed uint64, execEvents []event.Event, err error) {
	execState := s.branch(state)
	execCtx := s.makeContext(ctx, tx.GetSenders(), execState, gasLimit)
	// atomic execution of the all messages in a transaction, TODO: we should allow messages to fail in a specific mode
	for _, msg := range tx.GetMessages() {
		msgResp, err = s.handleMsg(execCtx, msg)
		if err != nil {
			return nil, 0, nil, err
		}
	}
	return msgResp, 0, execCtx.events, applyStateChanges(state, execState)
}

func (s STF[T]) beginBlock(ctx context.Context, state store.WritableState) (beginBlockEvents []event.Event, err error) {
	bbCtx := s.makeContext(ctx, []Identity{runtimeIdentity}, state, 0) // TODO: gas limit
	err = s.doBeginBlock(bbCtx)
	if err != nil {
		return nil, err
	}
	return bbCtx.events, nil
}

func (s STF[T]) endBlock(ctx context.Context, store store.WritableState, block appmanager.BlockRequest) (endBlockEvents []event.Event, err error) {
	ebCtx := s.makeContext(ctx, []Identity{runtimeIdentity}, store, 0) // TODO: gas limit
	err = s.doEndBlock(ebCtx)
	if err != nil {
		return nil, err
	}
	return ebCtx.events, nil
}

// Simulate simulates the execution of a tx on the provided state.
func (s STF[T]) Simulate(ctx context.Context, state store.ReadonlyState, gasLimit uint64, tx []byte) appmanager.TxResult {
	simulationState := s.branch(state)
	return s.deliverTx(ctx, simulationState, tx)
}

// ValidateTx will run only the validation steps required for a transaction.
// Validations are run over the provided state, with the provided gas limit.
func (s STF[T]) ValidateTx(ctx context.Context, state store.ReadonlyState, gasLimit uint64, txBytes []byte) appmanager.TxResult {
	tx, err := s.decodeTx(txBytes)
	if err != nil {
		return appmanager.TxResult{Error: err}
	}
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
func applyStateChanges(dst store.WritableState, src store.WritableState) error {
	changes, err := src.ChangeSets()
	if err != nil {
		return err
	}
	return dst.ApplyChangeSets(changes)
}
