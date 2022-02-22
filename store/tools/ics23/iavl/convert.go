package iavlproofs

import (
	"fmt"

	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/iavl"
)

// convertExistenceProof will convert the given proof into a valid
// existence proof, if that's what it is.
//
// This is the simplest case of the range proof and we will focus on
// demoing compatibility here
func convertExistenceProof(p *iavl.RangeProof, key, value []byte) (*ics23.ExistenceProof, error) {
	if len(p.Leaves) != 1 {
		return nil, fmt.Errorf("Existence proof requires RangeProof to have exactly one leaf")
	}
	return &ics23.ExistenceProof{
		Key:   key,
		Value: value,
		Leaf:  convertLeafOp(p.Leaves[0].Version),
		Path:  convertInnerOps(p.LeftPath),
	}, nil
}

func convertLeafOp(version int64) *ics23.LeafOp {
	// this is adapted from iavl/proof.go:proofLeafNode.Hash()
	prefix := aminoVarInt(0)
	prefix = append(prefix, aminoVarInt(1)...)
	prefix = append(prefix, aminoVarInt(version)...)

	return &ics23.LeafOp{
		Hash:         ics23.HashOp_SHA256,
		PrehashValue: ics23.HashOp_SHA256,
		Length:       ics23.LengthOp_VAR_PROTO,
		Prefix:       prefix,
	}
}

// we cannot get the proofInnerNode type, so we need to do the whole path in one function
func convertInnerOps(path iavl.PathToLeaf) []*ics23.InnerOp {
	steps := make([]*ics23.InnerOp, 0, len(path))

	// lengthByte is the length prefix prepended to each of the sha256 sub-hashes
	var lengthByte byte = 0x20

	// we need to go in reverse order, iavl starts from root to leaf,
	// we want to go up from the leaf to the root
	for i := len(path) - 1; i >= 0; i-- {
		// this is adapted from iavl/proof.go:proofInnerNode.Hash()
		prefix := aminoVarInt(int64(path[i].Height))
		prefix = append(prefix, aminoVarInt(path[i].Size)...)
		prefix = append(prefix, aminoVarInt(path[i].Version)...)

		var suffix []byte
		if len(path[i].Left) > 0 {
			// length prefixed left side
			prefix = append(prefix, lengthByte)
			prefix = append(prefix, path[i].Left...)
			// prepend the length prefix for child
			prefix = append(prefix, lengthByte)
		} else {
			// prepend the length prefix for child
			prefix = append(prefix, lengthByte)
			// length-prefixed right side
			suffix = []byte{lengthByte}
			suffix = append(suffix, path[i].Right...)
		}

		op := &ics23.InnerOp{
			Hash:   ics23.HashOp_SHA256,
			Prefix: prefix,
			Suffix: suffix,
		}
		steps = append(steps, op)
	}
	return steps
}

func aminoVarInt(orig int64) []byte {
	// amino-specific byte swizzling
	i := uint64(orig) << 1
	if orig < 0 {
		i = ^i
	}

	// avoid multiple allocs for normal case
	res := make([]byte, 0, 8)

	// standard protobuf encoding
	for i >= 1<<7 {
		res = append(res, uint8(i&0x7f|0x80))
		i >>= 7
	}
	res = append(res, uint8(i))
	return res
}
