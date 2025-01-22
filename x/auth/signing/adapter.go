package signing

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// V2AdaptableTx is an interface that wraps the GetSigningTxData method.
// GetSigningTxData returns an x/tx/signing.TxData representation of a transaction for use in signing
// interoperability with x/tx.
type V2AdaptableTx interface {
	GetSigningTxData() txsigning.TxData
}

// GetSignBytesAdapter returns the sign bytes for a given transaction and sign mode.  It accepts the arguments expected
// for signing in x/auth/tx and converts them to the arguments expected by the txsigning.HandlerMap, then applies
// HandlerMap.GetSignBytes to get the sign bytes.
func GetSignBytesAdapter(
	ctx context.Context,
	handlerMap *txsigning.HandlerMap,
	mode apisigning.SignMode,
	signerData txsigning.SignerData,
	tx sdk.Tx,
) ([]byte, error) {
	adaptableTx, ok := tx.(V2AdaptableTx)
	if !ok {
		return nil, fmt.Errorf("expected tx to be V2AdaptableTx, got %T", tx)
	}
	txData := adaptableTx.GetSigningTxData()

	var pubKey *anypb.Any
	if signerData.PubKey != nil {
		pubKey = &anypb.Any{
			TypeUrl: signerData.PubKey.TypeUrl,
			Value:   signerData.PubKey.Value,
		}
	}
	txSignerData := txsigning.SignerData{
		ChainID:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
		Address:       signerData.Address,
		PubKey:        pubKey,
	}
	// Generate the bytes to be signed.
	return handlerMap.GetSignBytes(ctx, mode, txSignerData, txData)
}
