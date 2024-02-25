package tx

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultTxEncoder returns a default protobuf TxEncoder using the provided Marshaler
func DefaultTxEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		gogoWrapper, ok := tx.(*gogoTxWrapper)
		if !ok {
			return nil, fmt.Errorf("unexpected tx type: %T", tx)
		}
		return marshalOption.Marshal(gogoWrapper.decodedTx.TxRaw)
	}
}

// DefaultJSONTxEncoder returns a default protobuf JSON TxEncoder using the provided Marshaler.
func DefaultJSONTxEncoder(cdc codec.Codec) sdk.TxEncoder {
	jsonMarshaler := protojson.MarshalOptions{
		Indent:         "",
		UseProtoNames:  true,
		UseEnumNumbers: false,
	}
	return func(tx sdk.Tx) ([]byte, error) {
		gogoWrapper, ok := tx.(*gogoTxWrapper)
		if !ok {
			return nil, fmt.Errorf("unexpected tx type: %T", tx)
		}
		return jsonMarshaler.Marshal(gogoWrapper.decodedTx.Tx)
	}
}
