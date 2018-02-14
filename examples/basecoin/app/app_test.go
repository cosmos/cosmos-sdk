package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	crypto "github.com/tendermint/go-crypto"
)

type testBasecoinApp struct {
	*BasecoinApp
	*bam.TestApp
}

func newTestBasecoinApp() *testBasecoinApp {
	app := NewBasecoinApp("")
	tba := &testBasecoinApp{
		BasecoinApp: app,
	}
	tba.TestApp = bam.NewTestApp(app.BaseApp)
	return tba
}

func TestSendMsg(t *testing.T) {
	tba := newTestBasecoinApp()
	tba.RunBeginBlock()
	defer tba.Close()

	// Construct a SendMsg.
	var msg = bank.SendMsg{
		Inputs: []bank.Input{
			{
				Address:  crypto.Address([]byte("input")),
				Coins:    sdk.Coins{{"atom", 10}},
				Sequence: 1,
			},
		},
		Outputs: []bank.Output{
			{
				Address: crypto.Address([]byte("output")),
				Coins:   sdk.Coins{{"atom", 10}},
			},
		},
	}

	// Run a Check on SendMsg.
	res := tba.RunCheckMsg(msg)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Run a Deliver on SendMsg.
	res = tba.RunDeliverMsg(msg)
	assert.Equal(t, sdk.CodeUnrecognizedAddress, res.Code, res.Log)
}

func TestGenesis(t *testing.T) {
	tba := newTestBasecoinApp()
	tba.RunBeginBlock()
	defer tba.Close()

	// construct some genesis bytes to reflect basecoin/types/AppAccount
	pk := crypto.GenPrivKeyEd25519().PubKey()
	addr := pk.Address()
	coins, err := sdk.ParseCoins("77foocoin,99barcoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addr,
		Coins:   coins,
	}
	acc := &types.AppAccount{baseAcc, "foobart"}

	genesisState := types.GenesisState{
		Accounts: []*types.GenesisAccount{
			types.NewGenesisAccount(acc),
		},
	}
	bytes, err := json.MarshalIndent(genesisState, "", "\t")

	app := tba.BasecoinApp
	ctx := app.BaseApp.NewContext(false, nil) // context for DeliverTx
	err = app.BaseApp.InitStater(ctx, bytes)
	require.Nil(t, err)

	res1 := app.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)
}
