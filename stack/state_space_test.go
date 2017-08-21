package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

// writerMid is a middleware that writes the given bytes on CheckTx and DeliverTx
type writerMid struct {
	name       string
	key, value []byte
	PassInitValidate
}

var _ Middleware = writerMid{}

func (w writerMid) Name() string { return w.name }

func (w writerMid) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, next sdk.Checker) (sdk.CheckResult, error) {
	store.Set(w.key, w.value)
	return next.CheckTx(ctx, store, tx)
}

func (w writerMid) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, next sdk.Deliver) (sdk.DeliverResult, error) {
	store.Set(w.key, w.value)
	return next.DeliverTx(ctx, store, tx)
}

func (w writerMid) InitState(l log.Logger, store state.SimpleDB, module,
	key, value string, next sdk.InitStater) (string, error) {
	store.Set([]byte(key), []byte(value))
	return next.InitState(l, store, module, key, value)
}

// writerHand is a handler that writes the given bytes on CheckTx and DeliverTx
type writerHand struct {
	name       string
	key, value []byte
	sdk.NopInitValidate
}

var _ sdk.Handler = writerHand{}

func (w writerHand) Name() string { return w.name }

func (w writerHand) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx) (sdk.CheckResult, error) {
	store.Set(w.key, w.value)
	return sdk.CheckResult{}, nil
}

func (w writerHand) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx) (sdk.DeliverResult, error) {
	store.Set(w.key, w.value)
	return sdk.DeliverResult{}, nil
}

func (w writerHand) InitState(l log.Logger, store state.SimpleDB, module,
	key, value string) (string, error) {
	store.Set([]byte(key), []byte(value))
	return "Success", nil
}

func TestStateSpace(t *testing.T) {
	cases := []struct {
		h        sdk.Handler
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
		sdk.TxMapper.RegisterImplementation(RawTx{}, tc.h.Name(), byte(50+i))

		// run various tests on this setup
		spaceCheck(t, i, app, tc.expected)
		spaceDeliver(t, i, app, tc.expected)
		// spaceOption(t, i, app, keys)
	}
}

func spaceCheck(t *testing.T, i int, app sdk.Handler, keys []data.Bytes) {
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

func spaceDeliver(t *testing.T, i int, app sdk.Handler, keys []data.Bytes) {
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
