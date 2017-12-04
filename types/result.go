package types

import (
	abci "github.com/tendermint/abci/types"
)

type KVPair struct {
	Key   []byte
	Value []byte
}

// Result is the union of ResponseDeliverTx and ResponseCheckTx.
type Result struct {

	// Code is the response code, is stored back on the chain.
	Code uint32

	// Data is any data returned from the app.
	Data []byte

	// Log is just debug information. NOTE: nondeterministic.
	Log string

	// GasAllocated is the maximum units of work we allow this tx to perform.
	GasAllocated int64

	// GasUsed is the amount of gas actually consumed. NOTE: not used.
	GasUsed int64

	// Tx fee amount and denom.
	FeeAmount int64
	FeeDenom  string

	// Changes to the validator set.
	ValSetDiff []abci.Validator

	// Tags are used for transaction indexing and pubsub.
	Tags []KVPair
}
