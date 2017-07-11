package stack

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/state"
)

func TestRecovery(t *testing.T) {
	assert := assert.New(t)

	// generic args here...
	ctx := NewContext("test-chain", 20, log.NewNopLogger())
	store := state.NewMemKVStore()
	tx := basecoin.Tx{}

	cases := []struct {
		msg      string // what to send to panic
		err      error  // what to send to panic
		expected string // expected text in panic
	}{
		{"buzz", nil, "buzz"},
		{"", errors.New("some text"), "some text"},
		{"text", errors.New("error"), "error"},
	}

	for idx, tc := range cases {
		i := strconv.Itoa(idx)
		fail := PanicHandler{Msg: tc.msg, Err: tc.err}
		rec := Recovery{}
		app := New(rec).Use(fail)

		// make sure check returns error, not a panic crash
		_, err := app.CheckTx(ctx, store, tx)
		if assert.NotNil(err, i) {
			assert.Equal(tc.expected, err.Error(), i)
		}

		// make sure deliver returns error, not a panic crash
		_, err = app.DeliverTx(ctx, store, tx)
		if assert.NotNil(err, i) {
			assert.Equal(tc.expected, err.Error(), i)
		}

	}
}
