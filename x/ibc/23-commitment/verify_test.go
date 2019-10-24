package commitment_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func TestBatchVerifyMembership(t *testing.T) {
	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db)

	iavlStoreKey := storetypes.NewKVStoreKey("iavlStoreKey")

	store.MountStoreWithDB(iavlStoreKey, storetypes.StoreTypeIAVL, nil)
	store.LoadVersion(0)

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavl.Store)
	for i := 0; i < 5; i++ {
		iavlStore.Set([]byte(fmt.Sprintf("KEY:%d", i)), []byte(fmt.Sprintf("VAL:%d", i)))
	}
	cid := store.Commit()

	header := abci.Header{AppHash: cid.Hash}
	ctx := sdk.NewContext(store, header, false, log.NewNopLogger())

	res := store.Query(abci.RequestQuery{
		Path:   "/iavlStoreKey/subspace",
		Data:   []byte("KEY:"),
		Height: cid.Version,
		Prove:  true,
	})
	require.NotNil(t, res.Proof)

	proof := commitment.Proof{
		Proof: res.Proof,
	}
	prefix := commitment.NewPrefix([]byte("iavlStoreKey"))

	keyBatches := [][]string{
		{"KEY:1"},                   // batch verify one key
		{"KEY:1", "KEY:2"},          // batch verify first 2 keys in subspace
		{"KEY:2", "KEY:3", "KEY:4"}, // batch verify middle 3 keys in subspace
		{"KEY:4", "KEY:5"},          // batch verify last 2 keys in subspace
		{"KEY:3", "KEY:2"},          // batch verify keys in reverse order
		{"KEY:4", "KEY:1"},          // batch verify non-contingous keys
		{"KEY:2", "KEY:3", "KEY:4", "KEY:1", "KEY:5"}, // batch verify all keys in random order
	}

	for i, batch := range keyBatches {
		items := make(map[string][]byte)

		for _, key := range batch {
			// key-pair must have form KEY:{str} => VAL:{str}
			splitKey := strings.Split(key, ":")
			items[key] = []byte(fmt.Sprintf("VAL:%s", splitKey[1]))
		}

		ok := commitment.BatchVerifyMembership(ctx, proof, prefix, items)

		require.True(t, ok, "Test case %d failed on batch verify", i)
	}
}
