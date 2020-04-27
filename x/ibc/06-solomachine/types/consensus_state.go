package types

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// ConsensusState defines a Solo Machine consensus state
type ConsensusState struct {
	Sequence uint64 `json:"sequence" yaml:"sequence"`

	PublicKey PublicKey `json:"public_key" yaml: "public_key"`
}

// ClientType return Solo Machine
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetHeight
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Sequence
}

// GetRoot
