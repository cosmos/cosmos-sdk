package types

import (
	"bytes"
	"net/url"

	ics23 "github.com/confio/ics23/go"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"

	"github.com/tendermint/tendermint/crypto/merkle"
)

// var representing the proofspecs for a SDK chain
var sdkSpecs = []*ics23.ProofSpec{ics23.IavlSpec, ics23.TendermintSpec}

// ICS 023 Merkle Types Implementation
//
// This file defines Merkle commitment types that implements ICS 023.

// Merkle proof implementation of the Proof interface
// Applied on SDK-based IBC implementation
var _ exported.Root = (*MerkleRoot)(nil)

// GetSDKSpecs is a getter function for the proofspecs of an sdk chain
func GetSDKSpecs() []*ics23.ProofSpec {
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

// VerifyMembership verifies the membership pf a merkle proof against the given root, path, and value.
func (proof MerkleProof) VerifyMembership(specs []*ics23.ProofSpec, root exported.Root, path exported.Path, value []byte) error {
	if err := proof.validateVerificationArgs(specs, root); err != nil {
		return err
	}

	// VerifyMembership specific argument validation
	mpath, ok := path.(MerklePath)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidProof, "path %v is not of type MerkleProof", path)
	}
	if len(mpath.KeyPath.Keys) != len(specs) {
		return sdkerrors.Wrapf(ErrInvalidProof, "path length %d not same as proof %d",
			len(mpath.KeyPath.Keys), len(specs))
	}
	if len(value) == 0 {
		return sdkerrors.Wrap(ErrInvalidProof, "empty value in membership proof")
	}

	// Convert Proof to []CommitmentProof
	proofs, err := convertProofs(proof)
	if err != nil {
		return err
	}

	// Since every proof in chain is a membership proof we can chain from index 0
	if err := verifyChainedMembershipProof(root.GetHash(), specs, proofs, mpath.KeyPath, value, 0); err != nil {
		return err
	}
	return nil
}

// VerifyNonMembership verifies the absence of a merkle proof against the given root and path.
// VerifyNonMembership verifies a chained proof where the absence of a given path is proven
// at the lowest subtree and then each subtree's inclusion is proved up to the final root.
func (proof MerkleProof) VerifyNonMembership(specs []*ics23.ProofSpec, root exported.Root, path exported.Path) error {
	if err := proof.validateVerificationArgs(specs, root); err != nil {
		return err
	}

	// VerifyNonMembership specific argument validation
	mpath, ok := path.(MerklePath)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidProof, "path %v is not of type MerkleProof", path)
	}
	if len(mpath.KeyPath.Keys) != len(specs) {
		return sdkerrors.Wrapf(ErrInvalidProof, "path length %d not same as proof %d",
			len(mpath.KeyPath.Keys), len(specs))
	}

	// Convert Proof to []CommitmentProof
	proofs, err := convertProofs(proof)
	if err != nil {
		return err
	}

	// VerifyNonMembership will verify the absence of key in lowest subtree, and then chain inclusion proofs
	// of all subroots up to final root
	subroot, err := proofs[0].Calculate()
	if err != nil {
		sdkerrors.Wrapf(ErrInvalidProof, "could not calculate root for proof index 0. %v", err)
	}
	key := mpath.KeyPath.GetKey(-1)
	if ok := ics23.VerifyNonMembership(specs[0], subroot, proofs[0], key); !ok {
		return sdkerrors.Wrapf(ErrInvalidProof, "could not verify absence of key %s", string(key))
	}

	// Verify chained membership proof starting from index 1 with value = subroot
	if err := verifyChainedMembershipProof(root.GetHash(), specs, proofs, mpath.KeyPath, subroot, 1); err != nil {
		return err
	}
	return nil
}

