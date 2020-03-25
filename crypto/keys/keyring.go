package keys

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/99designs/keyring"
	"github.com/pkg/errors"

	"github.com/tendermint/crypto/bcrypt"
	"github.com/tendermint/tendermint/crypto"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	"github.com/cosmos/cosmos-sdk/types"
)

const (
	BackendFile    = "file"
	BackendOS      = "os"
	BackendKWallet = "kwallet"
	BackendPass    = "pass"
	BackendTest    = "test"
)

const (
	keyringDirNameFmt     = "keyring-%s"
	testKeyringDirNameFmt = "keyring-test-%s"
)

var _ Keybase = keyringKeybase{}

// keyringKeybase implements the Keybase interface by using the Keyring library
// for account key persistence.
type keyringKeybase struct {
	base baseKeybase
	db   keyring.Keyring
}

var maxPassphraseEntryAttempts = 3

func newKeyringKeybase(db keyring.Keyring, opts ...KeybaseOption) Keybase {
	return keyringKeybase{
		db:   db,
		base: newBaseKeybase(opts...),
	}
}

// NewKeyring creates a new instance of a keyring. Keybase
// options can be applied when generating this new Keybase.
// Available backends are "os", "file", "test".
func NewKeyring(
	appName, backend, rootDir string, userInput io.Reader, opts ...KeybaseOption,
) (Keybase, error) {

	var db keyring.Keyring
	var err error

	switch backend {
	case BackendTest:
		db, err = keyring.Open(lkbToKeyringConfig(appName, rootDir, nil, true))
	case BackendFile:
		db, err = keyring.Open(newFileBackendKeyringConfig(appName, rootDir, userInput))
	case BackendOS:
		db, err = keyring.Open(lkbToKeyringConfig(appName, rootDir, userInput, false))
	case BackendKWallet:
		db, err = keyring.Open(newKWalletBackendKeyringConfig(appName, rootDir, userInput))
	case BackendPass:
		db, err = keyring.Open(newPassBackendKeyringConfig(appName, rootDir, userInput))
	default:
		return nil, fmt.Errorf("unknown keyring backend %v", backend)
	}
	if err != nil {
		return nil, err
	}

	return newKeyringKeybase(db, opts...), nil
}

// CreateMnemonic generates a new key and persists it to storage, encrypted
// using the provided password. It returns the generated mnemonic and the key Info.
// An error is returned if it fails to generate a key for the given algo type,
// or if another key is already stored under the same name.
func (kb keyringKeybase) CreateMnemonic(
	name string, language Language, passwd string, algo SigningAlgo,
) (info Info, mnemonic string, err error) {

	return kb.base.CreateMnemonic(kb, name, language, passwd, algo)
}

// CreateAccount converts a mnemonic to a private key and persists it, encrypted
// with the given password.
func (kb keyringKeybase) CreateAccount(
	name, mnemonic, bip39Passwd, encryptPasswd, hdPath string, algo SigningAlgo,
) (Info, error) {

	return kb.base.CreateAccount(kb, name, mnemonic, bip39Passwd, encryptPasswd, hdPath, algo)
}

// CreateLedger creates a new locally-stored reference to a Ledger keypair.
// It returns the created key info and an error if the Ledger could not be queried.
func (kb keyringKeybase) CreateLedger(
	name string, algo SigningAlgo, hrp string, account, index uint32,
) (Info, error) {

	return kb.base.CreateLedger(kb, name, algo, hrp, account, index)
}

// CreateOffline creates a new reference to an offline keypair. It returns the
// created key info.
func (kb keyringKeybase) CreateOffline(name string, pub tmcrypto.PubKey, algo SigningAlgo) (Info, error) {
	return kb.base.writeOfflineKey(kb, name, pub, algo), nil
}

// CreateMulti creates a new reference to a multisig (offline) keypair. It
// returns the created key Info object.
func (kb keyringKeybase) CreateMulti(name string, pub tmcrypto.PubKey) (Info, error) {
	return kb.base.writeMultisigKey(kb, name, pub), nil
}

