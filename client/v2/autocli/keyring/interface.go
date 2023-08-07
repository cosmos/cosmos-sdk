package keyring

type Keyring interface {
	// LookupKey returns the address of the key with the given name.
	LookupKey(name string) (string, error)
}
