package ibc

import (
	"github.com/tendermint/light-client/certifiers"

	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// newCertifier loads up the current state of this chain to make a proper
func newCertifier(chainID string, store state.KVStore) (*certifiers.InquiringCertifier, error) {
	// each chain has their own prefixed subspace
	space := stack.PrefixedStore(chainID, store)
	p := dbProvider{space}

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
	store state.KVStore
}

var _ certifiers.Provider = dbProvider{}

func (d dbProvider) StoreSeed(seed certifiers.Seed) error {
	return nil
}

func (d dbProvider) GetByHeight(h int) (certifiers.Seed, error) {
	return certifiers.Seed{}, certifiers.ErrSeedNotFound()
}
func (d dbProvider) GetByHash(hash []byte) (certifiers.Seed, error) {
	return certifiers.Seed{}, certifiers.ErrSeedNotFound()
}
