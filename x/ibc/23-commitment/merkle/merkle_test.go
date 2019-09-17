package merkle

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/state"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func defaultComponents() (sdk.StoreKey, sdk.Context, types.CommitMultiStore, *codec.Codec) {
	key := sdk.NewKVStoreKey("test")
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	cdc := codec.New()
	return key, ctx, cms, cdc
}

func commit(cms types.CommitMultiStore) Root {
	cid := cms.Commit()
	return NewRoot(cid.Hash)
}

// TestStore tests Merkle proof on the commitment.Store
// Sets/upates key-value pairs and prove with the query result proofs
func TestStore(t *testing.T) {
	k, ctx, cms, cdc := defaultComponents()
	storeName := k.Name()
	prefix := []byte{0x01, 0x03, 0x05, 0xAA, 0xBB}
	mapp := state.NewMapping(k, cdc, prefix)
	path := NewPath([][]byte{[]byte(storeName)}, prefix)

	m := make(map[string][]byte)
	kvpn := 10

	// Repeat to test on multiple commits
	for repeat := 0; repeat < 10; repeat++ {

		// Initializes random generated key-value pairs
		for i := 0; i < kvpn; i++ {
			k, v := make([]byte, 16), make([]byte, 16)
			rand.Read(k)
			rand.Read(v)
			m[string(k)] = v
			mapp.Value(k).Set(ctx, v)
		}

		// Commit store
		root := commit(cms)

		// Test query, and accumulate proofs
		proofs := make([]commitment.Proof, 0, kvpn)
		for k, v := range m {
			v0, p, err := QueryMultiStore(cms, storeName, prefix, []byte(k))
			require.NoError(t, err)
			require.Equal(t, cdc.MustMarshalBinaryBare(v), v0, "Queried value different at %d", repeat)
			proofs = append(proofs, p)
		}

		// Add some exclusion proofs
		for i := 0; i < 10; i++ {
			k := make([]byte, 64)
			rand.Read(k)
			v, p, err := QueryMultiStore(cms, storeName, prefix, k)
			require.NoError(t, err)
			require.Nil(t, v)
			proofs = append(proofs, p)
			m[string(k)] = []byte{}
		}

		cstore, err := commitment.NewStore(root, path, proofs)
		require.NoError(t, err)

		// Test commitment store
		for k, v := range m {
			if len(v) != 0 {
				require.True(t, cstore.Prove([]byte(k), cdc.MustMarshalBinaryBare(v)))
			} else {
				require.True(t, cstore.Prove([]byte(k), nil))
			}
		}

		// Modify existing data
		for k := range m {
			v := make([]byte, 64)
			rand.Read(v)
			m[k] = v
			mapp.Value([]byte(k)).Set(ctx, v)
		}

		root = commit(cms)

		// Test query, and accumulate proofs
		proofs = make([]commitment.Proof, 0, kvpn)
		for k, v := range m {
			v0, p, err := QueryMultiStore(cms, storeName, prefix, []byte(k))
			require.NoError(t, err)
			require.Equal(t, cdc.MustMarshalBinaryBare(v), v0)
			proofs = append(proofs, p)
		}

		cstore, err = commitment.NewStore(root, path, proofs)
		require.NoError(t, err)

		// Test commitment store
		for k, v := range m {
			if len(v) != 0 {
				require.True(t, cstore.Prove([]byte(k), cdc.MustMarshalBinaryBare(v)))
			} else {
				require.True(t, cstore.Prove([]byte(k), nil))
			}
		}
	}
}
