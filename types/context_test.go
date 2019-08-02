package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/log"
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

func defaultContext(key types.StoreKey) types.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, types.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := types.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	return ctx
}

func TestCacheContext(t *testing.T) {
	key := types.NewKVStoreKey(t.Name())
	k1 := []byte("hello")
	v1 := []byte("world")
	k2 := []byte("key")
	v2 := []byte("value")

	ctx := defaultContext(key)
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
	ctx := defaultContext(key)
	logger := NewMockLogger()
	ctx = ctx.WithLogger(logger)
	ctx.Logger().Debug("debug")
	ctx.Logger().Info("info")
	ctx.Logger().Error("error")
	require.Equal(t, *logger.logs, []string{"debug", "info", "error"})
}

type dummy int64

func (d dummy) Clone() interface{} {
	return d
}

// Testing saving/loading sdk type values to/from the context
func TestContextWithCustom(t *testing.T) {
	var ctx types.Context
	require.True(t, ctx.IsZero())

	header := abci.Header{}
	height := int64(1)
	chainid := "chainid"
	ischeck := true
	txbytes := []byte("txbytes")
	logger := NewMockLogger()
	voteinfos := []abci.VoteInfo{{}}
	meter := types.NewGasMeter(10000)
	minGasPrices := types.DecCoins{types.NewInt64DecCoin("feetoken", 1)}

	ctx = types.NewContext(nil, header, ischeck, logger)
	require.Equal(t, header, ctx.BlockHeader())

	ctx = ctx.
		WithBlockHeight(height).
		WithChainID(chainid).
		WithTxBytes(txbytes).
		WithVoteInfos(voteinfos).
		WithGasMeter(meter).
		WithMinGasPrices(minGasPrices)
	require.Equal(t, height, ctx.BlockHeight())
	require.Equal(t, chainid, ctx.ChainID())
	require.Equal(t, ischeck, ctx.IsCheckTx())
	require.Equal(t, txbytes, ctx.TxBytes())
	require.Equal(t, logger, ctx.Logger())
	require.Equal(t, voteinfos, ctx.VoteInfos())
	require.Equal(t, meter, ctx.GasMeter())
	require.Equal(t, minGasPrices, ctx.MinGasPrices())
}

// Testing saving/loading of header fields to/from the context
func TestContextHeader(t *testing.T) {
	var ctx types.Context

	height := int64(5)
	time := time.Now()
	addr := secp256k1.GenPrivKey().PubKey().Address()
	proposer := types.ConsAddress(addr)

	ctx = types.NewContext(nil, abci.Header{}, false, nil)

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
		h abci.Header
	}{
		"empty": {
			h: abci.Header{},
		},
		"height": {
			h: abci.Header{
				Height: 77,
			},
		},
		"time": {
			h: abci.Header{
				Time: time.Unix(12345677, 12345),
			},
		},
		"zero time": {
			h: abci.Header{
				Time: time.Unix(0, 0),
			},
		},
		"many items": {
			h: abci.Header{
				Height:  823,
				Time:    time.Unix(9999999999, 0),
				ChainID: "silly-demo",
			},
		},
		"many items with hash": {
			h: abci.Header{
				Height:        823,
				Time:          time.Unix(9999999999, 0),
				ChainID:       "silly-demo",
				AppHash:       []byte{5, 34, 11, 3, 23},
				ConsensusHash: []byte{11, 3, 23, 87, 3, 1},
			},
		},
	}

	for name, tc := range cases {
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
