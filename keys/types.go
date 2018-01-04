package keys

import (
	wire "github.com/tendermint/go-wire"

	crypto "github.com/tendermint/go-crypto"
)

// Info is the public information about a key
type Info struct {
	Name   string        `json:"name"`
	PubKey crypto.PubKey `json:"pubkey"`
}

// Address is a helper function to calculate the address from the pubkey
func (i Info) Address() []byte {
	return i.PubKey.Address()
}

func (i Info) bytes() []byte {
	return wire.BinaryBytes(i)
}

func readInfo(bs []byte) (info Info, err error) {
	err = wire.ReadBinaryBytes(bs, &info)
	return
}

func info(name string, privKey crypto.PrivKey) Info {
	return Info{
		Name:   name,
		PubKey: privKey.PubKey(),
	}
}

// Keybase allows simple CRUD on a keystore, as an aid to signing
type Keybase interface {
	// Sign some bytes
	Sign(name, passphrase string, msg []byte) (crypto.Signature, crypto.PubKey, error)
	// Create a new keypair
	Create(name, passphrase, algo string) (seedphrase string, _ Info, _ error)
	// Recover takes a seedphrase and loads in the key
	Recover(name, passphrase, algo, seedphrase string) (Info, error)
	List() ([]Info, error)
	Get(name string) (Info, error)
	Update(name, oldpass, newpass string) error
	Delete(name, passphrase string) error
}
