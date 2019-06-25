package merkle

import (
	"errors"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// merkle.Proof implementation of Proof
// Applied on SDK-based IBC implementation

var _ commitment.Root = Root{}

type Root struct {
	Hash      []byte
	KeyPrefix [][]byte
}

func NewRoot(hash []byte, prefixes [][]byte) Root {
	return Root{
		Hash:      hash,
		KeyPrefix: prefixes,
	}
}

func (Root) CommitmentKind() string {
	return "merkle"
}

func (r Root) Update(hash []byte) commitment.Root {
	return Root{
		Hash:      hash,
		KeyPrefix: r.KeyPrefix,
	}
}

var _ commitment.Proof = Proof{}

type Proof struct {
	Proof *merkle.Proof
	Key   []byte
}

func (Proof) CommitmentKind() string {
	return "merkle"
}

func (proof Proof) GetKey() []byte {
	return proof.Key
}

func (proof Proof) Verify(croot commitment.Root, value []byte) error {
	root, ok := croot.(Root)
	if !ok {
		return errors.New("invalid commitment root type")
	}

	keypath := merkle.KeyPath{}

	for _, key := range root.KeyPrefix {
		keypath = keypath.AppendKey(key, merkle.KeyEncodingHex)
	}

	keypath, err := PrefixKeyPath(keypath.String(), proof.Key)
	if err != nil {
		return err
	}

	// Hard coded for now
	runtime := rootmulti.DefaultProofRuntime()

	if value != nil {
		return runtime.VerifyValue(proof.Proof, root.Hash, keypath.String(), value)
	}
	return runtime.VerifyAbsence(proof.Proof, root.Hash, keypath.String())
}

func PrefixKeyPath(prefix string, key []byte) (res merkle.KeyPath, err error) {
	keys, err := merkle.KeyPathToKeys(prefix)
	if err != nil {
		return
	}

	keys[len(keys)-1] = append(keys[len(keys)-1], key...)

	for _, key := range keys {
		res = res.AppendKey(key, merkle.KeyEncodingHex)
	}

	return
}
