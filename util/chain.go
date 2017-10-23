package util

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

var (
	// ChainPattern must match any valid chain_id
	ChainPattern = regexp.MustCompile("^[A-Za-z0-9_-]+$")
)

// ChainedTx interface should be implemented by any
// message to bind it to one chain, with an optional
// expiration height
type ChainedTx interface {
	GetChain() ChainData
}

// ChainData is info from the message to bind it to
// chain and time
type ChainData struct {
	ChainID   string `json:"chain_id"`
	ExpiresAt uint64 `json:"expires_at"`
}

// Chain enforces that this tx was bound to the named chain
type Chain struct{}

var _ sdk.Decorator = Chain{}

// CheckTx makes sure we are on the proper chain
// - fulfills Decorator interface
func (c Chain) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx interface{}, next sdk.Checker) (res sdk.CheckResult, err error) {

	err = c.checkChainTx(ctx.ChainID(), ctx.BlockHeight(), tx)
	if err != nil {
		return res, err
	}
	return next.CheckTx(ctx, store, tx)
}

// DeliverTx makes sure we are on the proper chain
// - fulfills Decorator interface
func (c Chain) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx interface{}, next sdk.Deliverer) (res sdk.DeliverResult, err error) {

	err = c.checkChainTx(ctx.ChainID(), ctx.BlockHeight(), tx)
	if err != nil {
		return res, err
	}
	return next.DeliverTx(ctx, store, tx)
}

// checkChainTx makes sure the tx is a ChainedTx,
// it is on the proper chain, and it has not expired.
func (c Chain) checkChainTx(chainID string, height uint64, tx interface{}) error {
	// make sure it is a chaintx
	ctx, ok := tx.(ChainedTx)
	if !ok {
		return ErrNoChain()
	}

	data := ctx.GetChain()

	// compare against state
	if data.ChainID != chainID {
		return ErrWrongChain(data.ChainID)
	}
	if data.ExpiresAt != 0 && data.ExpiresAt <= height {
		return ErrExpired()
	}
	return nil
}
