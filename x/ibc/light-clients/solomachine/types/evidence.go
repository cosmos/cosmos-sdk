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
	_ evidenceexported.Evidence   = (*Evidence)(nil)
	_ clientexported.Misbehaviour = (*Evidence)(nil)
)

// ClientType is a Solo Machine light client.
func (ev Evidence) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (ev Evidence) GetClientID() string {
	return ev.ClientId
}

// Route implements Evidence interface.
func (ev Evidence) Route() string {
	return clienttypes.SubModuleName
}

// Type implements Evidence interface.
func (ev Evidence) Type() string {
	return "client_misbehaviour"
}

// String implements Evidence interface.
func (ev Evidence) String() string {
	out, _ := yaml.Marshal(ev)
	return string(out)
}

// Hash implements Evidence interface
func (ev Evidence) Hash() tmbytes.HexBytes {
	bz := SubModuleCdc.MustMarshalBinaryBare(&ev)
	return tmhash.Sum(bz)
}

// GetHeight returns the sequence at which misbehaviour occurred.
func (ev Evidence) GetHeight() int64 {
	return int64(ev.Sequence)
}

// ValidateBasic implements Evidence interface.
func (ev Evidence) ValidateBasic() error {
	if err := host.ClientIdentifierValidator(ev.ClientId); err != nil {
		return sdkerrors.Wrap(err, "invalid client identifier for solo machine")
	}

	if ev.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "sequence cannot be 0")
	}

	if err := ev.SignatureOne.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "signature one failed basic validation")
	}

	if err := ev.SignatureTwo.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "signature two failed basic validation")
	}

	// evidence signatures cannot be identical
	if bytes.Equal(ev.SignatureOne.Signature, ev.SignatureTwo.Signature) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "evidence signatures cannot be equal")
	}

	// message data signed cannot be identical
	if bytes.Equal(ev.SignatureOne.Data, ev.SignatureTwo.Data) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "evidence signature data must be signed over different messages")
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
