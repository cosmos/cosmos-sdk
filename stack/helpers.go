package stack

import (
	"github.com/pkg/errors"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

const (
	NameOK    = "ok"
	NameFail  = "fail"
	NamePanic = "panic"
)

// OKHandler just used to return okay to everything
type OKHandler struct {
	Log string
}

var _ basecoin.Handler = OKHandler{}

func (_ OKHandler) Name() string {
	return NameOK
}

// CheckTx always returns an empty success tx
func (ok OKHandler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return basecoin.Result{Log: ok.Log}, nil
}

// DeliverTx always returns an empty success tx
func (ok OKHandler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return basecoin.Result{Log: ok.Log}, nil
}

// FailHandler always returns an error
type FailHandler struct {
	Err error
}

var _ basecoin.Handler = FailHandler{}

func (_ FailHandler) Name() string {
	return NameFail
}

// CheckTx always returns the given error
func (f FailHandler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return res, errors.WithStack(f.Err)
}

// DeliverTx always returns the given error
func (f FailHandler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return res, errors.WithStack(f.Err)
}

// PanicHandler always panics, using the given error (first choice) or msg (fallback)
type PanicHandler struct {
	Msg string
	Err error
}

var _ basecoin.Handler = PanicHandler{}

func (_ PanicHandler) Name() string {
	return NamePanic
}

// CheckTx always panics
func (p PanicHandler) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	if p.Err != nil {
		panic(p.Err)
	}
	panic(p.Msg)
}

// DeliverTx always panics
func (p PanicHandler) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	if p.Err != nil {
		panic(p.Err)
	}
	panic(p.Msg)
}
