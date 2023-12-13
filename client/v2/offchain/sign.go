package offchain

import (
	"context"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/client/v2/internal/offchain"
	authsigning "cosmossdk.io/x/auth/signing"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
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

func getSignMode(signModeStr string) apisigning.SignMode {
	signMode := apisigning.SignMode_SIGN_MODE_UNSPECIFIED
	switch signModeStr {
	case flags.SignModeDirect:
		signMode = apisigning.SignMode_SIGN_MODE_DIRECT
	case flags.SignModeLegacyAminoJSON:
		signMode = apisigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	case flags.SignModeDirectAux:
		signMode = apisigning.SignMode_SIGN_MODE_DIRECT_AUX
	case flags.SignModeTextual:
		signMode = apisigning.SignMode_SIGN_MODE_TEXTUAL
	}
	return signMode
}

func Sign(ctx client.Context, rawBytes []byte, fromName, signMode, indent, encoding string, emitUnpopulated bool) (string, error) {
	digest, err := getEncoder(encoding)(rawBytes)
	if err != nil {
		return "", err
	}

	tx, err := sign(ctx, fromName, digest, getSignMode(signMode))
	if err != nil {
		return "", err
	}

	return marshalOffChainTx(tx, emitUnpopulated, indent)
}

func sign(ctx client.Context, fromName, digest string, signMode apisigning.SignMode) (*apitx.Tx, error) {
	keybase, err := keyring.NewAutoCLIKeyring(ctx.Keyring)
	if err != nil {
		return nil, err
	}

	pubKey, err := keybase.GetPubKey(fromName)
	if err != nil {
		return nil, err
	}

	addr, err := ctx.AddressCodec.BytesToString(pubKey.Address())
	if err != nil {
		return nil, err
	}

	msg := &offchain.MsgSignArbitraryData{
		AppDomain: version.AppName,
		Signer:    addr,
		Data:      digest,
	}

	txBuilder := newBuilder(ctx.Codec)
	err = txBuilder.setMsgs(msg)
	if err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		Address:       addr,
		ChainID:       ExpectedChainID,
		AccountNumber: ExpectedAccountNumber,
		Sequence:      ExpectedSequence,
		PubKey:        pubKey,
	}

	sigData := &SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}

	sig := OffchainSignature{
		PubKey:   pubKey,
		Data:     sigData,
		Sequence: ExpectedSequence,
	}

	sigs := []OffchainSignature{sig}
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}

	bytesToSign, err := getSignBytes(
		context.Background(), ctx.TxConfig.SignModeHandler(),
		signMode, signerData, txBuilder)
	if err != nil {
		return nil, err
	}

	signedBytes, err := keybase.Sign(fromName, bytesToSign, signMode)
	if err != nil {
		return nil, err
	}

	sigData.Signature = signedBytes

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}

	return txBuilder.GetProtoTx(), nil
}

func getSignBytes(ctx context.Context,
	handlerMap *txsigning.HandlerMap,
	mode apisigning.SignMode,
	signerData authsigning.SignerData,
	tx authsigning.V2AdaptableTx,
) ([]byte, error) {
	txData := tx.GetSigningTxData()

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

	return handlerMap.GetSignBytes(ctx, mode, txSignerData, txData)
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
