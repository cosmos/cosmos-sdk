package signing

import (
	"github.com/cosmos/cosmos-sdk/crypto/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

func VerifySignature(signingData types.SigningData, sigData types.SignatureData, tx sdk.Tx, handler types.SignModeHandler) bool {
	switch data := sigData.(type) {
	case *types.SingleSignatureData:
		signBytes, err := handler.GetSignBytes(data.SignMode, signingData, tx)
		if err != nil {
			return false
		}
		return signingData.PublicKey.VerifyBytes(signBytes, data.Signature)
	case *types.MultiSignatureData:
		multiPK, ok := signingData.PublicKey.(multisig.MultisigPubKey)
		if !ok {
			return false
		}
		return multiPK.VerifyMultisignature(func(mode types.SignMode) ([]byte, error) {
			return handler.GetSignBytes(mode, signingData, tx)
		}, data)
	}
	return false
}
