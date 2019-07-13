package commitment

type Root interface {
	CommitmentKind() string
}

type Path interface {
	CommitmentKind() string
	Pathify([]byte) []byte
}

type Proof interface {
	CommitmentKind() string
	GetKey() []byte
	Verify(Root, Path, []byte) error
}
