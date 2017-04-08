package app

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	eyescli "github.com/tendermint/merkleeyes/client"
)

const genesisFilepath = "./testdata/genesis.json"

func TestLoadGenesis(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	eyesCli := eyescli.NewLocalClient("", 0)
	app := NewBasecoin(eyesCli)
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
	assert.EqualValues(12345, acct.Balance[0].Amount)
	assert.EqualValues("blank", acct.Balance[0].Denom)

	// and public key is parsed properly
	apk := acct.PubKey.PubKey
	require.NotNil(apk)
	epk, ok := apk.(crypto.PubKeyEd25519)
	if assert.True(ok) {
		assert.EqualValues(pkbyte, epk[:])
	}
}

func TestParseGenesisList(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	bytes, err := cmn.ReadFile(genesisFilepath)
	require.Nil(err, "loading genesis file %+v", err)

	// the basecoin genesis go-data :)
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
