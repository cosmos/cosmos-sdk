package ibc

import (
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/certifiers"
	certerr "github.com/tendermint/light-client/certifiers/errors"

	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

const (
	prefixHash   = "v"
	prefixHeight = "h"
	prefixPacket = "p"
)

// newCertifier loads up the current state of this chain to make a proper certifier
// it will load the most recent height before block h if h is positive
// if h < 0, it will load the latest height
func newCertifier(store state.SimpleDB, chainID string, h int) (*certifiers.Inquiring, error) {
	// each chain has their own prefixed subspace
	p := newDBProvider(store)

	var fc certifiers.FullCommit
	var err error
	if h > 0 {
		// this gets the most recent verified commit below the specified height
		fc, err = p.GetByHeight(h)
	} else {
		// 0 or negative means start at latest commit
		fc, err = p.LatestCommit()
	}
	if err != nil {
		return nil, ErrHeaderNotFound(h)
	}

	// we have no source for untrusted keys, but use the db to load trusted history
	cert := certifiers.NewInquiring(chainID, fc, p,
		certifiers.NewMissingProvider())
	return cert, nil
}

// dbProvider wraps our kv store so it integrates with light-client verification
type dbProvider struct {
	byHash   state.SimpleDB
	byHeight *state.Span
}

func newDBProvider(store state.SimpleDB) *dbProvider {
	return &dbProvider{
		byHash:   stack.PrefixedStore(prefixHash, store),
		byHeight: state.NewSpan(stack.PrefixedStore(prefixHeight, store)),
	}
}

var _ certifiers.Provider = &dbProvider{}

func (d *dbProvider) StoreCommit(fc certifiers.FullCommit) error {
	// TODO: don't duplicate data....
	b := wire.BinaryBytes(fc)
	d.byHash.Set(fc.ValidatorsHash(), b)
	d.byHeight.Set(uint64(fc.Height()), b)
	return nil
}

func (d *dbProvider) LatestCommit() (fc certifiers.FullCommit, err error) {
	b, _ := d.byHeight.Top()
	if b == nil {
		return fc, certerr.ErrCommitNotFound()
	}
	err = wire.ReadBinaryBytes(b, &fc)
	return
}

func (d *dbProvider) GetByHeight(h int) (fc certifiers.FullCommit, err error) {
	b, _ := d.byHeight.LTE(uint64(h))
	if b == nil {
		return fc, certerr.ErrCommitNotFound()
	}
	err = wire.ReadBinaryBytes(b, &fc)
	return
}

func (d *dbProvider) GetByHash(hash []byte) (fc certifiers.FullCommit, err error) {
	b := d.byHash.Get(hash)
	if b == nil {
		return fc, certerr.ErrCommitNotFound()
	}
	err = wire.ReadBinaryBytes(b, &fc)
	return
}

// GetExactHeight is like GetByHeight, but returns an error instead of
// closest match if there is no exact match
func (d *dbProvider) GetExactHeight(h int) (fc certifiers.FullCommit, err error) {
	fc, err = d.GetByHeight(h)
	if err != nil {
		return
	}
	if fc.Height() != h {
		err = ErrHeaderNotFound(h)
	}
	return
}
