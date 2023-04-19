package signing

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/tx/decode"
	txsigning "cosmossdk.io/x/tx/signing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/registry"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// APISignModesToInternal converts a protobuf SignMode array to a signing.SignMode array.
func APISignModesToInternal(modes []signingv1beta1.SignMode) ([]signing.SignMode, error) {
	internalModes := make([]signing.SignMode, len(modes))
	for i, mode := range modes {
		internalMode, err := APISignModeToInternal(mode)
		if err != nil {
			return nil, err
		}
		internalModes[i] = internalMode
	}
	return internalModes, nil
}

// APISignModeToInternal converts a protobuf SignMode to a signing.SignMode.
func APISignModeToInternal(mode signingv1beta1.SignMode) (signing.SignMode, error) {
	switch mode {
	case signingv1beta1.SignMode_SIGN_MODE_DIRECT:
		return signing.SignMode_SIGN_MODE_DIRECT, nil
	case signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
		return signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, nil
	case signingv1beta1.SignMode_SIGN_MODE_TEXTUAL:
		return signing.SignMode_SIGN_MODE_TEXTUAL, nil
	case signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX:
		return signing.SignMode_SIGN_MODE_DIRECT_AUX, nil
	default:
		return signing.SignMode_SIGN_MODE_UNSPECIFIED, fmt.Errorf("unsupported sign mode %s", mode)
	}
}

// internalSignModeToAPI converts a signing.SignMode to a protobuf SignMode.
func internalSignModeToAPI(mode signing.SignMode) (signingv1beta1.SignMode, error) {
	switch mode {
	case signing.SignMode_SIGN_MODE_DIRECT:
		return signingv1beta1.SignMode_SIGN_MODE_DIRECT, nil
	case signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON:
		return signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, nil
	case signing.SignMode_SIGN_MODE_TEXTUAL:
		return signingv1beta1.SignMode_SIGN_MODE_TEXTUAL, nil
	case signing.SignMode_SIGN_MODE_DIRECT_AUX:
		return signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX, nil
	default:
		return signingv1beta1.SignMode_SIGN_MODE_UNSPECIFIED, fmt.Errorf("unsupported sign mode %s", mode)
	}
}

// VerifySignature verifies a transaction signature contained in SignatureData abstracting over different signing
// modes. It differs from VerifySignature in that it uses the new txsigning.TxData interface in x/tx.
func VerifySignature(
	ctx context.Context,
	pubKey cryptotypes.PubKey,
	signerData txsigning.SignerData,
	signatureData signing.SignatureData,
	handler *txsigning.HandlerMap,
	txData txsigning.TxData,
) error {
	switch data := signatureData.(type) {
	case *signing.SingleSignatureData:
		signMode, err := internalSignModeToAPI(data.SignMode)
		if err != nil {
			return err
		}
		signBytes, err := handler.GetSignBytes(ctx, signMode, signerData, txData)
		if err != nil {
			return err
		}
		if !pubKey.VerifySignature(signBytes, data.Signature) {
			return fmt.Errorf("unable to verify single signer signature")
		}
		return nil

	case *signing.MultiSignatureData:
		multiPK, ok := pubKey.(multisig.PubKey)
		if !ok {
			return fmt.Errorf("expected %T, got %T", (multisig.PubKey)(nil), pubKey)
		}
		err := multiPK.VerifyMultisignature(func(mode signing.SignMode) ([]byte, error) {
			signMode, err := internalSignModeToAPI(mode)
			if err != nil {
				return nil, err
			}
			return handler.GetSignBytes(ctx, signMode, signerData, txData)
		}, data)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unexpected SignatureData %T", signatureData)
	}
}

// GetSignBytesAdapter returns the sign bytes for a given transaction and sign mode.  It accepts the arguments expected
// for signing in x/auth/tx and converts them to the arguments expected by the txsigning.HandlerMap, then applies
// HandlerMap.GetSignBytes to get the sign bytes.
func GetSignBytesAdapter(
	ctx context.Context,
	encoder sdk.TxEncoder,
	handlerMap *txsigning.HandlerMap,
	mode signing.SignMode,
	signerData SignerData,
	tx sdk.Tx,
) ([]byte, error) {
	// round trip performance hit.
	// could be avoided if we had a way to get the bytes from the txBuilder.
	txBytes, err := encoder(tx)
	if err != nil {
		return nil, err
	}
	decodeCtx, err := decode.NewDecoder(decode.Options{ProtoFiles: registry.MergedProtoRegistry()})
	if err != nil {
		return nil, err
	}
	decodedTx, err := decodeCtx.Decode(txBytes)
	if err != nil {
		return nil, err
	}
	txData := txsigning.TxData{
		Body:          decodedTx.Tx.Body,
		AuthInfo:      decodedTx.Tx.AuthInfo,
		AuthInfoBytes: decodedTx.TxRaw.AuthInfoBytes,
		BodyBytes:     decodedTx.TxRaw.BodyBytes,
	}
	txSignMode, err := internalSignModeToAPI(mode)
	if err != nil {
		return nil, err
	}

	anyPk, err := codectypes.NewAnyWithValue(signerData.PubKey)
	if err != nil {
		return nil, err
	}

	txSignerData := txsigning.SignerData{
		ChainID:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
		Address:       signerData.Address,
		PubKey: &anypb.Any{
			TypeUrl: anyPk.TypeUrl,
			Value:   anyPk.Value,
		},
	}
	// Generate the bytes to be signed.
	return handlerMap.GetSignBytes(ctx, txSignMode, txSignerData, txData)
}
