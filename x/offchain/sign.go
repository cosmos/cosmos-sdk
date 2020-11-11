package offchain

import (
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

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
func (s Signer) Sign(privKey cryptotypes.PrivKey, msgs []msg) (authsigning.SigVerifiableTx, error) {
	// build unsigned tx
	builder := s.txConfig.NewTxBuilder()

	sdkMsgs := make([]sdk.Msg, len(msgs))
	for i, msg := range msgs {
		sdkMsgs[i] = msg
	}
	err := builder.SetMsgs(sdkMsgs...)
	if err != nil {
		return nil, err
	}

	// prepare transaction to sign
	signMode := s.txConfig.SignModeHandler().DefaultMode()

	signerData := authsigning.SignerData{
		ChainID:       ChainID,
		AccountNumber: AccountNumber,
		Sequence:      Sequence,
	}

	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}

	sig := signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     &sigData,
		Sequence: Sequence,
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
		Sequence: Sequence,
	}

	err = builder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}

	return builder.GetTx(), nil
}
