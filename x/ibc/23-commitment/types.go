package commitment

type Root interface {
	CommitmentKind() string
}

type Path interface {
	CommitmentKind() string
}

type Proof interface {
	CommitmentKind() string
	GetKey() []byte
	Verify(Root, Path, []byte) error
}
