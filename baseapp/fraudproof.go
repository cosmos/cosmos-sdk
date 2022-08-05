package baseapp

import "github.com/lazyledger/smt"

type FraudProof struct {
	blockHeight int64

	stateWitness map[string]StateWitness
}

type StateWitness struct {
	root        []byte
	WitnessData []WitnessData
}

type WitnessData struct {
	Key   []byte
	Value []byte
	proof smt.SparseMerkleProof
}
