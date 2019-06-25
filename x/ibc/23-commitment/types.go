package commitment

type Root interface {
	CommitmentKind() string
	Update([]byte) Root
}

type Proof interface {
	CommitmentKind() string
	GetKey() []byte
	Verify(Root, []byte) error
}
