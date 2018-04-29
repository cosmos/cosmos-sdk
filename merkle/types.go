package merkle

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// HashOp - identifier for hash operations
type HashOp uint8

// Defined HashOps
const (
	Nop = HashOp(iota)
	Ripemd160
	Sha224
	Sha256
	Sha284
	Sha512
)

// Node - inner node for merkle proof
type Node struct {
	Prefix []byte
	Suffix []byte
	Op     HashOp
}

// ExistsData - leaf node data for ExistsProof
type ExistsData struct {
	Prefix []byte
	Suffix []byte
	Op     HashOp
}

// ExistsProof - merkle proof for verifying existence of an element
type ExistsProof struct {
	PrefixLeaf  []byte
	PrefixInner []byte
	Data        ExistsData
	Nodes       []Node
	RootHash    []byte
}

// AbsentProof - merkle proof for verifying nonexistence of an element
type AbsentProof struct {
}

// RangeProof - merkle proof for verifying existence of a set of elements
type RangeProof struct {
}

// KeyProof - oneof ExistsProof/AbsentProof/RangeProof
type KeyProof interface {
	Verify(leaf []byte) error
	Root() []byte
}

// SubProof - merkle proof for verifying subtree structure
type SubProof struct {
	Proof KeyProof
	Infos [][]byte
}

// MultiProof - proof for verifying multi layer merkle tree
type MultiProof struct {
	KeyProof  KeyProof
	SubProofs []SubProof
}

// Bytes converts a MultiProof to a byte slice
func (p MultiProof) Bytes() ([]byte, error) {
	cdc := wire.NewCodec()
	RegisterCodec(cdc)
	return cdc.MarshalBinary(p)
}
