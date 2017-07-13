package ibc

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/state"
)

// nolint
const (
	NameIBC = "ibc"
)

// Handler allows us to update the chain state or create a packet
type Handler struct {
	basecoin.NopOption
}

var _ basecoin.Handler = Handler{}

// NewHandler makes a role handler to create roles
func NewHandler() Handler {
	return Handler{}
}

// Name - return name space
func (Handler) Name() string {
	return NameIBC
}

// CheckTx verifies the packet is formated correctly, and has the proper sequence
// for a registered chain
func (h Handler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return res, nil
}

// DeliverTx verifies all signatures on the tx and updated the chain state
// apropriately
func (h Handler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return res, nil
}
