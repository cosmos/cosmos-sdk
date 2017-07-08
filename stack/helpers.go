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
	ByteRawTx  = 0x1
	TypeRawTx  = "raw"
	rawMaxSize = 2000 * 1000
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(RawTx{}, TypeRawTx, ByteRawTx)
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
func (ok OKHandler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return basecoin.Result{Log: ok.Log}, nil
}

// DeliverTx always returns an empty success tx
func (ok OKHandler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return basecoin.Result{Log: ok.Log}, nil
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
func (EchoHandler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	data, err := data.ToWire(tx)
	return basecoin.Result{Data: data}, err
}

// DeliverTx always returns an empty success tx
func (EchoHandler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	data, err := data.ToWire(tx)
	return basecoin.Result{Data: data}, err
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
func (f FailHandler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	return res, errors.Wrap(f.Err)
}

// DeliverTx always returns the given error
func (f FailHandler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
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
func (p PanicHandler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	if p.Err != nil {
		panic(p.Err)
	}
	panic(p.Msg)
}

// DeliverTx always panics
func (p PanicHandler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	if p.Err != nil {
		panic(p.Err)
	}
	panic(p.Msg)
}
