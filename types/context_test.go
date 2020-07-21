package types_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
)

type MockLogger struct {
	logs *[]string
}

func NewMockLogger() MockLogger {
	logs := make([]string, 0)
	return MockLogger{
		&logs,
	}
}

func (l MockLogger) Debug(msg string, kvs ...interface{}) {
	*l.logs = append(*l.logs, msg)
}

func (l MockLogger) Info(msg string, kvs ...interface{}) {
	*l.logs = append(*l.logs, msg)
}

func (l MockLogger) Error(msg string, kvs ...interface{}) {
	*l.logs = append(*l.logs, msg)
}

func (l MockLogger) With(kvs ...interface{}) log.Logger {
	panic("not implemented")
}

func defaultContext(t *testing.T, key types.StoreKey) types.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, types.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	require.NoError(t, err)
	ctx := types.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())
	return ctx
}

func TestCacheContext(t *testing.T) {
	key := types.NewKVStoreKey(t.Name())
	k1 := []byte("hello")
	v1 := []byte("world")
	k2 := []byte("key")
	v2 := []byte("value")

	ctx := defaultContext(t, key)
	store := ctx.KVStore(key)
	store.Set(k1, v1)
	require.Equal(t, v1, store.Get(k1))
	require.Nil(t, store.Get(k2))

	cctx, write := ctx.CacheContext()
	cstore := cctx.KVStore(key)
	require.Equal(t, v1, cstore.Get(k1))
	require.Nil(t, cstore.Get(k2))

	cstore.Set(k2, v2)
	require.Equal(t, v2, cstore.Get(k2))
	require.Nil(t, store.Get(k2))

	write()

	require.Equal(t, v2, store.Get(k2))
}

func TestLogContext(t *testing.T) {
	key := types.NewKVStoreKey(t.Name())
	ctx := defaultContext(t, key)
	logger := NewMockLogger()
	ctx = ctx.WithLogger(logger)
	ctx.Logger().Debug("debug")
	ctx.Logger().Info("info")
	ctx.Logger().Error("error")
	require.Equal(t, *logger.logs, []string{"debug", "info", "error"})
}

type dummy int64 //nolint:unused

func (d dummy) Clone() interface{} {
	return d
}

// Testing saving/loading sdk type values to/from the context
func TestContextWithCustom(t *testing.T) {
	var ctx types.Context
	require.True(t, ctx.IsZero())

	header := tmproto.Header{}
	height := int64(1)
	chainid := "chainid"
	ischeck := true
	txbytes := []byte("txbytes")
	logger := NewMockLogger()
	voteinfos := []abci.VoteInfo{{}}
	meter := types.NewGasMeter(10000)
	blockGasMeter := types.NewGasMeter(20000)
	minGasPrices := types.DecCoins{types.NewInt64DecCoin("feetoken", 1)}

	ctx = types.NewContext(nil, header, ischeck, logger)
	require.Equal(t, header, ctx.BlockHeader())

	ctx = ctx.
		WithBlockHeight(height).
		WithChainID(chainid).
		WithTxBytes(txbytes).
		WithVoteInfos(voteinfos).
		WithGasMeter(meter).
		WithMinGasPrices(minGasPrices).
		WithBlockGasMeter(blockGasMeter)
	require.Equal(t, height, ctx.BlockHeight())
	require.Equal(t, chainid, ctx.ChainID())
	require.Equal(t, ischeck, ctx.IsCheckTx())
	require.Equal(t, txbytes, ctx.TxBytes())
	require.Equal(t, logger, ctx.Logger())
	require.Equal(t, voteinfos, ctx.VoteInfos())
	require.Equal(t, meter, ctx.GasMeter())
	require.Equal(t, minGasPrices, ctx.MinGasPrices())
	require.Equal(t, blockGasMeter, ctx.BlockGasMeter())

	require.False(t, ctx.WithIsCheckTx(false).IsCheckTx())

	// test IsReCheckTx
	require.False(t, ctx.IsReCheckTx())
	ctx = ctx.WithIsCheckTx(false)
	ctx = ctx.WithIsReCheckTx(true)
	require.True(t, ctx.IsCheckTx())
	require.True(t, ctx.IsReCheckTx())

	// test consensus param
	require.Nil(t, ctx.ConsensusParams())
	cp := &abci.ConsensusParams{}
	require.Equal(t, cp, ctx.WithConsensusParams(cp).ConsensusParams())

	// test inner context
	newContext := context.WithValue(ctx.Context(), "key", "value") //nolint:golint,staticcheck
	require.NotEqual(t, ctx.Context(), ctx.WithContext(newContext).Context())
}

// Testing saving/loading of header fields to/from the context
func TestContextHeader(t *testing.T) {
	var ctx types.Context

	height := int64(5)
	time := time.Now()
	addr := secp256k1.GenPrivKey().PubKey().Address()
	proposer := types.ConsAddress(addr)

	ctx = types.NewContext(nil, tmproto.Header{}, false, nil)

	ctx = ctx.
		WithBlockHeight(height).
		WithBlockTime(time).
		WithProposer(proposer)
	require.Equal(t, height, ctx.BlockHeight())
	require.Equal(t, height, ctx.BlockHeader().Height)
	require.Equal(t, time.UTC(), ctx.BlockHeader().Time)
	require.Equal(t, proposer.Bytes(), ctx.BlockHeader().ProposerAddress)
}

func TestContextHeaderClone(t *testing.T) {
	cases := map[string]struct {
		h tmproto.Header
	}{
		"empty": {
			h: tmproto.Header{},
		},
		"height": {
			h: tmproto.Header{
				Height: 77,
			},
		},
		"time": {
			h: tmproto.Header{
				Time: time.Unix(12345677, 12345),
			},
		},
		"zero time": {
			h: tmproto.Header{
				Time: time.Unix(0, 0),
			},
		},
		"many items": {
			h: tmproto.Header{
				Height:  823,
				Time:    time.Unix(9999999999, 0),
				ChainID: "silly-demo",
			},
		},
		"many items with hash": {
			h: tmproto.Header{
				Height:        823,
				Time:          time.Unix(9999999999, 0),
				ChainID:       "silly-demo",
				AppHash:       []byte{5, 34, 11, 3, 23},
				ConsensusHash: []byte{11, 3, 23, 87, 3, 1},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			ctx := types.NewContext(nil, tc.h, false, nil)
			require.Equal(t, tc.h.Height, ctx.BlockHeight())
			require.Equal(t, tc.h.Time.UTC(), ctx.BlockTime())

			// update only changes one field
			var newHeight int64 = 17
			ctx = ctx.WithBlockHeight(newHeight)
			require.Equal(t, newHeight, ctx.BlockHeight())
			require.Equal(t, tc.h.Time.UTC(), ctx.BlockTime())
		})
	}
}

func TestUnwrapSDKContext(t *testing.T) {
	sdkCtx := types.NewContext(nil, tmproto.Header{}, false, nil)
	ctx := types.WrapSDKContext(sdkCtx)
	sdkCtx2 := types.UnwrapSDKContext(ctx)
	require.Equal(t, sdkCtx, sdkCtx2)

	ctx = context.Background()
	require.Panics(t, func() {
		types.UnwrapSDKContext(ctx)
	})
}
