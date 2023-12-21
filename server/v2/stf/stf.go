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

var runtimeIdentity Identity = []byte("app-manager")

// STFAppManager is a struct that manages the state transition component of the app.
type STFAppManager[T transaction.Tx] struct {
	HandleMsg   func(ctx context.Context, msg Type) (msgResp Type, err error)
	HandleQuery func(ctx context.Context, req Type) (resp Type, err error)

	doBeginBlock func(ctx context.Context) error
	doEndBlock   func(ctx context.Context) error

	doTxValidation func(ctx context.Context, tx T) error

	decodeTx func(txBytes []byte) (T, error)

	Branch func(store store.ReadonlyStore) store.BranchStore // branch is a function that given a readonly store it returns a writable version of it.
}

// DeliverBlock is our state transition function.
// It takes a read only view of the state to apply the block to,
// executes the block and returns the block results and the new state.
func (s STFAppManager[T]) DeliverBlock(ctx context.Context, block appmanager.BlockRequest, state store.ReadonlyStore) (blockResult *appmanager.BlockResponse, newState store.BranchStore, err error) {
	blockResult = new(appmanager.BlockResponse)
	// creates a new branch store, from the readonly view of the state
	// that can be written to.
	newState = s.Branch(state)
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

func (s STFAppManager[T]) beginBlock(ctx context.Context, state store.BranchStore) (beginBlockEvents []event.Event, err error) {
	execCtx := s.MakeContext(ctx, []Identity{runtimeIdentity}, state, 0) // TODO: gas limit
	err = s.doBeginBlock(execCtx)
	if err != nil {
		return nil, err
	}
	// apply state changes
	changes, err := execCtx.store.ChangeSets()
	if err != nil {
		return nil, err
	}
	return execCtx.events, state.ApplyChangeSets(changes)
}

func (s STFAppManager[T]) deliverTx(ctx context.Context, state store.BranchStore, txBytes []byte) appmanager.TxResult {
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

// validateTx validates a transaction given the provided BranchStore and gas limit.
// If the validation is successful, state is committed
func (s STFAppManager[T]) validateTx(ctx context.Context, store store.BranchStore, gasLimit uint64, tx T) (gasUsed uint64, events []event.Event, err error) {
	validateCtx := s.MakeContext(ctx, tx.GetSenders(), store, gasLimit)
	err = s.doTxValidation(ctx, tx)
	if err != nil {
		return 0, nil, nil
	}
	// all went fine we can commit to state.
	changeSets, err := validateCtx.store.ChangeSets()
	if err != nil {
		return 0, nil, err
	}
	err = store.ApplyChangeSets(changeSets)
	if err != nil {
		return 0, nil, err
	}
	return validateCtx.gasUsed, validateCtx.events, nil
}

func (s STFAppManager[T]) execTx(ctx context.Context, store store.BranchStore, gasLimit uint64, tx T) (msgResp Type, gasUsed uint64, execEvents []event.Event, err error) {
	execCtx := s.MakeContext(ctx, tx.GetSenders(), store, gasLimit)
	// atomic execution of the all messages in a transaction, TODO: we should allow messages to fail in a specific mode
	for _, msg := range tx.GetMessages() {
		msgResp, err = s.HandleMsg(ctx, msg)
		if err != nil {
			return nil, 0, nil, err
		}

	}
	// get state changes and save them to the parent store
	changeSets, err := execCtx.store.ChangeSets()
	if err != nil {
		return nil, 0, nil, err
	}
	err = store.ApplyChangeSets(changeSets)
	if err != nil {
		return nil, 0, nil, err
	}
	return msgResp, 0, execCtx.events, nil
}

func (s STFAppManager[T]) endBlock(ctx context.Context, store store.BranchStore, block appmanager.BlockRequest) (endBlockEvents []event.Event, err error) {
	execCtx := s.MakeContext(ctx, []Identity{runtimeIdentity}, store, 0) // TODO: gas limit
	err = s.doBeginBlock(execCtx)
	if err != nil {
		return nil, err
	}
	// apply state changes
	changes, err := execCtx.store.ChangeSets()
	if err != nil {
		return nil, err
	}
	return execCtx.events, store.ApplyChangeSets(changes)
}

type executionContext struct {
	context.Context
	store    store.BranchStore
	gasUsed  uint64
	gasLimit uint64
	events   []event.Event
	sender   []Identity
}

func (s STFAppManager[T]) MakeContext(
	ctx context.Context,
	sender []Identity,
	store store.BranchStore,
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