// List returns the keys from storage in alphabetical order.
func (kb keyringKeybase) List() ([]Info, error) {
	var res []Info
	keys, err := kb.db.Keys()
	if err != nil {
		return nil, err
	}

	sort.Strings(keys)

	for _, key := range keys {
		if strings.HasSuffix(key, infoSuffix) {
			rawInfo, err := kb.db.Get(key)
			if err != nil {
				return nil, err
			}

			if len(rawInfo.Data) == 0 {
				return nil, keyerror.NewErrKeyNotFound(key)
			}

			info, err := unmarshalInfo(rawInfo.Data)
			if err != nil {
				return nil, err
			}

			res = append(res, info)
		}
	}

	return res, nil
}

// Get returns the public information about one key.
func (kb keyringKeybase) Get(name string) (Info, error) {
	key := infoKey(name)

	bs, err := kb.db.Get(string(key))
	if err != nil {
		return nil, err
	}

	if len(bs.Data) == 0 {
		return nil, keyerror.NewErrKeyNotFound(name)
	}

	return unmarshalInfo(bs.Data)
}

// GetByAddress fetches a key by address and returns its public information.
func (kb keyringKeybase) GetByAddress(address types.AccAddress) (Info, error) {
	ik, err := kb.db.Get(string(addrKey(address)))
	if err != nil {
		return nil, err
	}

	if len(ik.Data) == 0 {
		return nil, fmt.Errorf("key with address %s not found", address)
	}

	bs, err := kb.db.Get(string(ik.Data))
	if err != nil {
		return nil, err
	}

	return unmarshalInfo(bs.Data)
}

// Sign signs an arbitrary set of bytes with the named key. It returns an error
// if the key doesn't exist or the decryption fails.
func (kb keyringKeybase) Sign(name, passphrase string, msg []byte) (sig []byte, pub tmcrypto.PubKey, err error) {
	info, err := kb.Get(name)
	if err != nil {
		return
	}

	var priv tmcrypto.PrivKey

	switch i := info.(type) {
	case localInfo:
		if i.PrivKeyArmor == "" {
			return nil, nil, fmt.Errorf("private key not available")
		}

		priv, err = cryptoAmino.PrivKeyFromBytes([]byte(i.PrivKeyArmor))
		if err != nil {
			return nil, nil, err
		}

	case ledgerInfo:
		return kb.base.SignWithLedger(info, msg)

	case offlineInfo, multiInfo:
		return kb.base.DecodeSignature(info, msg)
	}

	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil
}

// ExportPrivateKeyObject exports an armored private key object.
func (kb keyringKeybase) ExportPrivateKeyObject(name string, passphrase string) (tmcrypto.PrivKey, error) {
	info, err := kb.Get(name)
	if err != nil {
		return nil, err
	}

	var priv tmcrypto.PrivKey

	switch linfo := info.(type) {
	case localInfo:
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return nil, err
		}

		priv, err = cryptoAmino.PrivKeyFromBytes([]byte(linfo.PrivKeyArmor))
		if err != nil {
			return nil, err
		}

	case ledgerInfo, offlineInfo, multiInfo:
		return nil, errors.New("only works on local private keys")
	}

	return priv, nil
}

// Export exports armored private key to the caller.
func (kb keyringKeybase) Export(name string) (armor string, err error) {
	bz, err := kb.db.Get(string(infoKey(name)))
	if err != nil {
		return "", err
	}

	if bz.Data == nil {
		return "", fmt.Errorf("no key to export with name: %s", name)
	}

	return mintkey.ArmorInfoBytes(bz.Data), nil
}

// ExportPubKey returns public keys in ASCII armored format. It retrieves an Info
// object by its name and return the public key in a portable format.
func (kb keyringKeybase) ExportPubKey(name string) (armor string, err error) {
	bz, err := kb.Get(name)
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name: %s", name)
	}

	return mintkey.ArmorPubKeyBytes(bz.GetPubKey().Bytes(), string(bz.GetAlgo())), nil
}

