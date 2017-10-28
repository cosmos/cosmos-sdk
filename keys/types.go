package keys

import (
	"fmt"
	"sort"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	data "github.com/tendermint/go-wire/data"
)

// Storage has many implementation, based on security and sharing requirements
// like disk-backed, mem-backed, vault, db, etc.
type Storage interface {
	Put(name string, key []byte, info Info) error
	Get(name string) (key []byte, info Info, err error)
	List() (Infos, error)
	Delete(name string) error
}

// Info is the public information about a key
type Info struct {
	Name    string        `json:"name"`
	Address data.Bytes    `json:"address"`
	PubKey  crypto.PubKey `json:"pubkey"`
}

func (i *Info) Format() Info {
	if !i.PubKey.Empty() {
		i.Address = i.PubKey.Address()
	}
	return *i
}

// Infos is a wrapper to allows alphabetical sorting of the keys
type Infos []Info

func (k Infos) Len() int           { return len(k) }
func (k Infos) Less(i, j int) bool { return k[i].Name < k[j].Name }
func (k Infos) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }
func (k Infos) Sort() {
	if k != nil {
		sort.Sort(k)
	}
}

// Signable represents any transaction we wish to send to tendermint core
// These methods allow us to sign arbitrary Tx with the KeyStore
type Signable interface {
	// SignBytes is the immutable data, which needs to be signed
	SignBytes() []byte

	// Sign will add a signature and pubkey.
	//
	// Depending on the Signable, one may be able to call this multiple times for multisig
	// Returns error if called with invalid data or too many times
	Sign(pubkey crypto.PubKey, sig crypto.Signature) error

	// Signers will return the public key(s) that signed if the signature
	// is valid, or an error if there is any issue with the signature,
	// including if there are no signatures
	Signers() ([]crypto.PubKey, error)

	// TxBytes returns the transaction data as well as all signatures
	// It should return an error if Sign was never called
	TxBytes() ([]byte, error)
}

// Signer allows one to use a keystore to sign transactions
type Signer interface {
	Sign(name, passphrase string, tx Signable) error
}

// Manager allows simple CRUD on a keystore, as an aid to signing
type Manager interface {
	Signer
	// Create also returns a seed phrase for cold-storage
	Create(name, passphrase, algo string) (Info, string, error)
	// Recover takes a seedphrase and loads in the private key
	Recover(name, passphrase, seedphrase string) (Info, error)
	List() (Infos, error)
	Get(name string) (Info, error)
	Update(name, oldpass, newpass string) error
	Delete(name, passphrase string) error
}

/**** MockSignable allows us to view data ***/

// MockSignable lets us wrap arbitrary data with a go-crypto signature
type MockSignable struct {
	Data      []byte
	PubKey    crypto.PubKey
	Signature crypto.Signature
}

var _ Signable = &MockSignable{}

// NewMockSignable sets the data to sign
func NewMockSignable(data []byte) *MockSignable {
	return &MockSignable{Data: data}
}

// TxBytes returns the full data with signatures
func (s *MockSignable) TxBytes() ([]byte, error) {
	return wire.BinaryBytes(s), nil
}

// SignBytes returns the original data passed into `NewSig`
func (s *MockSignable) SignBytes() []byte {
	return s.Data
}

// Sign will add a signature and pubkey.
//
// Depending on the Signable, one may be able to call this multiple times for multisig
// Returns error if called with invalid data or too many times
func (s *MockSignable) Sign(pubkey crypto.PubKey, sig crypto.Signature) error {
	s.PubKey = pubkey
	s.Signature = sig
	return nil
}

// Signers will return the public key(s) that signed if the signature
// is valid, or an error if there is any issue with the signature,
// including if there are no signatures
func (s *MockSignable) Signers() ([]crypto.PubKey, error) {
	if s.PubKey.Empty() {
		return nil, fmt.Errorf("no signers")
	}
	if !s.PubKey.VerifyBytes(s.SignBytes(), s.Signature) {
		return nil, fmt.Errorf("invalid signature")
	}
	return []crypto.PubKey{s.PubKey}, nil
}
