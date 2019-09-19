package client

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store/state"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func (obj Object) prefix() []byte {
	return bytes.Split(obj.ConsensusState.KeyBytes(), LocalRoot())[0]
}

func (obj Object) RootCLI(q state.ABCIQuerier, height uint64) (res commitment.Root, proof merkle.Proof, err error) {
	root := obj.Roots.Value(height)
	tmproof, err := root.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), root)
	return
}

func (obj Object) ConsensusStateCLI(q state.ABCIQuerier) (res ConsensusState, proof merkle.Proof, err error) {
	tmproof, err := obj.ConsensusState.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.ConsensusState)
	return
}

func (obj Object) FrozenCLI(q state.ABCIQuerier) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := obj.Frozen.Query(q)
	proof = merkle.NewProofFromValue(tmproof, obj.prefix(), obj.Frozen)
	return
}
