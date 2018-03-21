package staking

import (
	"fmt"

	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	capKey := sdk.NewKVStoreKey("capkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, capKey
}

func TestKeeperGetSet(t *testing.T) {
	ms, capKey := setupMultiStore()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	stakeKeeper := NewKeeper(capKey, bank.NewCoinKeeper(nil))
	addr := sdk.Address([]byte("some-address"))

	bi := stakeKeeper.getBondInfo(ctx, addr)
	assert.Equal(t, bi, bondInfo{})

	privKey := crypto.GenPrivKeyEd25519()

	bi = bondInfo{
		PubKey: privKey.PubKey(),
		Power:  int64(10),
	}
	fmt.Printf("Pubkey: %v\n", privKey.PubKey())
	stakeKeeper.setBondInfo(ctx, addr, bi)

	savedBi := stakeKeeper.getBondInfo(ctx, addr)
	assert.NotNil(t, savedBi)
	fmt.Printf("Bond Info: %v\n", savedBi)
	assert.Equal(t, int64(10), savedBi.Power)
}

func TestBonding(t *testing.T) {
	ms, capKey := setupMultiStore()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	stakeKeeper := NewKeeper(capKey, bank.NewCoinKeeper(nil))
	addr := sdk.Address([]byte("some-address"))
	privKey := crypto.GenPrivKeyEd25519()
	pubKey := privKey.PubKey()

	_, _, err := stakeKeeper.Unbond(ctx, addr)
	assert.Equal(t, err, ErrInvalidUnbond())

	_, err = stakeKeeper.Bond(ctx, addr, pubKey, 10)
	assert.Nil(t, err)

	power, err := stakeKeeper.Bond(ctx, addr, pubKey, 10)
	assert.Equal(t, int64(20), power)

	pk, _, err := stakeKeeper.Unbond(ctx, addr)
	assert.Nil(t, err)
	assert.Equal(t, pubKey, pk)

	_, _, err = stakeKeeper.Unbond(ctx, addr)
	assert.Equal(t, err, ErrInvalidUnbond())
}
