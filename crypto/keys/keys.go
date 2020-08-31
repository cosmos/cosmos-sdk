package keys

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	proto "github.com/gogo/protobuf/proto"
)

var (
	_ crypto.PubKey  = &Secp256K1PubKey{}
	_ crypto.PrivKey = &Secp256K1PrivKey{}
)

func (m *Secp256K1PubKey) Address() crypto.Address {
	return m.Key.Address()
}

func (m *Secp256K1PubKey) Bytes() []byte {
	return m.Key.Bytes()
}

func (m *Secp256K1PubKey) VerifySignature(msg []byte, sig []byte) bool {
	return m.Key.VerifySignature(msg, sig)
}

func (m *Secp256K1PubKey) Equals(key crypto.PubKey) bool {
	return m.Key.Equals(key)
}

func (m *Secp256K1PubKey) Type() string {
	return m.Key.Type()
}

func (m *Secp256K1PrivKey) Bytes() []byte {
	return m.Key.Bytes()
}

func (m *Secp256K1PrivKey) Sign(msg []byte) ([]byte, error) {
	return m.Key.Sign(msg)
}

func (m *Secp256K1PrivKey) PubKey() crypto.PubKey {
	return &Secp256K1PubKey{Key: m.Key.PubKey().(secp256k1.PubKey)}
}

func (m *Secp256K1PrivKey) Equals(key crypto.PrivKey) bool {
	return m.Key.Equals(key)
}

func (m *Secp256K1PrivKey) Type() string {
	return m.Key.Type()
}

func ProtoPubKeyToAminoPubKey(msg proto.Message) (crypto.PubKey, error) {
	switch msg := msg.(type) {
	case *Secp256K1PubKey:
		return msg.GetKey(), nil
	default:
		return nil, fmt.Errorf("unknown proto public key type: %v", msg)
	}
}

func AminoPubKeyToProtoPubKey(key crypto.PubKey) (proto.Message, error) {
	switch key := key.(type) {
	case secp256k1.PubKey:
		return &Secp256K1PubKey{Key: key}, nil
	default:
		return nil, fmt.Errorf("unknown public key type: %v", key)
	}
}
