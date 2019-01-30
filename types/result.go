package types

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Result is the union of ResponseDeliverTx and ResponseCheckTx.
type Result struct {

	// Code is the response code, is stored back on the chain.
	Code CodeType

	// Codespace is the string referring to the domain of an error
	Codespace CodespaceType

	// Data is any data returned from the app.
	Data []byte

	// Log is just debug information. NOTE: nondeterministic.
	Log string

	// GasWanted is the maximum units of work we allow this tx to perform.
	GasWanted uint64

	// GasUsed is the amount of gas actually consumed. NOTE: unimplemented
	GasUsed uint64

	// Tx fee amount and denom.
	FeeAmount int64
	FeeDenom  string

	// Tags are used for transaction indexing and pubsub.
	Tags Tags
}

// TODO: In the future, more codes may be OK.
func (res Result) IsOK() bool {
	return res.Code.IsOK()
}

// Is a version of ResponseDeliverTx where the tags are StringTags rather than []byte tags
type StringResponseDeliverTx struct {
	Height    int64      `json:"height"`
	TxHash    string     `json:"txhash"`
	Code      uint32     `json:"code,omitempty"`
	Data      []byte     `json:"data,omitempty"`
	Log       string     `json:"log,omitempty"`
	Info      string     `json:"info,omitempty"`
	GasWanted int64      `json:"gas_wanted,omitempty"`
	GasUsed   int64      `json:"gas_used,omitempty"`
	Tags      StringTags `json:"tags,omitempty"`
	Codespace string     `json:"codespace,omitempty"`
}

func NewStringResponseDeliverTx(res *ctypes.ResultBroadcastTxCommit) StringResponseDeliverTx {
	return StringResponseDeliverTx{
		Height:    res.Height,
		TxHash:    res.Hash.String(),
		Code:      res.DeliverTx.Code,
		Data:      res.DeliverTx.Data,
		Log:       res.DeliverTx.Log,
		Info:      res.DeliverTx.Info,
		GasWanted: res.DeliverTx.GasWanted,
		GasUsed:   res.DeliverTx.GasUsed,
		Tags:      TagsToStringTags(res.DeliverTx.Tags),
		Codespace: res.DeliverTx.Codespace,
	}
}
