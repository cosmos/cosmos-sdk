package app

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/genesis"
	"github.com/cosmos/cosmos-sdk/modules/coin"
)

const genesisFilepath = "./testdata/genesis.json"
const genesisAcctFilepath = "./testdata/genesis2.json"

// 2b is just like 2, but add carl who has inconsistent
// pubkey and address
const genesisBadAcctFilepath = "./testdata/genesis2b.json"

func TestLoadGenesisDoNotFailIfAppOptionsAreMissing(t *testing.T) {
	logger := log.TestingLogger()
	store, err := MockStoreApp("genesis", logger)
	require.Nil(t, err, "%+v", err)
	app := NewBaseApp(store, DefaultHandler("mycoin"), nil)

	err = genesis.LoadGenesis(app, "./testdata/genesis3.json")
	require.Nil(t, err, "%+v", err)
}

func TestLoadGenesisFailsWithUnknownOptions(t *testing.T) {
	require := require.New(t)

	logger := log.TestingLogger()
	store, err := MockStoreApp("genesis", logger)
	require.Nil(err, "%+v", err)

	app := NewBaseApp(store, DefaultHandler("mycoin"), nil)
	err = genesis.LoadGenesis(app, genesisFilepath)
	require.NotNil(err, "%+v", err)
}

// Fix for issue #89, change the parse format for accounts in genesis.json
func TestLoadGenesisAccountAddress(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	logger := log.TestingLogger()
	store, err := MockStoreApp("genesis", logger)
	require.Nil(err, "%+v", err)
	app := NewBaseApp(store, DefaultHandler("mycoin"), nil)

	err = genesis.LoadGenesis(app, genesisAcctFilepath)
	require.Nil(err, "%+v", err)

	// check the chain id
	assert.Equal("addr_accounts_chain", app.GetChainID())

	// make sure the accounts were set properly
	cases := []struct {
		addr      string
		exists    bool
		hasPubkey bool
		coins     coin.Coins
	}{
		// this comes from a public key, should be stored proper (alice)
		{"62035D628DE7543332544AA60D90D3693B6AD51B", true, true, coin.Coins{{"one", 111}}},
		// this comes from an address, should be stored proper (bob)
		{"C471FB670E44D219EE6DF2FC284BE38793ACBCE1", true, false, coin.Coins{{"two", 222}}},
		// this comes from a secp256k1 public key, should be stored proper (sam)
		{"979F080B1DD046C452C2A8A250D18646C6B669D4", true, true, coin.Coins{{"four", 444}}},
	}

	for i, tc := range cases {
		addr, err := hex.DecodeString(tc.addr)
		require.Nil(err, tc.addr)
		coins, err := getAddr(addr, app.Append())
		require.Nil(err, "%+v", err)
		if !tc.exists {
			assert.True(coins.IsZero(), "%d", i)
		} else if assert.False(coins.IsZero(), "%d", i) {
			// it should and does exist...
			assert.True(coins.IsValid())
			assert.Equal(tc.coins, coins)
		}
	}
}

// When you define an account in genesis with address
// and pubkey that don't match
func TestLoadGenesisAccountInconsistentAddress(t *testing.T) {
	require := require.New(t)

	logger := log.TestingLogger()
	store, err := MockStoreApp("genesis", logger)
	require.Nil(err, "%+v", err)
	app := NewBaseApp(store, DefaultHandler("mycoin"), nil)
	err = genesis.LoadGenesis(app, genesisBadAcctFilepath)
	require.NotNil(err)
}
