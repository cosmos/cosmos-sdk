package keys

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/99designs/keyring"
	"github.com/tendermint/crypto/bcrypt"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ Keybase = lazyKeybaseKeyring{}

	maxPassphraseEntryAttempts = 3
)

// lazyKeybaseKeyring implements a public wrapper around the keyringKeybase type
// and implements the Keybase interface.
type lazyKeybaseKeyring struct {
	name      string
	dir       string
	test      bool
	userInput io.Reader
}

// NewKeyring creates a new instance of a lazy keybase using a Keyring as
// the persistence layer.
func NewKeyring(name string, dir string, userInput io.Reader) Keybase {
	_, err := keyring.Open(keyring.Config{
		ServiceName: name,
	})
	if err != nil {
		panic(err)
	}

	return lazyKeybaseKeyring{name: name, dir: dir, userInput: userInput, test: false}
}

// NewTestKeyring creates a new instance of a keyring keybase
// for testing purposes that  does not prompt users for password.
func NewTestKeyring(name string, dir string) Keybase {
	if _, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{"file"},
		ServiceName:     name,
		FileDir:         dir,
	}); err != nil {
		panic(err)
	}

	return lazyKeybaseKeyring{name: name, dir: dir, test: true}
}

func (lkb lazyKeybaseKeyring) lkbToKeyringConfig() keyring.Config {
	if lkb.test {
		return keyring.Config{
			AllowedBackends:  []keyring.BackendType{"file"},
			ServiceName:      lkb.name,
			FileDir:          lkb.dir,
			FilePasswordFunc: fakePrompt,
		}
	}

	realPrompt := func(prompt string) (string, error) {
		keyhashStored := false
		keyhashFilePath := filepath.Join(lkb.dir, "keyhash")

		var keyhash []byte

		_, err := os.Stat(keyhashFilePath)
		switch {
		case err == nil:
			keyhash, err = ioutil.ReadFile(keyhashFilePath)
			if err != nil {
				return "", fmt.Errorf("failed to read %s: %v", keyhashFilePath, err)
			}

			keyhashStored = true

		case os.IsNotExist(err):
			keyhashStored = false

		default:
			return "", fmt.Errorf("failed to open %s: %v", keyhashFilePath, err)
		}

		failureCounter := 0
		for {
			failureCounter++
			if failureCounter > maxPassphraseEntryAttempts {
				return "", fmt.Errorf("too many failed passphrase attempts")
			}

			buf := bufio.NewReader(lkb.userInput)
			pass, err := input.GetPassword("Enter keyring passphrase:", buf)
			if err != nil {
				continue
			}

			if keyhashStored {
				if err := bcrypt.CompareHashAndPassword(keyhash, []byte(pass)); err != nil {
					fmt.Fprintln(os.Stderr, "incorrect passphrase")
					continue
				}
				return pass, nil
			}

			reEnteredPass, err := input.GetPassword("Re-enter keyring passphrase:", buf)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			if pass != reEnteredPass {
				fmt.Fprintln(os.Stderr, "passphrase do not match")
				continue
			}

			saltBytes := crypto.CRandBytes(16)
			passwordHash, err := bcrypt.GenerateFromPassword(saltBytes, []byte(pass), 2)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			if err := ioutil.WriteFile(lkb.dir+"/keyhash", passwordHash, 0555); err != nil {
				return "", err
			}

			return pass, nil
		}
	}

	return keyring.Config{
		ServiceName:      lkb.name,
		FileDir:          lkb.dir,
		FilePasswordFunc: realPrompt,
	}
}

func fakePrompt(prompt string) (string, error) {
	fmt.Fprintln(os.Stderr, "Fake Prompt for passphase. Testing only")
	return "test", nil
}

// List returns the keys from storage in alphabetical order.
func (lkb lazyKeybaseKeyring) List() ([]Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).List()
}

// Get returns the public information about one key.
func (lkb lazyKeybaseKeyring) Get(name string) (Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).Get(name)
}

// GetByAddress fetches a key by address and returns its public information.
func (lkb lazyKeybaseKeyring) GetByAddress(address sdk.AccAddress) (Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).GetByAddress(address)
}

// Delete removes key forever, but we must present the proper passphrase before
// deleting it (for security). It returns an error if the key doesn't exist or
// passphrases don't match. The passphrase is ignored when deleting references to
// offline and Ledger / HW wallet keys.
func (lkb lazyKeybaseKeyring) Delete(name, passphrase string, skipPass bool) error {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return err
	}

	return newKeyringKeybase(db).Delete(name, passphrase, skipPass)
}

