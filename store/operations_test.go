package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
	"testing"
	"time"
)

var (
	testSKName = "test"
	testKey    = []byte("testkey")
)

func upsert(ctx sdk.Context, sk sdk.StoreKey, cdc *amino.Codec, key []byte, val interface{}) {
	s := ctx.KVStore(sk)
	s.Set(key, cdc.MustMarshalBinaryBare(val))
}

func TestDel(t *testing.T) {
	cdc := amino.NewCodec()
	ctx, sk := mockApp(t)
	require.Error(t, Del(ctx, sk, testKey))
	upsert(ctx, sk, cdc, testKey, 12345)
	require.NoError(t, Del(ctx, sk, testKey))
	require.False(t, Has(ctx, sk, testKey))
}

func TestHas(t *testing.T) {
	cdc := amino.NewCodec()
	ctx, sk := mockApp(t)
	require.False(t, Has(ctx, sk, testKey))
	upsert(ctx, sk, cdc, testKey, 12345)
	require.True(t, Has(ctx, sk, testKey))
}

func mockApp(t *testing.T) (sdk.Context, sdk.StoreKey) {
	keys := sdk.NewKVStoreKeys(testSKName)
	ms := NewCommitMultiStore(db.NewMemDB())
	ms.MountStoreWithDB(keys[testSKName], sdk.StoreTypeIAVL, db.NewMemDB())
	require.NoError(t, ms.LoadVersion(0))
	hdr := abci.Header{ChainID: "unit-test-chain", Height: 1, Time: time.Unix(1558332092, 0)}
	return sdk.NewContext(ms, hdr, false, log.NewNopLogger()), keys[testSKName]
}

func TestIncrementSeq(t *testing.T) {
	ctx, sk := mockApp(t)
	require.True(t, sdk.OneUint().Equal(IncrementSeq(ctx, sk, testKey)))
	require.True(t, sdk.NewUint(2).Equal(IncrementSeq(ctx, sk, testKey)))
}

func TestGetSeq(t *testing.T) {
	ctx, sk := mockApp(t)
	require.True(t, sdk.ZeroUint().Equal(GetSeq(ctx, sk, testKey)))
	IncrementSeq(ctx, sk, testKey)
	require.True(t, sdk.OneUint().Equal(GetSeq(ctx, sk, testKey)))
}
