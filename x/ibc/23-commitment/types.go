package commitment

// Root is the interface for commitment root.
// A root is constructed from a set of key-value pairs,
// and the inclusion or non-inclusion of an arbitrary key-value pair
// can be proven with the proof.
type Root interface {
	CommitmentKind() string
}

// Path is the additional information provided to the verification function.
// Path represents the common "prefix" that a set of keys shares.
type Path interface {
	CommitmentKind() string
}

// Proof can prove whether the key-value pair is a part of the Root or not.
// Each proof has designated key-value pair it is able to prove.
// Proofs stores key but value is provided dynamically at the verification time.
type Proof interface {
	CommitmentKind() string
	GetKey() []byte
	Verify(Root, Path, []byte) error
}
