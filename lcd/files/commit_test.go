package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/cosmos/cosmos-sdk/lcd"
)

func tmpFile() string {
	suffix := cmn.RandStr(16)
	return filepath.Join(os.TempDir(), "fc-test-"+suffix)
}

func TestSerializeFullCommits(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// some constants
	appHash := []byte("some crazy thing")
	chainID := "ser-ial"
	h := int64(25)

	// build a fc
	keys := lcd.GenValKeys(5)
	vals := keys.ToValidators(10, 0)
	fc := keys.GenFullCommit(chainID, h, nil, vals, appHash, []byte("params"), []byte("results"), 0, 5)

	require.Equal(h, fc.Height())
	require.Equal(vals.Hash(), fc.ValidatorsHash())

	// try read/write with json
	jfile := tmpFile()
	defer os.Remove(jfile)
	jseed, err := LoadFullCommitJSON(jfile)
	assert.NotNil(err)
	err = SaveFullCommitJSON(fc, jfile)
	require.Nil(err)
	jseed, err = LoadFullCommitJSON(jfile)
	assert.Nil(err, "%+v", err)
	assert.Equal(h, jseed.Height())
	assert.Equal(vals.Hash(), jseed.ValidatorsHash())

	// try read/write with binary
	bfile := tmpFile()
	defer os.Remove(bfile)
	bseed, err := LoadFullCommit(bfile)
	assert.NotNil(err)
	err = SaveFullCommit(fc, bfile)
	require.Nil(err)
	bseed, err = LoadFullCommit(bfile)
	assert.Nil(err, "%+v", err)
	assert.Equal(h, bseed.Height())
	assert.Equal(vals.Hash(), bseed.ValidatorsHash())

	// make sure they don't read the other format (different)
	_, err = LoadFullCommit(jfile)
	assert.NotNil(err)
	_, err = LoadFullCommitJSON(bfile)
	assert.NotNil(err)
}
