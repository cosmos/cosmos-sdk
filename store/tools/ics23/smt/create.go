package smtproofs

import (
	"crypto/sha256"
	"fmt"

	ics23 "github.com/confio/ics23/go"
	"github.com/lazyledger/smt"
)

// PreimageMap represents an interface for accessing hashed tree paths and retrieving their
// corresponding preimages.
type PreimageMap interface {
	// KeyFor returns the preimage (key) for given path index.
	KeyFor(int) []byte
	// FindPath returns the ordered index of a given path, and whether it's contained in the tree.
	// If not found, the returned index is where the path would be inserted.
	FindPath([32]byte) (int, bool)
	// Len returns the number of mapped paths.
	Len() int
}

// CreateMembershipProof will produce a CommitmentProof that the given key (and queries value) exists in the SMT.
// If the key doesn't exist in the tree, this will return an error.
func CreateMembershipProof(tree *smt.SparseMerkleTree, key []byte) (*ics23.CommitmentProof, error) {
	exist, err := createExistenceProof(tree, key)
	if err != nil {
		return nil, err
	}
	proof := &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{
			Exist: exist,
		},
	}
	return proof, nil
}

func createExistenceProof(tree *smt.SparseMerkleTree, key []byte) (*ics23.ExistenceProof, error) {
	has, err := tree.Has(key)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("Cannot create ExistenceProof when key not in state")
	}
	value, err := tree.Get(key)
	if err != nil {
		return nil, err
	}
	proof, err := tree.Prove(key)
	if err != nil {
		return nil, err
	}
	path := sha256.Sum256(key)
	return &ics23.ExistenceProof{
		Key:   path[:],
		Value: value,
		Leaf:  ics23.SmtSpec.LeafSpec,
		Path:  convertInnerOps(path[:], proof.SideNodes),
	}, nil
}

// CreateNonMembershipProof will produce a CommitmentProof that the given key doesn't exist in the SMT.
// If the key exists in the tree, this will return an error.
func CreateNonMembershipProof(tree *smt.SparseMerkleTree, key []byte, preimages PreimageMap) (*ics23.CommitmentProof, error) {
	path := sha256.Sum256(key)
	has, err := tree.Has(key)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, fmt.Errorf("Cannot create NonExistenceProof when key in state")
	}
	nonexist := &ics23.NonExistenceProof{
		Key: path[:],
	}
	ix, found := preimages.FindPath(path)
	if found {
		return nil, fmt.Errorf("Found index for key not in state")
	}
	if ix > 0 {
		nonexist.Left, err = createExistenceProof(tree, preimages.KeyFor(ix-1))
		if err != nil {
			return nil, err
		}
	}
	if ix < preimages.Len() {
		nonexist.Right, err = createExistenceProof(tree, preimages.KeyFor(ix))
		if err != nil {
			return nil, err
		}
	}
	return &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Nonexist{
			nonexist,
		},
	}, nil
}

func convertInnerOps(path []byte, sideNodes [][]byte) []*ics23.InnerOp {
	depth := len(sideNodes)
	inners := make([]*ics23.InnerOp, 0, depth)
	for i := 0; i < len(sideNodes); i++ {
		op := &ics23.InnerOp{
			Hash:   ics23.HashOp_SHA256,
			Prefix: []byte{1},
		}
		if getBitAtFromMSB(path[:], depth-1-i) == 1 {
			// right child is on path
			op.Prefix = append(op.Prefix, sideNodes[i]...)
		} else {
			op.Suffix = sideNodes[i]
		}
		inners = append(inners, op)
	}
	return inners
}

// getBitAtFromMSB gets the bit at an offset from the most significant bit
// Copied from github.com/celestiaorg/smt
func getBitAtFromMSB(data []byte, position int) int {
	if int(data[position/8])&(1<<(8-1-uint(position)%8)) > 0 {
		return 1
	}
	return 0
}
