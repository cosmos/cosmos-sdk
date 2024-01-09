package store

import (
	"crypto/sha256"

	ics23 "github.com/cosmos/ics23/go"

	errorsmod "cosmossdk.io/errors"
)

// Proof operation types
const (
	ProofOpIAVLCommitment         = "ics23:iavl"
	ProofOpSimpleMerkleCommitment = "ics23:simple"
	ProofOpSMTCommitment          = "ics23:smt"
)

var (
	leafPrefix  = []byte{0}
	innerPrefix = []byte{1}

	// SimpleMerkleSpec is the ics23 proof spec for simple merkle proofs.
	SimpleMerkleSpec = &ics23.ProofSpec{
		LeafSpec: &ics23.LeafOp{
			Prefix:       leafPrefix,
			PrehashKey:   ics23.HashOp_NO_HASH,
			PrehashValue: ics23.HashOp_NO_HASH,
			Hash:         ics23.HashOp_SHA256,
			Length:       ics23.LengthOp_VAR_PROTO,
		},
		InnerSpec: &ics23.InnerSpec{
			ChildOrder:      []int32{0, 1},
			MinPrefixLength: 1,
			MaxPrefixLength: 1,
			ChildSize:       32,
			Hash:            ics23.HashOp_SHA256,
		},
	}
)

// CommitmentOp implements merkle.ProofOperator by wrapping an ics23 CommitmentProof.
// It also contains a Key field to determine which key the proof is proving.
// NOTE: CommitmentProof currently can either be ExistenceProof or NonexistenceProof
//
// Type and Spec are classified by the kind of merkle proof it represents allowing
// the code to be reused by more types. Spec is never on the wire, but mapped
// from type in the code.
type CommitmentOp struct {
	Type  string
	Key   []byte
	Spec  *ics23.ProofSpec
	Proof *ics23.CommitmentProof
}

func NewIAVLCommitmentOp(key []byte, proof *ics23.CommitmentProof) CommitmentOp {
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
		Spec:  SimpleMerkleSpec,
		Key:   key,
		Proof: proof,
	}
}

func NewSMTCommitmentOp(key []byte, proof *ics23.CommitmentProof) CommitmentOp {
	return CommitmentOp{
		Type:  ProofOpSMTCommitment,
		Spec:  ics23.SmtSpec,
		Key:   key,
		Proof: proof,
	}
}

func (op CommitmentOp) GetKey() []byte {
	return op.Key
}

// Run takes in a list of arguments and attempts to run the proof op against these
// arguments. Returns the root wrapped in [][]byte if the proof op succeeds with
// given args. If not, it will return an error.
//
// CommitmentOp will accept args of length 1 or length 0. If length 1 args is
// passed in, then CommitmentOp will attempt to prove the existence of the key
// with the value provided by args[0] using the embedded CommitmentProof and returns
// the CommitmentRoot of the proof. If length 0 args is passed in, then CommitmentOp
// will attempt to prove the absence of the key in the CommitmentOp and return the
// CommitmentRoot of the proof.
func (op CommitmentOp) Run(args [][]byte) ([][]byte, error) {
	// calculate root from proof
	root, err := op.Proof.Calculate()
	if err != nil {
		return nil, errorsmod.Wrapf(ErrInvalidProof, "could not calculate root for proof: %v", err)
	}

	// Only support an existence proof or nonexistence proof (batch proofs currently unsupported)
	switch len(args) {
	case 0:
		// Args are nil, so we verify the absence of the key.
		absent := ics23.VerifyNonMembership(op.Spec, root, op.Proof, op.Key)
		if !absent {
			return nil, errorsmod.Wrapf(ErrInvalidProof, "proof did not verify absence of key: %s", string(op.Key))
		}

	case 1:
		// Args is length 1, verify existence of key with value args[0]
		if !ics23.VerifyMembership(op.Spec, root, op.Proof, op.Key, args[0]) {
			return nil, errorsmod.Wrapf(ErrInvalidProof, "proof did not verify existence of key %s with given value %x", op.Key, args[0])
		}

	default:
		return nil, errorsmod.Wrapf(ErrInvalidProof, "args must be length 0 or 1, got: %d", len(args))
	}

	return [][]byte{root}, nil
}

// ProofFromByteSlices computes the proof from the given leaves.
func ProofFromByteSlices(leaves [][]byte, index int) (rootHash []byte, inners []*ics23.InnerOp) {
	if len(leaves) == 0 {
		return emptyHash(), nil
	}

	n := len(leaves)
	for n > 1 {
		if index < n-1 || index&1 == 1 {
			inner := &ics23.InnerOp{Hash: ics23.HashOp_SHA256}
			if index&1 == 0 {
				inner.Prefix = innerPrefix
				inner.Suffix = leaves[index^1]
			} else {
				inner.Prefix = append(innerPrefix, leaves[index^1]...)
			}
			inners = append(inners, inner)
		}
		for i := 0; i < n/2; i++ {
			leaves[i] = InnerHash(leaves[2*i], leaves[2*i+1])
		}
		if n&1 == 1 {
			leaves[n/2] = leaves[n-1]
		}
		n = (n + 1) / 2
		index /= 2
	}

	rootHash = leaves[0]
	return
}

// ConvertCommitmentOp converts the given merkle proof into an CommitmentOp.
func ConvertCommitmentOp(inners []*ics23.InnerOp, key, value []byte) CommitmentOp {
	return NewSimpleMerkleCommitmentOp(key, &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{
			Exist: &ics23.ExistenceProof{
				Key:   key,
				Value: value,
				Leaf:  SimpleMerkleSpec.LeafSpec,
				Path:  inners,
			},
		},
	})
}

func emptyHash() []byte {
	h := sha256.Sum256([]byte{})
	return h[:]
}

// LeafHash computes the hash of a leaf node.
func LeafHash(key, value []byte) ([]byte, error) {
	return SimpleMerkleSpec.LeafSpec.Apply(key, value)
}

// InnerHash computes the hash of an inner node.
func InnerHash(left, right []byte) []byte {
	data := make([]byte, len(innerPrefix)+len(left)+len(right))
	n := copy(data, innerPrefix)
	n += copy(data[n:], left)
	copy(data[n:], right)
	h := sha256.Sum256(data)
	return h[:]
}
