package client

type ValidityPredicateBase interface {
	ClientKind() Kind
	Height() int64
}

type Client interface {
	ClientKind() Kind
	Base() ValidityPredicateBase
	//	Root() CommitmentRoot
}

type Header interface {
	ClientKind() Kind
	//	Proof() HeaderProof
	Base() ValidityPredicateBase // can be nil
	// Root() CommitmentRoot
}

type ValidityPredicate func(Client, Header) (Client, error)

type Kind byte

const (
	Tendermint Kind = iota
)
