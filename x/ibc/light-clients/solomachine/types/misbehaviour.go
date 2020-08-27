package types

import (
	"bytes"

	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var (
	_ evidenceexported.Evidence   = (*Misbehaviour)(nil)
	_ clientexported.Misbehaviour = (*Misbehaviour)(nil)
)

// ClientType is a Solo Machine light client.
func (misbehaviour Misbehaviour) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (misbehaviour Misbehaviour) GetClientID() string {
	return misbehaviour.ClientId
}

// Route implements Evidence interface.
func (misbehaviour Misbehaviour) Route() string {
	return clienttypes.SubModuleName
}

// Type implements Evidence interface.
func (misbehaviour Misbehaviour) Type() string {
	return clientexported.TypeEvidenceClientMisbehaviour
}

// String implements Evidence interface.
func (misbehaviour Misbehaviour) String() string {
	out, _ := yaml.Marshal(misbehaviour)
	return string(out)
}

// Hash implements Evidence interface
func (misbehaviour Misbehaviour) Hash() tmbytes.HexBytes {
	bz := SubModuleCdc.MustMarshalBinaryBare(&misbehaviour)
	return tmhash.Sum(bz)
}

// GetHeight returns the sequence at which misbehaviour occurred.
func (misbehaviour Misbehaviour) GetHeight() int64 {
	return int64(misbehaviour.Sequence)
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

	return nil
}
