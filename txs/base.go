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
		RegisterImplementation(data.Bytes{}, TypeRaw, ByteRaw).
		RegisterImplementation(&Fee{}, TypeFees, ByteFees)
}

// WrapBytes converts data.Bytes into a Tx, so we
// can just pass raw bytes and display in hex in json
func WrapBytes(d []byte) basecoin.Tx {
	return basecoin.Tx{data.Bytes(d)}
}

/**** One Sig ****/

// OneSig lets us wrap arbitrary data with a go-crypto signature
type Fee struct {
	Tx  basecoin.Tx `json:"tx"`
	Fee types.Coin  `json:"fee"`
	// Gas types.Coin `json:"gas"`  // ?????
}

func NewFee(tx basecoin.Tx, fee types.Coin) *Fee {
	return &Fee{Tx: tx, Fee: fee}
}

func (f *Fee) Wrap() basecoin.Tx {
	return basecoin.Tx{f}
}
