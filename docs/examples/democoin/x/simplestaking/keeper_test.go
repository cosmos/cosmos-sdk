package simplestaking

import (
	"fmt"

	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type testInput struct {
	cdc    *codec.Codec
	ctx    sdk.Context
	capKey *sdk.KVStoreKey
	bk     bank.BaseKeeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	capKey := sdk.NewKVStoreKey("capkey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)
	ak := auth.NewAccountKeeper(cdc, capKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bk := bank.NewBaseKeeper(ak)
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)

	return testInput{cdc: cdc, ctx: ctx, capKey: capKey, bk: bk}
}

func TestKeeperGetSet(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	stakingKeeper := NewKeeper(input.capKey, input.bk, DefaultCodespace)
	addr := sdk.AccAddress([]byte("some-address"))

	bi := stakingKeeper.getBondInfo(ctx, addr)
	require.Equal(t, bi, bondInfo{})

	privKey := ed25519.GenPrivKey()

	bi = bondInfo{
		PubKey: privKey.PubKey(),
		Power:  int64(10),
	}
	fmt.Printf("Pubkey: %v\n", privKey.PubKey())
	stakingKeeper.setBondInfo(ctx, addr, bi)

	savedBi := stakingKeeper.getBondInfo(ctx, addr)
	require.NotNil(t, savedBi)
	fmt.Printf("Bond Info: %v\n", savedBi)
	require.Equal(t, int64(10), savedBi.Power)
}

func TestBonding(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	stakingKeeper := NewKeeper(input.capKey, input.bk, DefaultCodespace)
	addr := sdk.AccAddress([]byte("some-address"))
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	_, _, err := stakingKeeper.unbondWithoutCoins(ctx, addr)
	require.Equal(t, err, ErrInvalidUnbond(DefaultCodespace))

	_, err = stakingKeeper.bondWithoutCoins(ctx, addr, pubKey, sdk.NewInt64Coin(stakingtypes.DefaultBondDenom, 10))
	require.Nil(t, err)

	power, err := stakingKeeper.bondWithoutCoins(ctx, addr, pubKey, sdk.NewInt64Coin(stakingtypes.DefaultBondDenom, 10))
	require.Nil(t, err)
	require.Equal(t, int64(20), power)

	pk, _, err := stakingKeeper.unbondWithoutCoins(ctx, addr)
	require.Nil(t, err)
	require.Equal(t, pubKey, pk)

	_, _, err = stakingKeeper.unbondWithoutCoins(ctx, addr)
	require.Equal(t, err, ErrInvalidUnbond(DefaultCodespace))
}
