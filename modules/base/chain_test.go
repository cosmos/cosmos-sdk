package base

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

func TestChainValidate(t *testing.T) {
	assert := assert.New(t)
	raw := stack.NewRawTx([]byte{1, 2, 3, 4})

	cases := []struct {
		name    string
		expires uint64
		valid   bool
	}{
		{"hello", 0, true},
		{"one-2-three", 123, true},
		{"super!@#$%@", 0, false},
		{"WISH_2_be", 14, true},
		{"öhhh", 54, false},
	}

	for _, tc := range cases {
		tx := NewChainTx(tc.name, tc.expires, raw)
		err := tx.ValidateBasic()
		if tc.valid {
			assert.Nil(err, "%s: %+v", tc.name, err)
		} else {
			assert.NotNil(err, tc.name)
		}
	}

	empty := NewChainTx("okay", 0, basecoin.Tx{})
	err := empty.ValidateBasic()
	assert.NotNil(err)
}

func TestChain(t *testing.T) {
	assert := assert.New(t)
	msg := "got it"
	chainID := "my-chain"

	raw := stack.NewRawTx([]byte{1, 2, 3, 4})
	cases := []struct {
		tx       basecoin.Tx
		valid    bool
		errorMsg string
	}{
		{NewChainTx(chainID, 0, raw), true, ""},
		{NewChainTx("someone-else", 0, raw), false, "someone-else"},
		{raw, false, "No chain id provided"},
	}

	// generic args here...
	ctx := stack.NewContext(chainID, 100, log.NewNopLogger())
	store := state.NewMemKVStore()

	// build the stack
	ok := stack.OKHandler{Log: msg}
	app := stack.New(Chain{}).Use(ok)

	for idx, tc := range cases {
		i := strconv.Itoa(idx)

		// make sure check returns error, not a panic crash
		res, err := app.CheckTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", idx, err)
			assert.Equal(msg, res.Log, i)
		} else {
			if assert.NotNil(err, i) {
				assert.Contains(err.Error(), tc.errorMsg, i)
			}
		}

		// make sure deliver returns error, not a panic crash
		res, err = app.DeliverTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", idx, err)
			assert.Equal(msg, res.Log, i)
		} else {
			if assert.NotNil(err, i) {
				assert.Contains(err.Error(), tc.errorMsg, i)
			}
		}
	}
}
