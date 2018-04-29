package merkle

import (
	"bytes"
	"fmt"
	"hash"

	"golang.org/x/crypto/ripemd160"
)

// Verify proves the ExistsProof
func (p ExistsProof) Verify(leaf []byte) error {
	data := p.Data

	if data.Prefix != nil {
		leaf = append(data.Prefix, leaf...)
	}
	if data.Suffix != nil {
		leaf = append(leaf, data.Suffix...)
	}

	leaf = data.Op.Hash(leaf)

	for _, node := range p.Nodes {
		if node.Prefix != nil {
			leaf = append(node.Prefix, leaf...)
		}
		if node.Suffix != nil {
			leaf = append(leaf, node.Suffix...)
		}
		leaf = node.Op.Hash(leaf)
	}

	if !bytes.Equal(leaf, p.RootHash) {
		return fmt.Errorf("Verification failed")
	}

	return nil
}

// Root returns the root of the ExistsProof
func (p ExistsProof) Root() []byte {
	return p.RootHash
}

// Verify proves the AbsentProof
func (p AbsentProof) Verify(leaf []byte) error {
	panic("not implemented")
}

// Root returns the root of the AbsentProof
func (p AbsentProof) Root() []byte {
	panic("not implemented")
}

// Root returns the root of the MultiProof
func (p MultiProof) Root() []byte {
	if p.SubProofs == nil {
		return p.KeyProof.Root()
	}
	return p.SubProofs[len(p.SubProofs)-1].Proof.Root()

}

// Lifter defines how does a subroot becomes the leaf element of the higher tree
// The second argument will provide additional information
type Lifter func(int, [][]byte, []byte) ([]byte, error)

// Verify proves the MultiProof
func (p MultiProof) Verify(leaf []byte, lift Lifter, root []byte) (err error) {
	kp := p.KeyProof
	subroot := kp.Root()
	err = kp.Verify(leaf)
	if err != nil {
		return
	}

	for i, p := range p.SubProofs {
		var leaf []byte
		leaf, err = lift(i, p.Infos, subroot)

		if err != nil {
			return
		}

		err = p.Proof.Verify(leaf)
		if err != nil {
			return
		}
		subroot = p.Proof.Root()
	}

	if !bytes.Equal(subroot, root) {
		return fmt.Errorf("Root not match")
	}

	return nil

}

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
