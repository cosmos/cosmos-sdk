package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/state"
)

// writerMid is a middleware that writes the given bytes on CheckTx and DeliverTx
type writerMid struct {
	name       string
	key, value []byte
	PassInitValidate
}

var _ Middleware = writerMid{}

func (w writerMid) Name() string { return w.name }

func (w writerMid) CheckTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, next basecoin.Checker) (basecoin.CheckResult, error) {
	store.Set(w.key, w.value)
	return next.CheckTx(ctx, store, tx)
}

func (w writerMid) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, next basecoin.Deliver) (basecoin.DeliverResult, error) {
	store.Set(w.key, w.value)
	return next.DeliverTx(ctx, store, tx)
}

func (w writerMid) InitState(l log.Logger, store state.SimpleDB, module,
	key, value string, next basecoin.InitStater) (string, error) {
	store.Set([]byte(key), []byte(value))
	return next.InitState(l, store, module, key, value)
}

// writerHand is a handler that writes the given bytes on CheckTx and DeliverTx
type writerHand struct {
	name       string
	key, value []byte
	basecoin.NopInitValidate
}

var _ basecoin.Handler = writerHand{}

func (w writerHand) Name() string { return w.name }

func (w writerHand) CheckTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx) (basecoin.CheckResult, error) {
	store.Set(w.key, w.value)
	return basecoin.CheckResult{}, nil
}

func (w writerHand) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx) (basecoin.DeliverResult, error) {
	store.Set(w.key, w.value)
	return basecoin.DeliverResult{}, nil
}

func (w writerHand) InitState(l log.Logger, store state.SimpleDB, module,
	key, value string) (string, error) {
	store.Set([]byte(key), []byte(value))
	return "Success", nil
}

func TestStateSpace(t *testing.T) {
	cases := []struct {
		h        basecoin.Handler
		m        []Middleware
		expected []data.Bytes
	}{
		{
			writerHand{name: "foo", key: []byte{1, 2}, value: []byte("bar")},
			[]Middleware{
				writerMid{name: "bing", key: []byte{1, 2}, value: []byte("bang")},
			},
			[]data.Bytes{
				{'f', 'o', 'o', 0, 1, 2},
				{'b', 'i', 'n', 'g', 0, 1, 2},
			},
		},
	}

	for i, tc := range cases {
		// make an app with this setup
		d := NewDispatcher(WrapHandler(tc.h))
		app := New(tc.m...).Use(d)

		// register so RawTx is routed to this handler
		basecoin.TxMapper.RegisterImplementation(RawTx{}, tc.h.Name(), byte(50+i))

		// run various tests on this setup
		spaceCheck(t, i, app, tc.expected)
		spaceDeliver(t, i, app, tc.expected)
		// spaceOption(t, i, app, keys)
	}
}

func spaceCheck(t *testing.T, i int, app basecoin.Handler, keys []data.Bytes) {
	assert := assert.New(t)
	require := require.New(t)

	ctx := MockContext("chain", 100)
	store := state.NewMemKVStore()

	// run a tx
	_, err := app.CheckTx(ctx, store, NewRawTx([]byte{77}))
	require.Nil(err, "%d: %+v", i, err)

	// verify that the data was writen
	for j, k := range keys {
		v := store.Get(k)
		assert.NotEmpty(v, "%d / %d", i, j)
	}
}

func spaceDeliver(t *testing.T, i int, app basecoin.Handler, keys []data.Bytes) {
	assert := assert.New(t)
	require := require.New(t)

	ctx := MockContext("chain", 100)
	store := state.NewMemKVStore()

	// run a tx
	_, err := app.DeliverTx(ctx, store, NewRawTx([]byte{1, 56}))
	require.Nil(err, "%d: %+v", i, err)

	// verify that the data was writen
	for j, k := range keys {
		v := store.Get(k)
		assert.NotEmpty(v, "%d / %d", i, j)
	}
}
