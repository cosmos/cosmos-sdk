package ibc

import (
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/certifiers"

	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

const (
	prefixHash   = "v"
	prefixHeight = "h"
	prefixPacket = "p"
)

// newCertifier loads up the current state of this chain to make a proper
func newCertifier(chainID string, store state.KVStore) (*certifiers.InquiringCertifier, error) {
	// each chain has their own prefixed subspace
	space := stack.PrefixedStore(chainID, store)
	p := newDBProvider(space)

	// this gets the most recent verified seed
	seed, err := certifiers.LatestSeed(p)
	if err != nil {
		return nil, err
	}

	// we have no source for untrusted keys, but use the db to load trusted history
	cert := certifiers.NewInquiring(chainID, seed.Validators, p,
		certifiers.MissingProvider{})
	return cert, nil
}

// dbProvider wraps our kv store so it integrates with light-client verification
type dbProvider struct {
	byHash   state.KVStore
	byHeight *state.Span
}

func newDBProvider(store state.KVStore) *dbProvider {
	return &dbProvider{
		byHash:   stack.PrefixedStore(prefixHash, store),
		byHeight: state.NewSpan(stack.PrefixedStore(prefixHeight, store)),
	}
}

var _ certifiers.Provider = &dbProvider{}

func (d *dbProvider) StoreSeed(seed certifiers.Seed) error {
	// TODO: don't duplicate data....
	b := wire.BinaryBytes(seed)
	d.byHash.Set(seed.Hash(), b)
	d.byHeight.Set(uint64(seed.Height()), b)
	return nil
}

func (d *dbProvider) GetByHeight(h int) (seed certifiers.Seed, err error) {
	b, _ := d.byHeight.LTE(uint64(h))
	if b == nil {
		return seed, certifiers.ErrSeedNotFound()
	}
	err = wire.ReadBinaryBytes(b, &seed)
	return
}

func (d *dbProvider) GetByHash(hash []byte) (seed certifiers.Seed, err error) {
	b := d.byHash.Get(hash)
	if b == nil {
		return seed, certifiers.ErrSeedNotFound()
	}
	err = wire.ReadBinaryBytes(b, &seed)
	return
}
