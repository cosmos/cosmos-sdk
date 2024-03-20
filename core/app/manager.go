package app

import (
	"time"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
)

// QueryRequest represents a request to query data from the application.
type QueryRequest struct {
	Height int64  // The height at which to query the data.
	Path   string // The path to the data being queried.
	Data   []byte // The data to be queried.
}

// QueryResponse represents the response structure for a query.
type QueryResponse struct {
	Height int64  // The height of the query response.
	Value  []byte // The value of the query response.
}

// BlockRequest is a generic struct representing a request for a block.
// It contains information such as the block height, timestamp, hash, transactions,
// and consensus messages.
type BlockRequest[T transaction.Tx] struct {
	Height            uint64             // The height of the block.
	Time              time.Time          // The timestamp of the block.
	Hash              []byte             // The hash of the block.
	Txs               []T                // The transactions in the block.
	ConsensusMessages []transaction.Type // The consensus messages associated with the block.
}

// BlockResponse represents the response data structure returned by a block.
type BlockResponse struct {
	Apphash          []byte                        // Apphash is the application-specific hash of the block.
	ValidatorUpdates []appmodulev2.ValidatorUpdate // ValidatorUpdates contains the updates to the validators for the next block.
	PreBlockEvents   []event.Event                 // PreBlockEvents contains the events emitted before processing the block.
	BeginBlockEvents []event.Event                 // BeginBlockEvents contains the events emitted at the beginning of processing the block.
	TxResults        []TxResult                    // TxResults contains the results of executing transactions in the block.
	EndBlockEvents   []event.Event                 // EndBlockEvents contains the events emitted at the end of processing the block.
}

// RequestInitChain represents a request to initialize the chain.
type RequestInitChain struct {
	Time          time.Time                     // The time at which the request is made.
	ChainId       string                        // The ID of the chain.
	Validators    []appmodulev2.ValidatorUpdate // The updates to the validators.
	AppStateBytes []byte                        // The initial application state in bytes.
	InitialHeight int64                         // The initial height of the chain.
}

// ResponseInitChain represents the response returned by the InitChain method of the application manager.
type ResponseInitChain struct {
	Validators []appmodulev2.ValidatorUpdate // Validators contains the updates to the validators set.
	AppHash    []byte                        // AppHash contains the application hash.
}

// TxResult represents the result of a transaction execution.
type TxResult struct {
	Events    []event.Event      // Events emitted during the execution of the transaction.
	GasUsed   uint64             // Amount of gas used during the execution of the transaction.
	GasWanted uint64             // Amount of gas requested for the execution of the transaction.
	Resp      []transaction.Type // Response data returned by the transaction.
	Error     error              // Error occurred during the execution of the transaction, if any.
}
