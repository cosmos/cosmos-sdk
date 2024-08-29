package base

import (
	"fmt"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
)

// this file implements a general mechanism to plugin public keys to a baseaccount

// PubKey defines a generic pubkey.
type PubKey interface {
	gogoproto.Message
	VerifySignature(msg []byte, sig []byte) bool
}

type PubKeyG[T any] interface {
	*T
	PubKey
}

type pubKeyImpl struct {
	decode   func(b []byte) (PubKey, error)
	validate func(key PubKey) error
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
