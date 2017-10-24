package auth

import (
	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
)

//nolint
const (
	NameSigs = "sigs"
)

// Signatures parses out go-crypto signatures and adds permissions to the
// context for use inside the application
type Signatures struct {
}

var _ sdk.Decorator = Signatures{}

// SigPerm takes the binary address from PubKey.Address and makes it an Actor
func SigPerm(addr []byte) sdk.Actor {
	return sdk.NewActor(NameSigs, addr)
}

// Signable allows us to use txs.OneSig and txs.MultiSig (and others??)
type Signable interface {
	Signers() ([]crypto.PubKey, error)
}

// CheckTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) CheckTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Checker) (res sdk.CheckResult, err error) {

	sigs, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.CheckTx(ctx2, store, tx)
}

// DeliverTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) DeliverTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Deliverer) (res sdk.DeliverResult, err error) {

	sigs, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.DeliverTx(ctx2, store, tx)
}

func addSigners(ctx sdk.Context, sigs []crypto.PubKey) sdk.Context {
	perms := make([]sdk.Actor, len(sigs))
	for i, s := range sigs {
		perms[i] = SigPerm(s.Address())
	}
	// add the signers to the context and continue
	return ctx.WithPermissions(perms...)
}

func getSigners(tx interface{}) ([]crypto.PubKey, error) {
	stx, ok := tx.(Signable)
	if !ok {
		return nil, errors.ErrUnauthorized()
	}
	sig, err := stx.Signers()
	return sig, err
}
