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

func (p ExistsProof) SubVerify(leaf, root []byte) error {
	return nil
}

func (p AbsentProof) Verify(key, value, root []byte) error {
	panic("not implemented")
}

func (p AbsentProof) Root() []byte {
	panic("not implemented")
}

func (p MultiProof) Verify(leaf []byte, root []byte) (err error) {
	kp := p.KeyProof
	subroot := kp.Root()
	err = kp.Verify(leaf)
	if err != nil {
		return
	}

	for _, p := range p.SubProofs {
		err = p.SubVerify(subroot, p.RootHash)
		if err != nil {
			return
		}
		subroot = p.Root()
	}

	if !bytes.Equal(subroot, root) {
		return fmt.Errorf("Root not match")
	}

	return nil

}

func (op HashOp) Hash(bz []byte) (res []byte) {
	var hasher hash.Hash
	switch op {
	case RIPEMD160:
		hasher = ripemd160.New()
	default:
		panic("not implemented")
	}

	hasher.Write(bz)
	hasher.Sum(res)
	return
}
