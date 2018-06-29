package files_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/lcd"
	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
	"github.com/cosmos/cosmos-sdk/lcd/files"
)

func checkEqual(stored, loaded lcd.FullCommit, chainID string) error {
	err := loaded.ValidateBasic(chainID)
	if err != nil {
		return err
	}
	if !bytes.Equal(stored.ValidatorsHash(), loaded.ValidatorsHash()) {
		return errors.New("Different block hashes")
	}
	return nil
}

func TestFileProvider(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	dir, err := ioutil.TempDir("", "fileprovider-test")
	assert.Nil(err)
	defer os.RemoveAll(dir)
	p := files.NewProvider(dir)

	chainID := "test-files"
	appHash := []byte("some-data")
	keys := lcd.GenValKeys(5)
	count := 10

	// make a bunch of seeds...
	seeds := make([]lcd.FullCommit, count)
	for i := 0; i < count; i++ {
		// two seeds for each validator, to check how we handle dups
		// (10, 0), (10, 1), (10, 1), (10, 2), (10, 2), ...
		vals := keys.ToValidators(10, int64(count/2))
		h := int64(20 + 10*i)
		check := keys.GenCommit(chainID, h, nil, vals, appHash, []byte("params"), []byte("results"), 0, 5)
		seeds[i] = lcd.NewFullCommit(check, vals)
	}

	// check provider is empty
	seed, err := p.GetByHeight(20)
	require.NotNil(err)
	assert.True(lcdErr.IsCommitNotFoundErr(err))

	seed, err = p.GetByHash(seeds[3].ValidatorsHash())
	require.NotNil(err)
	assert.True(lcdErr.IsCommitNotFoundErr(err))

	// now add them all to the provider
	for _, s := range seeds {
		err = p.StoreCommit(s)
		require.Nil(err)
		// and make sure we can get it back
		s2, err := p.GetByHash(s.ValidatorsHash())
		assert.Nil(err)
		err = checkEqual(s, s2, chainID)
		assert.Nil(err)
		// by height as well
		s2, err = p.GetByHeight(s.Height())
		err = checkEqual(s, s2, chainID)
		assert.Nil(err)
	}

	// make sure we get the last hash if we overstep
	seed, err = p.GetByHeight(5000)
	if assert.Nil(err, "%+v", err) {
		assert.Equal(seeds[count-1].Height(), seed.Height())
		err = checkEqual(seeds[count-1], seed, chainID)
		assert.Nil(err)
	}

	// and middle ones as well
	seed, err = p.GetByHeight(47)
	if assert.Nil(err, "%+v", err) {
		// we only step by 10, so 40 must be the one below this
		assert.EqualValues(40, seed.Height())
	}

	// and proper error for too low
	_, err = p.GetByHeight(5)
	assert.NotNil(err)
	assert.True(lcdErr.IsCommitNotFoundErr(err))
}