// Import imports armored private key.
func (kb keyringKeybase) Import(name string, armor string) error {
	bz, _ := kb.Get(name)

	if bz != nil {
		pubkey := bz.GetPubKey()

		if len(pubkey.Bytes()) > 0 {
			return fmt.Errorf("cannot overwrite data for name: %s", name)
		}
	}

	infoBytes, err := mintkey.UnarmorInfoBytes(armor)
	if err != nil {
		return err
	}

	info, err := unmarshalInfo(infoBytes)
	if err != nil {
		return err
	}

	kb.writeInfo(name, info)

	err = kb.db.Set(keyring.Item{
		Key:  string(addrKey(info.GetAddress())),
		Data: infoKey(name),
	})
	if err != nil {
		return err
	}

	return nil
}

// ExportPrivKey returns a private key in ASCII armored format. An error is returned
// if the key does not exist or a wrong encryption passphrase is supplied.
func (kb keyringKeybase) ExportPrivKey(name, decryptPassphrase, encryptPassphrase string) (armor string, err error) {
	priv, err := kb.ExportPrivateKeyObject(name, decryptPassphrase)
	if err != nil {
		return "", err
	}

	info, err := kb.Get(name)
	if err != nil {
		return "", err
	}

	return mintkey.EncryptArmorPrivKey(priv, encryptPassphrase, string(info.GetAlgo())), nil
}

// ImportPrivKey imports a private key in ASCII armor format. An error is returned
// if a key with the same name exists or a wrong encryption passphrase is
// supplied.
func (kb keyringKeybase) ImportPrivKey(name, armor, passphrase string) error {
	if kb.HasKey(name) {
		return fmt.Errorf("cannot overwrite key: %s", name)
	}

	privKey, algo, err := mintkey.UnarmorDecryptPrivKey(armor, passphrase)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt private key")
	}

	// NOTE: The keyring keystore has no need for a passphrase.
	kb.writeLocalKey(name, privKey, "", SigningAlgo(algo))
	return nil
}

// HasKey returns whether the key exists in the keyring.
func (kb keyringKeybase) HasKey(name string) bool {
	bz, _ := kb.Get(name)
	return bz != nil
}

// ImportPubKey imports an ASCII-armored public key. It will store a new Info
// object holding a public key only, i.e. it will not be possible to sign with
// it as it lacks the secret key.
func (kb keyringKeybase) ImportPubKey(name string, armor string) error {
	bz, _ := kb.Get(name)
	if bz != nil {
		pubkey := bz.GetPubKey()

		if len(pubkey.Bytes()) > 0 {
			return fmt.Errorf("cannot overwrite data for name: %s", name)
		}
	}

	pubBytes, algo, err := mintkey.UnarmorPubKeyBytes(armor)
	if err != nil {
		return err
	}

	pubKey, err := cryptoAmino.PubKeyFromBytes(pubBytes)
	if err != nil {
		return err
	}

	kb.base.writeOfflineKey(kb, name, pubKey, SigningAlgo(algo))
	return nil
}

// Delete removes key forever, but we must present the proper passphrase before
// deleting it (for security). It returns an error if the key doesn't exist or
// passphrases don't match. The passphrase is ignored when deleting references to
// offline and Ledger / HW wallet keys.
func (kb keyringKeybase) Delete(name, _ string, _ bool) error {
	// verify we have the proper password before deleting
	info, err := kb.Get(name)
	if err != nil {
		return err
	}

	err = kb.db.Remove(string(addrKey(info.GetAddress())))
	if err != nil {
		return err
	}

	err = kb.db.Remove(string(infoKey(name)))
	if err != nil {
		return err
	}

	return nil
}

