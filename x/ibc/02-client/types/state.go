package types

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ClientState is a type that represents the state of a client.
// Any actor holding the Stage can access on and modify that client information.
type ClientState struct {
	// Client ID
	id string
	// Past state roots required to avoid race conditions between client updates
	// and proof-carrying transactions as defined in
	// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#utilising-past-roots
	Roots []ics23.Root `json:"roots" yaml:"roots"`
	// Consensus state bytes
	ConsensusState exported.ConsensusState `json:"consensus_state" yaml:"consensus_state"`
	// Boolean that states if the client is frozen when a misbehaviour proof is
	// submitted in the event of an equivocation.
	Frozen bool `json:"frozen" yaml:"frozen"`
}

// NewClientState creates a new ClientState instance
func NewClientState(id string) ClientState {
	return ClientState{
		id:     id,
		Frozen: false,
	}
}

// ID returns the client identifier
func (cs ClientState) ID() string {
	return cs.id
}

// func (state State) RootCLI(q cs.ABCIQuerier, height uint64) (res ics23.Root, proof merkle.Proof, err error) {
// 	root := cs.Roots.Value(height)
// 	tmProof, err := root.Query(q, &res)
// 	proof = merkle.NewProofFromValue(tmProof, cs.prefix(), root)
// 	return
// }

// func (state State) ConsensusStateCLI(q cs.ABCIQuerier) (res exported.ConsensusState, proof merkle.Proof, err error) {
// 	tmproof, err := cs.ConsensusState.Query(q, &res)
// 	proof = merkle.NewProofFromValue(tmproof, cs.prefix(), cs.ConsensusState)
// 	return
// }

// func (state State) FrozenCLI(q cs.ABCIQuerier) (res bool, proof merkle.Proof, err error) {
// 	res, tmproof, err := cs.Frozen.Query(q)
// 	proof = merkle.NewProofFromValue(tmproof, cs.prefix(), cs.Frozen)
// 	return
// }

// func (state State) prefix() []byte {
// 	return bytes.Split(cs.ConsensusState.KeyBytes(), []byte(SubModuleName+"/"))[0]
// }
