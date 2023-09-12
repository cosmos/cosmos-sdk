package keyring

import "github.com/cosmos/cosmos-sdk/crypto/keyring"

type Keyring interface {
	keyring.Keyring

	// LookupAddressByKeyName returns the address of the key with the given name.
	LookupAddressByKeyName(name string) ([]byte, error)
}
