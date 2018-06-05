package merkle

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/wire"
)

type OpType string

const (
	TMCoreOpType = OpType("tmcore")

	IAVLExistsOpType   = OpType("iavl_exists")
	IAVLAbsentOpType   = OpType("iavl_absent")
	SimpleExistsOpType = OpType("simple_exists")
)

type RawOp struct {
	Type OpType `json:"type"`
	Data string `json:"data"`
	Key  string `json:"key"`
}

func (ro RawOp) Encode() []byte {
	res, err := json.Marshal(ro)
	if err != nil {
		panic(err)
	}
	return res
}

func (ro *RawOp) Decode(bz []byte) error {
	return json.Unmarshal(bz, ro)
}

type OpDecoder func(RawOp) (Op, error)

func DefaultOpDecoder(ro RawOp) (res Op, err error) {
	switch ro.Type {
	case IAVLExistsOpType:
		var op IAVLExistsOp
		err = json.Unmarshal([]byte(ro.Data), &op)
		res = op
	case IAVLAbsentOpType:
		res = IAVLAbsentOp{ro.Key}
	case SimpleExistsOpType:
		var op SimpleExistsOp
		err = json.Unmarshal([]byte(ro.Data), &op)
		res = op
	default:
		err = fmt.Errorf("Cannot decode RawOp typeof %s", ro.Type)
	}
	return
}

type Op interface {
	Run([][]byte) ([][]byte, error)
	GetKey() string
	Raw() (RawOp, error)
}

type Proof []Op

func (p Proof) Verify(root []byte, value [][]byte, keys ...string) (err error) {
	for i, op := range p {
		key := op.GetKey()
		if key != "" {
			if keys[0] != key {
				return fmt.Errorf("Key mismatch on operation #%d: expected %+v but %+v", i, []byte(keys[0]), []byte(key))
			}
			keys = keys[1:]
		}
		value, err = op.Run(value)
		if err != nil {
			return
		}
	}
	if !bytes.Equal(root, value[0]) {
		return fmt.Errorf("Calculated root hash is invalid: expected %+v but %+v\n", root, value[0])
	}

	return nil
}

// TODO: extend encoding scheme to amino and protobuf
func (p Proof) Bytes(cdc *wire.Codec) (res []byte, err error) {
	rawops := make([]RawOp, len(p))
	for i, op := range p {
		rawops[i], err = op.Raw()
		if err != nil {
			return
		}
	}

	return json.Marshal(rawops)

}

func DecodeProof(cdc *wire.Codec, data []byte, decode OpDecoder) (res Proof, err error) {
	rawops := make([]RawOp, 0)
	if err = json.Unmarshal(data, &rawops); err != nil {
		return
	}

	res = make([]Op, len(rawops))
	for i, rawop := range rawops {
		res[i], err = decode(rawop)
		if err != nil {
			return
		}
	}
	return
}
