package commitment

type Root interface {
	CommitmentKind() string
	Update(RootUpdate) (Root, error)
}

type RootUpdate interface {
	CommitmentKind() string
}

type Proof interface {
	CommitmentKind() string
	GetKey() []byte
	Verify(Root, []byte) error
}
