package stack

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/state"
)

func TestOK(t *testing.T) {
	assert := assert.New(t)

	ctx := NewContext("test-chain", 20, log.NewNopLogger())
	store := state.NewMemKVStore()
	data := "this looks okay"
	tx := basecoin.Tx{}

	ok := OKHandler{Log: data}
	res, err := ok.CheckTx(ctx, store, tx)
	assert.Nil(err, "%+v", err)
	assert.Equal(data, res.Log)

	dres, err := ok.DeliverTx(ctx, store, tx)
	assert.Nil(err, "%+v", err)
	assert.Equal(data, dres.Log)
}

func TestFail(t *testing.T) {
	assert := assert.New(t)

	ctx := NewContext("test-chain", 20, log.NewNopLogger())
	store := state.NewMemKVStore()
	msg := "big problem"
	tx := basecoin.Tx{}

	fail := FailHandler{Err: errors.New(msg)}
	_, err := fail.CheckTx(ctx, store, tx)
	if assert.NotNil(err) {
		assert.Equal(msg, err.Error())
	}

	_, err = fail.DeliverTx(ctx, store, tx)
	if assert.NotNil(err) {
		assert.Equal(msg, err.Error())
	}
}

func TestPanic(t *testing.T) {
	assert := assert.New(t)

	ctx := NewContext("test-chain", 20, log.NewNopLogger())
	store := state.NewMemKVStore()
	msg := "system crash!"
	tx := basecoin.Tx{}

	fail := PanicHandler{Msg: msg}
	assert.Panics(func() { fail.CheckTx(ctx, store, tx) })
	assert.Panics(func() { fail.DeliverTx(ctx, store, tx) })
}

func TestCheck(t *testing.T) {
	assert := assert.New(t)

	ctx := MockContext("check-chain", 123)
	store := state.NewMemKVStore()
	h := CheckHandler{}

	a := basecoin.Actor{App: "foo", Address: []byte("baz")}
	b := basecoin.Actor{App: "si-ly", Address: []byte("bar")}

	cases := []struct {
		valid             bool
		signers, required []basecoin.Actor
	}{
		{true, nil, nil},
		{true, []basecoin.Actor{a}, []basecoin.Actor{a}},
		{true, []basecoin.Actor{a, b}, []basecoin.Actor{a}},
		{false, []basecoin.Actor{a}, []basecoin.Actor{a, b}},
		{false, []basecoin.Actor{a}, []basecoin.Actor{b}},
	}

	for i, tc := range cases {
		tx := CheckTx{tc.required}.Wrap()
		myCtx := ctx.WithPermissions(tc.signers...)
		_, err := h.CheckTx(myCtx, store, tx)
		_, err2 := h.DeliverTx(myCtx, store, tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
			assert.Nil(err2, "%d: %+v", i, err2)
		} else {
			assert.NotNil(err, "%d", i)
			assert.NotNil(err2, "%d", i)
		}
	}
}
