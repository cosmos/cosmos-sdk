package stack

import (
	"github.com/pkg/errors"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

const (
	NameVoid  = "void"
	NameFail  = "fail"
	NamePanic = "panic"
)

// voidHandler just used to return okay to everything
type voidHandler struct{}

var _ basecoin.Handler = voidHandler{}

func (_ voidHandler) Name() string {
	return NameVoid
}

// CheckTx always returns an empty success tx
func (_ voidHandler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return
}

// DeliverTx always returns an empty success tx
func (_ voidHandler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return
}

// failHandler always returns an error
type failHandler struct {
	err error
}

var _ basecoin.Handler = failHandler{}

func (_ failHandler) Name() string {
	return NameFail
}

// CheckTx always returns the given error
func (f failHandler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return res, errors.WithStack(f.err)
}

// DeliverTx always returns the given error
func (f failHandler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return res, errors.WithStack(f.err)
}

// panicHandler always panics, using the given error (first choice) or msg (fallback)
type panicHandler struct {
	msg string
	err error
}

var _ basecoin.Handler = panicHandler{}

func (_ panicHandler) Name() string {
	return NamePanic
}

// CheckTx always panics
func (p panicHandler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	if p.err != nil {
		panic(p.err)
	}
	panic(p.msg)
}

// DeliverTx always panics
func (p panicHandler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	if p.err != nil {
		panic(p.err)
	}
	panic(p.msg)
}
