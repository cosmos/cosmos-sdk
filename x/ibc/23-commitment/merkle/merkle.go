package merkle

import (
	"bytes"
	"errors"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
//	"github.com/cosmos/cosmos-sdk/store/state"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const merkleKind = "merkle"

// merkle.Proof implementation of Proof
// Applied on SDK-based IBC implementation
var _ commitment.Root = Root{}

// Root is Merkle root hash
type Root struct {
	Hash []byte
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

var _ commitment.Path = Path{}

// Path is merkle path prefixed to the key.
// The constructed key from the Path and the key will be append(Path.KeyPath, append(Path.KeyPrefix, key...))
type Path struct {
	// KeyPath is the list of keys prepended before the prefixed key
	KeyPath [][]byte
	// KeyPrefix is a byte slice prefixed before the key
	KeyPrefix []byte
}

// NewPath() constructs new Path
func NewPath(keypath [][]byte, keyprefix []byte) Path {
	return Path{
		KeyPath:   keypath,
		KeyPrefix: keyprefix,
	}
}

// Implements commitment.Path
func (Path) CommitmentKind() string {
	return merkleKind
}

var _ commitment.Proof = Proof{}

// Proof is Merkle proof with the key information.
type Proof struct {
	Proof *merkle.Proof
	Key   []byte
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
func (proof Proof) Verify(croot commitment.Root, cpath commitment.Path, value []byte) error {
	root, ok := croot.(Root)
	if !ok {
		return errors.New("invalid commitment root type")
	}

	path, ok := cpath.(Path)
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
	// TODO: check HasPrefix
	return Proof{proof, bytes.TrimPrefix(value.KeyBytes(), prefix)}
}
