package offchain

import (
	"fmt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

const (
	// ChainID defines the chain id an off-chain message must have
	ChainID = ""
	// AccountNumber defines the account number an off-chain message must have
	AccountNumber = 0
	// Sequence defines the sequence number an off-chain message must have
	Sequence = 0
)

// VerifyMessage asserts that the message implementation fits offchain specification correctly
func VerifyMessage(m sdk.Msg) error {
	// ensure the sdk.msg messages are of type offchain.msg
	// generally speaking we do not want to try to handle
	// any other type of transaction aside from the offchain ones
	// as they abide by different rules
	_, valid := m.(msg)
	if !valid {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid type: %T", m)
	}
	if m.Route() != Route {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "offchain messages route should be set to: %s", Route)
	}
	return nil
}

// NewVerifier is SignatureVerifier's constructor
func NewVerifier(signModeHandler authsigning.SignModeHandler) SignatureVerifier {
	return SignatureVerifier{signModeHandler: signModeHandler}
}

// SignatureVerifier takes care of verifying transactions given
// an instance of authsigning.SignModeHandler
type SignatureVerifier struct {
	signModeHandler authsigning.SignModeHandler
}

// Verify takes an sdk.Tx and verifies it
func (v SignatureVerifier) Verify(tx sdk.Tx) error {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "cannot verify tx of type %T", tx)
	}

	msgs := sigTx.GetMsgs()
	if len(msgs) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "no message provided")
	}
	for i, msg := range msgs {
		err := VerifyMessage(msg)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "message number %d is invalid: %s", i, err)
		}
	}

	signers := sigTx.GetPubKeys()
	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return fmt.Errorf("cannot verify: %w", err)
	}
	if len(signatures) != len(signers) {
		return fmt.Errorf("signatures and signers mismatch: %d <-> %d", len(signers), len(signatures))
	}
	for i, signature := range signatures {
		err := verifySignature(tx, signature, signers[i], v.signModeHandler)
		if err != nil {
			return fmt.Errorf("invalid signature %d: %w", i, err)
		}
	}
	return nil
}

// verifySignature verifies a single signature
func verifySignature(tx sdk.Tx, sig signing.SignatureV2, signer cryptotypes.PubKey, handler authsigning.SignModeHandler) error {
	// TODO: we're imposing chainID accountNumber and sequence, is there a way to verify those params beforehand?
	// TODO: so we can return a bad request error instead of an unauthorized sig one
	signerData := authsigning.SignerData{
		ChainID:       ChainID,
		AccountNumber: AccountNumber,
		Sequence:      Sequence,
	}
	err := authsigning.VerifySignature(signer, signerData, sig.Data, handler, tx)
	if err != nil {
		return err
	}
	return nil
}
