package keys

import (
	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

// Keybase allows simple CRUD on a keystore, as an aid to signing
type Keybase interface {
	// Sign some bytes
	Sign(name, passphrase string, msg []byte) (crypto.Signature, crypto.PubKey, error)
	// Create a new keypair
	Create(name, passphrase string, algo CryptoAlgo) (info Info, seed string, err error)
	// Recover takes a seedphrase and loads in the key
	Recover(name, passphrase, seedphrase string) (info Info, erro error)
	List() ([]Info, error)
	Get(name string) (Info, error)
	Update(name, oldpass, newpass string) error
	Delete(name, passphrase string) error

	Import(name string, armor string) (err error)
	Export(name string) (armor string, err error)
}

// Info is the public information about a key
type Info struct {
	Name         string        `json:"name"`
	PubKey       crypto.PubKey `json:"pubkey"`
	PrivKeyArmor string        `json:"privkey.armor"`
}

func newInfo(name string, pub crypto.PubKey, privArmor string) Info {
	return Info{
		Name:         name,
		PubKey:       pub,
		PrivKeyArmor: privArmor,
	}
}

// Address is a helper function to calculate the address from the pubkey
func (i Info) Address() []byte {
	return i.PubKey.Address()
}

func (i Info) bytes() []byte {
	bz, err := wire.MarshalBinary(i)
	if err != nil {
		panic(err)
	}
	return bz
}

func readInfo(bz []byte) (info Info, err error) {
	err = wire.UnmarshalBinary(bz, &info)
	return
}
