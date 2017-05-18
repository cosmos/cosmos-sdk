package txs

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire/data"
)

const (
	// for utils...
	ByteRaw   = 0x1
	ByteFees  = 0x2
	ByteMulti = 0x3

	// for signatures
	ByteSig      = 0x16
	ByteMultiSig = 0x17
)

const (
	// for utils...
	TypeRaw   = "raw"
	TypeFees  = "fee"
	TypeMulti = "multi"

	// for signatures
	TypeSig      = "sig"
	TypeMultiSig = "multisig"
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(Raw{}, TypeRaw, ByteRaw).
		RegisterImplementation(&Fee{}, TypeFees, ByteFees).
		RegisterImplementation(&MultiTx{}, TypeMulti, ByteMulti)
}

// Raw just contains bytes that can be hex-ified
type Raw struct {
	data.Bytes
}

func (r Raw) Wrap() basecoin.Tx {
	return basecoin.Tx{r}
}

func NewRaw(d []byte) Raw {
	return Raw{data.Bytes(d)}
}

/**** Fee ****/

// Fee attaches a fee payment to the embedded tx
type Fee struct {
	Tx    basecoin.Tx `json:"tx"`
	Fee   types.Coin  `json:"fee"`
	Payer data.Bytes  `json:"payer"` // the address who pays the fee
	// Gas types.Coin `json:"gas"`  // ?????
}

func NewFee(tx basecoin.Tx, fee types.Coin, addr []byte) *Fee {
	return &Fee{Tx: tx, Fee: fee, Payer: addr}
}

func (f *Fee) Wrap() basecoin.Tx {
	return basecoin.Tx{f}
}

/**** MultiTx  ******/
type MultiTx struct {
	Txs []basecoin.Tx `json:"txs"`
}

func NewMultiTx(txs ...basecoin.Tx) *MultiTx {
	return &MultiTx{Txs: txs}
}

func (mt *MultiTx) Wrap() basecoin.Tx {
	return basecoin.Tx{mt}
}
