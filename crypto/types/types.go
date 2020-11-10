package types

import (
	proto "github.com/gogo/protobuf/proto"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

// PubKey defines a public key and extends proto.Message.
type PubKey interface {
	proto.Message

	Address() Address
	Bytes() []byte
	VerifySignature(msg []byte, sig []byte) bool
	Equals(PubKey) bool
	Type() string
}

// PrivKey defines a private key and extends proto.Message.
type PrivKey interface {
	proto.Message

	Bytes() []byte
	Sign(msg []byte) ([]byte, error)
	PubKey() PubKey
	Equals(PrivKey) bool
	Type() string
}

type (
	Address = tmcrypto.Address
)
