package fee

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/coin"
)

// nolint
const (
	ByteFees = 0x21
	TypeFees = NameFee + "/tx"
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(Fee{}, TypeFees, ByteFees)
}

/**** Fee ****/

// Fee attaches a fee payment to the embedded tx
type Fee struct {
	// Gas coin.Coin `json:"gas"`  // ?????
	Fee   coin.Coin      `json:"fee"`
	Payer basecoin.Actor `json:"payer"` // the address who pays the fee
	Tx    basecoin.Tx    `json:"tx"`
}

// NewFee wraps a tx with a promised fee from this actor
func NewFee(tx basecoin.Tx, fee coin.Coin, payer basecoin.Actor) basecoin.Tx {
	return Fee{Tx: tx, Fee: fee, Payer: payer}.Wrap()
}

// nolint - TxInner Functions
func (f Fee) ValidateBasic() error {
	// TODO: more checks
	return f.Tx.ValidateBasic()
}
func (f Fee) Wrap() basecoin.Tx {
	return basecoin.Tx{f}
}
func (f Fee) Next() basecoin.Tx {
	return f.Tx
}
