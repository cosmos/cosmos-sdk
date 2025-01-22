package keyring

import (
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/address"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// autoCLIKeyring represents the keyring interface used by the AutoCLI.
// It purposely does not import the AutoCLI package to avoid circular dependencies.
type autoCLIKeyring interface {
	// List returns the names of all keys stored in the keyring.
	List() ([]string, error)

	// LookupAddressByKeyName returns the address of the key with the given name.
	LookupAddressByKeyName(name string) ([]byte, error)

	// GetPubKey returns the public key of the key with the given name.
	GetPubKey(name string) (cryptotypes.PubKey, error)

	// Sign signs the given bytes with the key with the given name.
	Sign(name string, msg []byte, signMode signingv1beta1.SignMode) ([]byte, error)

	// KeyType returns the type of the key.
	KeyType(name string) (uint, error)

	// KeyInfo given a key name or address returns key name, key address and key type.
	KeyInfo(name string) (string, string, uint, error)
}

// NewAutoCLIKeyring wraps the SDK keyring and makes it compatible with the AutoCLI keyring interfaces.
func NewAutoCLIKeyring(kr Keyring, ac address.Codec) (autoCLIKeyring, error) {
	return &autoCLIKeyringAdapter{kr, ac}, nil
}

type autoCLIKeyringAdapter struct {
	Keyring
	ac address.Codec
}

// List returns the names of all keys stored in the keyring.
func (a *autoCLIKeyringAdapter) List() ([]string, error) {
	list, err := a.Keyring.List()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(list))
	for i, key := range list {
		names[i] = key.Name
	}

	return names, nil
}

// LookupAddressByKeyName returns the address of a key stored in the keyring
func (a *autoCLIKeyringAdapter) LookupAddressByKeyName(name string) ([]byte, error) {
	record, err := a.Keyring.Key(name)
	if err != nil {
		return nil, err
	}

	addr, err := record.GetAddress()
	if err != nil {
		return nil, err
	}

	return addr, nil
}

// GetPubKey returns the public key of the key with the given name.
func (a *autoCLIKeyringAdapter) GetPubKey(name string) (cryptotypes.PubKey, error) {
	record, err := a.Keyring.Key(name)
	if err != nil {
		return nil, err
	}

	return record.GetPubKey()
}

// Sign signs the given bytes with the key with the given name.
func (a *autoCLIKeyringAdapter) Sign(name string, msg []byte, signMode signingv1beta1.SignMode) ([]byte, error) {
	record, err := a.Keyring.Key(name)
	if err != nil {
		return nil, err
	}

	signBytes, _, err := a.Keyring.Sign(record.Name, msg, signMode)
	return signBytes, err
}

// KeyType returns the type of the key with the given name.
func (a *autoCLIKeyringAdapter) KeyType(name string) (uint, error) {
	record, err := a.Keyring.Key(name)
	if err != nil {
		return 0, err
	}

	return uint(record.GetType()), nil
}

// KeyInfo returns key name, key address, and key type given a key name or address.
func (a *autoCLIKeyringAdapter) KeyInfo(nameOrAddr string) (string, string, uint, error) {
	addr, err := a.ac.StringToBytes(nameOrAddr)
	if err != nil {
		// If conversion fails, it's likely a name, not an address
		record, err := a.Keyring.Key(nameOrAddr)
		if err != nil {
			return "", "", 0, err
		}
		addr, err = record.GetAddress()
		if err != nil {
			return "", "", 0, err
		}
		addrStr, err := a.ac.BytesToString(addr)
		if err != nil {
			return "", "", 0, err
		}
		return record.Name, addrStr, uint(record.GetType()), nil
	}

	// If conversion succeeds, it's an address, get the key info by address
	record, err := a.Keyring.KeyByAddress(addr)
	if err != nil {
		return "", "", 0, err
	}

	return record.Name, nameOrAddr, uint(record.GetType()), nil
}
