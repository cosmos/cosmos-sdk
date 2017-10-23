package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/state"
	"github.com/tendermint/light-client/certifiers"
	certerr "github.com/tendermint/light-client/certifiers/errors"
)

func assertCommitsEqual(t *testing.T, fc, fc2 certifiers.FullCommit) {
	assert := assert.New(t)
	assert.Equal(fc.Height(), fc2.Height())
	assert.Equal(fc.ValidatorsHash(), fc2.ValidatorsHash())
	// TODO: more
}

func TestProviderStore(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// make a few commits
	keys := certifiers.GenValKeys(2)
	commits := makeCommits(keys, 4, "some-chain", "demo-store")

	// make a provider
	store := state.NewMemKVStore()
	p := newDBProvider(store)

	// check it...
	_, err := p.GetByHeight(20)
	require.NotNil(err)
	assert.True(certerr.IsCommitNotFoundErr(err))

	// add commits
	for _, fc := range commits {
		err = p.StoreCommit(fc)
		require.Nil(err)
	}

	// make sure we get it...
	fc := commits[0]
	val, err := p.GetByHeight(fc.Height())
	if assert.Nil(err) {
		assertCommitsEqual(t, fc, val)
	}

	// make sure we get higher
	val, err = p.GetByHeight(fc.Height() + 2)
	if assert.Nil(err) {
		assertCommitsEqual(t, fc, val)
	}

	// below is nothing
	_, err = p.GetByHeight(fc.Height() - 2)
	assert.True(certerr.IsCommitNotFoundErr(err))

	// make sure we get highest
	val, err = p.LatestCommit()
	if assert.Nil(err) {
		assertCommitsEqual(t, commits[3], val)
	}

	// make sure by hash also (note all have same hash, so overwritten)
	val, err = p.GetByHash(commits[1].ValidatorsHash())
	if assert.Nil(err) {
		assertCommitsEqual(t, commits[3], val)
	}
}

func TestDBProvider(t *testing.T) {
	store := state.NewMemKVStore()
	p := newDBProvider(store)
	checkProvider(t, p, "test-db", "bling")
}

func makeCommits(keys certifiers.ValKeys, count int, chainID, app string) []certifiers.FullCommit {
	appHash := []byte(app)
	commits := make([]certifiers.FullCommit, count)
	for i := 0; i < count; i++ {
		// two commits for each validator, to check how we handle dups
		// (10, 0), (10, 1), (10, 1), (10, 2), (10, 2), ...
		vals := keys.ToValidators(10, int64(count/2))
		h := 20 + 10*i
		commits[i] = keys.GenFullCommit(chainID, h, nil, vals, appHash, 0, len(keys))
	}
	return commits
}

func checkProvider(t *testing.T, p certifiers.Provider, chainID, app string) {
	assert, require := assert.New(t), require.New(t)
	keys := certifiers.GenValKeys(5)
	count := 10

	// make a bunch of commits...
	commits := makeCommits(keys, count, chainID, app)

	// check provider is empty
	fc, err := p.GetByHeight(20)
	require.NotNil(err)
	assert.True(certerr.IsCommitNotFoundErr(err))

	fc, err = p.GetByHash(commits[3].ValidatorsHash())
	require.NotNil(err)
	assert.True(certerr.IsCommitNotFoundErr(err))

	// now add them all to the provider
	for _, fc := range commits {
		err = p.StoreCommit(fc)
		require.Nil(err)
		// and make sure we can get it back
		fc2, err := p.GetByHash(fc.ValidatorsHash())
		assert.Nil(err)
		assertCommitsEqual(t, fc, fc2)
		// by height as well
		fc2, err = p.GetByHeight(fc.Height())
		assert.Nil(err)
		assertCommitsEqual(t, fc, fc2)
	}

	// make sure we get the last hash if we overstep
	fc, err = p.GetByHeight(5000)
	if assert.Nil(err) {
		assertCommitsEqual(t, commits[count-1], fc)
	}

	// and middle ones as well
	fc, err = p.GetByHeight(47)
	if assert.Nil(err) {
		// we only step by 10, so 40 must be the one below this
		assert.Equal(40, fc.Height())
	}

}
