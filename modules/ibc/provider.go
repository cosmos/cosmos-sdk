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

// newCertifier loads up the current state of this chain to make a proper certifier
// it will load the most recent height before block h if h is positive
// if h < 0, it will load the latest height
func newCertifier(store state.SimpleDB, chainID string, h int) (*certifiers.InquiringCertifier, error) {
	// each chain has their own prefixed subspace
	p := newDBProvider(store)

	var seed certifiers.Seed
	var err error
	if h > 0 {
		// this gets the most recent verified seed below the specified height
		seed, err = p.GetByHeight(h)
	} else {
		// 0 or negative means start at latest seed
		seed, err = certifiers.LatestSeed(p)
	}
	if err != nil {
		return nil, ErrHeaderNotFound(h)
	}

	// we have no source for untrusted keys, but use the db to load trusted history
	cert := certifiers.NewInquiring(chainID, seed, p,
		certifiers.MissingProvider{})
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

// GetExactHeight is like GetByHeight, but returns an error instead of
// closest match if there is no exact match
func (d *dbProvider) GetExactHeight(h int) (seed certifiers.Seed, err error) {
	seed, err = d.GetByHeight(h)
	if err != nil {
		return
	}
	if seed.Height() != h {
		err = ErrHeaderNotFound(h)
	}
	return
}
