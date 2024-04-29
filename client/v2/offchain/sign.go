package offchain

import (
	"context"
	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	tx "cosmossdk.io/client/v2/internal"
	"cosmossdk.io/client/v2/internal/offchain"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

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

	signMode = apisigning.SignMode_SIGN_MODE_TEXTUAL
)

// TxData is the data about a transaction that is necessary to generate sign bytes.
type TxData struct {
	// Body is the TxBody that will be part of the transaction.
	Body *apitx.TxBody

	// AuthInfo is the AuthInfo that will be part of the transaction.
	AuthInfo *apitx.AuthInfo

	// BodyBytes is the marshaled body bytes that will be part of TxRaw.
	BodyBytes []byte

	// AuthInfoBytes is the marshaled AuthInfo bytes that will be part of TxRaw.
	AuthInfoBytes []byte

	// BodyHasUnknownNonCriticals should be set to true if the transaction has been
	// decoded and found to have unknown non-critical fields. This is only needed
	// for amino JSON signing.
	BodyHasUnknownNonCriticals bool
}

type SignerData struct {
	Address       string
	ChainID       string
	AccountNumber uint64
	Sequence      uint64
	PubKey        cryptotypes.PubKey
}

// Sign signs given bytes using the specified encoder and SignMode.
func Sign(ctx tx.Context, rawBytes []byte, fromName, indent, encoding, output string, emitUnpopulated bool) (string, error) {
	encoder, err := getEncoder(encoding)
	if err != nil {
		return "", err
	}

	digest, err := encoder(rawBytes)
	if err != nil {
		return "", err
	}

	tx, err := sign(ctx, fromName, digest)
	if err != nil {
		return "", err
	}

	txMarshaller, err := getMarshaller(output, indent, emitUnpopulated)
	if err != nil {
		return "", err
	}

	return marshalOffChainTx(tx, txMarshaller)
}

// sign signs a digest with provided key and SignMode.
func sign(ctx tx.Context, fromName, digest string) (*apitx.Tx, error) {
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

	signerData := SignerData{
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
		context.Background(), ctx.TxConfig.SignModeHandler(), signerData, txBuilder)
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

	return txBuilder.GetTx(), nil
}

// getSignBytes gets the bytes to be signed for the given Tx and SignMode.
func getSignBytes(ctx context.Context,
	handlerMap *HandlerMap,
	signerData SignerData,
	tx *builder,
) ([]byte, error) {
	txData, err := tx.GetSigningTxData()
	if err != nil {
		return nil, err
	}

	txSignerData := SignerData{
		ChainID:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
		Address:       signerData.Address,
		PubKey:        signerData.PubKey,
	}

	return handlerMap.GetSignBytes(ctx, signMode, txSignerData, txData)
}
