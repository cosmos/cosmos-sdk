package client

type ValidityPredicateBase interface {
	ClientKind() ClientKind
	Height() int64
}

type ConsensusState interface {
	ClientKind() ClientKind
	Base() ValidityPredicateBase
	Root() CommitmentRoot
}

type Header interface {
	ClientKind() ClientKind
	Proof() HeaderProof
	Base() ValidityPredicateBase // can be nil
	Root() CommitmentRoot
}

type ValidityPredicate func(ConsensusState, Header) (ConsensusState, error)

type Client interface {
	ClientKind() ClientKind
	ValidityPredicate() ValidityPredicate
	ConsensusState() ConsensusState
}

type ClientKind byte

const (
	Tendermint ClientKind = iota
)
