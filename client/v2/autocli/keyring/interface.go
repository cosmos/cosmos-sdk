package keyring

type Keyring interface {
	// LookupAddressByKeyName returns the address of the key with the given name.
	LookupAddressByKeyName(name string) ([]byte, error)
}