// BatchVerifyMembership verifies a group of key value pairs against the given root
// NOTE: All items must be part of a batch proof in the first chained proof, i.e. items must all be part of smallest subtree in the chained proof
// NOTE: The path passed in must be the path from the root to the smallest subtree in the chained proof
// NOTE: Untested
func (proof MerkleProof) BatchVerifyMembership(specs []*ics23.ProofSpec, root exported.Root, path exported.Path, items map[string][]byte) error {
	if err := proof.validateVerificationArgs(specs, root); err != nil {
		return err
	}

	// Convert Proof to []CommitmentProof
	proofs, err := convertProofs(proof)
	if err != nil {
		return err
	}

	// VerifyNonMembership will verify the absence of key in lowest subtree, and then chain inclusion proofs
	// of all subroots up to final root
	subroot, err := proofs[0].Calculate()
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidProof, "could not calculate root for proof index 0: %v", err)
	}
	if ok := ics23.BatchVerifyMembership(specs[0], subroot, proofs[0], items); !ok {
		return sdkerrors.Wrapf(ErrInvalidProof, "could not verify batch items")
	}

	// BatchVerifyMembership specific argument validation
	// Path must only be defined if this is a chained proof
	if len(specs) > 1 {
		mpath, ok := path.(MerklePath)
		if !ok {
			return sdkerrors.Wrapf(ErrInvalidProof, "path %v is not of type MerkleProof", path)
		}
		// path length should be one less than specs, since lowest proof keys are in items
		if len(mpath.KeyPath.Keys) != len(specs)-1 {
			return sdkerrors.Wrapf(ErrInvalidProof, "path length %d not same as proof %d",
				len(mpath.KeyPath.Keys), len(specs))
		}

		// Since BatchedProof path does not include lowest subtree, exclude first proof from specs and proofs and start
		// chaining at index 0
		if err := verifyChainedMembershipProof(root.GetHash(), specs[1:], proofs[1:], mpath.KeyPath, subroot, 0); err != nil {
			return err
		}
	} else if !bytes.Equal(root.GetHash(), subroot) {
		// Since we are not chaining proofs, we must check first subroot equals given root
		return sdkerrors.Wrapf(ErrInvalidProof, "batched proof did not commit to expected root: %X, got: %X", root.GetHash(), subroot)
	}

	return nil
}

// BatchVerifyNonMembership verifies absence of a group of keys against the given root
// NOTE: All items must be part of a batch proof in the first chained proof, i.e. items must all be part of smallest subtree in the chained proof
// NOTE: The path passed in must be the path from the root to the smallest subtree in the chained proof
// NOTE: Untested
func (proof MerkleProof) BatchVerifyNonMembership(specs []*ics23.ProofSpec, root exported.Root, path exported.Path, items [][]byte) error {
	if err := proof.validateVerificationArgs(specs, root); err != nil {
		return err
	}

	// Convert Proof to []CommitmentProof
	proofs, err := convertProofs(proof)
	if err != nil {
		return err
	}

	// VerifyNonMembership will verify the absence of key in lowest subtree, and then chain inclusion proofs
	// of all subroots up to final root
	subroot, err := proofs[0].Calculate()
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidProof, "could not calculate root for proof index 0: %v", err)
	}
	if ok := ics23.BatchVerifyNonMembership(specs[0], subroot, proofs[0], items); !ok {
		return sdkerrors.Wrapf(ErrInvalidProof, "could not verify batch items")
	}

	// BatchVerifyNonMembership specific argument validation
	// Path must only be defined if this is a chained proof
	if len(specs) > 1 {
		mpath, ok := path.(MerklePath)
		if !ok {
			return sdkerrors.Wrapf(ErrInvalidProof, "path %v is not of type MerkleProof", path)
		}
		// path length should be one less than specs, since lowest proof keys are in items
		if len(mpath.KeyPath.Keys) != len(specs)-1 {
			return sdkerrors.Wrapf(ErrInvalidProof, "path length %d not same as proof %d",
				len(mpath.KeyPath.Keys), len(specs))
		}

		// Since BatchedProof path does not include lowest subtree, exclude first proof from specs and proofs and start
		// chaining at index 0
		if err := verifyChainedMembershipProof(root.GetHash(), specs[1:], proofs[1:], mpath.KeyPath, subroot, 0); err != nil {
			return err
		}
	} else if !bytes.Equal(root.GetHash(), subroot) {
		// Since we are not chaining proofs, we must check first subroot equals given root
		return sdkerrors.Wrapf(ErrInvalidProof, "batched proof did not commit to expected root: %X, got: %X", root.GetHash(), subroot)
	}

	return nil
}

