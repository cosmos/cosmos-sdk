package iavl

import (
	"errors"
	"fmt"

	ics23 "github.com/confio/ics23/go"
	"github.com/tendermint/tendermint/crypto/merkle"
)

const ProofOpIAVL = "iavlstore"

type IAVLOp struct {
	Key   []byte
	Proof *ics23.CommitmentProof
}

var _ merkle.ProofOperator = IAVLOp{}

func NewIAVLOp(key []byte, proof *ics23.CommitmentProof) IAVLOp {
	return IAVLOp{
		Key:   key,
		Proof: proof,
	}
}

func IAVLOpDecoder(pop merkle.ProofOp) (merkle.ProofOperator, error) {
	if pop.Type != ProofOpIAVL {
		return nil, errors.New(fmt.Sprintf("unexpected ProofOp.Type; got %v, want %v", pop.Type, ProofOpIAVL))
	}
	var op IAVLOp
	proof := &ics23.CommitmentProof{}
	err := proof.Unmarshal(pop.Data)
	if err != nil {
		return nil, err
	}
	op.Proof = proof

	// Get Key from proof for now
	if existProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Exist); !ok {
		op.Key = existProof.Exist.Key
	} else if nonexistProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Nonexist); !ok {
		op.Key = nonexistProof.Nonexist.Key
	} else {
		return nil, errors.New("Proof type unsupported")
	}
	return op, nil
}

func (op IAVLOp) GetKey() []byte {
	return op.Key
}

func (op IAVLOp) Run(args [][]byte) ([][]byte, error) {
	switch len(args) {
	case 0:
		// Args are nil, so we verify the absence of the key.
		nonexistProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Nonexist)
		if !ok {
			return nil, errors.New("proof is not a nonexistence proof and args is nil")
		}

		root, err := nonexistProof.Nonexist.Left.Calculate()
		if err != nil {
			return nil, errors.New("could not calculate root from nonexistence proof")
		}

		absent := ics23.VerifyNonMembership(ics23.IavlSpec, root, op.Proof, op.Key)
		if !absent {
			return nil, errors.New(fmt.Sprintf("proof did not verify absence of key: %s", string(op.Key)))
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
			return nil, errors.New(fmt.Sprintf("proof did not verify existence of key %s with given value %x", op.Key, args[0]))
		}

		return [][]byte{root}, nil
	default:
		return nil, errors.New(fmt.Sprintf("args must be length 0 or 1, got: %d", len(args)))
	}
}

func (op IAVLOp) ProofOp() merkle.ProofOp {
	bz, err := op.Proof.Marshal()
	if err != nil {
		panic(err.Error())
	}
	return merkle.ProofOp{
		Type: ProofOpIAVL,
		Key:  op.Key,
		Data: bz,
	}
}
