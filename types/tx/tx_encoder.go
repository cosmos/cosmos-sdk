package tx

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
)

func DefaultTxEncoder(marshaler codec.Marshaler) types.TxEncoder {
	return func(tx types.Tx) ([]byte, error) {
		wrapper, ok := tx.(builder)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", builder{}, tx)
		}

		raw := &TxRaw{
			BodyBytes:     wrapper.bodyBz,
			AuthInfoBytes: wrapper.authInfoBz,
			Signatures:    wrapper.tx.Signatures,
		}

		return marshaler.MarshalBinaryBare(raw)
	}
}
