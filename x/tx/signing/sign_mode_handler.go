package signing

import (
	"context"
)

// SignModeHandler is the interface that handlers for each sign mode should implement to generate sign bytes.
type SignModeHandler interface {
	// Mode is the sign mode supported by this handler
	Mode() SignMode

	// GetSignBytes returns the sign bytes for the provided SignerData and TxData, or an error.
	GetSignBytes(ctx context.Context, signerData SignerData, txData TxData) ([]byte, error)
}
