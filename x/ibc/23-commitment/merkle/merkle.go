package merkle

import (
	"bytes"
	"errors"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	//	"github.com/cosmos/cosmos-sdk/store/state"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ICS 023 Merkle Types Implementation
//
// This file defines Merkle commitment types that implements ICS 023.

const merkleKind = "merkle"

// merkle.Proof implementation of Proof
// Applied on SDK-based IBC implementation
var _ commitment.Root = Root{}

// Root is Merkle root hash
// In Cosmos-SDK, the AppHash of the Header becomes Root.
type Root struct {
	Hash []byte `json:"hash"`
}

// NewRoot constructs a new Root
func NewRoot(hash []byte) Root {
	return Root{
		Hash: hash,
	}
}

// Implements commitment.Root
func (Root) CommitmentKind() string {
	return merkleKind
}

var _ commitment.Prefix = Prefix{}

// Prefix is merkle path prefixed to the key.
// The constructed key from the Path and the key will be append(Path.KeyPath, append(Path.KeyPrefix, key...))
type Prefix struct {
	// KeyPath is the list of keys prepended before the prefixed key
	KeyPath [][]byte `json:"key_path"`
	// KeyPrefix is a byte slice prefixed before the key
	KeyPrefix []byte `json:"key_prefix"`
}

// NewPrefix constructs new Prefix instance
func NewPrefix(keypath [][]byte, keyprefix []byte) Prefix {
	return Prefix{
		KeyPath:   keypath,
		KeyPrefix: keyprefix,
	}
}

// Implements commitment.Prefix
func (Prefix) CommitmentKind() string {
	return merkleKind
}

func (prefix Prefix) Key(key []byte) []byte {
	return join(prefix.KeyPrefix, key)
}

var _ commitment.Proof = Proof{}

// Proof is Merkle proof with the key information.
type Proof struct {
	Proof *merkle.Proof `json:"proof"`
	Key   []byte        `json:"key"`
}

// Implements commitment.Proof
func (Proof) CommitmentKind() string {
	return merkleKind
}

// Returns the key of the proof
func (proof Proof) GetKey() []byte {
	return proof.Key
}

// Verify() proves the proof against the given root, path, and value.
func (proof Proof) Verify(croot commitment.Root, cpath commitment.Prefix, value []byte) error {
	root, ok := croot.(Root)
	if !ok {
		return errors.New("invalid commitment root type")
	}

	path, ok := cpath.(Prefix)
	if !ok {
		return errors.New("invalid commitment path type")
	}

	keypath := merkle.KeyPath{}
	for _, key := range path.KeyPath {
		keypath = keypath.AppendKey(key, merkle.KeyEncodingHex)
	}
	keypath = keypath.AppendKey(append(path.KeyPrefix, proof.Key...), merkle.KeyEncodingHex)

	// TODO: hard coded for now, should be extensible
	runtime := rootmulti.DefaultProofRuntime()

	if value != nil {
		return runtime.VerifyValue(proof.Proof, root.Hash, keypath.String(), value)
	}
	return runtime.VerifyAbsence(proof.Proof, root.Hash, keypath.String())
}

type Value interface {
	KeyBytes() []byte
}

func NewProofFromValue(proof *merkle.Proof, prefix []byte, value Value) Proof {
	return Proof{proof, bytes.TrimPrefix(value.KeyBytes(), prefix)}
}
