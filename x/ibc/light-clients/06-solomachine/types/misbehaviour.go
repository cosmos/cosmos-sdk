package types

import (
	"bytes"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var (
	_ exported.Misbehaviour = (*Misbehaviour)(nil)
)

// ClientType is a Solo Machine light client.
func (misbehaviour Misbehaviour) ClientType() string {
	return exported.Solomachine
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (misbehaviour Misbehaviour) GetClientID() string {
	return misbehaviour.ClientId
}

// Type implements Evidence interface.
func (misbehaviour Misbehaviour) Type() string {
	return exported.TypeClientMisbehaviour
}

// GetHeight returns the sequence at which misbehaviour occurred.
// Return exported.Height to satisfy interface
// Revision number is always 0 for a solo-machine
func (misbehaviour Misbehaviour) GetHeight() exported.Height {
	return clienttypes.NewHeight(0, misbehaviour.Sequence)
}

// ValidateBasic implements Evidence interface.
func (misbehaviour Misbehaviour) ValidateBasic() error {
	if err := host.ClientIdentifierValidator(misbehaviour.ClientId); err != nil {
		return sdkerrors.Wrap(err, "invalid client identifier for solo machine")
	}

	if misbehaviour.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidMisbehaviour, "sequence cannot be 0")
	}

	if err := misbehaviour.SignatureOne.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "signature one failed basic validation")
	}

	if err := misbehaviour.SignatureTwo.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "signature two failed basic validation")
	}

	// misbehaviour signatures cannot be identical
	if bytes.Equal(misbehaviour.SignatureOne.Signature, misbehaviour.SignatureTwo.Signature) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidMisbehaviour, "misbehaviour signatures cannot be equal")
	}

	// message data signed cannot be identical
	if bytes.Equal(misbehaviour.SignatureOne.Data, misbehaviour.SignatureTwo.Data) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidMisbehaviour, "misbehaviour signature data must be signed over different messages")
	}

	return nil
}

// ValidateBasic ensures that the signature and data fields are non-empty.
func (sd SignatureAndData) ValidateBasic() error {
	if len(sd.Signature) == 0 {
		return sdkerrors.Wrap(ErrInvalidSignatureAndData, "signature cannot be empty")
	}
	if len(sd.Data) == 0 {
		return sdkerrors.Wrap(ErrInvalidSignatureAndData, "data for signature cannot be empty")
	}
	if sd.DataType == UNSPECIFIED {
		return sdkerrors.Wrap(ErrInvalidSignatureAndData, "data type cannot be UNSPECIFIED")
	}
	if sd.Timestamp == 0 {
		return sdkerrors.Wrap(ErrInvalidSignatureAndData, "timestamp cannot be 0")
	}

	return nil
}
