package fee

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

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
	Fee   types.Coin     `json:"fee"`
	Payer basecoin.Actor `json:"payer"` // the address who pays the fee
	// Gas types.Coin `json:"gas"`  // ?????
}

func NewFee(tx basecoin.Tx, fee types.Coin, payer basecoin.Actor) *Fee {
	return &Fee{Tx: tx, Fee: fee, Payer: payer}
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
