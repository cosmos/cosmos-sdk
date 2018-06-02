package merkle

import (
	"bytes"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/ripemd160"

	"github.com/tendermint/iavl"

	"github.com/tendermint/tmlibs/merkle"

	"github.com/tendermint/go-wire"
)

type IAVLExistsOp struct {
	*iavl.KeyExistsProof `json:"key-exists-proof"`
	Key                  string `json:"key"`
}

func (op IAVLExistsOp) String() string {
	return "{ KeyExistsProof: " + op.KeyExistsProof.String() + " Key: " + op.Key + " }"
}

var _ Op = IAVLExistsOp{}

func (op IAVLExistsOp) iavlLeafNodeHash(value []byte) []byte {
	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	n, err := int(0), error(nil)

	wire.WriteInt8(0, buf, &n, &err)
	wire.WriteInt64(1, buf, &n, &err)
	wire.WriteInt64(op.Version, buf, &n, &err)
	wire.WriteByteSlice([]byte(op.Key), buf, &n, &err)
	wire.WriteByteSlice(value, buf, &n, &err)

	if err != nil {
		panic(fmt.Sprintf("Failed to hash proofLeafNode: %v", err))
	}
	hasher.Write(buf.Bytes())

	return hasher.Sum(nil)
}

func (op IAVLExistsOp) Run(value [][]byte) ([][]byte, error) {
	if len(value) != 1 {
		return nil, fmt.Errorf("Value size is not 1")
	}

	hash := op.iavlLeafNodeHash(value[0])
	for _, branch := range op.PathToKey.InnerNodes {
		hash = branch.Hash(hash)
	}

	return [][]byte{hash}, nil
}

func (op IAVLExistsOp) GetKey() string {
	return op.Key
}

func (op IAVLExistsOp) Raw() (res RawOp, err error) {
	bz, err := json.Marshal(op)
	if err != nil {
		return
	}

	return RawOp{
		Type: IAVLExistsOpType,
		Data: string(bz), // TODO
		Key:  op.Key,
	}, nil
}

type IAVLAbsentOp struct {
	Key string `json:"key"`
}

func (op IAVLAbsentOp) Run(value [][]byte) ([][]byte, error) {
	// TODO
	return nil, nil
}

func (op IAVLAbsentOp) GetKey() string {
	return op.Key
}

func (op IAVLAbsentOp) Raw() (res RawOp, err error) {
	bz, err := json.Marshal(op)
	if err != nil {
		return
	}

	return RawOp{
		Type: IAVLAbsentOpType,
		Data: string(bz), // TODO
		Key:  op.Key,
	}, nil
}

type SimpleExistsOp struct {
	*merkle.SimpleProof `json:"simple-proof"`
	Index               int `json:"index"`
	Total               int `json:"total"`
}

var _ Op = SimpleExistsOp{}

// copied from https://github.com/tendermint/tmlibs/blob/master/merkle/simple_proof.go#L61
func computeHashFromAunts(index int, total int, leafHash []byte, innerHashes [][]byte) []byte {
	if index >= total || index < 0 || total <= 0 {
		return nil
	}
	switch total {
	case 0:
		panic("Cannot call computeHashFromAunts() with 0 total")
	case 1:
		if len(innerHashes) != 0 {
			return nil
		}
		return leafHash
	default:
		if len(innerHashes) == 0 {
			return nil
		}
		numLeft := (total + 1) / 2
		if index < numLeft {
			leftHash := computeHashFromAunts(index, numLeft, leafHash, innerHashes[:len(innerHashes)-1])
			if leftHash == nil {
				return nil
			}
			return merkle.SimpleHashFromTwoHashes(leftHash, innerHashes[len(innerHashes)-1])
		}
		rightHash := computeHashFromAunts(index-numLeft, total-numLeft, leafHash, innerHashes[:len(innerHashes)-1])
		if rightHash == nil {
			return nil
		}
		return merkle.SimpleHashFromTwoHashes(innerHashes[len(innerHashes)-1], rightHash)
	}
}

func (op SimpleExistsOp) Run(value [][]byte) ([][]byte, error) {
	if len(value) != 1 {
		return nil, fmt.Errorf("Value size is not 1")
	}
	hash := value[0]
	return [][]byte{computeHashFromAunts(op.Index, op.Total, hash, op.Aunts)}, nil
}

func (op SimpleExistsOp) GetKey() string {
	return ""
}

func (op SimpleExistsOp) Raw() (res RawOp, err error) {
	bz, err := json.Marshal(op)
	if err != nil {
		return
	}

	return RawOp{
		Type: SimpleExistsOpType,
		Data: string(bz), // TODO
		Key:  "",
	}, nil
}
