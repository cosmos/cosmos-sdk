package types

import (
	"strings"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ICS 023 Merkle Types Implementation
//
// This file defines Merkle commitment types that implements ICS 023.

// TODO: use iota
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

// GetHash implements RootI interface
func (r Root) GetHash() []byte {
	return r.Hash
}

var _ exported.PrefixI = Prefix{}

// Prefix is merkle path prefixed to the key.
// The constructed key from the Path and the key will be append(Path.KeyPath, append(Path.KeyPrefix, key...))
type Prefix struct {
	KeyPrefix []byte `json:"key_prefix" yaml:"key_prefix"` // byte slice prefixed before the key
}

// NewPrefix constructs new Prefix instance
func NewPrefix(keyPrefix []byte) Prefix {
	return Prefix{
		KeyPrefix: keyPrefix,
	}
}

// CommitmentType implements PrefixI
func (Prefix) CommitmentType() string {
	return merkleKind
}

// Bytes returns the key prefix bytes
func (p Prefix) Bytes() []byte {
	return p.KeyPrefix
}

var _ exported.PathI = Path{}

// Path is the path used to verify commitment proofs, which can be an arbitrary
// structured object (defined by a commitment type).
type Path struct {
	KeyPath merkle.KeyPath `json:"key_path" yaml:"key_path"` // byte slice prefixed before the key
}

// NewPath creates a new CommitmentPath instance
func NewPath(keyPathStr []string) Path {
	merkleKeyPath := merkle.KeyPath{}
	for _, keyStr := range keyPathStr {
		merkleKeyPath = merkleKeyPath.AppendKey([]byte(keyStr), merkle.KeyEncodingHex)
	}

	return Path{
		KeyPath: merkleKeyPath,
	}
}

// CommitmentType implements PathI
func (Path) CommitmentType() string {
	return merkleKind
}

// String implements fmt.Stringer
func (p Path) String() string {
	return p.KeyPath.String()
}

// ApplyPrefix constructs a new commitment path from the arguments. It interprets
// the path argument in the context of the prefix argument.
//
// CONTRACT: provided path string MUST be a well formated path. See ICS24 for
// reference.
func ApplyPrefix(prefix exported.PrefixI, path string) Path {
	// Split paths by the separator
	pathSlice := strings.Split(path, "/")
	commitmentPath := NewPath(pathSlice)

	commitmentPath.KeyPath = commitmentPath.KeyPath.AppendKey(prefix.Bytes(), merkle.KeyEncodingHex)
	return commitmentPath
}

var _ exported.ProofI = Proof{}

// Proof is a wrapper type that contains a merkle proof.
// It demonstrates membership or non-membership for an element or set of elements,
// verifiable in conjunction with a known commitment root. Proofs should be
// succinct.
type Proof struct {
	Proof *merkle.Proof `json:"proof" yaml:"proof"`
}

// CommitmentType implements ProofI
func (Proof) CommitmentType() string {
	return merkleKind
}

// VerifyMembership verifies the membership pf a merkle proof against the given root, path, and value.
func (proof Proof) VerifyMembership(root exported.RootI, path exported.PathI, value []byte) bool {
	runtime := rootmulti.DefaultProofRuntime()
	err := runtime.VerifyValue(proof.Proof, root.GetHash(), path.String(), value)
	if err != nil {
		return false
	}

	return true
}

// VerifyNonMembership verifies the absence of a merkle proof against the given root and path.
func (proof Proof) VerifyNonMembership(root exported.RootI, path exported.PathI) bool {
	runtime := rootmulti.DefaultProofRuntime()
	err := runtime.VerifyAbsence(proof.Proof, root.GetHash(), path.String())
	if err != nil {
		return false
	}

	return true
}
