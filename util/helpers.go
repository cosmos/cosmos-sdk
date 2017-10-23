package util

import (
	"github.com/tendermint/go-wire/data"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	// ByteRawTx   = 0xF0
	// ByteCheckTx = 0xF1
	// ByteFailTx  = 0xF2

	// TypeRawTx   = NameOK + "/raw" // this will just say a-ok to RawTx
	// TypeCheckTx = NameCheck + "/tx"
	// TypeFailTx  = NameFail + "/tx"

	rawMaxSize = 2000 * 1000
)

func init() {
	// sdk.TxMapper.
	// 	RegisterImplementation(RawTx{}, TypeRawTx, ByteRawTx).
	// 	RegisterImplementation(CheckTx{}, TypeCheckTx, ByteCheckTx).
	// 	RegisterImplementation(FailTx{}, TypeFailTx, ByteFailTx)
}

// RawTx just contains bytes that can be hex-ified
type RawTx struct {
	Data data.Bytes
}

// ValidateBasic can ensure a limited size of tx
func (r RawTx) ValidateBasic() error {
	if len(r.Data) > rawMaxSize {
		return errors.ErrTooLarge()
	}
	return nil
}

// OKHandler just used to return okay to everything
type OKHandler struct {
	Log string
}

var _ sdk.Handler = OKHandler{}

// CheckTx always returns an empty success tx
func (ok OKHandler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx interface{}) (res sdk.CheckResult, err error) {
	return sdk.CheckResult{Log: ok.Log}, nil
}

// DeliverTx always returns an empty success tx
func (ok OKHandler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx interface{}) (res sdk.DeliverResult, err error) {
	return sdk.DeliverResult{Log: ok.Log}, nil
}

// EchoHandler returns success, echoing res.Data = tx bytes
type EchoHandler struct{}

var _ sdk.Handler = EchoHandler{}

// CheckTx returns input if RawTx comes in, otherwise panic
func (EchoHandler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	msg interface{}) (res sdk.CheckResult, err error) {
	raw := sdk.MustGetTx(msg).(RawTx)
	return sdk.CheckResult{Data: raw.Data}, nil
}

// DeliverTx returns input if RawTx comes in, otherwise panic
func (EchoHandler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	msg interface{}) (res sdk.DeliverResult, err error) {
	raw := sdk.MustGetTx(msg).(RawTx)
	return sdk.DeliverResult{Data: raw.Data}, nil
}

// FailHandler always returns an error
type FailHandler struct {
	Err error
}

var _ sdk.Handler = FailHandler{}

// CheckTx always returns the given error
func (f FailHandler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx interface{}) (res sdk.CheckResult, err error) {
	return res, errors.Wrap(f.Err)
}

// DeliverTx always returns the given error
func (f FailHandler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx interface{}) (res sdk.DeliverResult, err error) {
	return res, errors.Wrap(f.Err)
}

// PanicHandler always panics, using the given error (first choice) or msg (fallback)
type PanicHandler struct {
	Msg string
	Err error
}

var _ sdk.Handler = PanicHandler{}

// CheckTx always panics
func (p PanicHandler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx interface{}) (res sdk.CheckResult, err error) {

	if p.Err != nil {
		panic(p.Err)
	}
	panic(p.Msg)
}

// DeliverTx always panics
func (p PanicHandler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx interface{}) (res sdk.DeliverResult, err error) {

	if p.Err != nil {
		panic(p.Err)
	}
	panic(p.Msg)
}
