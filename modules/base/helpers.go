package base

import (
	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	NameVal   = "val"
	NamePrice = "price"

	TypeValChange = NameVal + "/change"
	ByteValChange = 0xfe

	TypePriceShow = NamePrice + "/show"
	BytePriceShow = 0xfd
)

func init() {
	sdk.TxMapper.
		RegisterImplementation(ValChangeTx{}, TypeValChange, ByteValChange).
		RegisterImplementation(PriceShowTx{}, TypePriceShow, BytePriceShow)
}

//--------------------------------
// Setup tx and handler for validation test cases

type ValSetHandler struct {
	sdk.NopCheck
	sdk.NopInitState
	sdk.NopInitValidate
}

var _ sdk.Handler = ValSetHandler{}

func (ValSetHandler) Name() string {
	return NameVal
}

func (ValSetHandler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx) (res sdk.DeliverResult, err error) {
	change, ok := tx.Unwrap().(ValChangeTx)
	if !ok {
		return res, errors.ErrUnknownTxType(tx)
	}
	res.Diff = change.Diff
	return
}

type ValChangeTx struct {
	Diff []*abci.Validator
}

func (v ValChangeTx) Wrap() sdk.Tx {
	return sdk.Tx{v}
}

func (v ValChangeTx) ValidateBasic() error { return nil }

//--------------------------------
// Setup tx and handler for testing checktx fees/gas

// PriceData is the data we ping back
var PriceData = []byte{0xCA, 0xFE}

// PriceHandler returns checktx results based on the input
type PriceHandler struct {
	sdk.NopInitState
	sdk.NopInitValidate
}

var _ sdk.Handler = PriceHandler{}

func (PriceHandler) Name() string {
	return NamePrice
}

func (PriceHandler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx) (res sdk.CheckResult, err error) {
	price, ok := tx.Unwrap().(PriceShowTx)
	if !ok {
		return res, errors.ErrUnknownTxType(tx)
	}
	res.GasAllocated = price.GasAllocated
	res.GasPayment = price.GasPayment
	res.Data = PriceData
	return
}

func (PriceHandler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx) (res sdk.DeliverResult, err error) {
	_, ok := tx.Unwrap().(PriceShowTx)
	if !ok {
		return res, errors.ErrUnknownTxType(tx)
	}
	res.Data = PriceData
	return
}

// PriceShowTx lets us bounce back a given fee/gas on CheckTx
type PriceShowTx struct {
	GasAllocated uint64
	GasPayment   uint64
}

func NewPriceShowTx(gasAllocated, gasPayment uint64) sdk.Tx {
	return PriceShowTx{GasAllocated: gasAllocated, GasPayment: gasPayment}.Wrap()
}

func (p PriceShowTx) Wrap() sdk.Tx {
	return sdk.Tx{p}
}

func (v PriceShowTx) ValidateBasic() error { return nil }
