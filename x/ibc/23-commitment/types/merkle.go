package types

import (
	"errors"
	"net/url"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// ICS 023 Merkle Types Implementation
//
// This file defines Merkle commitment types that implements ICS 023.

// Merkle proof implementation of the Proof interface
// Applied on SDK-based IBC implementation
var _ exported.Root = MerkleRoot{}

// MerkleRoot defines a merkle root hash.
// In the Cosmos SDK, the AppHash of a block header becomes the root.
type MerkleRoot struct {
	Hash []byte `json:"hash" yaml:"hash"`
}

// NewMerkleRoot constructs a new MerkleRoot
func NewMerkleRoot(hash []byte) MerkleRoot {
	return MerkleRoot{
		Hash: hash,
	}
}

// GetCommitmentType implements RootI interface
func (MerkleRoot) GetCommitmentType() exported.Type {
	return exported.Merkle
}

// GetHash implements RootI interface
func (mr MerkleRoot) GetHash() []byte {
	return mr.Hash
}

// IsEmpty returns true if the root is empty
func (mr MerkleRoot) IsEmpty() bool {
	return len(mr.GetHash()) == 0
}

var _ exported.Prefix = MerklePrefix{}

// MerklePrefix is merkle path prefixed to the key.
// The constructed key from the Path and the key will be append(Path.KeyPath, append(Path.KeyPrefix, key...))
type MerklePrefix struct {
	KeyPrefix []byte `json:"key_prefix" yaml:"key_prefix"` // byte slice prefixed before the key
}

// NewMerklePrefix constructs new MerklePrefix instance
func NewMerklePrefix(keyPrefix []byte) MerklePrefix {
	return MerklePrefix{
		KeyPrefix: keyPrefix,
	}
}

// GetCommitmentType implements Prefix interface
func (MerklePrefix) GetCommitmentType() exported.Type {
	return exported.Merkle
}

// Bytes returns the key prefix bytes
func (mp MerklePrefix) Bytes() []byte {
	return mp.KeyPrefix
}

// IsEmpty returns true if the prefix is empty
func (mp MerklePrefix) IsEmpty() bool {
	return len(mp.Bytes()) == 0
}

var _ exported.Path = MerklePath{}

// MerklePath is the path used to verify commitment proofs, which can be an arbitrary
// structured object (defined by a commitment type).
type MerklePath struct {
	KeyPath merkle.KeyPath `json:"key_path" yaml:"key_path"` // byte slice prefixed before the key
}

// NewMerklePath creates a new MerklePath instance
func NewMerklePath(keyPathStr []string) MerklePath {
	merkleKeyPath := merkle.KeyPath{}
	for _, keyStr := range keyPathStr {
		merkleKeyPath = merkleKeyPath.AppendKey([]byte(keyStr), merkle.KeyEncodingURL)
	}

	return MerklePath{
		KeyPath: merkleKeyPath,
	}
}

// GetCommitmentType implements PathI
func (MerklePath) GetCommitmentType() exported.Type {
	return exported.Merkle
}

// String implements fmt.Stringer.
func (mp MerklePath) String() string {
	return mp.KeyPath.String()
}

// Pretty returns the unescaped path of the URL string.
func (mp MerklePath) Pretty() string {
	path, err := url.PathUnescape(mp.KeyPath.String())
	if err != nil {
		panic(err)
	}
	return path
}

// IsEmpty returns true if the path is empty
func (mp MerklePath) IsEmpty() bool {
	return len(mp.KeyPath) == 0
}

// ApplyPrefix constructs a new commitment path from the arguments. It interprets
// the path argument in the context of the prefix argument.
//
// CONTRACT: provided path string MUST be a well formated path. See ICS24 for
// reference.
func ApplyPrefix(prefix exported.Prefix, path string) (MerklePath, error) {
	err := host.DefaultPathValidator(path)
	if err != nil {
		return MerklePath{}, err
	}

	if prefix == nil || prefix.IsEmpty() {
		return MerklePath{}, errors.New("prefix can't be empty")
	}
	return NewMerklePath([]string{string(prefix.Bytes()), path}), nil
}

var _ exported.Proof = MerkleProof{}

// MerkleProof is a wrapper type that contains a merkle proof.
// It demonstrates membership or non-membership for an element or set of elements,
// verifiable in conjunction with a known commitment root. Proofs should be
// succinct.
type MerkleProof struct {
	Proof *merkle.Proof `json:"proof" yaml:"proof"`
}

// GetCommitmentType implements ProofI
func (MerkleProof) GetCommitmentType() exported.Type {
	return exported.Merkle
}

// VerifyMembership verifies the membership pf a merkle proof against the given root, path, and value.
func (proof MerkleProof) VerifyMembership(root exported.Root, path exported.Path, value []byte) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() || path == nil || path.IsEmpty() || len(value) == 0 {
		return errors.New("empty params or proof")
	}

	runtime := rootmulti.DefaultProofRuntime()
	return runtime.VerifyValue(proof.Proof, root.GetHash(), path.String(), value)
}

// VerifyNonMembership verifies the absence of a merkle proof against the given root and path.
func (proof MerkleProof) VerifyNonMembership(root exported.Root, path exported.Path) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() || path == nil || path.IsEmpty() {
		return errors.New("empty params or proof")
	}

	runtime := rootmulti.DefaultProofRuntime()
	return runtime.VerifyAbsence(proof.Proof, root.GetHash(), path.String())
}

// IsEmpty returns true if the root is empty
func (proof MerkleProof) IsEmpty() bool {
	return (proof == MerkleProof{}) || proof.Proof == nil
}

// ValidateBasic checks if the proof is empty.
func (proof MerkleProof) ValidateBasic() error {
	if proof.IsEmpty() {
		return ErrInvalidProof
	}
	return nil
}
