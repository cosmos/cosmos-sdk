package app

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

func newBasecoinApp() *BasecoinApp {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return NewBasecoinApp(logger, db)
}

func TestSendMsg(t *testing.T) {
	bapp := newBasecoinApp()

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

	priv := crypto.GenPrivKeyEd25519()
	sig := priv.Sign(msg.GetSignBytes())
	tx := sdk.NewStdTx(msg, []sdk.StdSignature{{
		PubKey:    priv.PubKey(),
		Signature: sig,
	}})

	// Run a Check
	res := bapp.Check(tx)
	assert.Equal(t, sdk.CodeUnrecognizedAddress, res.Code, res.Log)

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeUnrecognizedAddress, res.Code, res.Log)
}

func TestGenesis(t *testing.T) {
	bapp := newBasecoinApp()

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
	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")

	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{vals, stateBytes})

	// a checkTx context
	ctx := bapp.BaseApp.NewContext(true, nil)

	res1 := bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)
}
