package commitment

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// ICS 023 Merkle Types Implementation
//
// This file defines Merkle commitment types that implements ICS 023.

// Merkle proof implementation of the Proof interface
// Applied on SDK-based IBC implementation
var _ RootI = Root{}

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

// GetCommitmentType implements RootI interface
func (Root) GetCommitmentType() Type {
	return Merkle
}

// GetHash implements RootI interface
func (r Root) GetHash() []byte {
	return r.Hash
}

// IsEmpty returns true if the root is empty
func (r Root) IsEmpty() bool {
	return len(r.GetHash()) == 0
}

var _ PrefixI = Prefix{}

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

// GetCommitmentType implements PrefixI
func (Prefix) GetCommitmentType() Type {
	return Merkle
}

// Bytes returns the key prefix bytes
func (p Prefix) Bytes() []byte {
	return p.KeyPrefix
}

// IsEmpty returns true if the prefix is empty
func (p Prefix) IsEmpty() bool {
	return len(p.Bytes()) == 0
}

var _ PathI = Path{}

// Path is the path used to verify commitment proofs, which can be an arbitrary
// structured object (defined by a commitment type).
type Path struct {
	KeyPath merkle.KeyPath `json:"key_path" yaml:"key_path"` // byte slice prefixed before the key
}

// NewPath creates a new CommitmentPath instance
func NewPath(keyPathStr []string) Path {
	merkleKeyPath := merkle.KeyPath{}
	for _, keyStr := range keyPathStr {
		merkleKeyPath = merkleKeyPath.AppendKey([]byte(keyStr), merkle.KeyEncodingURL)
	}

	return Path{
		KeyPath: merkleKeyPath,
	}
}

// GetCommitmentType implements PathI
func (Path) GetCommitmentType() Type {
	return Merkle
}

// String implements fmt.Stringer.
func (p Path) String() string {
	return p.KeyPath.String()
}

// Pretty returns the unescaped path of the URL string.
func (p Path) Pretty() string {
	path, err := url.PathUnescape(p.KeyPath.String())
	if err != nil {
		panic(err)
	}
	return path
}

// IsEmpty returns true if the path is empty
func (p Path) IsEmpty() bool {
	return len(p.KeyPath) == 0
}

// ApplyPrefix constructs a new commitment path from the arguments. It interprets
// the path argument in the context of the prefix argument.
//
// CONTRACT: provided path string MUST be a well formated path. See ICS24 for
// reference.
func ApplyPrefix(prefix PrefixI, path string) (Path, error) {
	err := host.DefaultPathValidator(path)
	if err != nil {
		return Path{}, err
	}

	if prefix == nil || prefix.IsEmpty() {
		return Path{}, errors.New("prefix can't be empty")
	}
	return NewPath([]string{string(prefix.Bytes()), path}), nil
}

var _ ProofI = Proof{}

// Proof is a wrapper type that contains a merkle proof.
// It demonstrates membership or non-membership for an element or set of elements,
// verifiable in conjunction with a known commitment root. Proofs should be
// succinct.
type Proof struct {
	Proof *merkle.Proof `json:"proof" yaml:"proof"`
}

// GetCommitmentType implements ProofI
func (Proof) GetCommitmentType() Type {
	return Merkle
}

// VerifyMembership verifies the membership pf a merkle proof against the given root, path, and value.
func (proof Proof) VerifyMembership(root RootI, path PathI, value []byte) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() || path == nil || path.IsEmpty() || len(value) == 0 {
		return errors.New("empty params or proof")
	}

	runtime := rootmulti.DefaultProofRuntime()
	fmt.Println(path.String())
	return runtime.VerifyValue(proof.Proof, root.GetHash(), path.String(), value)
}

// VerifyNonMembership verifies the absence of a merkle proof against the given root and path.
func (proof Proof) VerifyNonMembership(root RootI, path PathI) error {
	if proof.IsEmpty() || root == nil || root.IsEmpty() || path == nil || path.IsEmpty() {
		return errors.New("empty params or proof")
	}

	runtime := rootmulti.DefaultProofRuntime()
	return runtime.VerifyAbsence(proof.Proof, root.GetHash(), path.String())
}

// IsEmpty returns true if the root is empty
func (proof Proof) IsEmpty() bool {
	return (proof == Proof{}) || proof.Proof == nil
}

// ValidateBasic checks if the proof is empty.
func (proof Proof) ValidateBasic() error {
	if proof.IsEmpty() {
		return ErrInvalidProof
	}
	return nil
}
