package merkle

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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

func commit(cms types.CommitMultiStore, root Root) Root {
	cid := cms.Commit()
	return root.Update(cid.Hash).(Root)
}

func TestStore(t *testing.T) {
	k, ctx, cms, _ := defaultComponents()
	kvstore := ctx.KVStore(k)
	root := Root{KeyPath: [][]byte{[]byte("test")}, KeyPrefix: []byte{0x01, 0x03, 0x05}}

	kvstore.Set(root.Key([]byte("hello")), []byte("world"))
	kvstore.Set(root.Key([]byte("merkle")), []byte("tree"))
	kvstore.Set(root.Key([]byte("block")), []byte("chain"))

	root = commit(cms, root)

	c1, v1, p1 := root.Query(cms, []byte("hello"))
	require.Equal(t, uint32(0), c1)
	require.Equal(t, []byte("world"), v1)
	c2, v2, p2 := root.Query(cms, []byte("merkle"))
	require.Equal(t, uint32(0), c2)
	require.Equal(t, []byte("tree"), v2)
	c3, v3, p3 := root.Query(cms, []byte("block"))
	require.Equal(t, uint32(0), c3)
	require.Equal(t, []byte("chain"), v3)

	cstore, err := commitment.NewStore(root, []commitment.Proof{p1, p2, p3})
	require.NoError(t, err)

	require.True(t, cstore.Prove([]byte("hello"), []byte("world")))
	require.True(t, cstore.Prove([]byte("merkle"), []byte("tree")))
	require.True(t, cstore.Prove([]byte("block"), []byte("chain")))

	kvstore.Set(root.Key([]byte("12345")), []byte("67890"))
	kvstore.Set(root.Key([]byte("qwerty")), []byte("zxcv"))
	kvstore.Set(root.Key([]byte("hello")), []byte("dlrow"))

	root = commit(cms, root)

	c1, v1, p1 = root.Query(cms, []byte("12345"))
	require.Equal(t, uint32(0), c1)
	require.Equal(t, []byte("67890"), v1)
	c2, v2, p2 = root.Query(cms, []byte("qwerty"))
	require.Equal(t, uint32(0), c2)
	require.Equal(t, []byte("zxcv"), v2)
	c3, v3, p3 = root.Query(cms, []byte("hello"))
	require.Equal(t, uint32(0), c3)
	require.Equal(t, []byte("dlrow"), v3)

	cstore, err = commitment.NewStore(root, []commitment.Proof{p1, p2, p3})
	require.NoError(t, err)

	require.True(t, cstore.Prove([]byte("12345"), []byte("67890")))
	require.True(t, cstore.Prove([]byte("qwerty"), []byte("zxcv")))
	require.True(t, cstore.Prove([]byte("hello"), []byte("dlrow")))
}