// verifyChainedMembershipProof takes a list of proofs and specs and verifies each proof sequentially ensuring that the value is committed to
// by first proof and each subsequent subroot is committed to by the next subroot and checking that the final calculated root is equal to the given roothash.
// The proofs and specs are passed in from lowest subtree to the highest subtree, but the keys are passed in from highest subtree to lowest.
// The index specifies what index to start chaining the membership proofs, this is useful since the lowest proof may not be a membership proof, thus we
// will want to start the membership proof chaining from index 1 with value being the lowest subroot
func verifyChainedMembershipProof(root []byte, specs []*ics23.ProofSpec, proofs []*ics23.CommitmentProof, keys KeyPath, value []byte, index int) error {
	var (
		subroot []byte
		err     error
	)
	// Initialize subroot to value since the proofs list may be empty.
	// This may happen if this call is verifying intermediate proofs after the lowest proof has been executed.
	// In this case, there may be no intermediate proofs to verify and we just check that lowest proof root equals final root
	subroot = value
	for i := index; i < len(proofs); i++ {
		subroot, err = proofs[i].Calculate()
		if err != nil {
			return sdkerrors.Wrapf(ErrInvalidProof, "could not calculate proof root at index %d. %v", i, err)
		}
		// Since keys are passed in from highest to lowest, we must grab their indices in reverse order
		// from the proofs and specs which are lowest to highest
		key := keys.GetKey(-1 * (i + 1))
		if ok := ics23.VerifyMembership(specs[i], subroot, proofs[i], key, value); !ok {
			return sdkerrors.Wrapf(ErrInvalidProof, "chained membership proof failed to verify membership of value: %X in subroot %X at index %d for proof %v",
				value, subroot, i, proofs[i])
		}
		// Set value to subroot so that we verify next proof in chain commits to this subroot
		value = subroot
	}
	// Check that chained proof root equals passed-in root
	if !bytes.Equal(root, subroot) {
		return sdkerrors.Wrapf(ErrInvalidProof, "proof did not commit to expected root: %X, got: %X", root, subroot)
	}
	return nil
}

// convertProofs converts a MerkleProof into []*ics23.CommitmentProof
func convertProofs(mproof MerkleProof) ([]*ics23.CommitmentProof, error) {
	// Unmarshal all proof ops to CommitmentProof
	proofs := make([]*ics23.CommitmentProof, len(mproof.Proof.Ops))
	for i, op := range mproof.Proof.Ops {
		var p ics23.CommitmentProof
		err := p.Unmarshal(op.Data)
		if err != nil {
			return nil, sdkerrors.Wrapf(ErrInvalidMerkleProof, "could not unmarshal proof op into CommitmentProof at index: %d", i)
		}
		proofs[i] = &p
	}
	return proofs, nil
}

// Empty returns true if the root is empty
func (proof MerkleProof) Empty() bool {
	return proof.Equal(nil) || proof.Equal(MerkleProof{}) || proof.Proof.Equal(nil) || proof.Proof.Equal(merkle.Proof{})
}

// ValidateBasic checks if the proof is empty.
func (proof MerkleProof) ValidateBasic() error {
	if proof.Empty() {
		return ErrInvalidProof
	}
	return nil
}

// validateVerificationArgs verifies the proof arguments are valid
func (proof MerkleProof) validateVerificationArgs(specs []*ics23.ProofSpec, root exported.Root) error {
	if proof.Empty() {
		return sdkerrors.Wrap(ErrInvalidMerkleProof, "proof cannot be empty")
	}

	if root == nil || root.Empty() {
		return sdkerrors.Wrap(ErrInvalidMerkleProof, "root cannot be empty")
	}

	if len(specs) != len(proof.Proof.Ops) {
		return sdkerrors.Wrapf(ErrInvalidMerkleProof,
			"length of specs: %d not equal to length of proof: %d",
			len(specs), len(proof.Proof.Ops))
	}

	for i, spec := range specs {
		if spec == nil {
			return sdkerrors.Wrapf(ErrInvalidProof, "spec at position %d is nil", i)
		}
	}
	return nil
}
