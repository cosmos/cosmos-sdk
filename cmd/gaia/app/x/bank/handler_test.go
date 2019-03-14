package bank

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
)

var (
	addrs = []sdk.AccAddress{
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
	}

	initAmt = sdk.NewInt(atomsToUatoms * 100)
)

type testInput struct {
	ctx        sdk.Context
	accKeeper  auth.AccountKeeper
	bankKeeper bank.Keeper
}

func newTestCodec() *codec.Codec {
	cdc := codec.New()

	bank.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

func createTestInput(t *testing.T) testInput {
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tKeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	cdc := newTestCodec()
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ctx := sdk.NewContext(ms, abci.Header{Time: time.Now().UTC()}, false, log.NewNopLogger())

	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tKeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)

	require.NoError(t, ms.LoadLatestVersion())

	paramsKeeper := params.NewKeeper(cdc, keyParams, tKeyParams)
	accKeeper := auth.NewAccountKeeper(
		cdc,
		keyAcc,
		paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	bankKeeper := bank.NewBaseKeeper(
		accKeeper,
		paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)

	for _, addr := range addrs {
		_, _, err := bankKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewCoin(uatomDenom, initAmt)})
		require.NoError(t, err)
	}

	return testInput{ctx, accKeeper, bankKeeper}
}

func TestHandlerMsgSendTransfersDisabled(t *testing.T) {
	input := createTestInput(t)
	input.bankKeeper.SetSendEnabled(input.ctx, false)

	handler := NewHandler(input.bankKeeper)
	amt := sdk.NewInt(atomsToUatoms * 5)
	msg := bank.NewMsgSend(addrs[0], addrs[1], sdk.Coins{sdk.NewCoin(uatomDenom, amt)})

	res := handler(input.ctx, msg)
	require.False(t, res.IsOK(), "expected failed message execution: %v", res.Log)

	from := input.accKeeper.GetAccount(input.ctx, addrs[0])
	require.Equal(t, from.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, initAmt)})

	to := input.accKeeper.GetAccount(input.ctx, addrs[1])
	require.Equal(t, to.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, initAmt)})
}

func TestHandlerMsgSendTransfersEnabled(t *testing.T) {
	input := createTestInput(t)
	input.bankKeeper.SetSendEnabled(input.ctx, true)

	handler := NewHandler(input.bankKeeper)
	amt := sdk.NewInt(atomsToUatoms * 5)
	msg := bank.NewMsgSend(addrs[0], addrs[1], sdk.Coins{sdk.NewCoin(uatomDenom, amt)})

	res := handler(input.ctx, msg)
	require.True(t, res.IsOK(), "expected successful message execution: %v", res.Log)

	from := input.accKeeper.GetAccount(input.ctx, addrs[0])
	balance := initAmt.Sub(amt)
	require.Equal(t, from.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, balance)})

	to := input.accKeeper.GetAccount(input.ctx, addrs[1])
	balance = initAmt.Add(amt)
	require.Equal(t, to.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, balance)})
}

func TestHandlerMsgMultiSendTransfersDisabledBurn(t *testing.T) {
	input := createTestInput(t)
	input.bankKeeper.SetSendEnabled(input.ctx, false)

	handler := NewHandler(input.bankKeeper)
	totalAmt := sdk.NewInt(10 * atomsToUatoms)
	burnAmt := sdk.NewInt(9 * atomsToUatoms)
	transAmt := sdk.NewInt(1 * atomsToUatoms)
	msg := bank.NewMsgMultiSend(
		[]bank.Input{
			bank.NewInput(addrs[0], sdk.Coins{sdk.NewCoin(uatomDenom, totalAmt)}),
		},
		[]bank.Output{
			bank.NewOutput(BurnedCoinsAccAddr, sdk.Coins{sdk.NewCoin(uatomDenom, burnAmt)}),
			bank.NewOutput(addrs[1], sdk.Coins{sdk.NewCoin(uatomDenom, transAmt)}),
		},
	)

	res := handler(input.ctx, msg)
	require.True(t, res.IsOK(), "expected successful message execution: %v", res.Log)

	from := input.accKeeper.GetAccount(input.ctx, addrs[0])
	balance := initAmt.Sub(totalAmt)
	require.Equal(t, from.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, balance)})

	to := input.accKeeper.GetAccount(input.ctx, addrs[1])
	balance = initAmt.Add(transAmt)
	require.Equal(t, to.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, balance)})

	burn := input.accKeeper.GetAccount(input.ctx, BurnedCoinsAccAddr)
	require.Equal(t, burn.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, burnAmt)})
}

func TestHandlerMsgMultiSendTransfersDisabledInvalidBurn(t *testing.T) {
	input := createTestInput(t)
	input.bankKeeper.SetSendEnabled(input.ctx, false)

	handler := NewHandler(input.bankKeeper)
	totalAmt := sdk.NewInt(15 * atomsToUatoms)
	burnAmt := sdk.NewInt(10 * atomsToUatoms)
	transAmt := sdk.NewInt(5 * atomsToUatoms)
	msg := bank.NewMsgMultiSend(
		[]bank.Input{
			bank.NewInput(addrs[0], sdk.Coins{sdk.NewCoin(uatomDenom, totalAmt)}),
		},
		[]bank.Output{
			bank.NewOutput(BurnedCoinsAccAddr, sdk.Coins{sdk.NewCoin(uatomDenom, burnAmt)}),
			bank.NewOutput(addrs[1], sdk.Coins{sdk.NewCoin(uatomDenom, transAmt)}),
		},
	)

	res := handler(input.ctx, msg)
	require.False(t, res.IsOK(), "expected failed message execution: %v", res.Log)

	from := input.accKeeper.GetAccount(input.ctx, addrs[0])
	require.Equal(t, from.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, initAmt)})

	to := input.accKeeper.GetAccount(input.ctx, addrs[1])
	require.Equal(t, to.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, initAmt)})
}

func TestHandlerMsgMultiSendTransfersEnabled(t *testing.T) {
	input := createTestInput(t)
	input.bankKeeper.SetSendEnabled(input.ctx, true)

	handler := NewHandler(input.bankKeeper)
	totalAmt := sdk.NewInt(10 * atomsToUatoms)
	outAmt := sdk.NewInt(5 * atomsToUatoms)
	msg := bank.NewMsgMultiSend(
		[]bank.Input{
			bank.NewInput(addrs[0], sdk.Coins{sdk.NewCoin(uatomDenom, totalAmt)}),
		},
		[]bank.Output{
			bank.NewOutput(addrs[1], sdk.Coins{sdk.NewCoin(uatomDenom, outAmt)}),
			bank.NewOutput(addrs[2], sdk.Coins{sdk.NewCoin(uatomDenom, outAmt)}),
		},
	)

	res := handler(input.ctx, msg)
	require.True(t, res.IsOK(), "expected successful message execution: %v", res.Log)

	from := input.accKeeper.GetAccount(input.ctx, addrs[0])
	balance := initAmt.Sub(totalAmt)
	require.Equal(t, from.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, balance)})

	out1 := input.accKeeper.GetAccount(input.ctx, addrs[1])
	balance = initAmt.Add(outAmt)
	require.Equal(t, out1.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, balance)})

	out2 := input.accKeeper.GetAccount(input.ctx, addrs[2])
	balance = initAmt.Add(outAmt)
	require.Equal(t, out2.GetCoins(), sdk.Coins{sdk.NewCoin(uatomDenom, balance)})
}
