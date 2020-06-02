package iavl

import (
	"errors"
	"fmt"

	ics23 "github.com/confio/ics23/go"
	"github.com/tendermint/tendermint/crypto/merkle"
)

const ProofOpIAVLCommitment = "iavlstore"

// IavlOP implements merkle.ProofOperator by wrapping an ics23 CommitmentProof
// It also contains a Key field to determine which key the proof is proving.
// NOTE: CommitmentProof currently can either be ExistenceProof or NonexistenceProof
type CommitmentOp struct {
	Key   []byte
	Proof *ics23.CommitmentProof
}

var _ merkle.ProofOperator = CommitmentOp{}

func NewCommitmentOp(key []byte, proof *ics23.CommitmentProof) CommitmentOp {
	return CommitmentOp{
		Key:   key,
		Proof: proof,
	}
}

// CommitmentOpDecoder takes a merkle.ProofOp and attempt to decode it into a CommitmentOp ProofOperator
// The proofOp.Data is just a marshalled CommitmentProof. The Key of the CommitmentOp is extracted
// from the unmarshalled proof
func CommitmentOpDecoder(pop merkle.ProofOp) (merkle.ProofOperator, error) {
	if pop.Type != ProofOpIAVLCommitment {
		return nil, fmt.Errorf("unexpected ProofOp.Type; got %v, want %v", pop.Type, ProofOpIAVLCommitment)
	}
	var op CommitmentOp
	proof := &ics23.CommitmentProof{}
	err := proof.Unmarshal(pop.Data)
	if err != nil {
		return nil, err
	}
	op.Proof = proof

	// Get Key from proof for now
	if existProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Exist); ok {
		op.Key = existProof.Exist.Key
	} else if nonexistProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Nonexist); ok {
		op.Key = nonexistProof.Nonexist.Key
	} else {
		return nil, errors.New("proof type unsupported")
	}
	return op, nil
}

func (op CommitmentOp) GetKey() []byte {
	return op.Key
}

func (op CommitmentOp) Run(args [][]byte) ([][]byte, error) {
	// Only support an existence proof or nonexistence proof (batch proofs currently unsupported)
	switch len(args) {
	case 0:
		// Args are nil, so we verify the absence of the key.
		nonexistProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Nonexist)
		if !ok {
			return nil, errors.New("proof is not a nonexistence proof and args is nil")
		}

		// get root from either left or right existence proof. Note they must have the same root if both exist
		// and at least one proof must be non-nil
		root, err := nonexistProof.Nonexist.Left.Calculate()
		if err != nil {
			// Left proof may be nil, check right proof
			root, err = nonexistProof.Nonexist.Right.Calculate()
			if err != nil {
				return nil, errors.New("could not calculate root from nonexistence proof")
			}
		}

		absent := ics23.VerifyNonMembership(ics23.IavlSpec, root, op.Proof, op.Key)
		if !absent {
			return nil, fmt.Errorf("proof did not verify absence of key: %s", string(op.Key))
		}

		return [][]byte{root}, nil

	case 1:
		// Args is length 1, verify existence of key with value args[0]
		existProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Exist)
		if !ok {
			return nil, errors.New("proof is not a existence proof and args is length 1")
		}
		// For subtree verification, we simply calculate the root from the proof and use it to prove
		// against the value
		root, err := existProof.Exist.Calculate()
		if err != nil {
			return nil, errors.New("could not calculate root from existence proof")
		}

		exists := ics23.VerifyMembership(ics23.IavlSpec, root, op.Proof, op.Key, args[0])
		if !exists {
			return nil, fmt.Errorf("proof did not verify existence of key %s with given value %x", op.Key, args[0])
		}

		return [][]byte{root}, nil
	default:
		return nil, fmt.Errorf("args must be length 0 or 1, got: %d", len(args))
	}
}

func (op CommitmentOp) ProofOp() merkle.ProofOp {
	bz, err := op.Proof.Marshal()
	if err != nil {
		panic(err.Error())
	}
	return merkle.ProofOp{
		Type: ProofOpIAVLCommitment,
		Key:  op.Key,
		Data: bz,
	}
}
