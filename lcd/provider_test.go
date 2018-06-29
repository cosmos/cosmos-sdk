// nolint: vetshadow
package lcd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/lcd"
	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
)

// missingProvider doesn't store anything, always a miss
// Designed as a mock for testing
type missingProvider struct{}

// NewMissingProvider returns a provider which does not store anything and always misses.
func NewMissingProvider() lcd.Provider {
	return missingProvider{}
}

func (missingProvider) StoreCommit(lcd.FullCommit) error { return nil }
func (missingProvider) GetByHeight(int64) (lcd.FullCommit, error) {
	return lcd.FullCommit{}, lcdErr.ErrCommitNotFound()
}
func (missingProvider) GetByHash([]byte) (lcd.FullCommit, error) {
	return lcd.FullCommit{}, lcdErr.ErrCommitNotFound()
}
func (missingProvider) LatestCommit() (lcd.FullCommit, error) {
	return lcd.FullCommit{}, lcdErr.ErrCommitNotFound()
}

func TestMemProvider(t *testing.T) {
	p := lcd.NewMemStoreProvider()
	checkProvider(t, p, "test-mem", "empty")
}

func TestCacheProvider(t *testing.T) {
	p := lcd.NewCacheProvider(
		NewMissingProvider(),
		lcd.NewMemStoreProvider(),
		NewMissingProvider(),
	)
	checkProvider(t, p, "test-cache", "kjfhekfhkewhgit")
}

func checkProvider(t *testing.T, p lcd.Provider, chainID, app string) {
	assert, require := assert.New(t), require.New(t)
	appHash := []byte(app)
	keys := lcd.GenValKeys(5)
	count := 10

	// make a bunch of commits...
	commits := make([]lcd.FullCommit, count)
	for i := 0; i < count; i++ {
		// two commits for each validator, to check how we handle dups
		// (10, 0), (10, 1), (10, 1), (10, 2), (10, 2), ...
		vals := keys.ToValidators(10, int64(count/2))
		h := int64(20 + 10*i)
		commits[i] = keys.GenFullCommit(chainID, h, nil, vals, appHash, []byte("params"), []byte("results"), 0, 5)
	}

	// check provider is empty
	fc, err := p.GetByHeight(20)
	require.NotNil(err)
	assert.True(lcdErr.IsCommitNotFoundErr(err))

	fc, err = p.GetByHash(commits[3].ValidatorsHash())
	require.NotNil(err)
	assert.True(lcdErr.IsCommitNotFoundErr(err))

	// now add them all to the provider
	for _, s := range commits {
		err = p.StoreCommit(s)
		require.Nil(err)
		// and make sure we can get it back
		s2, err := p.GetByHash(s.ValidatorsHash())
		assert.Nil(err)
		assert.Equal(s, s2)
		// by height as well
		s2, err = p.GetByHeight(s.Height())
		assert.Nil(err)
		assert.Equal(s, s2)
	}

	// make sure we get the last hash if we overstep
	fc, err = p.GetByHeight(5000)
	if assert.Nil(err) {
		assert.Equal(commits[count-1].Height(), fc.Height())
		assert.Equal(commits[count-1], fc)
	}

	// and middle ones as well
	fc, err = p.GetByHeight(47)
	if assert.Nil(err) {
		// we only step by 10, so 40 must be the one below this
		assert.EqualValues(40, fc.Height())
	}

}

// this will make a get height, and if it is good, set the data as well
func checkGetHeight(t *testing.T, p lcd.Provider, ask, expect int64) {
	fc, err := p.GetByHeight(ask)
	require.Nil(t, err, "GetByHeight")
	if assert.Equal(t, expect, fc.Height()) {
		err = p.StoreCommit(fc)
		require.Nil(t, err, "StoreCommit")
	}
}

func TestCacheGetsBestHeight(t *testing.T) {
	// assert, require := assert.New(t), require.New(t)
	require := require.New(t)

	// we will write data to the second level of the cache (p2),
	// and see what gets cached, stored in
	p := lcd.NewMemStoreProvider()
	p2 := lcd.NewMemStoreProvider()
	cp := lcd.NewCacheProvider(p, p2)

	chainID := "cache-best-height"
	appHash := []byte("01234567")
	keys := lcd.GenValKeys(5)
	count := 10

	// set a bunch of commits
	for i := 0; i < count; i++ {
		vals := keys.ToValidators(10, int64(count/2))
		h := int64(10 * (i + 1))
		fc := keys.GenFullCommit(chainID, h, nil, vals, appHash, []byte("params"), []byte("results"), 0, 5)
		err := p2.StoreCommit(fc)
		require.NoError(err)
	}

	// let's get a few heights from the cache and set them proper
	checkGetHeight(t, cp, 57, 50)
	checkGetHeight(t, cp, 33, 30)

	// make sure they are set in p as well (but nothing else)
	checkGetHeight(t, p, 44, 30)
	checkGetHeight(t, p, 50, 50)
	checkGetHeight(t, p, 99, 50)

	// now, query the cache for a higher value
	checkGetHeight(t, p2, 99, 90)
	checkGetHeight(t, cp, 99, 90)
}
