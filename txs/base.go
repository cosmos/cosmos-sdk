package txs

import (
	"github.com/tendermint/basecoin"
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

// let's register data.Bytes as a "raw" tx, for tests or
// other data we don't want to post....
func init() {
	basecoin.TxMapper.RegisterImplementation(data.Bytes{}, TypeRaw, ByteRaw)
}

func WrapBytes(d []byte) basecoin.Tx {
	return basecoin.Tx{data.Bytes(d)}
}
