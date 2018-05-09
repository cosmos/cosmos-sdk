package baseapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	abci "github.com/tendermint/abci/types"
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

func TestContextGetOpShouldNeverPanic(t *testing.T) {
	var ms MultiStore
	ctx := NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	indices := []int64{
		-10, 1, 0, 10, 20,
	}

	for _, index := range indices {
		_, _ = ctx.GetOp(index)
	}
}

func defaultContext(key StoreKey) Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := NewContext(cms, abci.Header{}, false, nil, log.NewNopLogger())
	return ctx
}

func TestCacheContext(t *testing.T) {
	key := NewKVStoreKey(t.Name())
	k1 := []byte("hello")
	v1 := []byte("world")
	k2 := []byte("key")
	v2 := []byte("value")

	ctx := defaultContext(key)
	store := ctx.KVStore(key)
	store.Set(k1, v1)
	assert.Equal(t, v1, store.Get(k1))
	assert.Nil(t, store.Get(k2))

	cctx, write := ctx.CacheContext()
	cstore := cctx.KVStore(key)
	assert.Equal(t, v1, cstore.Get(k1))
	assert.Nil(t, cstore.Get(k2))

	cstore.Set(k2, v2)
	assert.Equal(t, v2, cstore.Get(k2))
	assert.Nil(t, store.Get(k2))

	write()

	assert.Equal(t, v2, store.Get(k2))
}

func TestLogContext(t *testing.T) {
	key := NewKVStoreKey(t.Name())
	ctx := defaultContext(key)
	logger := NewMockLogger()
	ctx = ctx.WithLogger(logger)
	ctx.Logger().Debug("debug")
	ctx.Logger().Info("info")
	ctx.Logger().Error("error")
	require.Equal(t, *logger.logs, []string{"debug", "info", "error"})
}
