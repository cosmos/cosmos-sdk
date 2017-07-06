package fee

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/coin"
)

// nolint
const (
	ByteFees = 0x20
	TypeFees = "fee"
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(&Fee{}, TypeFees, ByteFees)
}

/**** Fee ****/

// Fee attaches a fee payment to the embedded tx
type Fee struct {
	Tx    basecoin.Tx    `json:"tx"`
	Fee   coin.Coin      `json:"fee"`
	Payer basecoin.Actor `json:"payer"` // the address who pays the fee
	// Gas coin.Coin `json:"gas"`  // ?????
}

func NewFee(tx basecoin.Tx, fee coin.Coin, payer basecoin.Actor) basecoin.Tx {
	return (&Fee{Tx: tx, Fee: fee, Payer: payer}).Wrap()
}

func (f *Fee) ValidateBasic() error {
	// TODO: more checks
	return f.Tx.ValidateBasic()
}

func (f *Fee) Wrap() basecoin.Tx {
	return basecoin.Tx{f}
}

func (f *Fee) Next() basecoin.Tx {
	return f.Tx
}
