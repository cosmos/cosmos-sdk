package util

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestRecovery(t *testing.T) {
	assert := assert.New(t)

	// generic args here...
	ctx := MockContext("test-chain", 20)
	store := state.NewMemKVStore()
	tx := 0 // we ignore it, so it can be anything

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
		app := sdk.ChainDecorators(rec).WithHandler(fail)

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
