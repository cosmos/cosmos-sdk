package types

import (
	"github.com/tendermint/tendermint/crypto"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// ConsensusState defines a Solo Machine consensus state
type ConsensusState struct {
	Sequence uint64 `json:"sequence" yaml:"sequence"`

	PublicKey crypto.PublicKey `json:"public_key" yaml: "public_key"`
}

// ClientType returns Solo Machine type
func (ConsensusState) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetHeight returns the sequence number
func (cs ConsensusState) GetHeight() uint64 {
	return cs.Sequence
}

// GetRoot

// ValidateBasic
