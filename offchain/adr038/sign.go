package adr038

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var (
	errNoTxConfig        = errors.New("no tx config provided")
	errNoKeyringProvided = errors.New("no keyring provided")
)

// SignatureProvider offers an abstraction over private keys
// which can be constructed from raw private keys or keyrings
// it is a subset of the methods exposed by cryptotypes.PrivKey
type SignatureProvider interface {
	Sign(msg []byte) (signedBytes []byte, err error)
	PubKey() cryptotypes.PubKey
}

// keyringWrapper implements SignatureProvider over a keyring
type keyringWrapper struct {
	uid     string
	pubKey  cryptotypes.PubKey
	keyring keyring.Keyring
}

func (k keyringWrapper) Sign(msg []byte) (signedBytes []byte, err error) {
	signedBytes, _, err = k.keyring.Sign(k.uid, msg)
	return
}

func (k keyringWrapper) PubKey() cryptotypes.PubKey {
	return k.pubKey
}

func newKeyRingWrapper(uid string, keyring keyring.Keyring) (keyringWrapper, error) {
	// assert from name exists
	info, err := keyring.Key(uid)
	if err != nil {
		return keyringWrapper{}, err
	}
	pubKey, err := info.GetPubKey()
	if err != nil {
		return keyringWrapper{}, err
	}
	wrapper := keyringWrapper{uid: uid, pubKey: pubKey}
	return wrapper, nil
}

// NewSignerFromClientContext builds an offchain message signer from a client context
func NewSignerFromClientContext(clientCtx client.Context) (Signer, error) {
	if clientCtx.TxConfig == nil {
		return Signer{}, errNoTxConfig
	}
	if clientCtx.Keyring == nil {
		return Signer{}, errNoKeyringProvided
	}
	privKey, err := newKeyRingWrapper(clientCtx.GetFromName(), clientCtx.Keyring)
	if err != nil {
		return Signer{}, err
	}
	return NewSigner(clientCtx.TxConfig, privKey), nil
}

// NewSigner is Signer's constructor
func NewSigner(txConfig client.TxConfig, provider SignatureProvider) Signer {
	return Signer{
		txConfig: txConfig,
		privKey:  provider,
	}
}

// Signer defines an offchain messages signer
type Signer struct {
	txConfig client.TxConfig
	privKey  SignatureProvider
}

// Sign produces a signed tx given a private key
// and the msgs we're aiming to sign
func (s Signer) Sign(msgs []sdk.Msg) (authsigning.SigVerifiableTx, error) {
	if len(msgs) == 0 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "no msg provided")
	}
	// build unsigned tx
	builder := s.txConfig.NewTxBuilder()

	for i, msg := range msgs {
		err := verifyMessage(msg)
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "message number %d is invalid: %s", i, err)
		}
	}
	err := builder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	// prepare transaction to sign
	signMode := s.txConfig.SignModeHandler().DefaultMode()

	signerData := authsigning.SignerData{
		ChainID:       ExpectedChainID,
		AccountNumber: ExpectedAccountNumber,
		Sequence:      ExpectedSequence,
	}

	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}

	sig := signing.SignatureV2{
		PubKey:   s.privKey.PubKey(),
		Data:     &sigData,
		Sequence: ExpectedSequence,
	}
	err = builder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}

	bytesToSign, err := s.txConfig.SignModeHandler().
		GetSignBytes(
			signMode,
			signerData,
			builder.GetTx(),
		)
	if err != nil {
		return nil, err
	}

	signedBytes, err := s.privKey.Sign(bytesToSign)
	if err != nil {
		return nil, err
	}

	sigData = signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: signedBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   s.privKey.PubKey(),
		Data:     &sigData,
		Sequence: ExpectedSequence,
	}

	err = builder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}

	return builder.GetTx(), nil
}
