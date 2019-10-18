package types

import (
	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ICS 023 Merkle Types Implementation
//
// This file defines Merkle commitment types that implements ICS 023.

const merkleKind = "merkle"

// merkle.Proof implementation of Proof
// Applied on SDK-based IBC implementation
var _ exported.RootI = Root{}

// Root defines a merkle root hash.
// In the Cosmos SDK, the AppHash of a block header becomes the Root.
type Root struct {
	Hash []byte `json:"hash" yaml:"hash"`
}

// NewRoot constructs a new Root
func NewRoot(hash []byte) Root {
	return Root{
		Hash: hash,
	}
}

// CommitmentType implements RootI interface
func (Root) CommitmentType() string {
	return merkleKind
}

// Bytes implements RootI interface
func (r Root) Bytes() []byte {
	return r.Hash
}

var _ exported.PrefixI = Prefix{}

// TODO: applyPrefix()

// Prefix is merkle path prefixed to the key.
// The constructed key from the Path and the key will be append(Path.KeyPath, append(Path.KeyPrefix, key...))
type Prefix struct {
	// KeyPath is the list of keys prepended before the prefixed key
	KeyPath [][]byte `json:"key_path" yaml:"key_path"`
	// KeyPrefix is a byte slice prefixed before the key
	KeyPrefix []byte `json:"key_prefix" yaml:"key_prefix"`
}

// NewPrefix constructs new Prefix instance
func NewPrefix(keypath [][]byte, keyprefix []byte) Prefix {
	return Prefix{
		KeyPath:   keypath,
		KeyPrefix: keyprefix,
	}
}

// CommitmentType implements PrefixI
func (Prefix) CommitmentType() string {
	return merkleKind
}

// Key returns the full commitment prefix key
func (prefix Prefix) Key(key []byte) []byte {
	return join(prefix.KeyPrefix, key)
}

var _ exported.ProofI = Proof{}

// Proof is a wrapper type that contains a merkle proof and the key used to verify .
// It demonstrates membership or non-membership for an element or set of elements,
// verifiable in conjunction with a known commitment root. Proofs should be
// succinct.
type Proof struct {
	Proof *merkle.Proof `json:"proof" yaml:"proof"`
	Key   []byte        `json:"key" yaml:"key"`
}

// CommitmentType implements ProofI
func (Proof) CommitmentType() string {
	return merkleKind
}

// GetKey returns the key of the commitment proof
func (proof Proof) GetKey() []byte {
	return proof.Key
}

// GetRawProof returns the raw merkle proof
func (proof Proof) GetRawProof() *merkle.Proof {
	return proof.Proof
}

// VerifyMembership proves the proof against the given root, path, and value.
func (proof Proof) VerifyMembership(commitmentRoot exported.RootI, commitmentPrefix exported.PrefixI, value []byte) bool {
	root, ok := commitmentRoot.(Root)
	if !ok {
		return false
	}

	path, ok := commitmentPrefix.(Prefix)
	if !ok {
		return false
	}

	keypath := merkle.KeyPath{}
	for _, key := range path.KeyPath {
		keypath = keypath.AppendKey(key, merkle.KeyEncodingHex)
	}
	keypath = keypath.AppendKey(append(path.KeyPrefix, proof.Key...), merkle.KeyEncodingHex)

	runtime := rootmulti.DefaultProofRuntime()

	err := runtime.VerifyValue(proof.Proof, root.Hash, keypath.String(), value)
	if err != nil {
		return false
	}

	return true
}

// VerifyAbsence verifies the absence of a proof against the given root, path.
func (proof Proof) VerifyAbsence(commitmentRoot exported.RootI, commitmentPrefix exported.PrefixI) bool {
	root, ok := commitmentRoot.(Root)
	if !ok {
		return false
	}

	path, ok := commitmentPrefix.(Prefix)
	if !ok {
		return false
	}

	keypath := merkle.KeyPath{}
	for _, key := range path.KeyPath {
		keypath = keypath.AppendKey(key, merkle.KeyEncodingHex)
	}
	keypath = keypath.AppendKey(append(path.KeyPrefix, proof.Key...), merkle.KeyEncodingHex)

	runtime := rootmulti.DefaultProofRuntime()

	err := runtime.VerifyAbsence(proof.Proof, root.Hash, keypath.String())
	if err != nil {
		return false
	}
	return true
}
