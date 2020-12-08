package types

import codectypes "github.com/cosmos/cosmos-sdk/codec/types"

// ProofInfo contains the upgraded client and consensus state along with the associated proofs
// This will be useful for relayers who want to upgrade IBC clients of this chain on other counterparty chains
type ProofInfo struct {
	ClientState    *codectypes.Any `json:"client_state"`
	ClientProof    []byte          `json:"client_proof"`
	ConsensusState *codectypes.Any `json:"consensus_state"`
	ConsensusProof []byte          `json:"consensus_proof"`
}
