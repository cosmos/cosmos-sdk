package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
)

func TestVerifyMembership(t *testing.T) {
	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db)

	iavlStoreKey := storetypes.NewKVStoreKey("iavlStoreKey")

	store.MountStoreWithDB(iavlStoreKey, storetypes.StoreTypeIAVL, nil)
	store.LoadVersion(0)

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavl.Store)
	iavlStore.Set([]byte("MYKEY"), []byte("MYVALUE"))

	res := store.Query(abci.RequestQuery{
		Path:  "/iavlStoreKey/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	proof := types.Proof{
		Proof: res.Proof,
	}

}
