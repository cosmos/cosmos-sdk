package types

import (
	"net/url"

	ics23 "github.com/confio/ics23/go"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
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
		KeyPaths: []KeyPath{merkleKeyPath},
	}
}

// GetCommitmentType implements PathI
func (MerklePath) GetCommitmentType() exported.Type {
	return exported.Merkle
}

// String implements fmt.Stringer.
func (mp MerklePath) String() string {
	if len(mp.KeyPaths) == 0 {
		return ""
	}
	pathStr := mp.KeyPaths[0].String()
	for i, p := range mp.KeyPaths {
		if i != 0 {
			pathStr += "/" + p.String()
		}
	}
	return pathStr
}

// Pretty returns the unescaped path of the URL string.
func (mp MerklePath) Pretty() string {
	path, err := url.PathUnescape(mp.String())
	if err != nil {
		panic(err)
	}
	return path
}

// IsEmpty returns true if the path is empty
func (mp MerklePath) IsEmpty() bool {
	return len(mp.KeyPaths) == 0
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

// NewMerkleProof constructs a new MerkleProof from ICS23 CommitmentProofs
func NewMerkleProof(proofs ...*ics23.CommitmentProof) MerkleProof {
	return MerkleProof{
		Proofs: proofs,
	}
}

// GetCommitmentType implements ProofI
func (MerkleProof) GetCommitmentType() exported.Type {
	return exported.Merkle
}

// VerifyMembership verifies the membership pf a merkle proof against the given root, path, and value.
func (proof MerkleProof) VerifyMembership(specs []*ics23.ProofSpec, root exported.Root, path exported.Path, value []byte) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() || path == nil || path.IsEmpty() || len(value) == 0 {
		return sdkerrors.Wrap(ErrInvalidProof, "empty params or proof")
	}

	mpath, ok := path.(MerklePath)
	if !ok {
		return sdkerrors.Wrap(ErrInvalidProof, "path is not a merkle path for a merkle proof")
	}

	if len(proof.Proofs) != len(mpath.KeyPaths) || len(proof.Proofs) != len(specs) {
		return sdkerrors.Wrapf(ErrInvalidProof, "invalid chained proof. chained proof length %d, spec length %d, path length %d must all be equal",
			len(proof.Proofs), len(specs), len(mpath.KeyPaths))
	}

	for i, p := range proof.Proofs {
		// While the proofs go from lowest subtree to the final tree, the path is defined from the root
		// down to the leaf. Thus, we must pass in subpaths in reverse order during chained proof verification
		subpath := []byte(mpath.KeyPaths[len(mpath.KeyPaths)-i-1].String())
		existProof, ok := p.Proof.(*ics23.CommitmentProof_Exist)
		if !ok {
			return sdkerrors.Wrap(ErrInvalidProof, "proof is not an existence proof")
		}
		// For subtree verification, we simply calculate the root from the proof and use it to prove
		// against the value
		subroot, err := existProof.Exist.Calculate()
		if err != nil {
			return sdkerrors.Wrap(ErrInvalidProof, err.Error())
		}
		if i != len(proof.Proofs)-1 {
			if !ics23.VerifyMembership(specs[i], subroot, p, subpath, value) {
				return sdkerrors.Wrapf(ErrInvalidProof, "invalid proof for path: %s", path.String())
			}
		} else {
			// For the final verification, we prove inclusion against the root that was passed into function
			// rather than calculating subroot in order to verify that the given value was committed to by
			// the given root
			if !ics23.VerifyMembership(specs[i], root.GetHash(), p, subpath, value) {
				return sdkerrors.Wrapf(ErrInvalidProof, "invalid proof for path: %s", path.String())
			}
		}
		// Each subtree root becomes the proving value for the next proof in the chaining process
		value = subroot
	}
	return nil
}

