package keys

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/crypto/bcrypt"
	"github.com/tendermint/tendermint/crypto"
)

var _ Keybase = lazyKeybaseKeyring{}

type lazyKeybaseKeyring struct {
	name           string
	dir            string
	test           bool
	userInput      io.Reader
	cachedPassword string
}

// NewKeybaseKeyring creates a new instance of a lazy keybase.
func NewKeybaseKeyring(name string, dir string, userInput io.Reader, test bool) Keybase {

	_, err := keyring.Open(keyring.Config{
		//Keyring with encrypted application data
		ServiceName: name,
	})
	if err != nil {
		panic(err)
	}

	return lazyKeybaseKeyring{name: name, dir: dir, userInput: userInput, test: test}
}

func (lkb *lazyKeybaseKeyring) lkbToKeyringConfig() keyring.Config {
	if lkb.test {
		return keyring.Config{
			AllowedBackends: []keyring.BackendType{"file"},
			//Keyring with encrypted application data
			ServiceName:      lkb.name,
			FileDir:          lkb.dir,
			FilePasswordFunc: fakePrompt,
		}
	}

	realPrompt := func(prompt string) (string, error) {

		buf := bufio.NewReader(lkb.userInput)

		var keyhash []byte
		keyhashStored := false

		if _, err := os.Stat(lkb.dir + "/keyhash"); err == nil {
			keyhash, err = ioutil.ReadFile(lkb.dir + "/keyhash")
			keyhashStored = true
		} else if os.IsNotExist(err) {
			keyhashStored = false
		} else {
			return "", err
		}

		if lkb.cachedPassword != "" {
			return lkb.cachedPassword, nil
		}

		failureCounter := 0
		for {
			failureCounter++

			if failureCounter > 10 {
				return "", fmt.Errorf("Too Many Failed Passphrase attempts")
			}

			pass, err := input.GetPassword("Enter keyring files passphrase for your keys:", buf)

			if err != nil {
				continue
			}

			if !keyhashStored {

				reEnteredPass, err := input.GetPassword("Re-enter keyring files passphrase for your keys:", buf)
				if err != nil {
					fmt.Println(err)
					continue
				}

				if pass != reEnteredPass {
					fmt.Println("Passwords do not match")
					continue
				}

				saltBytes := crypto.CRandBytes(16)
				passwordHash, err := bcrypt.GenerateFromPassword(saltBytes, []byte(pass), 2)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = ioutil.WriteFile(lkb.dir+"/keyhash", passwordHash, 0555)
				if err != nil {
					return "", err
				}
				lkb.cachedPassword = pass
				return pass, nil

			} else {
				err = bcrypt.CompareHashAndPassword(keyhash, []byte(pass))
				if err != nil {
					fmt.Println("incorrect password")
					continue
				}
				lkb.cachedPassword = pass
				return pass, nil
			}
		}

	}

	return keyring.Config{
		//Keyring with encrypted application data
		ServiceName:      lkb.name,
		FileDir:          lkb.dir,
		FilePasswordFunc: realPrompt,
	}
}

func fakePrompt(prompt string) (string, error) {

	fmt.Println("Fake Prompt for passphase. Testing only")
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
