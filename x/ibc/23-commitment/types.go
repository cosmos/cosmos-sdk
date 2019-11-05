package commitment

// ICS 023 Types Implementation
//
// This file includes types defined under
// https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments

// spec:Path and spec:Value are defined as bytestring

// RootI implements spec:CommitmentRoot.
// A root is constructed from a set of key-value pairs,
// and the inclusion or non-inclusion of an arbitrary key-value pair
// can be proven with the proof.
type RootI interface {
	GetCommitmentType() Type
	GetHash() []byte
}

// PrefixI implements spec:CommitmentPrefix.
// Prefix represents the common "prefix" that a set of keys shares.
type PrefixI interface {
	GetCommitmentType() Type
	Bytes() []byte
}

// PathI implements spec:CommitmentPath.
// A path is the additional information provided to the verification function.
type PathI interface {
	GetCommitmentType() Type
	String() string
}

// ProofI implements spec:CommitmentProof.
// Proof can prove whether the key-value pair is a part of the Root or not.
// Each proof has designated key-value pair it is able to prove.
// Proofs includes key but value is provided dynamically at the verification time.
type ProofI interface {
	GetCommitmentType() Type
	VerifyMembership(RootI, PathI, []byte) bool
	VerifyNonMembership(RootI, PathI) bool
}

// Type defines the type of the commitment
type Type byte

// Registered commitment types
const (
	Merkle Type = iota + 1 // 1
)

// string representation of the commitment types
const (
	TypeMerkle string = "merkle"
)

// String implements the Stringer interface
func (ct Type) String() string {
	switch ct {
	case Merkle:
		return TypeMerkle

	default:
		return ""
	}
}
