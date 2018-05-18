package merkle

import (
	"bytes"
	"errors"
	"hash"

	"golang.org/x/crypto/ripemd160"

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

// Hash hashes the byte slice as defined in the HashOp
func (op HashOp) Hash(bz []byte) (res []byte) {
	var hasher hash.Hash
	switch op {
	case Nop:
		return bz
	case Ripemd160:
		hasher = ripemd160.New()
	default:
		panic("not implemented")
	}

	hasher.Write(bz)
	return hasher.Sum(nil)
}

// Node - inner node for merkle proof
type Node struct {
	Prefix []byte
	Suffix []byte
	Op     HashOp
}

// ExistsProof - merkle proof for verifying existence of an element
type ExistsProof []Node

func (p ExistsProof) Run(data []byte) ([]byte, error) {
	for _, node := range p {
		data = node.Op.Hash(append(append(node.Prefix, data...), node.Suffix...))
	}
	return data, nil
}

// KeyProof - oneof ExistsProof/AbsentProof/RangeProof
type KeyProof interface {
	Run([]byte) ([]byte, error)
}

type Wrapper interface {
	Wrap(string, []byte) []byte
}

// SubProof - merkle proof for verifying subtree structure
type SubProof struct {
	ExistsProof
	Wrapper
	IsDescriptor bool
	Key          string
}

// MultiProof - proof for verifying multi layer merkle tree
type MultiProof struct {
	KeyProof  KeyProof
	SubProofs []SubProof
}

func (p MultiProof) Verify(data []byte, root []byte, keys ...string) error {
	data, err := p.KeyProof.Run(data)
	if err != nil {
		return err
	}

	for _, sp := range p.SubProofs {
		if sp.IsDescriptor {
			if keys == nil {
				return errors.New("Subproof length not match with given keys length")
			}
			if keys[0] != sp.Key {
				return errors.New("Subproof key not match with given key")
			}
			keys = keys[1:]
		}
		data = sp.Wrap(sp.Key, data)
		data, err = sp.Run(data)
		if err != nil {
			return err
		}
	}

	if !bytes.Equal(data, root) {
		return errors.New("Calculated root not match with given root")
	}

	return nil
}

// Bytes converts a MultiProof to a byte slice
func (p MultiProof) Bytes() ([]byte, error) {
	cdc := wire.NewCodec()
	RegisterCodec(cdc)
	return cdc.MarshalBinary(p)
}
