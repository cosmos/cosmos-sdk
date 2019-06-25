package commitment

type Root interface {
	CommitmentKind() string
}

type Proof interface {
	CommitmentKind() string
	GetKey() []byte
	Verify(Root, []byte) error
}
