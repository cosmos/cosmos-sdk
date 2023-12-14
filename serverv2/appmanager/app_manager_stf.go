package appmanager

import (
	"context"
	"time"
)

type executionContext struct {
	context.Context
	store    BranchStore
	gasUsed  uint64
	gasLimit uint64
	events   []Event
}

type TxDecoder interface {
	Decode([]byte) (Tx, error)
}

type Tx interface {
	GetMessage() Type
	GetSender() Identity
	GetGasLimit() uint64
}

type Block struct {
	Height            uint64
	Time              time.Time
	Hash              []byte
	Txs               [][]byte
	ConsensusMessages []Type // <= proto.Message
}

type BlockResponse struct {
	AppHash                    Hash
	BeginBlockEvents           []Event
	TxResults                  []TxResult
	EndBlockEvents             []Event
	ConsensusMessagesResponses []Type
}

type TxResult struct {
	Events  []Event
	GasUsed uint64

	Resp  Type
	Error error
}

// STFAppManager is a struct that manages the state transition component of the app.
type STFAppManager struct {
	msgRouter   MsgRouter
	queryRouter QueryRouter

	doBeginBlock func(ctx context.Context) error
	doEndBlock   func(ctx context.Context) error

	doTxValidation func(ctx context.Context, tx Tx) error

	txDecoder TxDecoder

	store  Store
	branch func(store ReadonlyStore) BranchStore

	blockGasLimit uint64
}

// DeliverBlock is our state transition function.
func (s STFAppManager) DeliverBlock(ctx context.Context, block Block) (blockResult *BlockResponse, err error) {
	blockResult = new(BlockResponse)
	// create a new readonly view of the store in the new block.
	readonlyBlockStore := s.store.ReadonlyWithVersion(block.Height)
	// creates a new branch store, from the readonly view of the state
	// that can be written to.
	blockStore := s.branch(readonlyBlockStore)
	// begin block
	beginBlockEvents, err := s.beginBlock(ctx, blockStore)
	if err != nil {
		return nil, err
	}
	// execute txs
	txResults := make([]TxResult, len(block.Txs))
	for i, txBytes := range block.Txs {
		txResults[i] = s.deliverTx(ctx, blockStore, txBytes)
	}
	// end block
	endBlockEvents, err := s.endBlock(ctx, blockStore, block)
	if err != nil {
		return nil, err
	}

	// commit to storage
	changeSet, err := blockStore.ChangeSets()
	if err != nil {
		return nil, err
	}
	commitmentHash, err := s.store.CommitChanges(changeSet)
	if err != nil {
		return nil, err
	}
	return &BlockResponse{
		AppHash:          commitmentHash,
		BeginBlockEvents: beginBlockEvents,
		TxResults:        txResults,
		EndBlockEvents:   endBlockEvents,
	}, nil
}

func (s STFAppManager) beginBlock(ctx context.Context, store BranchStore) (beginBlockEvents []Event, err error) {
	execCtx := s.makeContext(ctx, store, 0) // TODO: gas limit
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

func (s STFAppManager) deliverTx(ctx context.Context, blockStore BranchStore, txBytes []byte) TxResult {
	// decode TX
	tx, err := s.txDecoder.Decode(txBytes)
	if err != nil {
		return TxResult{
			Error: err,
		}
	}
	// validate tx
	validateGas, validationEvents, err := s.validateTx(ctx, blockStore, tx.GetGasLimit(), tx)
	if err != nil {
		return TxResult{
			Error: err,
		}
	}
	// exec tx
	execResp, execGas, execEvents, err := s.execTx(ctx, blockStore, tx.GetGasLimit()-validateGas, tx)
	if err != nil {
		return TxResult{
			Events:  validationEvents,
			GasUsed: validateGas + execGas,
			Error:   err,
		}
	}
	return TxResult{
		Events:  append(validationEvents, execEvents...),
		GasUsed: execGas + validateGas,
		Resp:    execResp,
		Error:   nil,
	}
}

// validateTx validates a transaction given the provided BranchStore and gas limit.
// If the validation is successful, state is committed
func (s STFAppManager) validateTx(ctx context.Context, store BranchStore, gasLimit uint64, tx Tx) (gasUsed uint64, events []Event, err error) {
	validateCtx := s.makeContext(ctx, store, gasLimit)
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

func (s STFAppManager) execTx(ctx context.Context, store BranchStore, gasLimit uint64, tx Tx) (msgResp Type, gasUsed uint64, execEvents []Event, err error) {
	execCtx := s.makeContext(ctx, store, gasLimit)
	msgResp, err = s.msgRouter.Handle(ctx, tx.GetMessage())
	if err != nil {
		return nil, 0, nil, err
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

func (s STFAppManager) endBlock(ctx context.Context, store BranchStore, block Block) (endBlockEvents []Event, err error) {
	execCtx := s.makeContext(ctx, store, 0) // TODO: gas limit
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

func (s STFAppManager) makeContext(
	ctx context.Context,
	store BranchStore,
	gasLimit uint64,
) *executionContext {
	return &executionContext{
		Context:  ctx,
		store:    store,
		gasUsed:  0,
		gasLimit: gasLimit,
		events:   make([]Event, 0),
	}
}
