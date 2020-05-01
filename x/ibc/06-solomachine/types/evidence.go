package types

import (
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var (
	_ evidenceexported.Evidence   = Evidence{}
	_ clientexported.Misbehaviour = Evidence{}
)

// Evidence is proof of misbehaviour for a solo machine which consists of a sequence
// and two signatures over different messages at that sequence.
type Evidence struct {
	ClientID     string           `json:"client_id" yaml:"client_id"`
	Sequence     uint64           `json:"sequence" yaml:"sequnce"`
	SignatureOne SignatureAndData `json:"signature_one" yaml:"signature_one"`
	SignatureTwo SignatureAndData `json:"signature_one" yam;:"signature_two"`
}

// ClientType is a Solo Machine light client
func (ev Evidence) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetClientID returns the ID of the client that committed a misbehaviour
func (ev Evidence) GetClientID() string {
	return ev.ClientID
}

// Route implements Evidence interface
func (ev Evidence) Route() string {
	return clienttypes.SubModuleName
}

// Type implements Evidence interface
func (ev Evidence) Type() string {
	return "client_misbehaviour"
}

// String implements Evidence interface
func (ev Evidence) String() string {
	bz, err := yaml.Marshal(ev)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// GetHeight returns the sequence at which misbehaviour occurred
func (ev Evidence) GetHeight() int64 {
	return ev.Sequence
}

// ValidateBasic implements Evidence interface
func (ev Evidence) ValidateBasic() error {
	if err := host.DefaultClientIdentifierValidator(ev.ClientID); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	if sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidSequence, "sequence cannot be 0")
	}

	// signatures should not be equal

	// data of signatures must be different
}

// SignatureAndData is a signature and the data signed over to create the signature.
type SignatureAndData struct {
	Signature []byte `json:"signature" yaml:"signature"`
	Data      []byte `json:"data" yaml:"data"`
}
