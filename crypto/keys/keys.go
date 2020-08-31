package keys

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
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

// ProtoPubKeyToAminoPubKey converts the proto.Message into a crypto.PubKey
func ProtoPubKeyToAminoPubKey(msg proto.Message) (crypto.PubKey, error) {
	switch msg := msg.(type) {
	case *Secp256K1PubKey:
		return msg.GetKey(), nil
	case *MultisigThresholdPubKey:
		keys := make([]crypto.PubKey, len(msg.PubKeys))
		for i, any := range msg.PubKeys {
			k, ok := any.GetCachedValue().(crypto.PubKey)
			if !ok {
				return nil, fmt.Errorf("expected crypto.PubKey")
			}
			keys[i] = k
		}
		return multisig.PubKeyMultisigThreshold{K: uint(msg.K), PubKeys: keys}, nil
	default:
		return nil, fmt.Errorf("unknown proto public key type: %v", msg)
	}
}

// AminoPubKeyToProtoPubKey converts the crypto.PubKey into a proto.Message
func AminoPubKeyToProtoPubKey(key crypto.PubKey) (proto.Message, error) {
	switch key := key.(type) {
	case secp256k1.PubKey:
		return &Secp256K1PubKey{Key: key}, nil
	case multisig.PubKeyMultisigThreshold:
		keys := make([]*types.Any, len(key.PubKeys))
		for i, k := range key.PubKeys {
			msg, err := AminoPubKeyToProtoPubKey(k)
			if err != nil {
				return nil, err
			}
			any, err := types.NewAnyWithValue(msg)
			if err != nil {
				return nil, err
			}
			keys[i] = any
		}
		return &MultisigThresholdPubKey{K: uint32(key.K), PubKeys: keys}, nil
	default:
		return nil, fmt.Errorf("unknown public key type: %v", key)
	}
}
