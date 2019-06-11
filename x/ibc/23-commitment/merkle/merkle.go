package merkle

import (
	"errors"

	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// merkle.Proof implementation of Proof
// Applied on SDK-based IBC implementation

var _ commitment.Root = Root{}

type Root = []byte

var _ commitment.Proof = Proof{}

type Proof struct {
	Proof *merkle.Proof
	Key   []byte
}

func (proof Proof) GetKey() []byte {
	return proof.Key
}

func (proof Proof) Verify(croot commitment.Root, value []byte) error {
	root, ok := croot.(Root)
	if !ok {
		return errors.New("invalid commitment root type")
	}

	keypath, err := PrefixKeyPath(SDKPrefix().String(), proof.Key)
	if err != nil {
		return err
	}
	// Hard coded for now
	runtime := merkle.DefaultProofRuntime()
	runtime.RegisterOpDecoder(iavl.ProofOpIAVLAbsence, iavl.IAVLAbsenceOpDecoder)
	runtime.RegisterOpDecoder(iavl.ProofOpIAVLValue, iavl.IAVLValueOpDecoder)

	if value != nil {
		return runtime.VerifyValue(proof.Proof, root, keypath.String(), value)
	}
	return runtime.VerifyAbsence(proof.Proof, root, keypath.String())
}

// Hard coded for now
func SDKPrefix() merkle.KeyPath {
	return new(merkle.KeyPath).
		AppendKey([]byte("ibc"), merkle.KeyEncodingHex).
		AppendKey([]byte{0x00}, merkle.KeyEncodingHex)
}

func PrefixKeyPath(prefix string, key []byte) (res merkle.KeyPath, err error) {
	keys, err := merkle.KeyPathToKeys(prefix)
	if err != nil {
		return
	}

	keys[len(keys)-1] = append(keys[len(keys)], key...)

	for _, key := range keys {
		res = res.AppendKey(key, merkle.KeyEncodingHex)
	}

	return
}
