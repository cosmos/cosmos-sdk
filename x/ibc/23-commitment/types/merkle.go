package types

import (
	"net/url"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"

	"github.com/tendermint/tendermint/crypto/merkle"
)

// var representing the proofspecs for a SDK chain
var sdkSpecs = []string{storetypes.ProofOpIAVLCommitment, storetypes.ProofOpSimpleMerkleCommitment}

// ICS 023 Merkle Types Implementation
//
// This file defines Merkle commitment types that implements ICS 023.

// Merkle proof implementation of the Proof interface
// Applied on SDK-based IBC implementation
var _ exported.Root = (*MerkleRoot)(nil)

// GetSDKSpecs is a getter function for the proofspecs of an sdk chain
func GetSDKSpecs() []string {
	return sdkSpecs
}

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

// IsEmpty returns true if the root is empty
func (mr MerkleRoot) IsEmpty() bool {
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

// IsEmpty returns true if the prefix is empty
func (mp MerklePrefix) IsEmpty() bool {
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

// IsEmpty returns true if the path is empty
func (mp MerklePath) IsEmpty() bool {
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

	if prefix == nil || prefix.IsEmpty() {
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
func (proof MerkleProof) VerifyMembership(specs []string, root exported.Root, path exported.Path, value []byte) error {
	if err := proof.validateVerificationArgs(specs, root); err != nil {
		return err
	}

	// VerifyMembership specific argument validation
	if path == nil || path.IsEmpty() {
		return sdkerrors.Wrap(ErrInvalidProof, "empty path")
	}
	if len(value) == 0 {
		return sdkerrors.Wrap(ErrInvalidProof, "empty value in membership proof")
	}

	runtime := rootmulti.DefaultProofRuntime()
	return runtime.VerifyValue(proof.Proof, root.GetHash(), path.String(), value)
}

// VerifyNonMembership verifies the absence of a merkle proof against the given root and path.
func (proof MerkleProof) VerifyNonMembership(specs []string, root exported.Root, path exported.Path) error {
	if err := proof.validateVerificationArgs(specs, root); err != nil {
		return err
	}

	// VerifyNonMembership specific argument validation
	if path == nil || path.IsEmpty() {
		return sdkerrors.Wrap(ErrInvalidProof, "empty path")
	}

	runtime := rootmulti.DefaultProofRuntime()
	return runtime.VerifyAbsence(proof.Proof, root.GetHash(), path.String())
}

// BatchVerifyMembership verifies a group of key value pairs against the given root
// NOTE: Untested
func (proof MerkleProof) BatchVerifyMembership(specs []string, root exported.Root, items map[string][]byte) error {
	if err := proof.validateVerificationArgs(specs, root); err != nil {
		return err
	}

	// Verify each item separately against same proof
	runtime := rootmulti.DefaultProofRuntime()
	for path, value := range items {
		// check value is not empty
		if len(value) == 0 {
			return sdkerrors.Wrap(ErrInvalidProof, "empty value in batched membership proof")
		}

		if err := runtime.VerifyValue(proof.Proof, root.GetHash(), path, value); err != nil {
			return sdkerrors.Wrapf(ErrInvalidProof, "verification failed for path: %s, value: %x. Error: %v",
				path, value, err)
		}
	}
	return nil
}

// BatchVerifyNonMembership verifies absence of a group of keys against the given root
// NOTE: Untested
func (proof MerkleProof) BatchVerifyNonMembership(specs []string, root exported.Root, items []string) error {
	if err := proof.validateVerificationArgs(specs, root); err != nil {
		return err
	}

	// Verify each item separately against same proof
	runtime := rootmulti.DefaultProofRuntime()
	for _, path := range items {
		if err := runtime.VerifyAbsence(proof.Proof, root.GetHash(), path); err != nil {
			return sdkerrors.Wrapf(ErrInvalidProof, "absence verification failed for path: %s. Error: %v",
				path, err)
		}
	}
	return nil
}

// IsEmpty returns true if the root is empty
func (proof MerkleProof) IsEmpty() bool {
	return proof.Proof.Equal(nil) || proof.Equal(MerkleProof{}) || proof.Proof.Equal(nil) || proof.Proof.Equal(merkle.Proof{})
}

// ValidateBasic checks if the proof is empty.
func (proof MerkleProof) ValidateBasic() error {
	if proof.IsEmpty() {
		return ErrInvalidProof
	}
	return nil
}

// validateVerificationArgs verifies the proof arguments are valid
func (proof MerkleProof) validateVerificationArgs(specs []string, root exported.Root) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() {
		return sdkerrors.Wrap(ErrInvalidMerkleProof, "empty params or proof")
	}

	if len(specs) != len(proof.Proof.Ops) {
		return sdkerrors.Wrapf(ErrInvalidMerkleProof,
			"length of specs: %d not equal to length of proof: %d",
			len(specs), len(proof.Proof.Ops))
	}

	for i, op := range proof.Proof.Ops {
		if op.Type != specs[i] {
			return sdkerrors.Wrapf(ErrInvalidMerkleProof,
				"proof does not match expected format at position %d, expected: %s, got: %s",
				i, specs[i], op.Type)
		}
	}
	return nil
}
