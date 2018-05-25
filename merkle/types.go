package merkle

import (
	"bytes"
	"fmt"
)

type OpType string

const (
	TMCoreOpType = OpType("tmcore_wrapper")

	IAVLExistsOpType   = OpType("iavl_exists_proof")
	IAVLAbsentOpType   = OpType("iavl_absent_proof")
	SimpleExistsOpType = OpType("simple_exists_proof")
)

type RawOp struct {
	Type OpType
	Data [][]byte
	Key  string
}

type OpDecoder func(RawOp) Op

func DefaultLayerDecoder(ro RawOp) Op {
	switch ro.Type {
	/*	case TMCoreOpType:
		return TMCoreOp{}*/
	case IAVLExistsOpType:
		return IAVLExistsOp{}
	case IAVLAbsentOpType:
		//		proof = IAVLAbsentProof{}
		return nil
	case SimpleExistsOpType:
		return SimpleExistsOp{}
	default:
		return nil
	}
}

type Op interface {
	Run([][]byte) ([][]byte, error)
	GetKey() string
	Raw() RawOp
}

type Proof []Op

func (p Proof) Verify(root []byte, value [][]byte, keys ...string) (err error) {
	for i, op := range p {
		key := op.GetKey()
		if key != "" {
			if keys[0] != key {
				return fmt.Errorf("Key mismatch on operation #%d", i)
			}
			keys = keys[1:]
		}
		value, err = op.Run(value)
		if err != nil {
			return
		}
	}
	if !bytes.Equal(root, value[0]) {
		return fmt.Errorf("Calculated root hash is invalid")
	}

	return nil
}

func (p Proof) Bytes() []byte {
	return nil // TODO
}
