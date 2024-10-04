package keyring

import (
	"context"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"

	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// KeyringContextKey is the key used to store the keyring in the context.
// The keyring must be wrapped using the KeyringImpl.
var KeyringContextKey keyringContextKey

type keyringContextKey struct{}

var _ Keyring = &KeyringImpl{}

type KeyringImpl struct {
	k Keyring
}

// NewKeyringInContext returns a new context with the keyring set.
func NewKeyringInContext(ctx context.Context, k Keyring) context.Context {
	return context.WithValue(ctx, KeyringContextKey, NewKeyringImpl(k))
}

func NewKeyringImpl(k Keyring) *KeyringImpl {
	return &KeyringImpl{k: k}
}

// GetPubKey implements Keyring.
func (k *KeyringImpl) GetPubKey(name string) (types.PubKey, error) {
	return k.k.GetPubKey(name)
}

// List implements Keyring.
func (k *KeyringImpl) List() ([]string, error) {
	return k.k.List()
}

// LookupAddressByKeyName implements Keyring.
func (k *KeyringImpl) LookupAddressByKeyName(name string) ([]byte, error) {
	return k.k.LookupAddressByKeyName(name)
}

// Sign implements Keyring.
func (k *KeyringImpl) Sign(name string, msg []byte, signMode signingv1beta1.SignMode) ([]byte, error) {
	return k.k.Sign(name, msg, signMode)
}

// KeyType returns the type of the key.
func (k *KeyringImpl) KeyType(name string) (uint, error) {
	return k.k.KeyType(name)
}

// KeyInfo given a key name or address returns key name, key address and key type.
func (k *KeyringImpl) KeyInfo(nameOrAddr string) (string, string, uint, error) {
	return k.k.KeyInfo(nameOrAddr)
}
