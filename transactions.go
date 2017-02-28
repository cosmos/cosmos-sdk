package keys

import (
	"sort"

	crypto "github.com/tendermint/go-crypto"
)

// KeyInfo is the public information about a key
type KeyInfo struct {
	Name   string
	PubKey crypto.PubKeyS
}

// KeyInfos is a wrapper to allows alphabetical sorting of the keys
type KeyInfos []KeyInfo

func (k KeyInfos) Len() int           { return len(k) }
func (k KeyInfos) Less(i, j int) bool { return k[i].Name < k[j].Name }
func (k KeyInfos) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }
func (k KeyInfos) Sort() {
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

// KeyManager allows simple CRUD on a keystore, as an aid to signing
type KeyManager interface {
	Create(name, passphrase string) error
	List() (KeyInfos, error)
	Get(name string) (KeyInfo, error)
	Update(name, oldpass, newpass string) error
	Delete(name, passphrase string) error
}
