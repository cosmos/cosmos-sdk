package aminojson

import (
	"context"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/tx/signing"
)

type SignModeHandler struct{}

func (s SignModeHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

func (s SignModeHandler) GetSignBytes(ctx context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

var _ signing.SignModeHandler = (*SignModeHandler)(nil)
