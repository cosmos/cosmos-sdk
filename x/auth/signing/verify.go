package signing

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// VerifySignature verifies a transaction signature contained in SignatureData abstracting over different signing modes
// and single vs multi-signatures.
func VerifySignature(pubKey cryptotypes.PubKey, signerData SignerData, sigData signing.SignatureData, handler SignModeHandler, tx sdk.Tx) error {
	switch data := sigData.(type) {
	case *signing.SingleSignatureData:
		signBytes, err := handler.GetSignBytes(data.SignMode, signerData, tx)
		if err != nil {
			return err
		}

		if data.SignMode == signing.SignMode_SIGN_MODE_EIP_191 {
			// do this to not have to register a new type of pubkey like here:
			// https://github.com/scrtlabs/cosmos-sdk/blob/07817ad365/crypto/keys/secp256k1/keys.pb.go#L120

			secp256k1PubKey, ok := pubKey.(*secp256k1.PubKey)
			if !ok {
				return fmt.Errorf("eip191 sign mode requires pubkey to be of type cosmos.crypto.secp256k1.PubKey")
			}

			if !secp256k1PubKey.VerifySignatureEip191(signBytes, data.Signature) {
				return fmt.Errorf("unable to verify single signer eip191 signature %s for signBytes %s", hex.EncodeToString(data.Signature), hex.EncodeToString(signBytes))
			}
		} else if !pubKey.VerifySignature(signBytes, data.Signature) {
			return fmt.Errorf("unable to verify single signer signature %s for signBytes %s", hex.EncodeToString(data.Signature), hex.EncodeToString(signBytes))
		}
		return nil

	case *signing.MultiSignatureData:
		multiPK, ok := pubKey.(multisig.PubKey)
		if !ok {
			return fmt.Errorf("expected %T, got %T", (multisig.PubKey)(nil), pubKey)
		}
		err := multiPK.VerifyMultisignature(func(mode signing.SignMode) ([]byte, error) {
			return handler.GetSignBytes(mode, signerData, tx)
		}, data)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unexpected SignatureData %T", sigData)
	}
}
