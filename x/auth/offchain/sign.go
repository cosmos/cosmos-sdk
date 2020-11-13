package offchain

import (
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// NewSigner is Signer's constructor
func NewSigner(txConfig client.TxConfig) Signer {
	return Signer{
		txConfig: txConfig,
	}
}

// Signer defines an offchain messages signer
type Signer struct {
	txConfig client.TxConfig
}

// Sign produces a signed tx given a private key
// and the msgs we're aiming to sign
func (s Signer) Sign(privKey cryptotypes.PrivKey, msgs []sdk.Msg) (authsigning.SigVerifiableTx, error) {
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
		PubKey:   privKey.PubKey(),
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

	signedBytes, err := privKey.Sign(bytesToSign)
	if err != nil {
		return nil, err
	}

	sigData = signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: signedBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     &sigData,
		Sequence: ExpectedSequence,
	}

	err = builder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}

	return builder.GetTx(), nil
}
