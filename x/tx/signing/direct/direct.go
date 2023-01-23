package direct

import (
	"context"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/x/tx/signing"
)

type signModeDirectHandler struct{}

func (h signModeDirectHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (signModeDirectHandler) GetSignBytes(_ context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {
	return proto.Marshal(&txv1beta1.SignDoc{
		BodyBytes:     txData.TxRaw.BodyBytes,
		AuthInfoBytes: txData.TxRaw.AuthInfoBytes,
		ChainId:       signerData.ChainId,
		AccountNumber: signerData.AccountNumber,
	})
}

var _ signing.SignModeHandler = signModeDirectHandler{}
