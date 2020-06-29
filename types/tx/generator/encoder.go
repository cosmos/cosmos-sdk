package generator

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	tx2 "github.com/cosmos/cosmos-sdk/types/tx"
)

func DefaultTxEncoder(marshaler codec.Marshaler) types.TxEncoder {
	return func(tx types.Tx) ([]byte, error) {
		wrapper, ok := tx.(*builder)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", builder{}, tx)
		}

		raw := &tx2.TxRaw{
			BodyBytes:     wrapper.GetBodyBytes(),
			AuthInfoBytes: wrapper.GetAuthInfoBytes(),
			Signatures:    wrapper.tx.Signatures,
		}

		return marshaler.MarshalBinaryBare(raw)
	}
}

func DefaultJSONTxEncoder(marshaler codec.Marshaler) types.TxEncoder {
	return func(tx types.Tx) ([]byte, error) {
		wrapper, ok := tx.(*builder)
		if !ok {
			return nil, fmt.Errorf("expected %T, got %T", builder{}, tx)
		}

		return marshaler.MarshalJSON(wrapper.tx)
	}
}
