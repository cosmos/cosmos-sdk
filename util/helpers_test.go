package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestOK(t *testing.T) {
	assert := assert.New(t)

	ctx := MockContext("test-chain", 20)
	store := state.NewMemKVStore()
	data := "this looks okay"
	tx := sdk.Tx{}

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

	ctx := MockContext("test-chain", 20)
	store := state.NewMemKVStore()
	msg := "big problem"
	tx := sdk.Tx{}

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

	ctx := MockContext("test-chain", 20)
	store := state.NewMemKVStore()
	msg := "system crash!"
	tx := sdk.Tx{}

	fail := PanicHandler{Msg: msg}
	assert.Panics(func() { fail.CheckTx(ctx, store, tx) })
	assert.Panics(func() { fail.DeliverTx(ctx, store, tx) })
}
