package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/state"
	"github.com/tendermint/light-client/certifiers"
)

func assertSeedEqual(t *testing.T, s, s2 certifiers.FullCommit) {
	assert := assert.New(t)
	assert.Equal(s.Height(), s2.Height())
	assert.Equal(s.Hash(), s2.Hash())
	// TODO: more
}

func TestProviderStore(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// make a few seeds
	keys := certifiers.GenValKeys(2)
	seeds := makeSeeds(keys, 4, "some-chain", "demo-store")

	// make a provider
	store := state.NewMemKVStore()
	p := newDBProvider(store)

	// check it...
	_, err := p.GetByHeight(20)
	require.NotNil(err)
	assert.True(certifiers.IsSeedNotFoundErr(err))

	// add a seed
	for _, s := range seeds {
		err = p.StoreSeed(s)
		require.Nil(err)
	}

	// make sure we get it...
	s := seeds[0]
	val, err := p.GetByHeight(s.Height())
	if assert.Nil(err) {
		assertSeedEqual(t, s, val)
	}

	// make sure we get higher
	val, err = p.GetByHeight(s.Height() + 2)
	if assert.Nil(err) {
		assertSeedEqual(t, s, val)
	}

	// below is nothing
	_, err = p.GetByHeight(s.Height() - 2)
	assert.True(certifiers.IsSeedNotFoundErr(err))

	// make sure we get highest
	val, err = certifiers.LatestSeed(p)
	if assert.Nil(err) {
		assertSeedEqual(t, seeds[3], val)
	}

	// make sure by hash also (note all have same hash, so overwritten)
	val, err = p.GetByHash(seeds[1].Hash())
	if assert.Nil(err) {
		assertSeedEqual(t, seeds[3], val)
	}
}

func TestDBProvider(t *testing.T) {
	store := state.NewMemKVStore()
	p := newDBProvider(store)
	checkProvider(t, p, "test-db", "bling")
}

func makeSeeds(keys certifiers.ValKeys, count int, chainID, app string) []certifiers.FullCommit {
	appHash := []byte(app)
	seeds := make([]certifiers.FullCommit, count)
	for i := 0; i < count; i++ {
		// two seeds for each validator, to check how we handle dups
		// (10, 0), (10, 1), (10, 1), (10, 2), (10, 2), ...
		vals := keys.ToValidators(10, int64(count/2))
		h := 20 + 10*i
		check := keys.GenCheckpoint(chainID, h, nil, vals, appHash, 0, len(keys))
		seeds[i] = certifiers.FullCommit{check, vals}
	}
	return seeds
}

func checkProvider(t *testing.T, p certifiers.Provider, chainID, app string) {
	assert, require := assert.New(t), require.New(t)
	keys := certifiers.GenValKeys(5)
	count := 10

	// make a bunch of seeds...
	seeds := makeSeeds(keys, count, chainID, app)

	// check provider is empty
	seed, err := p.GetByHeight(20)
	require.NotNil(err)
	assert.True(certifiers.IsSeedNotFoundErr(err))

	seed, err = p.GetByHash(seeds[3].Hash())
	require.NotNil(err)
	assert.True(certifiers.IsSeedNotFoundErr(err))

	// now add them all to the provider
	for _, s := range seeds {
		err = p.StoreSeed(s)
		require.Nil(err)
		// and make sure we can get it back
		s2, err := p.GetByHash(s.Hash())
		assert.Nil(err)
		assertSeedEqual(t, s, s2)
		// by height as well
		s2, err = p.GetByHeight(s.Height())
		assert.Nil(err)
		assertSeedEqual(t, s, s2)
	}

	// make sure we get the last hash if we overstep
	seed, err = p.GetByHeight(5000)
	if assert.Nil(err) {
		assertSeedEqual(t, seeds[count-1], seed)
	}

	// and middle ones as well
	seed, err = p.GetByHeight(47)
	if assert.Nil(err) {
		// we only step by 10, so 40 must be the one below this
		assert.Equal(40, seed.Height())
	}

}
