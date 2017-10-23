package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/state"
	"github.com/tendermint/light-client/certifiers"
	certerr "github.com/tendermint/light-client/certifiers/errors"
)

func assertSeedEqual(t *testing.T, s, s2 certifiers.FullCommit) {
	assert := assert.New(t)
	assert.Equal(s.Height(), s2.Height())
	assert.Equal(s.ValidatorsHash(), s2.ValidatorsHash())
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
	assert.True(certerr.IsCommitNotFoundErr(err))

	// add a seed
	for _, s := range seeds {
		err = p.StoreCommit(s)
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
	assert.True(certerr.IsCommitNotFoundErr(err))

	// make sure we get highest
	val, err = p.LatestCommit()
	if assert.Nil(err) {
		assertSeedEqual(t, seeds[3], val)
	}

	// make sure by hash also (note all have same hash, so overwritten)
	val, err = p.GetByHash(seeds[1].ValidatorsHash())
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
		seeds[i] = keys.GenFullCommit(chainID, h, nil, vals, appHash, 0, len(keys))
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
	assert.True(certerr.IsCommitNotFoundErr(err))

	seed, err = p.GetByHash(seeds[3].ValidatorsHash())
	require.NotNil(err)
	assert.True(certerr.IsCommitNotFoundErr(err))

	// now add them all to the provider
	for _, s := range seeds {
		err = p.StoreCommit(s)
		require.Nil(err)
		// and make sure we can get it back
		s2, err := p.GetByHash(s.ValidatorsHash())
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
