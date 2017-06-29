package stack

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
)

func TestChain(t *testing.T) {
	assert := assert.New(t)
	msg := "got it"
	chainID := "my-chain"

	raw := txs.NewRaw([]byte{1, 2, 3, 4}).Wrap()
	cases := []struct {
		tx       basecoin.Tx
		valid    bool
		errorMsg string
	}{
		{txs.NewChain(chainID, raw).Wrap(), true, ""},
		{txs.NewChain("someone-else", raw).Wrap(), false, "Tx belongs to different chain - someone-else"},
		{raw, false, "No chain id provided"},
	}

	// generic args here...
	ctx := NewContext(log.NewNopLogger())
	store := types.NewMemKVStore()

	// build the stack
	ok := OKHandler{msg}
	app := New(Chain{chainID}).Use(ok)

	for idx, tc := range cases {
		i := strconv.Itoa(idx)

		// make sure check returns error, not a panic crash
		res, err := app.CheckTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", idx, err)
			assert.Equal(msg, res.Log, i)
		} else {
			if assert.NotNil(err, i) {
				assert.Equal(tc.errorMsg, err.Error(), i)
			}
		}

		// make sure deliver returns error, not a panic crash
		res, err = app.DeliverTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", idx, err)
			assert.Equal(msg, res.Log, i)
		} else {
			if assert.NotNil(err, i) {
				assert.Equal(tc.errorMsg, err.Error(), i)
			}
		}
	}
}
