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
	cid := store.Commit()

	res := store.Query(abci.RequestQuery{
		Path:  "/iavlStoreKey/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	proof := types.Proof{
		Proof: res.Proof,
	}

	cases := []struct {
		root       []byte
		pathArr    []string
		value      []byte
		shouldPass bool
		errString  string
	}{
		{cid.Hash, []string{"iavlStoreKey", "MYKEY"}, []byte("MYVALUE"), true, "valid membership proof failed"},                               // valid proof
		{cid.Hash, []string{"iavlStoreKey", "MYKEY"}, []byte("WRONGVALUE"), false, "invalid membership proof with wrong value passed"},        // invalid proof with wrong value
		{cid.Hash, []string{"iavlStoreKey", "MYKEY"}, []byte(nil), false, "invalid membership proof with wrong value passed"},                 // invalid proof with nil value
		{cid.Hash, []string{"iavlStoreKey", "NOTMYKEY"}, []byte("MYVALUE"), false, "invalid membership proof with wrong key passed"},          // invalid proof with wrong key
		{cid.Hash, []string{"iavlStoreKey", "MYKEY", "MYKEY"}, []byte("MYVALUE"), false, "invalid membership proof with wrong path passed"},   // invalid proof with wrong path
		{cid.Hash, []string{"iavlStoreKey"}, []byte("MYVALUE"), false, "invalid membership proof with wrong path passed"},                     // invalid proof with wrong path
		{cid.Hash, []string{"MYKEY"}, []byte("MYVALUE"), false, "invalid membership proof with wrong path passed"},                            // invalid proof with wrong path
		{cid.Hash, []string{"otherStoreKey", "MYKEY"}, []byte("MYVALUE"), false, "invalid membership proof with wrong store prefix passed"},   // invalid proof with wrong store prefix
		{[]byte("WRONGROOT"), []string{"iavlStoreKey", "MYKEY"}, []byte("MYVALUE"), false, "invalid membership proof with wrong root passed"}, // invalid proof with wrong root
		{[]byte(nil), []string{"iavlStoreKey", "MYKEY"}, []byte("MYVALUE"), false, "invalid membership proof with nil root passed"},           // invalid proof with nil root
	}

	for i, tc := range cases {
		root := types.NewRoot(tc.root)
		path := types.NewPath(tc.pathArr)

		ok := proof.VerifyMembership(root, path, tc.value)

		require.True(t, ok == tc.shouldPass, "Test case %d failed: %s", i, tc.errString)
	}

}

func TestVerifyNonMembership(t *testing.T) {
	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db)

	iavlStoreKey := storetypes.NewKVStoreKey("iavlStoreKey")

	store.MountStoreWithDB(iavlStoreKey, storetypes.StoreTypeIAVL, nil)
	store.LoadVersion(0)

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavl.Store)
	iavlStore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit()

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/iavlStoreKey/key", // required path to get key/value+proof
		Data:  []byte("MYABSENTKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	proof := types.Proof{
		Proof: res.Proof,
	}

	cases := []struct {
		root       []byte
		pathArr    []string
		shouldPass bool
		errString  string
	}{
		{cid.Hash, []string{"iavlStoreKey", "MYABSENTKEY"}, true, "valid non-membership proof failed"},                               // valid proof
		{cid.Hash, []string{"iavlStoreKey", "MYKEY"}, false, "invalid non-membership proof with wrong key passed"},                   // invalid proof with existent key
		{cid.Hash, []string{"iavlStoreKey", "MYKEY", "MYABSENTKEY"}, false, "invalid non-membership proof with wrong path passed"},   // invalid proof with wrong path
		{cid.Hash, []string{"iavlStoreKey", "MYABSENTKEY", "MYKEY"}, false, "invalid non-membership proof with wrong path passed"},   // invalid proof with wrong path
		{cid.Hash, []string{"iavlStoreKey"}, false, "invalid non-membership proof with wrong path passed"},                           // invalid proof with wrong path
		{cid.Hash, []string{"MYABSENTKEY"}, false, "invalid non-membership proof with wrong path passed"},                            // invalid proof with wrong path
		{cid.Hash, []string{"otherStoreKey", "MYABSENTKEY"}, false, "invalid non-membership proof with wrong store prefix passed"},   // invalid proof with wrong store prefix
		{[]byte("WRONGROOT"), []string{"iavlStoreKey", "MYABSENTKEY"}, false, "invalid non-membership proof with wrong root passed"}, // invalid proof with wrong root
		{[]byte(nil), []string{"iavlStoreKey", "MYABSENTKEY"}, false, "invalid non-membership proof with nil root passed"},           // invalid proof with nil root
	}

	for i, tc := range cases {
		root := types.NewRoot(tc.root)
		path := types.NewPath(tc.pathArr)

		ok := proof.VerifyNonMembership(root, path)

		require.True(t, ok == tc.shouldPass, "Test case %d failed: %s", i, tc.errString)
	}

}

func TestApplyPrefix(t *testing.T) {
	prefix := types.NewPrefix([]byte("storePrefixKey"))

	pathStr := "path1/path2/path3/key"

	prefixedPath := types.ApplyPrefix(prefix, pathStr)

	require.Equal(t, "/storePrefixKey/path1/path2/path3/key", prefixedPath.String(), "Prefixed path incorrect")
}
