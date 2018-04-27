package merkle

import (
	"bytes"
	"fmt"
	"hash"

	"golang.org/x/crypto/ripemd160"
)

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

func (p ExistsProof) Root() []byte {
	return p.RootHash
}

func (p AbsentProof) Verify(leaf []byte) error {
	panic("not implemented")
}

func (p AbsentProof) Root() []byte {
	panic("not implemented")
}

func (p MultiProof) Root() []byte {
	if p.SubProofs == nil {
		return p.KeyProof.Root()
	} else {
		return p.SubProofs[len(p.SubProofs)-1].Proof.Root()
	}
}

type Inner func(int, [][]byte, []byte) ([]byte, error)

func (p MultiProof) Verify(leaf []byte, iv Inner, root []byte) (err error) {
	kp := p.KeyProof
	subroot := kp.Root()
	err = kp.Verify(leaf)
	if err != nil {
		return
	}

	for i, p := range p.SubProofs {
		var leaf []byte
		leaf, err = iv(i, p.Infos, subroot)

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

func (op HashOp) Hash(bz []byte) (res []byte) {
	var hasher hash.Hash
	switch op {
	case NOP:
		return bz
	case RIPEMD160:
		hasher = ripemd160.New()
	default:
		panic("not implemented")
	}

	hasher.Write(bz)
	return hasher.Sum(nil)
}
