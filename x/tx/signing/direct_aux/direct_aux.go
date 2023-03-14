package direct_aux

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/tx/signing"
)

// SignModeHandler is the SIGN_MODE_DIRECT_AUX implementation of signing.SignModeHandler.
type signModeHandler struct {
	signersContext *signing.GetSignersContext
}

// NewSignModeHandler returns a new SignModeHandler.
func NewSignModeHandler(signersContext *signing.GetSignersContext) signModeHandler {
	return signModeHandler{signersContext: signersContext}
}

var _ signing.SignModeHandler = signModeHandler{}

// Mode implements signing.SignModeHandler.Mode.
func (h signModeHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX
}

func (h signModeHandler) getFirstSigner(txData signing.TxData) (string, error) {
	for _, msg := range txData.Body.Messages {
		signer, err := h.signersContext.GetSigners(msg)
		if err != nil {
			return "", err
		}
		return signer[0], nil
	}
	return "", fmt.Errorf("no signer found")
}

// GetSignBytes implements signing.SignModeHandler.GetSignBytes.
func (h signModeHandler) GetSignBytes(
	_ context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {

	// replicate https://github.com/cosmos/cosmos-sdk/blob/4a6a1e3cb8de459891cb0495052589673d14ef51/x/auth/tx/builder.go#L142
	feePayer := txData.AuthInfo.Fee.Payer
	if feePayer == "" {
		fp, err := h.getFirstSigner(txData)
		if err != nil {
			return nil, err
		}
		feePayer = fp
	}
	if feePayer == signerData.Address {
		return nil, fmt.Errorf("fee payer %s cannot sign with %s: unauthorized",
			feePayer, signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX)
	}

	signDocDirectAux := &txv1beta1.SignDocDirectAux{
		BodyBytes:     txData.BodyBytes,
		PublicKey:     signerData.PubKey,
		ChainId:       signerData.ChainId,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
		Tip:           txData.AuthInfo.Tip,
	}
	return proto.Marshal(signDocDirectAux)
}
