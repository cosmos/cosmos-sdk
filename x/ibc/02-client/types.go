package client

import (
	"bytes"
)

type ValidityPredicateBase interface {
	ClientKind() Kind
	Height() int64
	Marshal() []byte
	Unmarshal([]byte) error
}

// ConsensusState
type Client interface {
	ClientKind() Kind
	Base() ValidityPredicateBase
	//	Root() CommitmentRoot
}

func Equal(client1, client2 Client) bool {
	return client1.ClientKind() == client2.ClientKind() &&
		bytes.Equal(client1.Base().Marshal(), client2.Base().Marshal())
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