// Update changes the passphrase with which an already stored key is encrypted.
// The oldpass must be the current passphrase used for encryption, getNewpass is
// a function to get the passphrase to permanently replace the current passphrase.
func (kb keyringKeybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	info, err := kb.Get(name)
	if err != nil {
		return err
	}

	switch linfo := info.(type) {
	case localInfo:
		key, _, err := mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, oldpass)
		if err != nil {
			return err
		}

		newpass, err := getNewpass()
		if err != nil {
			return err
		}

		kb.writeLocalKey(name, key, newpass, linfo.GetAlgo())
		return nil

	default:
		return fmt.Errorf("locally stored key required; received: %v", reflect.TypeOf(info).String())
	}
}

// SupportedAlgos returns a list of supported signing algorithms.
func (kb keyringKeybase) SupportedAlgos() []SigningAlgo {
	return kb.base.SupportedAlgos()
}

// SupportedAlgosLedger returns a list of supported ledger signing algorithms.
func (kb keyringKeybase) SupportedAlgosLedger() []SigningAlgo {
	return kb.base.SupportedAlgosLedger()
}

// CloseDB releases the lock and closes the storage backend.
func (kb keyringKeybase) CloseDB() {}

func (kb keyringKeybase) writeLocalKey(name string, priv tmcrypto.PrivKey, _ string, algo SigningAlgo) Info {
	// encrypt private key using keyring
	pub := priv.PubKey()
	info := newLocalInfo(name, pub, string(priv.Bytes()), algo)

	kb.writeInfo(name, info)
	return info
}

func (kb keyringKeybase) writeInfo(name string, info Info) {
	// write the info by key
	key := infoKey(name)
	serializedInfo := marshalInfo(info)

	err := kb.db.Set(keyring.Item{
		Key:  string(key),
		Data: serializedInfo,
	})
	if err != nil {
		panic(err)
	}

	err = kb.db.Set(keyring.Item{
		Key:  string(addrKey(info.GetAddress())),
		Data: key,
	})
	if err != nil {
		panic(err)
	}
}

func lkbToKeyringConfig(appName, dir string, buf io.Reader, test bool) keyring.Config {
	if test {
		return keyring.Config{
			AllowedBackends: []keyring.BackendType{keyring.FileBackend},
			ServiceName:     appName,
			FileDir:         filepath.Join(dir, fmt.Sprintf(testKeyringDirNameFmt, appName)),
			FilePasswordFunc: func(_ string) (string, error) {
				return "test", nil
			},
		}
	}

	return keyring.Config{
		ServiceName:      appName,
		FileDir:          dir,
		FilePasswordFunc: newRealPrompt(dir, buf),
	}
}

func newKWalletBackendKeyringConfig(appName, _ string, _ io.Reader) keyring.Config {
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.KWalletBackend},
		ServiceName:     "kdewallet",
		KWalletAppID:    appName,
		KWalletFolder:   "",
	}
}

func newPassBackendKeyringConfig(appName, dir string, _ io.Reader) keyring.Config {
	prefix := filepath.Join(dir, fmt.Sprintf(keyringDirNameFmt, appName))
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.PassBackend},
		ServiceName:     appName,
		PassPrefix:      prefix,
	}
}

func newFileBackendKeyringConfig(name, dir string, buf io.Reader) keyring.Config {
	fileDir := filepath.Join(dir, fmt.Sprintf(keyringDirNameFmt, name))
	return keyring.Config{
		AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
		ServiceName:      name,
		FileDir:          fileDir,
		FilePasswordFunc: newRealPrompt(fileDir, buf),
	}
}

func newRealPrompt(dir string, buf io.Reader) func(string) (string, error) {
	return func(prompt string) (string, error) {
		keyhashStored := false
		keyhashFilePath := filepath.Join(dir, "keyhash")

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

			buf := bufio.NewReader(buf)
			pass, err := input.GetPassword("Enter keyring passphrase:", buf)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
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

			if err := ioutil.WriteFile(dir+"/keyhash", passwordHash, 0555); err != nil {
				return "", err
			}

			return pass, nil
		}
	}
}
