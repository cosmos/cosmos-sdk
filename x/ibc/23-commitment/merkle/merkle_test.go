package merkle

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
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
	k, ctx, cms, _ := defaultComponents()
	kvstore := ctx.KVStore(k)
	path := Path{KeyPath: [][]byte{[]byte("test")}, KeyPrefix: []byte{0x01, 0x03, 0x05}}

	m := make(map[string][]byte)
	kvpn := 1000

	// Repeat 100 times to test on multiple commits
	for repeat := 0; repeat < 10; repeat++ {

		// Initializes random generated key-value pairs
		for i := 0; i < kvpn; i++ {
			k, v := make([]byte, 64), make([]byte, 64)
			rand.Read(k)
			rand.Read(v)
			m[string(k)] = v
			kvstore.Set(path.Key(k), v)
		}

		// Commit store
		root := commit(cms)

		// Test query, and accumulate proofs
		proofs := make([]commitment.Proof, 0, kvpn)
		for k, v := range m {
			c, v0, p := path.QueryMultiStore(cms, []byte(k))
			require.Equal(t, uint32(0), c)
			require.Equal(t, v, v0)
			proofs = append(proofs, p)
		}

		// Add some exclusion proofs
		for i := 0; i < 100; i++ {
			k := make([]byte, 64)
			rand.Read(k)
			c, v, p := path.QueryMultiStore(cms, k)
			require.Equal(t, uint32(0), c)
			require.Nil(t, v)
			proofs = append(proofs, p)
			m[string(k)] = []byte{}
		}

		cstore, err := commitment.NewStore(root, path, proofs)
		require.NoError(t, err)

		// Test commitment store
		for k, v := range m {
			if len(v) != 0 {
				require.True(t, cstore.Prove([]byte(k), v))
			} else {
				require.True(t, cstore.Prove([]byte(k), nil))
			}
		}

		// Modify existing data
		for k := range m {
			v := make([]byte, 64)
			rand.Read(v)
			m[k] = v
			kvstore.Set(path.Key([]byte(k)), v)
		}

		root = commit(cms)

		// Test query, and accumulate proofs
		proofs = make([]commitment.Proof, 0, kvpn)
		for k, v := range m {
			c, v0, p := path.QueryMultiStore(cms, []byte(k))
			require.Equal(t, uint32(0), c)
			require.Equal(t, v, v0)
			proofs = append(proofs, p)
		}

		cstore, err = commitment.NewStore(root, path, proofs)
		require.NoError(t, err)

		// Test commitment store
		for k, v := range m {
			if len(v) != 0 {
				require.True(t, cstore.Prove([]byte(k), v))
			} else {
				require.True(t, cstore.Prove([]byte(k), nil))
			}
		}
	}
}
