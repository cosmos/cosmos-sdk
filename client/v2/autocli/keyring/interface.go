package keyring

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	txsigning "github.com/cosmos/cosmos-sdk/x/tx/signing"
)

// Keyring is an interface used for signing transactions.
// It aims to be simplistic and easy to use.
type Keyring interface {
	// List returns the names of all keys stored in the keyring.
	List() ([]string, error)

	// LookupAddressByKeyName returns the address of the key with the given name.
	LookupAddressByKeyName(name string) ([]byte, error)

	// GetPubKey returns the public key of the key with the given name.
	GetPubKey(name string) (cryptotypes.PubKey, error)

	// Sign signs the given bytes with the key with the given name.
	Sign(name string, msg []byte, signMode txsigning.SignMode) ([]byte, error)
}
