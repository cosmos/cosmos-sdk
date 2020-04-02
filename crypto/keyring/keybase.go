package keyring

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types"
)

// Keybase exposes operations on a generic keystore
type Keybase interface {
	// CRUD on the keystore
	List() ([]Info, error)
	// Get returns the public information about one key.
	Get(name string) (Info, error)
	// Get performs a by-address lookup and returns the public
	// information about one key if there's any.
	GetByAddress(address types.AccAddress) (Info, error)
	// Delete removes a key.
	Delete(name, passphrase string, skipPass bool) error
	// Sign bytes, looking up the private key to use.
	Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error)

	// CreateMnemonic generates a new mnemonic, derives a hierarchical deterministic
	// key from that. and persists it to storage, encrypted using the provided password.
	// It returns the generated mnemonic and the key Info. It returns an error if it fails to
	// generate a key for the given algo type, or if another key is already stored under the
	// same name.
	CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error)

	// CreateAccount converts a mnemonic to a private key and BIP 32 HD Path
	// and persists it, encrypted with the given password.
	CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd, hdPath string, algo SigningAlgo) (Info, error)

	// CreateLedger creates, stores, and returns a new Ledger key reference
	CreateLedger(name string, algo SigningAlgo, hrp string, account, index uint32) (info Info, err error)

	// CreateOffline creates, stores, and returns a new offline key reference
	CreateOffline(name string, pubkey crypto.PubKey, algo SigningAlgo) (info Info, err error)

	// CreateMulti creates, stores, and returns a new multsig (offline) key reference
	CreateMulti(name string, pubkey crypto.PubKey) (info Info, err error)

	// Import imports ASCII armored Info objects.
	Import(name string, armor string) (err error)

	// ImportPrivKey imports a private key in ASCII armor format.
	// It returns an error if a key with the same name exists or a wrong encryption passphrase is
	// supplied.
	ImportPrivKey(name, armor, passphrase string) error

	// ImportPubKey imports ASCII-armored public keys.
	// Store a new Info object holding a public key only, i.e. it will
	// not be possible to sign with it as it lacks the secret key.
	ImportPubKey(name string, armor string) (err error)

	// Export exports an Info object in ASCII armored format.
	Export(name string) (armor string, err error)

	// ExportPubKey returns public keys in ASCII armored format.
	// Retrieve a Info object by its name and return the public key in
	// a portable format.
	ExportPubKey(name string) (armor string, err error)

	// ExportPrivKey returns a private key in ASCII armored format.
	// It returns an error if the key does not exist or a wrong encryption passphrase is supplied.
	ExportPrivKey(name, decryptPassphrase, encryptPassphrase string) (armor string, err error)

	// ExportPrivateKeyObject *only* works on locally-stored keys. Temporary method until we redo the exporting API
	ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error)

	// SupportedAlgos returns a list of signing algorithms supported by the keybase
	SupportedAlgos() []SigningAlgo

	// SupportedAlgosLedger returns a list of signing algorithms supported by the keybase's ledger integration
	SupportedAlgosLedger() []SigningAlgo
}
