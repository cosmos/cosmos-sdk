package stack

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

func TestOK(t *testing.T) {
	assert := assert.New(t)

	ctx := NewContext("test-chain", log.NewNopLogger())
	store := types.NewMemKVStore()
	data := "this looks okay"
	tx := basecoin.Tx{}

	ok := OKHandler{Log: data}
	res, err := ok.CheckTx(ctx, store, tx)
	assert.Nil(err, "%+v", err)
	assert.Equal(data, res.Log)

	res, err = ok.DeliverTx(ctx, store, tx)
	assert.Nil(err, "%+v", err)
	assert.Equal(data, res.Log)
}

func TestFail(t *testing.T) {
	assert := assert.New(t)

	ctx := NewContext("test-chain", log.NewNopLogger())
	store := types.NewMemKVStore()
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

	ctx := NewContext("test-chain", log.NewNopLogger())
	store := types.NewMemKVStore()
	msg := "system crash!"
	tx := basecoin.Tx{}

	fail := PanicHandler{Msg: msg}
	assert.Panics(func() { fail.CheckTx(ctx, store, tx) })
	assert.Panics(func() { fail.DeliverTx(ctx, store, tx) })
}
