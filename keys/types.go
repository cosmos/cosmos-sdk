package keys

import (
	crypto "github.com/tendermint/go-crypto"
)

// Keybase exposes operations on a generic keystore
type Keybase interface {

	// CRUD on the keystore
	List() ([]Info, error)
	Get(name string) (Info, error)
	Delete(name, passphrase string) error

	// Sign some bytes, looking up the private key to use
	Sign(name, passphrase string, msg []byte) (crypto.Signature, crypto.PubKey, error)

	// Create a new locally-stored keypair, returning the mnemonic
	CreateMnemonic(name, passphrase string, algo SignAlgo) (info Info, seed string, err error)
	// Recover takes a seedphrase and loads in the key
	Recover(name, passphrase, seedphrase string) (info Info, erro error)

	// Create, store, and return a new Ledger key reference
	CreateLedger(name string, path crypto.DerivationPath, algo SignAlgo) (info Info, err error)

	// Create, store, and return a new offline key reference
	CreateOffline(name string, pubkey crypto.PubKey) (info Info, err error)

	// The following operations will *only* work on locally-stored keys
	Update(name, oldpass, newpass string) error
	Import(name string, armor string) (err error)
	ImportPubKey(name string, armor string) (err error)
	Export(name string) (armor string, err error)
	ExportPubKey(name string) (armor string, err error)
}

// Publically exposed information about a keypair
type Info interface {
	// Human-readable type for key listing
	GetType() string
	// Name of the key
	GetName() string
	// Public key
	GetPubKey() crypto.PubKey
}

var _ Info = &localInfo{}
var _ Info = &ledgerInfo{}
var _ Info = &offlineInfo{}

// localInfo is the public information about a locally stored key
type localInfo struct {
	Name         string        `json:"name"`
	PubKey       crypto.PubKey `json:"pubkey"`
	PrivKeyArmor string        `json:"privkey.armor"`
}

func newLocalInfo(name string, pub crypto.PubKey, privArmor string) Info {
	return &localInfo{
		Name:         name,
		PubKey:       pub,
		PrivKeyArmor: privArmor,
	}
}

func (i localInfo) GetType() string {
	return "local"
}

func (i localInfo) GetName() string {
	return i.Name
}

func (i localInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// ledgerInfo is the public information about a Ledger key
type ledgerInfo struct {
	Name   string                `json:"name"`
	PubKey crypto.PubKey         `json:"pubkey"`
	Path   crypto.DerivationPath `json:"path"`
}

func newLedgerInfo(name string, pub crypto.PubKey, path crypto.DerivationPath) Info {
	return &ledgerInfo{
		Name:   name,
		PubKey: pub,
		Path:   path,
	}
}

func (i ledgerInfo) GetType() string {
	return "ledger"
}

func (i ledgerInfo) GetName() string {
	return i.Name
}

func (i ledgerInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// offlineInfo is the public information about an offline key
type offlineInfo struct {
	Name   string        `json:"name"`
	PubKey crypto.PubKey `json:"pubkey"`
}

func newOfflineInfo(name string, pub crypto.PubKey) Info {
	return &offlineInfo{
		Name:   name,
		PubKey: pub,
	}
}

func (i offlineInfo) GetType() string {
	return "offline"
}

func (i offlineInfo) GetName() string {
	return i.Name
}

func (i offlineInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// encoding info
func writeInfo(i Info) []byte {
	return cdc.MustMarshalBinary(i)
}

// decoding info
func readInfo(bz []byte) (info Info, err error) {
	err = cdc.UnmarshalBinary(bz, &info)
	return
}
