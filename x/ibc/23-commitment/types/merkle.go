package types

import (
	"net/url"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"

	"github.com/tendermint/tendermint/crypto/merkle"
)

// ICS 023 Merkle Types Implementation
//
// This file defines Merkle commitment types that implements ICS 023.

// Merkle proof implementation of the Proof interface
// Applied on SDK-based IBC implementation
var _ exported.Root = (*MerkleRoot)(nil)

// NewMerkleRoot constructs a new MerkleRoot
func NewMerkleRoot(hash []byte) MerkleRoot {
	return MerkleRoot{
		Hash: hash,
	}
}

// GetHash implements RootI interface
func (mr MerkleRoot) GetHash() []byte {
	return mr.Hash
}

// GetCommitmentType implements RootI interface
func (MerkleRoot) GetCommitmentType() exported.Type {
	return exported.Merkle
}

// Empty returns true if the root is empty
func (mr MerkleRoot) Empty() bool {
	return len(mr.GetHash()) == 0
}

var _ exported.Prefix = (*MerklePrefix)(nil)

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

// Empty returns true if the prefix is empty
func (mp MerklePrefix) Empty() bool {
	return len(mp.Bytes()) == 0
}

var _ exported.Path = (*MerklePath)(nil)

// NewMerklePath creates a new MerklePath instance
func NewMerklePath(keyPathStr []string) MerklePath {
	merkleKeyPath := KeyPath{}
	for _, keyStr := range keyPathStr {
		merkleKeyPath = merkleKeyPath.AppendKey([]byte(keyStr), URL)
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

// Empty returns true if the path is empty
func (mp MerklePath) Empty() bool {
	return len(mp.KeyPath.Keys) == 0
}

// ApplyPrefix constructs a new commitment path from the arguments. It interprets
// the path argument in the context of the prefix argument.
//
// CONTRACT: provided path string MUST be a well formated path. See ICS24 for
// reference.
func ApplyPrefix(prefix exported.Prefix, path string) (MerklePath, error) {
	err := host.PathValidator(path)
	if err != nil {
		return MerklePath{}, err
	}

	if prefix == nil || prefix.Empty() {
		return MerklePath{}, sdkerrors.Wrap(ErrInvalidPrefix, "prefix can't be empty")
	}
	return NewMerklePath([]string{string(prefix.Bytes()), path}), nil
}

var _ exported.Proof = (*MerkleProof)(nil)

// GetCommitmentType implements ProofI
func (MerkleProof) GetCommitmentType() exported.Type {
	return exported.Merkle
}

// VerifyMembership verifies the membership pf a merkle proof against the given root, path, and value.
func (proof MerkleProof) VerifyMembership(root exported.Root, path exported.Path, value []byte) error {
	if proof.Empty() || root == nil || root.Empty() || path == nil || path.Empty() || len(value) == 0 {
		return sdkerrors.Wrap(ErrInvalidMerkleProof, "empty params or proof")
	}

	runtime := rootmulti.DefaultProofRuntime()
	return runtime.VerifyValue(proof.Proof, root.GetHash(), path.String(), value)
}

// VerifyNonMembership verifies the absence of a merkle proof against the given root and path.
func (proof MerkleProof) VerifyNonMembership(root exported.Root, path exported.Path) error {
	if proof.Empty() || root == nil || root.Empty() || path == nil || path.Empty() {
		return sdkerrors.Wrap(ErrInvalidMerkleProof, "empty params or proof")
	}

	runtime := rootmulti.DefaultProofRuntime()
	return runtime.VerifyAbsence(proof.Proof, root.GetHash(), path.String())
}

// Empty returns true if the root is empty
func (proof MerkleProof) Empty() bool {
	return proof.Proof.Equal(nil) || proof.Equal(MerkleProof{}) || proof.Proof.Equal(nil) || proof.Proof.Equal(merkle.Proof{})
}

// ValidateBasic checks if the proof is empty.
func (proof MerkleProof) ValidateBasic() error {
	if proof.Empty() {
		return ErrInvalidProof
	}
	return nil
}
