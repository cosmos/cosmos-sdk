package app

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/basecoin/types"
	crypto "github.com/tendermint/go-crypto"
	eyescli "github.com/tendermint/merkleeyes/client"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"
)

const genesisFilepath = "./testdata/genesis.json"
const genesisAcctFilepath = "./testdata/genesis2.json"

func TestLoadGenesisDoNotFailIfAppOptionsAreMissing(t *testing.T) {
	eyesCli := eyescli.NewLocalClient("", 0)
	app := NewBasecoin(eyesCli, log.TestingLogger())
	err := app.LoadGenesis("./testdata/genesis3.json")
	require.Nil(t, err, "%+v", err)
}

func TestLoadGenesis(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	eyesCli := eyescli.NewLocalClient("", 0)
	app := NewBasecoin(eyesCli, log.TestingLogger())
	err := app.LoadGenesis(genesisFilepath)
	require.Nil(err, "%+v", err)

	// check the chain id
	assert.Equal("foo_bar_chain", app.GetState().GetChainID())

	// and check the account info - previously calculated values
	addr, _ := hex.DecodeString("eb98e0688217cfdeb70eddf4b33cdcc37fc53197")
	pkbyte, _ := hex.DecodeString("6880db93598e283a67c4d88fc67a8858aa2de70f713fe94a5109e29c137100c2")

	acct := app.GetState().GetAccount(addr)
	require.NotNil(acct)

	// make sure balance is proper
	assert.Equal(2, len(acct.Balance))
	assert.True(acct.Balance.IsValid())
	// note, that we now sort them to be valid
	assert.EqualValues(654321, acct.Balance[0].Amount)
	assert.EqualValues("ETH", acct.Balance[0].Denom)
	assert.EqualValues(12345, acct.Balance[1].Amount)
	assert.EqualValues("blank", acct.Balance[1].Denom)

	// and public key is parsed properly
	apk := acct.PubKey.Unwrap()
	require.NotNil(apk)
	epk, ok := apk.(crypto.PubKeyEd25519)
	if assert.True(ok) {
		assert.EqualValues(pkbyte, epk[:])
	}
}

// Fix for issue #89, change the parse format for accounts in genesis.json
func TestLoadGenesisAccountAddress(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	eyesCli := eyescli.NewLocalClient("", 0)
	app := NewBasecoin(eyesCli, log.TestingLogger())
	err := app.LoadGenesis(genesisAcctFilepath)
	require.Nil(err, "%+v", err)

	// check the chain id
	assert.Equal("addr_accounts_chain", app.GetState().GetChainID())

	// make sure the accounts were set properly
	cases := []struct {
		addr      string
		exists    bool
		hasPubkey bool
		coins     types.Coins
	}{
		// this comes from a public key, should be stored proper (alice)
		{"62035D628DE7543332544AA60D90D3693B6AD51B", true, true, types.Coins{{"one", 111}}},
		// this comes from an address, should be stored proper (bob)
		{"C471FB670E44D219EE6DF2FC284BE38793ACBCE1", true, false, types.Coins{{"two", 222}}},
		// this one had a mismatched address and pubkey, should not store under either (carl)
		{"1234ABCDD18E8EFE3FFC4B0506BF9BF8E5B0D9E9", false, false, nil}, // this is given addr
		{"700BEC5ED18E8EFE3FFC4B0506BF9BF8E5B0D9E9", false, false, nil}, // this is addr of the given pubkey
		// this comes from a secp256k1 public key, should be stored proper (sam)
		{"979F080B1DD046C452C2A8A250D18646C6B669D4", true, true, types.Coins{{"four", 444}}},
	}

	for _, tc := range cases {
		addr, err := hex.DecodeString(tc.addr)
		require.Nil(err, tc.addr)
		acct := app.GetState().GetAccount(addr)
		if !tc.exists {
			assert.Nil(acct, tc.addr)
		} else if assert.NotNil(acct, tc.addr) {
			// it should and does exist...
			assert.True(acct.Balance.IsValid())
			assert.Equal(tc.coins, acct.Balance)
			assert.Equal(!tc.hasPubkey, acct.PubKey.Empty(), tc.addr)
		}
	}
}

func TestParseGenesisList(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	bytes, err := cmn.ReadFile(genesisFilepath)
	require.Nil(err, "loading genesis file %+v", err)

	// the basecoin genesis go-wire/data :)
	genDoc := new(FullGenesisDoc)
	err = json.Unmarshal(bytes, genDoc)
	require.Nil(err, "unmarshaling genesis file %+v", err)

	pluginOpts, err := parseGenesisList(genDoc.AppOptions.PluginOptions)
	require.Nil(err, "%+v", err)
	genDoc.AppOptions.pluginOptions = pluginOpts

	assert.Equal(genDoc.AppOptions.pluginOptions[0].Key, "plugin1/key1")
	assert.Equal(genDoc.AppOptions.pluginOptions[1].Key, "plugin1/key2")
	assert.Equal(genDoc.AppOptions.pluginOptions[0].Value, "value1")
	assert.Equal(genDoc.AppOptions.pluginOptions[1].Value, "value2")
}
