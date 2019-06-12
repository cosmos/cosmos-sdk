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
	key := sdk.NewKVStoreKey("ibc")
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

func key(str string) []byte {
	return append([]byte{0x00}, []byte(str)...)
}

func query(t *testing.T, cms types.CommitMultiStore, k string) (value []byte, proof Proof) {
	qres := cms.(types.Queryable).Query(abci.RequestQuery{Path: "/ibc/key", Data: key(k), Prove: true})
	require.Equal(t, uint32(0), qres.Code, qres.Log)
	value = qres.Value
	proof = Proof{
		Key:   []byte(k),
		Proof: qres.Proof,
	}
	return
}

func TestStore(t *testing.T) {
	k, ctx, cms, _ := defaultComponents()
	kvstore := ctx.KVStore(k)

	kvstore.Set(key("hello"), []byte("world"))
	kvstore.Set(key("merkle"), []byte("tree"))
	kvstore.Set(key("block"), []byte("chain"))

	cid := cms.Commit()

	v1, p1 := query(t, cms, "hello")
	require.Equal(t, []byte("world"), v1)
	v2, p2 := query(t, cms, "merkle")
	require.Equal(t, []byte("tree"), v2)
	v3, p3 := query(t, cms, "block")
	require.Equal(t, []byte("chain"), v3)

	cstore, err := commitment.NewStore(cid.Hash, []commitment.Proof{p1, p2, p3})
	require.NoError(t, err)

	require.True(t, cstore.Prove([]byte("hello"), []byte("world")))
	require.True(t, cstore.Prove([]byte("merkle"), []byte("tree")))
	require.True(t, cstore.Prove([]byte("block"), []byte("chain")))
}
