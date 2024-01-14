package keyring

import (
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// Options are used to configure a keyring.
type Options struct{}

type Option func(*Options)

// KeyType reflects a human-readable type for key listing.
type KeyType uint

type Record struct{}

// Keyring is an interface used for signing transactions.
// It aims to be simplistic and easy to use.
type Keyring interface {
	// List returns the names of all keys stored in the keyring.
	List() ([]string, error)

	// LookupAddressByKeyName returns the address of the key with the given name.
	LookupAddressByKeyName(name string) ([]byte, error)

	// GetPubKey returns the public key of the key with the given name.
	GetPubKey(name string) (cryptotypes.PubKey, error)

	// Key and KeyByAddress return keys by uid and address respectively.
	Key(uid string) (*Record, error)
	KeyByAddress(address []byte) (*Record, error)

	GetRecordAddress(record *Record) ([]byte, error)
	GetRecordName(record *Record) string
	GetRecordType(record *Record) KeyType

	// Sign signs the given bytes with the key with the given name.
	Sign(name string, msg []byte, signMode signingv1beta1.SignMode) ([]byte, error)
}
