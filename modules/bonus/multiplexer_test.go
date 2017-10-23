package bonus

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/stretchr/testify/assert"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"
)

func TestMultiplexer(t *testing.T) {
	assert := assert.New(t)
	msg := "diddly"
	chainID := "multi-verse"
	height := uint64(100)

	// Generic args here...
	store := state.NewMemKVStore()
	ctx := stack.NewContext(chainID, height, log.NewNopLogger())

	// Build the stack
	app := stack.
		New(Multiplexer{}).
		Dispatch(
			stack.WrapHandler(stack.OKHandler{Log: msg}),
			stack.WrapHandler(PriceHandler{}),
		)

	raw := stack.NewRawTx([]byte{1, 2, 3, 4})
	fail := stack.NewFailTx()
	price1 := NewPriceShowTx(123, 456)
	price2 := NewPriceShowTx(1000, 2000)
	price3 := NewPriceShowTx(11, 0)

	join := func(data ...[]byte) []byte {
		return wire.BinaryBytes(data)
	}

	cases := [...]struct {
		tx           sdk.Tx
		valid        bool
		gasAllocated uint64
		gasPayment   uint64
		log          string
		data         data.Bytes
	}{
		// test the components without multiplexer (no effect)
		0: {raw, true, 0, 0, msg, nil},
		1: {price1, true, 123, 456, "", PriceData},
		2: {fail, false, 0, 0, "", nil},
		// test multiplexer on error
		3: {NewMultiTx(raw, fail, price1), false, 0, 0, "", nil},
		// test combining info on multiplexer
		4: {NewMultiTx(price1, raw), true, 123, 456, "\n" + msg, join(PriceData, nil)},
		// add lots of prices
		5: {NewMultiTx(price1, price2, price3), true, 1134, 2456, "\n\n", join(PriceData, PriceData, PriceData)},
		// combine multiple  logs
		6: {NewMultiTx(raw, price3, raw), true, 11, 0, msg + "\n\n" + msg, join(nil, PriceData, nil)},
	}

	for i, tc := range cases {
		cres, err := app.CheckTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
			assert.Equal(tc.log, cres.Log, "%d", i)
			assert.Equal(tc.data, cres.Data, "%d", i)
			assert.Equal(tc.gasAllocated, cres.GasAllocated, "%d", i)
			assert.Equal(tc.gasPayment, cres.GasPayment, "%d", i)
		} else {
			assert.NotNil(err, "%d", i)
		}

		// make sure deliver returns error, not a panic crash
		dres, err := app.DeliverTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
			assert.Equal(tc.log, dres.Log, "%d", i)
			assert.Equal(tc.data, dres.Data, "%d", i)
		} else {
			assert.NotNil(err, "%d", i)
		}
	}
}