// VerifyNonMembership verifies the absence of a merkle proof against the given root and path.
// VerifyNonMembership verifies a chained proof where the absence of a given path is proven
// at the lowest subtree and then each subtree's inclusion is proved up to the final root.
func (proof MerkleProof) VerifyNonMembership(specs []*ics23.ProofSpec, root exported.Root, path exported.Path) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() || path == nil || path.IsEmpty() {
		return sdkerrors.Wrap(ErrInvalidProof, "empty params or proof")
	}

	mpath, ok := path.(MerklePath)
	if !ok {
		return sdkerrors.Wrap(ErrInvalidProof, "path is not a merkle path for a merkle proof")
	}

	if len(proof.Proofs) != len(mpath.KeyPaths) || len(proof.Proofs) != len(specs) {
		return sdkerrors.Wrapf(ErrInvalidProof, "invalid chained proof. chained proof length %d, spec length %d, path length %d must all be equal",
			len(proof.Proofs), len(specs), len(mpath.KeyPaths))
	}

	var value, subroot []byte
	var err error
	for i, p := range proof.Proofs {
		// While the proofs go from lowest subtree to the final tree, the path is defined from the root
		// down to the leaf. Thus, we must pass in subpaths in reverse order during chained proof verification
		subpath := []byte(mpath.KeyPaths[len(mpath.KeyPaths)-i-1].String())
		if i == 0 {
			// The first proof, thus the proof for the lowest subtree, is a nonexistence proof.
			// Thus, we calculate the root from proof and then prove nonexistence of the path against this root
			nonexistProof, ok := p.Proof.(*ics23.CommitmentProof_Nonexist)
			if !ok {
				return sdkerrors.Wrap(ErrInvalidProof, "proof is not a nonexistence proof")
			}
			subroot, err = nonexistProof.Nonexist.Left.Calculate()
			if err != nil {
				return sdkerrors.Wrap(ErrInvalidProof, err.Error())
			}

			if !ics23.VerifyNonMembership(specs[i], subroot, p, subpath) {
				return sdkerrors.Wrapf(ErrInvalidProof, "invalid proof for path: %s", path.String())
			}
		} else {
			// Each subsequent proof is a proof of inclusion of the **previous** subtree's root
			if i != len(proof.Proofs)-1 {
				// For intermediate proofs, we calculate the subroot from the proof and prove the previous subtree's
				// root in this higher subroot
				existProof, ok := p.Proof.(*ics23.CommitmentProof_Exist)
				if !ok {
					return sdkerrors.Wrap(ErrInvalidProof, "proof is not an existence proof")
				}
				subroot, err = existProof.Exist.Calculate()
				if err != nil {
					return sdkerrors.Wrap(ErrInvalidProof, err.Error())
				}

				if !ics23.VerifyMembership(specs[i], subroot, p, subpath, value) {
					return sdkerrors.Wrapf(ErrInvalidProof, "invalid proof for path: %s", path.String())
				}
			} else {
				// The final proof in the chain will prove inclusion against the given root.
				if !ics23.VerifyMembership(specs[i], root.GetHash(), p, subpath, value) {
					return sdkerrors.Wrapf(ErrInvalidProof, "invalid proof for path: %s", path.String())
				}
			}
		}
		// Each subtree root becomes the proving value for the next proof in the chaining process
		value = subroot
	}

	return nil

}

// BatchVerifyMembership verifies existence of a batch of items in a subtree for a given root with all items under the given path.
// NOTE: All items must be part of a batch proof in the first chained proof, i.e. items must all be part of smallest subtree in the chained proof
// NOTE: The path passed in must be the path from the root to the smallest subtree in the chained proof
func (proof MerkleProof) BatchVerifyMembership(specs []*ics23.ProofSpec, root exported.Root, path exported.Path, items map[string][]byte) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() || path == nil || path.IsEmpty() {
		return sdkerrors.Wrap(ErrInvalidProof, "empty params or proof")
	}

	mpath, ok := path.(MerklePath)
	if !ok {
		return sdkerrors.Wrap(ErrInvalidProof, "path is not a merkle path for a merkle proof")
	}

	// Path length must be 1 less than proof length since paths for last proof are provided by item keys
	if len(proof.Proofs) != len(mpath.KeyPaths)-1 || len(proof.Proofs) != len(specs) {
		return sdkerrors.Wrapf(ErrInvalidProof, "invalid chained proof. chained proof length %d, spec length %d, path length %d must all be equal",
			len(proof.Proofs), len(specs), len(mpath.KeyPaths)+1)
	}

	var value, subroot []byte
	var err error
	for i, p := range proof.Proofs {
		// First proof in chain must be a batch proof containing existence proofs for all items
		if i == 0 {
			// For subtree verification, we simply calculate the root from the proof and use it to prove
			// against the value
			bproof, ok := p.Proof.(*ics23.CommitmentProof_Batch)
			if ok {
				if len(bproof.Batch.GetEntries()) == 0 || bproof.Batch.GetEntries()[0] == nil {
					return sdkerrors.Wrap(ErrInvalidProof, "batch proof has empty entry")
				}
				bexist := bproof.Batch.GetEntries()[0].GetExist()
				if bexist == nil {
					return sdkerrors.Wrap(ErrInvalidProof, "batch proof is not an existence proof")
				}
				subroot, err = bexist.Calculate()
			} else {
				return sdkerrors.Wrap(ErrInvalidProof, "not a batch proof, compressed batch proof currently unimplemented")
			}

			// Batch verify all items against calculated root of subtree
			if !ics23.BatchVerifyMembership(specs[i], subroot, p, items) {
				return sdkerrors.Wrap(ErrInvalidProof, "batch verification failed")
			}
		} else {
			// While the proofs go from lowest subtree to the final tree, the path is defined from the root
			// down to the leaf. Thus, we must pass in subpaths in reverse order during chained proof verification
			// NOTE: The path passed in to function only goes to the penultimate subtree since the paths for the final tree
			// are the keys in the items map
			subpath := []byte(mpath.KeyPaths[len(mpath.KeyPaths)-i].String())
			if i != len(proof.Proofs)-1 {
				// For intermediate proofs, we calculate the subroot from the proof and prove the previous subtree's
				// root in this higher subroot
				existProof, ok := p.Proof.(*ics23.CommitmentProof_Exist)
				if !ok {
					return sdkerrors.Wrap(ErrInvalidProof, "proof is not an existence proof")
				}
				subroot, err = existProof.Exist.Calculate()
				if err != nil {
					return sdkerrors.Wrap(ErrInvalidProof, err.Error())
				}

				if !ics23.VerifyMembership(specs[i], subroot, p, subpath, value) {
					return sdkerrors.Wrapf(ErrInvalidProof, "invalid proof for path: %s", path.String())
				}
			} else {
				// The final proof in the chain will prove inclusion against the given root.
				if !ics23.VerifyMembership(specs[i], root.GetHash(), p, subpath, value) {
					return sdkerrors.Wrapf(ErrInvalidProof, "invalid batch proof for path: %s", path.String())
				}
			}
		}
		// Each subtree root becomes the proving value for the next proof in the chaining process
		value = subroot
	}
	return nil
}

