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

	fmt.Printf("hashed': %+v\n", leaf)

	leaf = data.Op.Hash(leaf)

	fmt.Printf("OneHash: %+v\n", leaf)

	for _, node := range p.Nodes {
		if node.Prefix != nil {
			leaf = append(node.Prefix, leaf...)
		}
		if node.Suffix != nil {
			leaf = append(leaf, node.Suffix...)
		}
		fmt.Printf("hashed': %+v\n", leaf)
		leaf = node.Op.Hash(leaf)

		fmt.Printf("AftHash: %+v\n", leaf)
	}

	if !bytes.Equal(leaf, p.RootHash) {
		return fmt.Errorf("Verification failed")
	}

	return nil
}

func (p ExistsProof) Root() []byte {
	return p.RootHash
}

func (p AbsentProof) Verify(key, value, root []byte) error {
	panic("not implemented")
}

func (p AbsentProof) Root() []byte {
	panic("not implemented")
}

func (p MultiProof) Root() []byte {
	if p.SubProofs == nil {
		return p.KeyProof.Root()
	} else {
		return p.SubProofs[len(p.SubProofs)-1].Root()
	}
}

func (p MultiProof) Verify(leaf []byte, root []byte, inners ...[]byte) (err error) {
	fmt.Printf("Proof: %+v\n", p)

	kp := p.KeyProof
	subroot := kp.Root()
	err = kp.Verify(leaf)
	if err != nil {
		return
	}

	if len(p.SubProofs) != len(inners) {
		return fmt.Errorf("Subproof roots length not match")
	}

	for i, p := range p.SubProofs {
		err = p.Verify(subroot)
		if err != nil {
			return
		}
		subroot = p.Root()
		if !bytes.Equal(subroot, inners[i]) {
			return fmt.Errorf("Subproof root #%d not match", i)
		}
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
	return hasher.Sum(nil)
}
