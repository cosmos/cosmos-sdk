package simplestake

import (
	"fmt"

	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	authKey := sdk.NewKVStoreKey("authkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, authKey
}

func TestCoinKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	accountMapper := auth.NewMapper(authKey, &auth.BaseAccount{})
	coinKeeper := NewCoinKeeper(accountMapper)
	accountMapper.SetAccount(accountMapper.NewAccountWithAddress())

	assert.Equal(t, sdk.Coins{}, coinKeeper, ...)


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

func GenerateKeys() (crypto.PrivKey, crypto.PubKey, sdk.Address) {
	privKey := crypto.GenPrivKeyEd25519()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	return (privKey, pubKey, addr)
}