func (proof MerkleProof) BatchVerifyNonMembership(specs []*ics23.ProofSpec, root exported.Root, path exported.Path, keys [][]byte) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() || path == nil || path.IsEmpty() {
		return sdkerrors.Wrap(ErrInvalidProof, "empty params or proof")
	}

	mpath, ok := path.(MerklePath)
	if !ok {
		return sdkerrors.Wrap(ErrInvalidProof, "path is not a merkle path for a merkle proof")
	}

	// Path length must be 1 less than proof length since paths for last proof are provided by item keys
	if len(proof.Proofs) != len(mpath.KeyPaths) || len(proof.Proofs) != len(specs) {
		return sdkerrors.Wrapf(ErrInvalidProof, "invalid chained proof. chained proof length %d, spec length %d, path length %d must all be equal",
			len(proof.Proofs), len(specs), len(mpath.KeyPaths)+1)
	}

	var value, subroot []byte
	var err error
	for i, p := range proof.Proofs {
		// First proof in chain must be a batch proof containing nonexistence proofs for all items
		if i == 0 {
			// For subtree verification, we simply calculate the root from the proof and use it to prove
			// nonexistence of the key
			bproof, ok := p.Proof.(*ics23.CommitmentProof_Batch)
			if ok {
				if len(bproof.Batch.GetEntries()) == 0 || bproof.Batch.GetEntries()[0] == nil {
					return sdkerrors.Wrap(ErrInvalidProof, "batch proof has empty entry")
				}
				bnonexist := bproof.Batch.GetEntries()[0].GetNonexist()
				if bnonexist == nil {
					return sdkerrors.Wrap(ErrInvalidProof, "batch proof is not a nonexistence proof")
				}
				subroot, err = bnonexist.Left.Calculate()
			} else {
				return sdkerrors.Wrap(ErrInvalidProof, "not a batch proof, compressed batch proof currently unimplemented")
			}

			// Batch verify nonexistence of all items against calculated root of subtree
			if !ics23.BatchVerifyNonMembership(specs[i], subroot, p, keys) {
				return sdkerrors.Wrap(ErrInvalidProof, "batch verification failed")
			}
		} else {
			// While the proofs go from lowest subtree to the final tree, the path is defined from the root
			// down to the leaf. Thus, we must pass in subpaths in reverse order during chained proof verification
			// NOTE: The path passed in to function only goes to the penultimate subtree since the paths for the final tree
			// are the keys in the items map
			subpath := []byte(mpath.KeyPaths[len(mpath.KeyPaths)-i].String())
			if i != len(proof.Proofs)-1 {
				// For intermediate proofs, we calculate the subroot from the proof and prove the previous subtree's
				// root in this higher subroot
				existProof, ok := p.Proof.(*ics23.CommitmentProof_Exist)
				if !ok {
					return sdkerrors.Wrap(ErrInvalidProof, "proof is not an existence proof")
				}
				subroot, err = existProof.Exist.Calculate()
				if err != nil {
					return sdkerrors.Wrap(ErrInvalidProof, err.Error())
				}

				if !ics23.VerifyMembership(specs[i], subroot, p, subpath, value) {
					return sdkerrors.Wrapf(ErrInvalidProof, "invalid proof for path: %s", path.String())
				}
			} else {
				// The final proof in the chain will prove inclusion against the given root.
				if !ics23.VerifyMembership(specs[i], root.GetHash(), p, subpath, value) {
					return sdkerrors.Wrapf(ErrInvalidProof, "invalid batch proof for path: %s", path.String())
				}
			}
		}
		// Each subtree root becomes the proving value for the next proof in the chaining process
		value = subroot
	}
	return nil
}

// IsEmpty returns true if MerkleProof is empty
func (proof MerkleProof) IsEmpty() bool {
	if len(proof.Proofs) == 0 {
		return true
	}
	for _, p := range proof.Proofs {
		if p == nil {
			return true
		}
	}

	return false
}

// ValidateBasic checks if the proof is empty or malformed.
func (proof MerkleProof) ValidateBasic() error {
	if proof.IsEmpty() {
		return ErrInvalidProof
	}
	return nil
}
