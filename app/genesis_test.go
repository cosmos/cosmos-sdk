package app

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-crypto"
	eyescli "github.com/tendermint/merkleeyes/client"
)

func TestLoadGenesis(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	eyesCli := eyescli.NewLocalClient("", 0)
	app := NewBasecoin(eyesCli)
	err := app.LoadGenesis("./testdata/genesis.json")
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
