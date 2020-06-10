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
	_ evidenceexported.Evidence   = Evidence{}
	_ clientexported.Misbehaviour = Evidence{}
)

// Evidence is proof of misbehaviour for a solo machine which consists of a sequence
// and two signatures over different messages at that sequence.
type Evidence struct {
	ClientID     string           `json:"client_id" yaml:"client_id"`
	Sequence     uint64           `json:"sequence" yaml:"sequence"`
	SignatureOne SignatureAndData `json:"signature_one" yaml:"signature_one"`
	SignatureTwo SignatureAndData `json:"signature_two" yam;:"signature_two"`
}

// ClientType is a Solo Machine light client.
func (ev Evidence) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (ev Evidence) GetClientID() string {
	return ev.ClientID
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
	bz, err := yaml.Marshal(ev)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// Hash implements Evidence interface
func (ev Evidence) Hash() tmbytes.HexBytes {
	bz := SubModuleCdc.MustMarshalBinaryBare(ev)
	return tmhash.Sum(bz)
}

// GetHeight returns the sequence at which misbehaviour occurred.
func (ev Evidence) GetHeight() int64 {
	return int64(ev.Sequence)
}

// ValidateBasic implements Evidence interface.
func (ev Evidence) ValidateBasic() error {
	if err := host.ClientIdentifierValidator(ev.ClientID); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	if ev.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "sequence cannot be 0")
	}

	if err := ev.SignatureOne.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	if err := ev.SignatureTwo.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	// evidence signatures cannot be identical
	if bytes.Equal(ev.SignatureOne.Signature, ev.SignatureTwo.Signature) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "evidence signatures cannot be equal")
	}

	// message data signed cannot be identical
	if bytes.Equal(ev.SignatureOne.Data, ev.SignatureTwo.Data) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "evidence signatures must be signed over different messages")
	}

	return nil
}

// SignatureAndData is a signature and the data signed over to create the signature.
type SignatureAndData struct {
	Signature []byte `json:"signature" yaml:"signature"`
	Data      []byte `json:"data" yaml:"data"`
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
