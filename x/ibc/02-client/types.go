package client

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// TODO: types in this file should be (de/)serialized with proto in the future

type AminoMarshaler interface {
	MarshalAmino() (string, error)
	UnmarshalAmino(string) error
}

type ValidityPredicateBase interface {
	Kind() Kind
	GetHeight() int64
	Equal(ValidityPredicateBase) bool
}

// ConsensusState
type Client interface {
	Kind() Kind
	GetBase() ValidityPredicateBase
	GetRoot() commitment.Root
	Validate(Header) (Client, error) // ValidityPredicate
}

func Equal(client1, client2 Client) bool {
	return client1.Kind() == client2.Kind() &&
		client1.GetBase().Equal(client2.GetBase())
}

type Header interface {
	Kind() Kind
	//	Proof() HeaderProof
	GetBase() ValidityPredicateBase // can be nil
	GetRoot() commitment.Root
}

// XXX: Kind should be enum?

type Kind byte

const (
	Tendermint Kind = iota
)
