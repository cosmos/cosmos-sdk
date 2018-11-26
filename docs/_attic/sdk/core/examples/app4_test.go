package app

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// Create and return App4 instance
func newTestChain() *bapp.BaseApp {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return NewApp4(logger, db)
}

// Initialize all provided addresses with 100 testCoin
func InitTestChain(bc *bapp.BaseApp, chainID string, addrs ...sdk.AccAddress) {
	var accounts []*GenesisAccount
	for _, addr := range addrs {
		acc := GenesisAccount{
			Address: addr,
			Coins:   sdk.Coins{sdk.NewCoin("testCoin", sdk.NewInt(100))},
		}
		accounts = append(accounts, &acc)
	}
	accountState := GenesisState{accounts}
	genState, err := json.Marshal(accountState)
	if err != nil {
		panic(err)
	}
	bc.InitChain(abci.RequestInitChain{ChainId: chainID, AppStateBytes: genState})
}

// Generate basic SpendMsg with one input and output
func GenerateSpendMsg(sender, receiver sdk.AccAddress, amount sdk.Coins) bank.MsgSend {
	return bank.MsgSend{
		Inputs:  []bank.Input{{sender, amount}},
		Outputs: []bank.Output{{receiver, amount}},
	}
}

// Test spending nonexistent funds fails
func TestBadMsg(t *testing.T) {
	bc := newTestChain()

	// Create privkeys and addresses
	priv1 := ed25519.GenPrivKey()
	priv2 := ed25519.GenPrivKey()
	addr1 := priv1.PubKey().Address().Bytes()
	addr2 := priv2.PubKey().Address().Bytes()

	// Attempt to spend non-existent funds
	msg := GenerateSpendMsg(addr1, addr2, sdk.Coins{sdk.NewCoin("testCoin", sdk.NewInt(100))})

	// Construct transaction
	fee := auth.StdFee{
		Gas:    1000000000000000,
		Amount: sdk.Coins{sdk.NewCoin("testCoin", sdk.NewInt(0))},
	}
	signBytes := auth.StdSignBytes("test-chain", 0, 0, fee, []sdk.Msg{msg}, "")
	sig, err := priv1.Sign(signBytes)
	if err != nil {
		panic(err)
	}
	sigs := []auth.StdSignature{{
		PubKey:        priv1.PubKey(),
		Signature:     sig,
		AccountNumber: 0,
		Sequence:      0,
	}}

	tx := auth.StdTx{
		Msgs:       []sdk.Msg{msg},
		Fee:        fee,
		Signatures: sigs,
		Memo:       "",
	}

	bc.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: "test-chain"}})

	// Deliver the transaction
	res := bc.Deliver(tx)

	// Check that tx failed
	require.False(t, res.IsOK(), "Invalid tx passed")

}

func TestMsgSend(t *testing.T) {
	bc := newTestChain()

	priv1 := ed25519.GenPrivKey()
	priv2 := ed25519.GenPrivKey()
	addr1 := priv1.PubKey().Address().Bytes()
	addr2 := priv2.PubKey().Address().Bytes()

	InitTestChain(bc, "test-chain", addr1)

	// Send funds to addr2
	msg := GenerateSpendMsg(addr1, addr2, sdk.Coins{sdk.NewCoin("testCoin", sdk.NewInt(100))})

	fee := auth.StdFee{
		Gas:    1000000000000000,
		Amount: sdk.Coins{sdk.NewCoin("testCoin", sdk.NewInt(0))},
	}
	signBytes := auth.StdSignBytes("test-chain", 0, 0, fee, []sdk.Msg{msg}, "")
	sig, err := priv1.Sign(signBytes)
	if err != nil {
		panic(err)
	}
	sigs := []auth.StdSignature{{
		PubKey:        priv1.PubKey(),
		Signature:     sig,
		AccountNumber: 0,
		Sequence:      0,
	}}

	tx := auth.StdTx{
		Msgs:       []sdk.Msg{msg},
		Fee:        fee,
		Signatures: sigs,
		Memo:       "",
	}

	bc.BeginBlock(abci.RequestBeginBlock{})

	res := bc.Deliver(tx)

	require.True(t, res.IsOK(), res.Log)

}
