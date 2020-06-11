package tx

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/std"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

func TestTxWrapper(t *testing.T) {
	// TODO:
	// - verify that body and authInfo bytes encoded with DefaultTxEncoder and decoded with DefaultTxDecoder can be
	//   retrieved from GetBodyBytes and GetAuthInfoBytes
	// - create a TxWrapper using NewTxWrapper and:
	//   - verify that calling the SetBody results in the correct GetBodyBytes
	//   - verify that calling the SetAuthInfo results in the correct GetAuthInfoBytes and GetPubKeys
	//   - verify no nil panics

	tx := NewTxWrapper(codec.New(), std.DefaultPublicKeyCodec{})

	txBody := TxBody{}
	tx.SetBody(&txBody)
	memo := "memo"
	protoTx, ok := tx.(types.ProtoTx)

	if !ok {
		return nil, fmt.Errorf("can only get direct sign bytes for a ProtoTx, got %T", tx)
	}

	bodyBz := protoTx.GetBodyBytes()
	// facing import cycle error
}
