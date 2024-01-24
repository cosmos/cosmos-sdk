package tx

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultJSONTxDecoder returns a default protobuf JSON TxDecoder using the provided Marshaler.
func DefaultJSONTxDecoder(cdc codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		panic("impl")
	}
}
