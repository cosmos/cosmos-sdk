package baseapp

import (
	"github.com/lazyledger/smt"
	"github.com/tendermint/tendermint/crypto/merkle"
)

// Represents a single-round fraudProof
type FraudProof struct {
	// The block height to load state of
	blockHeight int64

	// A map from module name to state witness
	stateWitness map[string]StateWitness
}

// State witness with a list of all witness data
type StateWitness struct {
	// store level proof
	proof    merkle.ProofOperator
	rootHash []byte
	// List of witness data
	WitnessData []WitnessData
}

// Witness data containing a key/value pair and a SMT proof for said key/value pair
type WitnessData struct {
	Key   []byte
	Value []byte
	proof smt.SparseMerkleProof
}
