package adr036

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// SignatureProvider offers an abstraction over private keys
// which can be constructed from raw private keys or keyrings
// it is a subset of the methods exposed by cryptotypes.PrivKey
type SignatureProvider interface {
	Sign(msg []byte, signMode signing.SignMode) (signedBytes []byte, err error)
	PubKey() cryptotypes.PubKey
}

// keyringWrapper implements SignatureProvider over a keyring
type keyringWrapper struct {
	uid     string
	pubKey  cryptotypes.PubKey
	keyring keyring.Keyring
}

func (k keyringWrapper) Sign(msg []byte, signMode signing.SignMode) (signedBytes []byte, err error) {
	signedBytes, _, err = k.keyring.Sign(k.uid, msg, signMode)
	return
}

func (k keyringWrapper) PubKey() cryptotypes.PubKey {
	return k.pubKey
}

func newKeyRingWrapper(uid string, keyring keyring.Keyring) (SignatureProvider, error) {
	// assert from name exists
	info, err := keyring.Key(uid)
	if err != nil {
		return nil, err
	}
	pubKey, err := info.GetPubKey()
	if err != nil {
		return nil, err
	}
	return &keyringWrapper{uid: uid, pubKey: pubKey, keyring: keyring}, nil
}

// OffChainSigner abstraction over
type OffChainSigner struct {
	signer   SignatureProvider
	txConfig client.TxConfig
}

// NewOffChainSignerFromClientContext builds an offchain message signer from a client context
func NewOffChainSignerFromClientContext(clientCtx client.Context) (*OffChainSigner, error) {
	if clientCtx.TxConfig == nil {
		return nil, errors.New("txconfig must be set")
	}
	if clientCtx.Keyring == nil {
		return nil, errors.New("keybase must be set")
	}
	privKey, err := newKeyRingWrapper(clientCtx.GetFromName(), clientCtx.Keyring)
	if err != nil {
		return nil, err
	}
	return NewOffChainSigner(privKey, clientCtx.TxConfig), nil
}

func NewOffChainSigner(priv SignatureProvider, txConfig client.TxConfig) *OffChainSigner {
	return &OffChainSigner{
		signer:   priv,
		txConfig: txConfig,
	}
}

// Sign produces a signed tx given
func (s OffChainSigner) Sign(ctx context.Context, msgs []sdk.Msg, signMode signing.SignMode) (authsigning.SigVerifiableTx, error) {
	err := validateMsgs(msgs)
	if err != nil {
		return nil, err
	}

	return s.singMsg(ctx, msgs, signMode)
}

func (s OffChainSigner) singMsg(ctx context.Context, msgs []sdk.Msg, signmode signing.SignMode) (authsigning.SigVerifiableTx, error) {
	// Build unsigned tx
	tx := s.txConfig.NewTxBuilder()
	if err := tx.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	pubKey := s.signer.PubKey()
	signerData := authsigning.SignerData{
		Address:       sdk.AccAddress(pubKey.Address()).String(),
		ChainID:       ExpectedChainID,
		AccountNumber: ExpectedAccountNumber,
		Sequence:      ExpectedSequence,
		PubKey:        pubKey,
	}

	sigData := signing.SingleSignatureData{
		SignMode:  signmode,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: ExpectedSequence,
	}

	sigs := []signing.SignatureV2{sig}

	err := tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(
		ctx, s.txConfig.SignModeHandler(),
		signmode, signerData, tx.GetTx())
	if err != nil {
		return nil, err
	}

	signedBytes, err := s.signer.Sign(bytesToSign, signmode)
	if err != nil {
		return nil, err
	}

	sigData = signing.SingleSignatureData{
		SignMode:  signmode,
		Signature: signedBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: ExpectedSequence,
	}
	err = tx.SetSignatures(sig)
	if err != nil {
		return nil, err
	}

	signingTx := tx.GetTx()

	return signingTx, nil
}

func validateMsgs(msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return errors.New("no msg provided")
	}
	for _, msg := range msgs {
		if sdk.MsgTypeURL(msg) != sdk.MsgTypeURL(&MsgSignArbitraryData{}) {
			return errors.New("not arbitrary data message")
		}
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}
		if err := m.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}
