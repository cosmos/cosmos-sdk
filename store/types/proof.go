package types

import (
	ics23 "github.com/confio/ics23/go"
	"github.com/tendermint/tendermint/crypto/merkle"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	ProofOpIAVLCommitment         = "ics23:iavl"
	ProofOpSimpleMerkleCommitment = "ics23:simple"
)

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

// CommitmentOpDecoder takes a merkle.ProofOp and attempts to decode it into a CommitmentOp ProofOperator
// The proofOp.Data is just a marshalled CommitmentProof. The Key of the CommitmentOp is extracted
// from the unmarshalled proof.
func CommitmentOpDecoder(pop merkle.ProofOp) (merkle.ProofOperator, error) {
	var spec *ics23.ProofSpec
	switch pop.Type {
	case ProofOpIAVLCommitment:
		spec = ics23.IavlSpec
	case ProofOpSimpleMerkleCommitment:
		spec = ics23.TendermintSpec
	default:
		return nil, sdkerrors.Wrapf(ErrInvalidProof, "unexpected ProofOp.Type; got %s, want supported ics23 subtypes 'ProofOpIAVLCommitment' or 'ProofOpSimpleMerkleCommitment'", pop.Type)
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

// Run takes in a list of arguments and attempts to run the proof op against these arguments
// Returns the root wrapped in [][]byte if the proof op succeeds with given args. If not,
// it will return an error.
//
// CommitmentOp will accept args of length 1 or length 0
// If length 1 args is passed in, then CommitmentOp will attempt to prove the existence of the key
// with the value provided by args[0] using the embedded CommitmentProof and return the CommitmentRoot of the proof
// If length 0 args is passed in, then CommitmentOp will attempt to prove the absence of the key
// in the CommitmentOp and return the CommitmentRoot of the proof
func (op CommitmentOp) Run(args [][]byte) ([][]byte, error) {
	// Only support an existence proof or nonexistence proof (batch proofs currently unsupported)
	switch len(args) {
	case 0:
		// Args are nil, so we verify the absence of the key.
		nonexistProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Nonexist)
		if !ok {
			return nil, sdkerrors.Wrap(ErrInvalidProof, "proof is not a nonexistence proof and args is nil")
		}

		// get root from either left or right existence proof. Note they must have the same root if both exist
		// and at least one proof must be non-nil
		var (
			root []byte
			err  error
		)
		switch {
		// check left proof to calculate root
		case nonexistProof.Nonexist.Left != nil:
			root, err = nonexistProof.Nonexist.Left.Calculate()
			if err != nil {
				return nil, sdkerrors.Wrap(ErrInvalidProof, "could not calculate root from nonexistence proof")
			}
		case nonexistProof.Nonexist.Right != nil:
			// Left proof is nil, check right proof
			root, err = nonexistProof.Nonexist.Right.Calculate()
			if err != nil {
				return nil, sdkerrors.Wrap(ErrInvalidProof, "could not calculate root from nonexistence proof")
			}
		default:
			// both left and right existence proofs are empty
			// this only proves absence against a nil root (empty store)
			return [][]byte{nil}, nil
		}

		absent := ics23.VerifyNonMembership(op.Spec, root, op.Proof, op.Key)
		if !absent {
			return nil, sdkerrors.Wrapf(ErrInvalidProof, "proof did not verify absence of key: %s", string(op.Key))
		}

		return [][]byte{root}, nil

	case 1:
		// Args is length 1, verify existence of key with value args[0]
		existProof, ok := op.Proof.Proof.(*ics23.CommitmentProof_Exist)
		if !ok {
			return nil, sdkerrors.Wrap(ErrInvalidProof, "proof is not a existence proof and args is length 1")
		}
		// For subtree verification, we simply calculate the root from the proof and use it to prove
		// against the value
		root, err := existProof.Exist.Calculate()
		if err != nil {
			return nil, sdkerrors.Wrap(ErrInvalidProof, "could not calculate root from existence proof")
		}

		if !ics23.VerifyMembership(op.Spec, root, op.Proof, op.Key, args[0]) {
			return nil, sdkerrors.Wrapf(ErrInvalidProof, "proof did not verify existence of key %s with given value %x", op.Key, args[0])
		}

		return [][]byte{root}, nil
	default:
		return nil, sdkerrors.Wrapf(ErrInvalidProof, "args must be length 0 or 1, got: %d", len(args))
	}
}

// ProofOp implements ProofOperator interface and converts a CommitmentOp
// into a merkle.ProofOp format that can later be decoded by CommitmentOpDecoder
// back into a CommitmentOp for proof verification
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
