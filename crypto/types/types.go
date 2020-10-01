package types

import (
	proto "github.com/gogo/protobuf/proto"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

// PubKey interface extends proto.Message
// and tendermint crypto.PubKey
type PubKey interface {
	proto.Message
	tmcrypto.PubKey
}

// PrivKey interface extends proto.Message
// and tendermint crypto.PrivKey
type PrivKey interface {
	proto.Message
	tmcrypto.PrivKey
}

type (
	Address = tmcrypto.Address
)

// IntoTmPubKey allows our own PubKey types be converted into Tendermint's
// pubkey types.
type IntoTmPubKey interface {
	AsTmPubKey() tmcrypto.PubKey
}
