package handlers

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

const (
	NameVoid = "void"
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
