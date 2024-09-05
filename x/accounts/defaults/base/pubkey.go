package base

import (
	"fmt"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	dcrd_secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

// this file implements a general mechanism to plugin public keys to a baseaccount

// PubKey defines a generic pubkey.
type PubKey interface {
	transaction.Msg
	VerifySignature(msg, sig []byte) bool
}

type PubKeyG[T any] interface {
	*T
	PubKey
}

type pubKeyImpl struct {
	decode   func(b []byte) (PubKey, error)
	validate func(key PubKey) error
}

func WithSecp256K1PubKey() Option {
	return WithPubKeyWithValidationFunc(func(pt *secp256k1.PubKey) error {
		_, err := dcrd_secp256k1.ParsePubKey(pt.Key)
		return err
	})
}

func WithPubKey[T any, PT PubKeyG[T]]() Option {
	return WithPubKeyWithValidationFunc[T, PT](func(_ PT) error {
		return nil
	})
}

func WithPubKeyWithValidationFunc[T any, PT PubKeyG[T]](validateFn func(PT) error) Option {
	pkImpl := pubKeyImpl{
		decode: func(b []byte) (PubKey, error) {
			key := PT(new(T))
			err := gogoproto.Unmarshal(b, key)
			if err != nil {
				return nil, err
			}
			return key, nil
		},
		validate: func(k PubKey) error {
			concrete, ok := k.(PT)
			if !ok {
				return fmt.Errorf("invalid pubkey type passed for validation, wanted: %T, got: %T", concrete, k)
			}
			return validateFn(concrete)
		},
	}
	return func(a *Account) {
		a.supportedPubKeys[gogoproto.MessageName(PT(new(T)))] = pkImpl
	}
}

func nameFromTypeURL(url string) string {
	name := url
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		name = name[i+len("/"):]
	}
	return name
}
