package types

import (
	proto "github.com/gogo/protobuf/proto"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

// PubKey defines a public key and
// extends proto.Message.
type PubKey interface {
	proto.Message

	Address() Address
	Bytes() []byte
	VerifySignature(msg []byte, sig []byte) bool
	Equals(PubKey) bool
	Type() string
}

// BasePrivKey defines a private key
type BasePrivKey interface {
	Bytes() []byte
	Sign(msg []byte) ([]byte, error)
	PubKey() PubKey
	Equals(BasePrivKey) bool
	Type() string
}

// PrivKey extends proto.Message and
// BasePrivKey.
type PrivKey interface {
	proto.Message
	BasePrivKey
}

type (
	Address = tmcrypto.Address
)
