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
type SignModeHandler struct{}

var _ signing.SignModeHandler = SignModeHandler{}

// Mode implements signing.SignModeHandler.Mode.
func (h SignModeHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX
}

// GetSignBytes implements signing.SignModeHandler.GetSignBytes.
func (SignModeHandler) GetSignBytes(
	_ context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {

	// replicate https://github.com/cosmos/cosmos-sdk/blob/4a6a1e3cb8de459891cb0495052589673d14ef51/x/auth/tx/builder.go#L142
	feePayer := txData.AuthInfo.Fee.Payer
	if feePayer == "" {
		signers, err := signing.NewGetSignersContext(signing.GetSignersOptions{}).GetSigners(txData.Body)
		if err != nil {
			return nil, err
		}
		feePayer = signers[0]
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
