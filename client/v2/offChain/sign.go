package offChain

import (
	"context"
	"google.golang.org/protobuf/encoding/protojson"

	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/client/v2/internal/offChain"
	authsigning "cosmossdk.io/x/auth/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
)

const (
	// ExpectedChainID defines the chain id an off-chain message must have
	ExpectedChainID = ""
	// ExpectedAccountNumber defines the account number an off-chain message must have
	ExpectedAccountNumber = 0
	// ExpectedSequence defines the sequence number an off-chain message must have
	ExpectedSequence = 0
)

type encodingFunc = func([]byte) (string, error)

func noEncoding(digest []byte) (string, error) {
	return string(digest), nil
}

func getSignMode(signModeStr string) signing.SignMode {
	signMode := signing.SignMode_SIGN_MODE_UNSPECIFIED
	switch signModeStr {
	case flags.SignModeDirect:
		signMode = signing.SignMode_SIGN_MODE_DIRECT
	case flags.SignModeLegacyAminoJSON:
		signMode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	case flags.SignModeDirectAux:
		signMode = signing.SignMode_SIGN_MODE_DIRECT_AUX
	case flags.SignModeTextual:
		signMode = signing.SignMode_SIGN_MODE_TEXTUAL
	}
	return signMode
}

func Sign(ctx client.Context, rawBytes []byte, fromName, signMode string) (string, error) {
	digest, err := noEncoding(rawBytes)
	if err != nil {
		return "", err
	}

	tx, err := sign(ctx, fromName, digest, getSignMode(signMode))
	if err != nil {
		return "", err
	}

	return marshalOffChainTx(tx, true, "  ")
}

func sign(ctx client.Context, fromName, digest string, signMode signing.SignMode) (*apitx.Tx, error) {
	keybase := ctx.Keyring
	r, err := keybase.Key(fromName)
	if err != nil {
		return nil, err
	}

	pubKey, err := r.GetPubKey()
	if err != nil {
		return nil, err
	}

	msg := &offChain.MsgSignArbitraryData{
		AppDomain:     version.AppName,
		SignerAddress: types.AccAddress(pubKey.Address()).String(),
		Data:          digest,
	}

	txBuilder := newBuilder(ctx.Codec)
	err = txBuilder.setMsgs(msg)
	if err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		Address:       types.AccAddress(pubKey.Address()).String(),
		ChainID:       ExpectedChainID,
		AccountNumber: ExpectedAccountNumber,
		Sequence:      ExpectedSequence,
		PubKey:        pubKey,
	}
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: ExpectedSequence,
	}

	sigs := []signing.SignatureV2{sig}
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(
		context.Background(), ctx.TxConfig.SignModeHandler(),
		signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	signedBytes, _, err := keybase.Sign(fromName, bytesToSign, signMode)
	if err != nil {
		return nil, err
	}

	sigData = signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: signedBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: ExpectedSequence,
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}

	return txBuilder.GetProtoTx(), nil
}

func marshalOffChainTx(tx *apitx.Tx, emitUnpopulated bool, indent string) (string, error) {
	bytesTx, err := protojson.MarshalOptions{
		EmitUnpopulated: emitUnpopulated,
		Indent:          indent,
	}.Marshal(tx)
	if err != nil {
		return "", err
	}
	return string(bytesTx), nil
}
