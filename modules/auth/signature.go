package auth

import (
	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	NameSigs = "sigs"
)

// Signatures parses out go-crypto signatures and adds permissions to the
// context for use inside the application
type Signatures struct {
	stack.PassInitState
	stack.PassInitValidate
}

// Name of the module - fulfills Middleware interface
func (Signatures) Name() string {
	return NameSigs
}

var _ stack.Middleware = Signatures{}

// SigPerm takes the binary address from PubKey.Address and makes it an Actor
func SigPerm(addr []byte) sdk.Actor {
	return sdk.NewActor(NameSigs, addr)
}

// Signable allows us to use txs.OneSig and txs.MultiSig (and others??)
type Signable interface {
	sdk.TxLayer
	Signers() ([]crypto.PubKey, error)
}

// CheckTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.CheckTx(ctx2, store, tnext)
}

// DeliverTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.DeliverTx(ctx2, store, tnext)
}

func addSigners(ctx sdk.Context, sigs []crypto.PubKey) sdk.Context {
	perms := make([]sdk.Actor, len(sigs))
	for i, s := range sigs {
		perms[i] = SigPerm(s.Address())
	}
	// add the signers to the context and continue
	return ctx.WithPermissions(perms...)
}

func getSigners(tx sdk.Tx) ([]crypto.PubKey, sdk.Tx, error) {
	stx, ok := tx.Unwrap().(Signable)
	if !ok {
		return nil, sdk.Tx{}, errors.ErrUnauthorized()
	}
	sig, err := stx.Signers()
	return sig, stx.Next(), err
}
