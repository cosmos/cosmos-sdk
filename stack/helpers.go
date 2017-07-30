package stack

import (
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

//nolint
const (
	NameOK    = "ok"
	NameFail  = "fail"
	NamePanic = "panic"
	NameEcho  = "echo"
)

//nolint
const (
	ByteRawTx   = 0xF0
	ByteCheckTx = 0xF1
	ByteFailTx  = 0xF2

	TypeRawTx   = NameOK + "/raw" // this will just say a-ok to RawTx
	TypeCheckTx = NameCheck + "/tx"
	TypeFailTx  = NameFail + "/tx"

	rawMaxSize = 2000 * 1000
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(RawTx{}, TypeRawTx, ByteRawTx).
		RegisterImplementation(CheckTx{}, TypeCheckTx, ByteCheckTx).
		RegisterImplementation(FailTx{}, TypeFailTx, ByteFailTx)
}

// RawTx just contains bytes that can be hex-ified
type RawTx struct {
	data.Bytes
}

var _ basecoin.TxInner = RawTx{}

// nolint
func NewRawTx(d []byte) basecoin.Tx {
	return RawTx{data.Bytes(d)}.Wrap()
}
func (r RawTx) Wrap() basecoin.Tx {
	return basecoin.Tx{r}
}
func (r RawTx) ValidateBasic() error {
	if len(r.Bytes) > rawMaxSize {
		return errors.ErrTooLarge()
	}
	return nil
}

// CheckTx contains a list of permissions to be tested
type CheckTx struct {
	Required []basecoin.Actor
}

var _ basecoin.TxInner = CheckTx{}

// nolint
func NewCheckTx(req []basecoin.Actor) basecoin.Tx {
	return CheckTx{req}.Wrap()
}
func (c CheckTx) Wrap() basecoin.Tx {
	return basecoin.Tx{c}
}
func (CheckTx) ValidateBasic() error {
	return nil
}

// FailTx just gets routed to filaure
type FailTx struct{}

var _ basecoin.TxInner = FailTx{}

func NewFailTx() basecoin.Tx {
	return FailTx{}.Wrap()
}

func (f FailTx) Wrap() basecoin.Tx {
	return basecoin.Tx{f}
}
func (r FailTx) ValidateBasic() error {
	return nil
}

// OKHandler just used to return okay to everything
type OKHandler struct {
	Log string
	basecoin.NopOption
}

var _ basecoin.Handler = OKHandler{}

// Name - return handler's name
func (OKHandler) Name() string {
	return NameOK
}

// CheckTx always returns an empty success tx
func (ok OKHandler) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.CheckResult, err error) {
	return basecoin.CheckResult{Log: ok.Log}, nil
}

// DeliverTx always returns an empty success tx
func (ok OKHandler) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.DeliverResult, err error) {
	return basecoin.DeliverResult{Log: ok.Log}, nil
}

// EchoHandler returns success, echoing res.Data = tx bytes
type EchoHandler struct {
	basecoin.NopOption
}

var _ basecoin.Handler = EchoHandler{}

// Name - return handler's name
func (EchoHandler) Name() string {
	return NameEcho
}

// CheckTx always returns an empty success tx
func (EchoHandler) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.CheckResult, err error) {
	data, err := data.ToWire(tx)
	return basecoin.CheckResult{Data: data}, err
}

// DeliverTx always returns an empty success tx
func (EchoHandler) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.DeliverResult, err error) {
	data, err := data.ToWire(tx)
	return basecoin.DeliverResult{Data: data}, err
}

// FailHandler always returns an error
type FailHandler struct {
	Err error
	basecoin.NopOption
}

var _ basecoin.Handler = FailHandler{}

// Name - return handler's name
func (FailHandler) Name() string {
	return NameFail
}

// CheckTx always returns the given error
func (f FailHandler) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.CheckResult, err error) {
	return res, errors.Wrap(f.Err)
}

// DeliverTx always returns the given error
func (f FailHandler) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.DeliverResult, err error) {
	return res, errors.Wrap(f.Err)
}

// PanicHandler always panics, using the given error (first choice) or msg (fallback)
type PanicHandler struct {
	Msg string
	Err error
	basecoin.NopOption
}

var _ basecoin.Handler = PanicHandler{}

// Name - return handler's name
func (PanicHandler) Name() string {
	return NamePanic
}

// CheckTx always panics
func (p PanicHandler) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.CheckResult, err error) {
	if p.Err != nil {
		panic(p.Err)
	}
	panic(p.Msg)
}

// DeliverTx always panics
func (p PanicHandler) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.DeliverResult, err error) {
	if p.Err != nil {
		panic(p.Err)
	}
	panic(p.Msg)
}

// CheckHandler accepts CheckTx and verifies the permissions
type CheckHandler struct {
	basecoin.NopOption
}

var _ basecoin.Handler = CheckHandler{}

// Name - return handler's name
func (CheckHandler) Name() string {
	return NameCheck
}

// CheckTx verifies the permissions
func (c CheckHandler) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.CheckResult, err error) {
	check, ok := tx.Unwrap().(CheckTx)
	if !ok {
		return res, errors.ErrUnknownTxType(tx)
	}

	for _, perm := range check.Required {
		if !ctx.HasPermission(perm) {
			return res, errors.ErrUnauthorized()
		}
	}
	return res, nil
}

// DeliverTx verifies the permissions
func (c CheckHandler) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.DeliverResult, err error) {
	check, ok := tx.Unwrap().(CheckTx)
	if !ok {
		return res, errors.ErrUnknownTxType(tx)
	}

	for _, perm := range check.Required {
		if !ctx.HasPermission(perm) {
			return res, errors.ErrUnauthorized()
		}
	}
	return res, nil
}
