package tx

import (
	"fmt"
	gogoproto "github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/protobuf/proto"
)

// DefaultTxEncoder returns a default protobuf TxEncoder using the provided Marshaler
func DefaultTxEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		gogoWrapper, ok := tx.(*gogoTxWrapper)
		if !ok {
			return nil, fmt.Errorf("unexpected tx type: %T", tx)
		}
		return marshalOption.Marshal(gogoWrapper.TxRaw)
	}
}

// DefaultJSONTxEncoder returns a default protobuf JSON TxEncoder using the provided Marshaler.
func DefaultJSONTxEncoder(cdc codec.Codec) sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		gogoWrapper, ok := tx.(*gogoTxWrapper)
		if !ok {
			return nil, fmt.Errorf("unexpected tx type: %T", tx)
		}
		bz, err := proto.Marshal(gogoWrapper.Tx)
		if err != nil {
			return nil, err
		}
		v1Tx := &sdktx.Tx{}
		err = gogoproto.Unmarshal(bz, v1Tx)
		if err != nil {
			return nil, err
		}
		return cdc.MarshalJSON(v1Tx)
	}
}
