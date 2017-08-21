package fee

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
)

// nolint
const (
	ByteFees = 0x28
	TypeFees = NameFee + "/tx"
)

func init() {
	sdk.TxMapper.
		RegisterImplementation(Fee{}, TypeFees, ByteFees)
}

/**** Fee ****/

// Fee attaches a fee payment to the embedded tx
type Fee struct {
	// Gas coin.Coin `json:"gas"`  // ?????
	Fee   coin.Coin      `json:"fee"`
	Payer sdk.Actor `json:"payer"` // the address who pays the fee
	Tx    sdk.Tx    `json:"tx"`
}

// NewFee wraps a tx with a promised fee from this actor
func NewFee(tx sdk.Tx, fee coin.Coin, payer sdk.Actor) sdk.Tx {
	return Fee{Tx: tx, Fee: fee, Payer: payer}.Wrap()
}

// nolint - TxInner Functions
func (f Fee) ValidateBasic() error {
	// TODO: more checks
	return f.Tx.ValidateBasic()
}
func (f Fee) Wrap() sdk.Tx {
	return sdk.Tx{f}
}
func (f Fee) Next() sdk.Tx {
	return f.Tx
}
