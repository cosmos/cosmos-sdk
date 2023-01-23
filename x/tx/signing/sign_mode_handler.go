package signing

import (
	"context"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
)

type SignModeHandler interface {
	// Mode is the sign mode supported by this handler
	Mode() signingv1beta1.SignMode

	// GetSignBytes returns the sign bytes for the provided SignerData and TxData, or an error.
	GetSignBytes(ctx context.Context, signerData SignerData, txData TxData) ([]byte, error)
}
