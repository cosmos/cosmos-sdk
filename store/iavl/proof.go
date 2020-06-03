package iavl

import (
	"errors"
	"fmt"

	ics23 "github.com/confio/ics23/go"
	"github.com/tendermint/tendermint/crypto/merkle"
)

const ProofOpIAVLCommitment = "ics23:iavl"
const ProofOpSimpleMerkleCommitment = "ics23:simple"

// CommitmentOp implements merkle.ProofOperator by wrapping an ics23 CommitmentProof
// It also contains a Key field to determine which key the proof is proving.
// NOTE: CommitmentProof currently can either be ExistenceProof or NonexistenceProof
//
// Type and Spec are classified by the kind of merkle proof it represents allowing
// the code to be reused by more types. Spec is never on the wire, but mapped from type in the code.
type CommitmentOp struct {
	Type  string
	Spec  *ics23.ProofSpec
	Key   []byte
	Proof *ics23.CommitmentProof
}

var _ merkle.ProofOperator = CommitmentOp{}

func NewIavlCommitmentOp(key []byte, proof *ics23.CommitmentProof) CommitmentOp {
	return CommitmentOp{
		Type:  ProofOpIAVLCommitment,
		Spec:  ics23.IavlSpec,
		Key:   key,
		Proof: proof,
	}
}

func NewSimpleMerkleCommitmentOp(key []byte, proof *ics23.CommitmentProof) CommitmentOp {
	return CommitmentOp{
		Type:  ProofOpSimpleMerkleCommitment,
		Spec:  ics23.TendermintSpec,
		Key:   key,
		Proof: proof,
	}
}

// CommitmentOpDecoder takes a merkle.ProofOp and attempt to decode it into a CommitmentOp ProofOperator
// The proofOp.Data is just a marshalled CommitmentProof. The Key of the CommitmentOp is extracted
// from the unmarshalled proof
func CommitmentOpDecoder(pop merkle.ProofOp) (merkle.ProofOperator, error) {
	var spec *ics23.ProofSpec
	switch pop.Type {
	case ProofOpIAVLCommitment:
		spec = ics23.IavlSpec
	case ProofOpSimpleMerkleCommitment:
		spec = ics23.TendermintSpec
	default:
		return nil, fmt.Errorf("unexpected ProofOp.Type; got %v, want supported ics23 subtype", pop.Type)
	}

	proof := &ics23.CommitmentProof{}
	err := proof.Unmarshal(pop.Data)
	if err != nil {
		return nil, err
	}

	op := CommitmentOp{
		Type:  pop.Type,
		Key:   pop.Key,
		Spec:  spec,
		Proof: proof,
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

		absent := ics23.VerifyNonMembership(op.Spec, root, op.Proof, op.Key)
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

		exists := ics23.VerifyMembership(op.Spec, root, op.Proof, op.Key, args[0])
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
		Type: op.Type,
		Key:  op.Key,
		Data: bz,
	}
}
