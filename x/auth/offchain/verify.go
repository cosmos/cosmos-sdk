package offchain

import (
	"errors"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// errors are private and used only for testing purposes
var (
	errInvalidType  = errors.New("invalid type")
	errInvalidRoute = errors.New("invalid route")
)

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
		err := verifyMessage(msg)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "message number %d is invalid: %s", i, err)
		}
	}

	signers, err := sigTx.GetPubKeys()
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, err.Error())
	}

	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return fmt.Errorf("cannot verify: %w", err)
	}
	if len(signatures) != len(signers) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "signatures and signers mismatch: %d <-> %d", len(signers), len(signatures))
	}
	for i, signature := range signatures {
		err := verifySignature(tx, signature, signers[i], v.signModeHandler)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid signature %d: %s", i, err)
		}
	}
	return nil
}

// verifySignature verifies a single signature
func verifySignature(tx sdk.Tx, sig signing.SignatureV2, signer cryptotypes.PubKey, handler authsigning.SignModeHandler) error {
	// TODO: we're imposing chainID accountNumber and sequence, is there a way to verify those params beforehand?
	// TODO: so we can return a bad request error instead of an unauthorized sig one
	signerData := authsigning.SignerData{
		ChainID:       ExpectedChainID,
		AccountNumber: ExpectedAccountNumber,
		Sequence:      ExpectedSequence,
	}
	err := authsigning.VerifySignature(signer, signerData, sig.Data, handler, tx)
	if err != nil {
		return err
	}
	return nil
}

// verifyMessage asserts that the message implementation fits offchain specification correctly
func verifyMessage(m sdk.Msg) error {
	// ensure the sdk.msg messages are of type offchain.msg
	// generally speaking we do not want to try to handle
	// any other type of transaction aside from the offchain ones
	// as they abide by different rules
	_, valid := m.(msg)
	if !valid {
		return fmt.Errorf("%w: %T", errInvalidType, m)
	}
	if m.Route() != ExpectedRoute {
		return fmt.Errorf("%w: %s", errInvalidRoute, m.Route())
	}
	return nil
}
