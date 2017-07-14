package ibc

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// nolint
const (
	NameIBC = "ibc"
)

// Handler allows us to update the chain state or create a packet
//
// TODO: require auth for registration, the authorized actor (or role)
// should be defined in the handler, and set via SetOption
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
	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case RegisterChainTx:
		return h.initSeed(ctx, store, t)
	case UpdateChainTx:
		return h.updateSeed(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// DeliverTx verifies all signatures on the tx and updated the chain state
// apropriately
func (h Handler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case RegisterChainTx:
		// TODO: do we want some permissioning for this???
		return h.initSeed(ctx, store, t)
	case UpdateChainTx:
		return h.updateSeed(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// initSeed imports the first seed for this chain and accepts it as the root of trust
func (h Handler) initSeed(ctx basecoin.Context, store state.KVStore,
	t RegisterChainTx) (res basecoin.Result, err error) {

	chainID := t.ChainID()
	s := NewChainSet(store)
	err = s.Register(chainID, ctx.BlockHeight(), t.Seed.Height())
	if err != nil {
		return res, err
	}

	space := stack.PrefixedStore(chainID, store)
	provider := newDBProvider(space)
	err = provider.StoreSeed(t.Seed)
	return res, err
}

// updateSeed checks the seed against the existing chain data and rejects it if it
// doesn't fit (or no chain data)
func (h Handler) updateSeed(ctx basecoin.Context, store state.KVStore,
	t UpdateChainTx) (res basecoin.Result, err error) {

	chainID := t.ChainID()
	if !NewChainSet(store).Exists([]byte(chainID)) {
		return res, ErrNotRegistered(chainID)
	}

	// load the certifier for this chain
	seed := t.Seed
	space := stack.PrefixedStore(chainID, store)
	cert, err := newCertifier(space, chainID, seed.Height())
	if err != nil {
		return res, err
	}

	// this will import the seed if it is valid in the current context
	err = cert.Update(seed.Checkpoint, seed.Validators)
	return res, err
}
