package auth

import (
	crypto "github.com/tendermint/go-crypto"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

//nolint
const (
	NameSigs = "sigs"
)

// Signatures parses out go-crypto signatures and adds permissions to the
// context for use inside the application
type Signatures struct {
	stack.PassOption
}

// Name of the module - fulfills Middleware interface
func (Signatures) Name() string {
	return NameSigs
}

var _ stack.Middleware = Signatures{}

// SigPerm takes the binary address from PubKey.Address and makes it an Actor
func SigPerm(addr []byte) basecoin.Actor {
	return basecoin.NewActor(NameSigs, addr)
}

// Signable allows us to use txs.OneSig and txs.MultiSig (and others??)
type Signable interface {
	basecoin.TxLayer
	Signers() ([]crypto.PubKey, error)
}

// CheckTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Checker) (res basecoin.CheckResult, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.CheckTx(ctx2, store, tnext)
}

// DeliverTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.DeliverTx(ctx2, store, tnext)
}

func addSigners(ctx basecoin.Context, sigs []crypto.PubKey) basecoin.Context {
	perms := make([]basecoin.Actor, len(sigs))
	for i, s := range sigs {
		perms[i] = SigPerm(s.Address())
	}
	// add the signers to the context and continue
	return ctx.WithPermissions(perms...)
}

func getSigners(tx basecoin.Tx) ([]crypto.PubKey, basecoin.Tx, error) {
	stx, ok := tx.Unwrap().(Signable)
	if !ok {
		return nil, basecoin.Tx{}, errors.ErrUnauthorized()
	}
	sig, err := stx.Signers()
	return sig, stx.Next(), err
}
