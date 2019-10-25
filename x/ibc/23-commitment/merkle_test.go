package commitment_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"

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

	proof := commitment.Proof{
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
		root := commitment.NewRoot(tc.root)
		path := commitment.NewPath(tc.pathArr)

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

	proof := commitment.Proof{
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
		root := commitment.NewRoot(tc.root)
		path := commitment.NewPath(tc.pathArr)

		ok := proof.VerifyNonMembership(root, path)

		require.True(t, ok == tc.shouldPass, "Test case %d failed: %s", i, tc.errString)
	}

}

func TestApplyPrefix(t *testing.T) {
	prefix := commitment.NewPrefix([]byte("storePrefixKey"))

	pathStr := "path1/path2/path3/key"

	prefixedPath, err := commitment.ApplyPrefix(prefix, pathStr)
	require.Nil(t, err, "valid prefix returns error")

	require.Equal(t, "/storePrefixKey/path1/path2/path3/key", prefixedPath.String(), "Prefixed path incorrect")

	// invalid prefix contains non-alphanumeric character
	invalidPathStr := "invalid-path/doesitfail?/hopefully"
	invalidPath, err := commitment.ApplyPrefix(prefix, invalidPathStr)
	require.NotNil(t, err, "invalid prefix does not returns error")
	require.Equal(t, commitment.Path{}, invalidPath, "invalid prefix returns valid Path on ApplyPrefix")
}
