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

type lazyKeybaseKeyring struct {
	name      string
	dir       string
	test      bool
	userInput io.Reader
}

// NewKeybaseKeyring creates a new instance of a lazy keybase using a Keyring as
// the persistence layer.
func NewKeybaseKeyring(name string, dir string, userInput io.Reader, test bool) Keybase {
	_, err := keyring.Open(keyring.Config{
		ServiceName: name,
	})
	if err != nil {
		panic(err)
	}

	return lazyKeybaseKeyring{name: name, dir: dir, userInput: userInput, test: test}
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

func (lkb lazyKeybaseKeyring) List() ([]Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).List()
}

func (lkb lazyKeybaseKeyring) Get(name string) (Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).Get(name)
}

func (lkb lazyKeybaseKeyring) GetByAddress(address sdk.AccAddress) (Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).GetByAddress(address)
}

func (lkb lazyKeybaseKeyring) Delete(name, passphrase string, skipPass bool) error {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return err
	}

	return newKeyringKeybase(db).Delete(name, passphrase, skipPass)
}

func (lkb lazyKeybaseKeyring) Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return nil, nil, err
	}

	return newKeyringKeybase(db).Sign(name, passphrase, msg)
}

func (lkb lazyKeybaseKeyring) CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return nil, "", err
	}

	return newKeyringKeybase(db).CreateMnemonic(name, language, passwd, algo)
}

func (lkb lazyKeybaseKeyring) CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd, account, index)
}

func (lkb lazyKeybaseKeyring) Derive(name, mnemonic, bip39Passwd, encryptPasswd string, params hd.BIP44Params) (Info, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).Derive(name, mnemonic, bip39Passwd, encryptPasswd, params)
}

func (lkb lazyKeybaseKeyring) CreateLedger(name string, algo SigningAlgo, hrp string, account uint32, index uint32) (info Info, err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).CreateLedger(name, algo, hrp, account, index)
}

func (lkb lazyKeybaseKeyring) CreateOffline(name string, pubkey crypto.PubKey) (info Info, err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).CreateOffline(name, pubkey)
}

func (lkb lazyKeybaseKeyring) CreateMulti(name string, pubkey crypto.PubKey) (info Info, err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).CreateMulti(name, pubkey)
}

func (lkb lazyKeybaseKeyring) Update(name, oldpass string, getNewpass func() (string, error)) error {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return err
	}

	return newKeyringKeybase(db).Update(name, oldpass, getNewpass)
}

func (lkb lazyKeybaseKeyring) Import(name string, armor string) (err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return err
	}

	return newKeyringKeybase(db).Import(name, armor)
}

func (lkb lazyKeybaseKeyring) ImportPrivKey(name string, armor string, passphrase string) error {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return err
	}

	return newKeyringKeybase(db).ImportPrivKey(name, armor, passphrase)
}

func (lkb lazyKeybaseKeyring) ImportPubKey(name string, armor string) (err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return err
	}

	return newKeyringKeybase(db).ImportPubKey(name, armor)
}

func (lkb lazyKeybaseKeyring) Export(name string) (armor string, err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return "", err
	}

	return newKeyringKeybase(db).Export(name)
}

func (lkb lazyKeybaseKeyring) ExportPubKey(name string) (armor string, err error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return "", err
	}

	return newKeyringKeybase(db).ExportPubKey(name)
}

func (lkb lazyKeybaseKeyring) ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error) {
	db, err := keyring.Open(lkb.lkbToKeyringConfig())

	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db).ExportPrivateKeyObject(name, passphrase)
}

func (lkb lazyKeybaseKeyring) ExportPrivKey(name string, decryptPassphrase string,
	encryptPassphrase string) (armor string, err error) {

	db, err := keyring.Open(lkb.lkbToKeyringConfig())
	if err != nil {
		return "", err
	}

	return newKeyringKeybase(db).ExportPrivKey(name, decryptPassphrase, encryptPassphrase)
}

func (lkb lazyKeybaseKeyring) CloseDB() {}
