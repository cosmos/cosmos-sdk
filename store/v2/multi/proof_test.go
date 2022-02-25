package multi

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
)

// We hash keys produce SMT paths, so reflect that here
func keyPath(prefix, key string) string {
	hashed := sha256.Sum256([]byte(key))
	return prefix + string(hashed[:])
}

func TestVerifySMTStoreProof(t *testing.T) {
	// Create main tree for testing.
	txn := memdb.NewDB().ReadWriter()
	store := smt.NewStore(txn)
	store.Set([]byte("MYKEY"), []byte("MYVALUE"))
	root := store.Root()

	res, err := proveKey(store, []byte("MYKEY"))
	require.NoError(t, err)

	// Verify good proof.
	prt := DefaultProofRuntime()
	err = prt.VerifyValue(res, root, keyPath("/", "MYKEY"), []byte("MYVALUE"))
	require.NoError(t, err)

	// Fail to verify bad proofs.
	err = prt.VerifyValue(res, root, keyPath("/", "MYKEY_NOT"), []byte("MYVALUE"))
	require.Error(t, err)

	err = prt.VerifyValue(res, root, keyPath("/", "MYKEY/MYKEY"), []byte("MYVALUE"))
	require.Error(t, err)

	err = prt.VerifyValue(res, root, keyPath("", "MYKEY"), []byte("MYVALUE"))
	require.Error(t, err)

	err = prt.VerifyValue(res, root, keyPath("/", "MYKEY"), []byte("MYVALUE_NOT"))
	require.Error(t, err)

	err = prt.VerifyValue(res, root, keyPath("/", "MYKEY"), []byte(nil))
	require.Error(t, err)
}

func TestVerifyMultiStoreQueryProof(t *testing.T) {
	db := memdb.NewDB()
	store, err := NewStore(db, storeParams1(t))
	require.NoError(t, err)

	substore := store.GetKVStore(skey_1)
	substore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit()

	res := store.Query(abci.RequestQuery{
		Path:  "/store1/key", // required path to get key/value+proof
		Data:  []byte("MYKEY"),
		Prove: true,
	})
	require.NotNil(t, res.ProofOps)

	// Verify good proofs.
	prt := DefaultProofRuntime()
	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("/store1/", "MYKEY"), []byte("MYVALUE"))
	require.NoError(t, err)

	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("/store1/", "MYKEY"), []byte("MYVALUE"))
	require.NoError(t, err)

	// Fail to verify bad proofs.
	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("/store1/", "MYKEY_NOT"), []byte("MYVALUE"))
	require.Error(t, err)

	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("/store1/MYKEY/", "MYKEY"), []byte("MYVALUE"))
	require.Error(t, err)

	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("store1/", "MYKEY"), []byte("MYVALUE"))
	require.Error(t, err)

	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("/", "MYKEY"), []byte("MYVALUE"))
	require.Error(t, err)

	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("/store1/", "MYKEY"), []byte("MYVALUE_NOT"))
	require.Error(t, err)

	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("/store1/", "MYKEY"), []byte(nil))
	require.Error(t, err)
}

func TestVerifyMultiStoreQueryProofAbsence(t *testing.T) {
	db := memdb.NewDB()
	store, err := NewStore(db, storeParams1(t))
	require.NoError(t, err)

	substore := store.GetKVStore(skey_1)
	substore.Set([]byte("MYKEY"), []byte("MYVALUE"))
	cid := store.Commit()

	res := store.Query(abci.RequestQuery{
		Path:  "/store1/key", // required path to get key/value+proof
		Data:  []byte("MYABSENTKEY"),
		Prove: true,
	})
	require.NotNil(t, res.ProofOps)

	// Verify good proof.
	prt := DefaultProofRuntime()
	err = prt.VerifyAbsence(res.ProofOps, cid.Hash, keyPath("/store1/", "MYABSENTKEY"))
	require.NoError(t, err)

	// Fail to verify bad proofs.
	prt = DefaultProofRuntime()
	err = prt.VerifyAbsence(res.ProofOps, cid.Hash, keyPath("/", "MYABSENTKEY"))
	require.Error(t, err)

	prt = DefaultProofRuntime()
	err = prt.VerifyValue(res.ProofOps, cid.Hash, keyPath("/store1/", "MYABSENTKEY"), []byte(""))
	require.Error(t, err)
}
