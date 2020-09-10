package keys

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	proto "github.com/gogo/protobuf/proto"
)

// ProtoPubKeyToAminoPubKey converts the proto.Message into a crypto.PubKey
func ProtoPubKeyToAminoPubKey(msg proto.Message) (crypto.PubKey, error) {
	switch msg := msg.(type) {
	case *Secp256K1PubKey:
		return msg.GetKey(), nil
	case *LegacyAminoMultisigThresholdPubKey:
		keys := make([]crypto.PubKey, len(msg.PubKeys))
		for i, any := range msg.PubKeys {
			cachedKey, ok := any.GetCachedValue().(proto.Message)
			if !ok {
				return nil, fmt.Errorf("can't proto marshal %T", any.GetCachedValue())
			}
			key, err := ProtoPubKeyToAminoPubKey(cachedKey)
			if err != nil {
				return nil, err
			}
			keys[i] = key
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
		return &LegacyAminoMultisigThresholdPubKey{K: uint32(key.K), PubKeys: keys}, nil
	default:
		return nil, fmt.Errorf("unknown public key type: %v", key)
	}
}