// Sign signs an arbitrary set of bytes with the named key. It returns an error
// if the key doesn't exist or the decryption fails.
func (lkb lazyKeybaseKeyring) Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, nil, err
	}

	return newKeyringKeybase(db).Sign(name, passphrase, msg)
}

// CreateMnemonic generates a new key and persists it to storage, encrypted
// using the provided password. It returns the generated mnemonic and the key Info.
// An error is returned if it fails to generate a key for the given algo type,
// or if another key is already stored under the same name.
func (lkb lazyKeybaseKeyring) CreateMnemonic(
	name string, language Language, passwd string, algo SigningAlgo,
) (info Info, seed string, err error) {

	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, "", err
	}

	return newKeyringKeybase(db).CreateMnemonic(name, language, passwd, algo)
}

// CreateAccount converts a mnemonic to a private key and persists it, encrypted
// with the given password.
func (lkb lazyKeybaseKeyring) CreateAccount(
	name, mnemonic, bip39Passwd, encryptPasswd string, account, index uint32,
) (Info, error) {

	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd, account, index)
}

// Derive computes a BIP39 seed from th mnemonic and bip39Passphrase. It creates
// a private key from the seed using the BIP44 params.
func (lkb lazyKeybaseKeyring) Derive(
	name, mnemonic, bip39Passwd, encryptPasswd string, params hd.BIP44Params,
) (Info, error) {

	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).Derive(name, mnemonic, bip39Passwd, encryptPasswd, params)
}

// CreateLedger creates a new locally-stored reference to a Ledger keypair.
// It returns the created key info and an error if the Ledger could not be queried.
func (lkb lazyKeybaseKeyring) CreateLedger(
	name string, algo SigningAlgo, hrp string, account, index uint32,
) (Info, error) {

	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).CreateLedger(name, algo, hrp, account, index)
}

// CreateOffline creates a new reference to an offline keypair. It returns the
// created key info.
func (lkb lazyKeybaseKeyring) CreateOffline(name string, pubkey crypto.PubKey) (Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).CreateOffline(name, pubkey)
}

// CreateMulti creates a new reference to a multisig (offline) keypair. It
// returns the created key Info object.
func (lkb lazyKeybaseKeyring) CreateMulti(name string, pubkey crypto.PubKey) (Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).CreateMulti(name, pubkey)
}

// Update changes the passphrase with which an already stored key is encrypted.
// The oldpass must be the current passphrase used for encryption, getNewpass is
// a function to get the passphrase to permanently replace the current passphrase.
func (lkb lazyKeybaseKeyring) Update(name, oldpass string, getNewpass func() (string, error)) error {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return err
	}

	return newKeyringKeybase(db).Update(name, oldpass, getNewpass)
}

// Import imports armored private key.
func (lkb lazyKeybaseKeyring) Import(name, armor string) error {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return err
	}

	return newKeyringKeybase(db).Import(name, armor)
}

// ImportPrivKey imports a private key in ASCII armor format. An error is returned
// if a key with the same name exists or a wrong encryption passphrase is
// supplied.
func (lkb lazyKeybaseKeyring) ImportPrivKey(name, armor, passphrase string) error {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return err
	}

	return newKeyringKeybase(db).ImportPrivKey(name, armor, passphrase)
}

// ImportPubKey imports an ASCII-armored public key. It will store a new Info
// object holding a public key only, i.e. it will not be possible to sign with
// it as it lacks the secret key.
func (lkb lazyKeybaseKeyring) ImportPubKey(name, armor string) error {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return err
	}

	return newKeyringKeybase(db).ImportPubKey(name, armor)
}

// Export exports armored a private key for the given name.
func (lkb lazyKeybaseKeyring) Export(name string) (armor string, err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return "", err
	}

	return newKeyringKeybase(db).Export(name)
}

// ExportPubKey returns public keys in ASCII armored format. It retrieves an Info
// object by its name and return the public key in a portable format.
func (lkb lazyKeybaseKeyring) ExportPubKey(name string) (armor string, err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return "", err
	}

	return newKeyringKeybase(db).ExportPubKey(name)
}

// ExportPrivateKeyObject exports an armored private key object.
func (lkb lazyKeybaseKeyring) ExportPrivateKeyObject(name, passphrase string) (crypto.PrivKey, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).ExportPrivateKeyObject(name, passphrase)
}

// ExportPrivKey returns a private key in ASCII armored format. An error is returned
// if the key does not exist or a wrong encryption passphrase is supplied.
func (lkb lazyKeybaseKeyring) ExportPrivKey(name, decryptPassphrase, encryptPassphrase string) (string, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return "", err
	}

	return newKeyringKeybase(db).ExportPrivKey(name, decryptPassphrase, encryptPassphrase)
}

func (lkb lazyKeybaseKeyring) CloseDB() {}
