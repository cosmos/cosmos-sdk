package multi

import (
	"crypto/sha256"

	"github.com/tendermint/tendermint/crypto/merkle"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"

	types "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
)

// DefaultProofRuntime returns a ProofRuntime supporting SMT and simple merkle proofs.
func DefaultProofRuntime() (prt *merkle.ProofRuntime) {
	prt = merkle.NewProofRuntime()
	prt.RegisterOpDecoder(types.ProofOpSMTCommitment, types.CommitmentOpDecoder)
	prt.RegisterOpDecoder(types.ProofOpSimpleMerkleCommitment, types.CommitmentOpDecoder)
	return prt
}

// Prove commitment of key within an smt store and return ProofOps
func proveKey(s *smt.Store, key []byte) (*tmcrypto.ProofOps, error) {
	var ret tmcrypto.ProofOps
	keyProof, err := s.GetProofICS23(key)
	if err != nil {
		return nil, err
	}
	hkey := sha256.Sum256(key)
	ret.Ops = append(ret.Ops, types.NewSmtCommitmentOp(hkey[:], keyProof).ProofOp())
	return &ret, nil
}

// GetProof returns ProofOps containing: a proof for the given key within this substore;
// and a proof of the substore's existence within the MultiStore.
func (s *viewSubstore) GetProof(key []byte) (*tmcrypto.ProofOps, error) {
	ret, err := proveKey(s.stateCommitmentStore, key)
	if err != nil {
		return nil, err
	}

	// Prove commitment of substore within root store
	storeHashes, err := s.root.getMerkleRoots()
	if err != nil {
		return nil, err
	}
	storeProof, err := types.ProofOpFromMap(storeHashes, s.name)
	if err != nil {
		return nil, err
	}
	ret.Ops = append(ret.Ops, storeProof)
	return ret, nil
}
