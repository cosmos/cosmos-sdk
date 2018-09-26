package simplestake

import (
	"fmt"

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
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	authKey := sdk.NewKVStoreKey("authkey")
	capKey := sdk.NewKVStoreKey("capkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, authKey, capKey
}

func TestKeeperGetSet(t *testing.T) {
	ms, authKey, capKey := setupMultiStore()
	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	stakeKeeper := NewKeeper(capKey, bank.NewBaseKeeper(accountMapper), DefaultCodespace)
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	addr := sdk.AccAddress([]byte("some-address"))

	bi := stakeKeeper.getBondInfo(ctx, addr)
	require.Equal(t, bi, bondInfo{})

	privKey := ed25519.GenPrivKey()

	bi = bondInfo{
		PubKey: privKey.PubKey(),
		Power:  int64(10),
	}
	fmt.Printf("Pubkey: %v\n", privKey.PubKey())
	stakeKeeper.setBondInfo(ctx, addr, bi)

	savedBi := stakeKeeper.getBondInfo(ctx, addr)
	require.NotNil(t, savedBi)
	fmt.Printf("Bond Info: %v\n", savedBi)
	require.Equal(t, int64(10), savedBi.Power)
}

func TestBonding(t *testing.T) {
	ms, authKey, capKey := setupMultiStore()
	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())

	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountMapper)
	stakeKeeper := NewKeeper(capKey, bankKeeper, DefaultCodespace)
	addr := sdk.AccAddress([]byte("some-address"))
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	_, _, err := stakeKeeper.unbondWithoutCoins(ctx, addr)
	require.Equal(t, err, ErrInvalidUnbond(DefaultCodespace))

	_, err = stakeKeeper.bondWithoutCoins(ctx, addr, pubKey, sdk.NewInt64Coin("steak", 10))
	require.Nil(t, err)

	power, err := stakeKeeper.bondWithoutCoins(ctx, addr, pubKey, sdk.NewInt64Coin("steak", 10))
	require.Nil(t, err)
	require.Equal(t, int64(20), power)

	pk, _, err := stakeKeeper.unbondWithoutCoins(ctx, addr)
	require.Nil(t, err)
	require.Equal(t, pubKey, pk)

	_, _, err = stakeKeeper.unbondWithoutCoins(ctx, addr)
	require.Equal(t, err, ErrInvalidUnbond(DefaultCodespace))
}
