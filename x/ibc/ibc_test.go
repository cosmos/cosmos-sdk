package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// AccountKeeper(/Keeper) and IBCMapper should use different StoreKey later

type testInput struct {
	cdc    *codec.Codec
	ctx    sdk.Context
	ak     auth.AccountKeeper
	bk     bank.BaseKeeper
	ibcKey *sdk.KVStoreKey
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()
	cdc := makeCodec()

	ibcKey := sdk.NewKVStoreKey("ibcCapKey")
	authCapKey := sdk.NewKVStoreKey("authCapKey")
	fckCapKey := sdk.NewKVStoreKey("fckCapKey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(ibcKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(fckCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)
	ak := auth.NewAccountKeeper(
		cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount,
	)
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, auth.DefaultParams())

	return testInput{cdc: cdc, ctx: ctx, ak: ak, bk: bk, ibcKey: ibcKey}
}

func makeCodec() *codec.Codec {
	var cdc = codec.New()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.MsgSend{}, "test/ibc/Send", nil)
	cdc.RegisterConcrete(IBCTransferMsg{}, "test/ibc/IBCTransferMsg", nil)
	cdc.RegisterConcrete(IBCReceiveMsg{}, "test/ibc/IBCReceiveMsg", nil)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/ibc/Account", nil)
	codec.RegisterCrypto(cdc)

	cdc.Seal()

	return cdc
}

func newAddress() sdk.AccAddress {
	return sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
}

func getCoins(ck bank.Keeper, ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, sdk.Error) {
	zero := sdk.Coins(nil)
	coins, _, err := ck.AddCoins(ctx, addr, zero)
	return coins, err
}

func TestIBC(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	src := newAddress()
	dest := newAddress()
	chainid := "ibcchain"
	zero := sdk.Coins(nil)
	mycoins := sdk.Coins{sdk.NewInt64Coin("mycoin", 10)}

	coins, _, err := input.bk.AddCoins(ctx, src, mycoins)
	require.Nil(t, err)
	require.Equal(t, mycoins, coins)

	ibcm := NewMapper(input.cdc, input.ibcKey, DefaultCodespace)
	h := NewHandler(ibcm, input.bk)
	packet := IBCPacket{
		SrcAddr:   src,
		DestAddr:  dest,
		Coins:     mycoins,
		SrcChain:  chainid,
		DestChain: chainid,
	}

	store := ctx.KVStore(input.ibcKey)

	var msg sdk.Msg
	var res sdk.Result
	var egl uint64
	var igs uint64

	egl = ibcm.getEgressLength(store, chainid)
	require.Equal(t, egl, uint64(0))

	msg = IBCTransferMsg{
		IBCPacket: packet,
	}
	res = h(ctx, msg)
	require.True(t, res.IsOK())

	coins, err = getCoins(input.bk, ctx, src)
	require.Nil(t, err)
	require.Equal(t, zero, coins)

	egl = ibcm.getEgressLength(store, chainid)
	require.Equal(t, egl, uint64(1))

	igs = ibcm.GetIngressSequence(ctx, chainid)
	require.Equal(t, igs, uint64(0))

	msg = IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   src,
		Sequence:  0,
	}
	res = h(ctx, msg)
	require.True(t, res.IsOK())

	coins, err = getCoins(input.bk, ctx, dest)
	require.Nil(t, err)
	require.Equal(t, mycoins, coins)

	igs = ibcm.GetIngressSequence(ctx, chainid)
	require.Equal(t, igs, uint64(1))

	res = h(ctx, msg)
	require.False(t, res.IsOK())

	igs = ibcm.GetIngressSequence(ctx, chainid)
	require.Equal(t, igs, uint64(1))
}
