package merkle

import (
	"github.com/tendermint/iavl"

	"github.com/tendermint/go-crypto/merkle"
)

func FromIAVLKey(proof iavl.KeyProof, key string) Op {
	switch proof := proof.(type) {
	case *iavl.KeyExistsProof:
		return FromIAVLExists(proof, key)
	case *iavl.KeyAbsentProof:
		return FromIAVLAbsent(proof, key)
	default:
		return nil
	}
}

func FromIAVLExists(proof *iavl.KeyExistsProof, key string) Op {
	return IAVLExistsOp{proof, key}
}

func FromIAVLAbsent(proof *iavl.KeyAbsentProof, key string) Op {
	return IAVLAbsentOp{key} // TODO
}

func FromSimpleProof(proof *merkle.SimpleProof, index int, total int) Op {
	return SimpleExistsOp{
		SimpleProof: proof,
		Index:       index,
		Total:       total,
	}
}
