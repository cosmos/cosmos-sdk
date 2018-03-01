package app

import (
	"encoding/json"
	"fmt"
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

	// Construct a SendMsg
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

	// just marshal/unmarshal!
	cdc := MakeCodec()
	txBytes, err := cdc.MarshalBinary(tx)
	require.NoError(t, err)

	// Run a Check
	cres := bapp.CheckTx(txBytes)
	assert.Equal(t, sdk.CodeUnrecognizedAddress,
		sdk.CodeType(cres.Code), cres.Log)

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	dres := bapp.DeliverTx(txBytes)
	assert.Equal(t, sdk.CodeUnrecognizedAddress,
		sdk.CodeType(dres.Code), dres.Log)
}

func TestGenesis(t *testing.T) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	bapp := NewBasecoinApp(logger, db)

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
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
	bapp.Commit()

	// A checkTx context
	ctx := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)

	// reload app and ensure the account is still there
	bapp = NewBasecoinApp(logger, db)
	ctx = bapp.BaseApp.NewContext(true, abci.Header{})
	res1 = bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)

}

func TestSendMsgWithAccounts(t *testing.T) {
	bapp := newBasecoinApp()

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	// First key goes in genesis, used for sending
	priv1 := crypto.GenPrivKeyEd25519()
	pk1 := priv1.PubKey()
	addr1 := pk1.Address()

	// Second key receies
	pk2 := crypto.GenPrivKeyEd25519().PubKey()
	addr2 := pk2.Address()

	// Give 77 foocoin to the first key
	coins, err := sdk.ParseCoins("77foocoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	acc1 := &types.AppAccount{baseAcc, "foobart"}

	// Construct genesis state
	genesisState := types.GenesisState{
		Accounts: []*types.GenesisAccount{
			types.NewGenesisAccount(acc1),
		},
	}
	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")

	// Initialize the chain
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{vals, stateBytes})
	bapp.Commit()

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})

	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc1, res1)

	// Construct a SendMsg
	var msg = bank.SendMsg{
		Inputs: []bank.Input{
			{
				Address:  crypto.Address(addr1),
				Coins:    sdk.Coins{{"foocoin", 10}},
				Sequence: 1,
			},
		},
		Outputs: []bank.Output{
			{
				Address: crypto.Address(addr2),
				Coins:   sdk.Coins{{"foocoin", 10}},
			},
		},
	}

	// Sign the tx
	sig := priv1.Sign(msg.GetSignBytes())
	tx := sdk.NewStdTx(msg, []sdk.StdSignature{{
		PubKey:    priv1.PubKey(),
		Signature: sig,
	}})

	// Run a Check
	res := bapp.Check(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// A deliverTx context
	ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})

	// Check balances
	res2 := bapp.accountMapper.GetAccount(ctxDeliver, addr1)
	res3 := bapp.accountMapper.GetAccount(ctxDeliver, addr2)

	assert.Equal(t, fmt.Sprintf("%v", res2.GetCoins()), "67foocoin")
	assert.Equal(t, fmt.Sprintf("%v", res3.GetCoins()), "10foocoin")
}
