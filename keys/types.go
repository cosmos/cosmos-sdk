package keys

import (
	crypto "github.com/tendermint/go-crypto"
)

// Info is the public information about a key
type Info struct {
	Name   string        `json:"name"`
	PubKey crypto.PubKey `json:"pubkey"`
}

// Keybase allows simple CRUD on a keystore, as an aid to signing
type Keybase interface {
	// Sign some bytes
	Sign(name, passphrase string, msg []byte) (crypto.Signature, crypto.PubKey, error)
	// Create a new keypair
	Create(name, passphrase, algo string) (_ Info, seedphrase string, _ error)
	// Recover takes a seedphrase and loads in the key
	Recover(name, passphrase, seedphrase string) (Info, error)
	List() ([]Info, error)
	Get(name string) (Info, error)
	Update(name, oldpass, newpass string) error
	Delete(name, passphrase string) error
}
