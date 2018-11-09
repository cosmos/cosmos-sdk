package store

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestVerifyIAVLStoreQueryProof(t *testing.T) {
	// Create main tree for testing.
	db := dbm.NewMemDB()
	iStore, err := LoadIAVLStore(db, CommitID{}, sdk.PruneNothing)
	store := iStore.(*iavlStore)
	require.Nil(t, err)
	store.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit()

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	// Verify proof.
	prt := DefaultProofRuntime()
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte("MYVALUE"))
	require.Nil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY_NOT", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte("MYVALUE_NOT"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte(nil))
	require.NotNil(t, err)
}

func TestVerifyMultiStoreQueryProof(t *testing.T) {
	// Create main tree for testing.
	db := dbm.NewMemDB()
	store := NewCommitMultiStore(db)
	iavlStoreKey := sdk.NewKVStoreKey("iavlStoreKey")

	store.MountStoreWithDB(iavlStoreKey, sdk.StoreTypeIAVL, nil)
	store.LoadVersion(0)

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavlStore)
	iavlStore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit()

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/iavlStoreKey/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	// Verify proof.
	prt := DefaultProofRuntime()
	err := prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte("MYVALUE"))
	require.Nil(t, err)

	// Verify proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte("MYVALUE"))
	require.Nil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY_NOT", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "iavlStoreKey/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte("MYVALUE_NOT"))
	require.NotNil(t, err)

	// Verify (bad) proof.
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte(nil))
	require.NotNil(t, err)
}

func TestVerifyMultiStoreQueryProofEmptyStore(t *testing.T) {
	// Create main tree for testing.
	db := dbm.NewMemDB()
	store := NewCommitMultiStore(db)
	iavlStoreKey := sdk.NewKVStoreKey("iavlStoreKey")

	store.MountStoreWithDB(iavlStoreKey, sdk.StoreTypeIAVL, nil)
	store.LoadVersion(0)
	cid := store.Commit() // Commit with empty iavl store.

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/iavlStoreKey/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	// Verify proof.
	prt := DefaultProofRuntime()
	err := prt.VerifyAbsence(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY")
	require.Nil(t, err)

	// Verify (bad) proof.
	prt = DefaultProofRuntime()
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYKEY", []byte("MYVALUE"))
	require.NotNil(t, err)
}

func TestVerifyMultiStoreQueryProofAbsence(t *testing.T) {
	// Create main tree for testing.
	db := dbm.NewMemDB()
	store := NewCommitMultiStore(db)
	iavlStoreKey := sdk.NewKVStoreKey("iavlStoreKey")

	store.MountStoreWithDB(iavlStoreKey, sdk.StoreTypeIAVL, nil)
	store.LoadVersion(0)

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavlStore)
	iavlStore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit() // Commit with empty iavl store.

	// Get Proof
	res := store.Query(abci.RequestQuery{
		Path:  "/iavlStoreKey/key", // required path to get key/value+proof
		Data:  []byte("MYABSENTKEY"),
		Prove: true,
	})
	require.NotNil(t, res.Proof)

	// Verify proof.
	prt := DefaultProofRuntime()
	err := prt.VerifyAbsence(res.Proof, cid.Hash, "/iavlStoreKey/MYABSENTKEY")
	require.Nil(t, err)

	// Verify (bad) proof.
	prt = DefaultProofRuntime()
	err = prt.VerifyAbsence(res.Proof, cid.Hash, "/MYABSENTKEY")
	require.NotNil(t, err)

	// Verify (bad) proof.
	prt = DefaultProofRuntime()
	err = prt.VerifyValue(res.Proof, cid.Hash, "/iavlStoreKey/MYABSENTKEY", []byte(""))
	require.NotNil(t, err)
}
