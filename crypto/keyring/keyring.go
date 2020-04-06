package keyring

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/99designs/keyring"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/tendermint/crypto/bcrypt"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	BackendFile    = "file"
	BackendOS      = "os"
	BackendKWallet = "kwallet"
	BackendPass    = "pass"
	BackendTest    = "test"
	BackendMemory  = "memory"
)

const (
	keyringFileDirName = "keyring-file"
	keyringTestDirName = "keyring-test"
	passKeyringPrefix  = "keyring-%s"
)

var (
	_                          Keyring = &keystore{}
	maxPassphraseEntryAttempts         = 3
)

// Keyring exposes operations over a backend supported by github.com/99designs/keyring.
type Keyring interface {
	// List all keys.
	List() ([]Info, error)

	// Key and KeyByAddress return keys by uid and address respectively.
	Key(uid string) (Info, error)
	KeyByAddress(address sdk.Address) (Info, error)

	// Delete and DeleteByAddress remove keys from the keyring.
	Delete(uid string) error
	DeleteByAddress(address sdk.Address) error

	// NewMnemonic generates a new mnemonic, derives a hierarchical deterministic
	// key from that, and persists it to the storage. Returns the generated mnemonic and the key
	// Info. It returns an error if it fails to generate a key for the given algo type, or if
	// another key is already stored under the same name.
	NewMnemonic(uid string, language Language, hdPath string, algo SignatureAlgo) (Info, string, error)

	// NewAccount converts a mnemonic to a private key and BIP-39 HD Path and persists it.
	NewAccount(uid, mnemonic, bip39Passwd, hdPath string, algo SignatureAlgo) (Info, error)

	// SaveLedgerKey retrieves a public key reference from a Ledger device and persists it.
	SaveLedgerKey(uid string, algo SignatureAlgo, hrp string, coinType, account, index uint32) (Info, error)

	// SavePubKey stores a public key and returns the persisted Info structure.
	SavePubKey(uid string, pubkey tmcrypto.PubKey, algo hd.PubKeyType) (Info, error)

	// SaveMultisig stores and returns a new multsig (offline) key reference.
	SaveMultisig(uid string, pubkey tmcrypto.PubKey) (Info, error)

	Signer

	Importer
	Exporter
}

// Signer is implemented by key stores that want to provide signing capabilities.
type Signer interface {
	// Sign sign byte messages with a user key.
	Sign(uid string, msg []byte) ([]byte, tmcrypto.PubKey, error)

	// SignByAddress sign byte messages with a user key providing the address.
	SignByAddress(address sdk.Address, msg []byte) ([]byte, tmcrypto.PubKey, error)
}

// Importer is implemented by key stores that support import of public and private keys.
type Importer interface {
	// ImportPrivKey imports ASCII armored passphrase-encrypted private keys.
	ImportPrivKey(uid, armor, passphrase string) error
	// ImportPubKey imports ASCII armored public keys.
	ImportPubKey(uid string, armor string) error
}

// Exporter is implemented by key stores that support export of public and private keys.
type Exporter interface {
	// Export public key
	ExportPubKeyArmor(uid string) (string, error)
	ExportPubKeyArmorByAddress(address sdk.Address) (string, error)
	// ExportPrivKey returns a private key in ASCII armored format.
	// It returns an error if the key does not exist or a wrong encryption passphrase is supplied.
	ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error)
	ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error)
}

// Option overrides keyring configuration options.
type Option func(options *Options)

//Options define the options of the Keyring
type Options struct {
	SupportedAlgos       SigningAlgoList
	SupportedAlgosLedger SigningAlgoList
}

// NewInMemory creates a transient keyring useful for testing
// purposes and on-the-fly key generation.
// Keybase options can be applied when generating this new Keybase.
func NewInMemory(opts ...Option) Keyring {
	return newKeystore(keyring.NewArrayKeyring(nil), opts...)
}

// NewKeyring creates a new instance of a keyring.
// Keyring ptions can be applied when generating the new instance.
// Available backends are "os", "file", "kwallet", "memory", "pass", "test".
func New(
	appName, backend, rootDir string, userInput io.Reader, opts ...Option,
) (Keyring, error) {

	var db keyring.Keyring
	var err error

	switch backend {
	case BackendMemory:
		return NewInMemory(opts...), err
	case BackendTest:
		db, err = keyring.Open(newTestBackendKeyringConfig(appName, rootDir))
	case BackendFile:
		db, err = keyring.Open(newFileBackendKeyringConfig(appName, rootDir, userInput))
	case BackendOS:
		db, err = keyring.Open(newOSBackendKeyringConfig(appName, rootDir, userInput))
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

	return newKeystore(db, opts...), nil
}

type keystore struct {
	db      keyring.Keyring
	options Options
}

func newKeystore(kr keyring.Keyring, opts ...Option) keystore {
	// Default options for keybase
	options := Options{
		SupportedAlgos:       SigningAlgoList{hd.Secp256k1},
		SupportedAlgosLedger: SigningAlgoList{hd.Secp256k1},
	}

	for _, optionFn := range opts {
		optionFn(&options)
	}

	return keystore{kr, options}
}

func (ks keystore) ExportPubKeyArmor(uid string) (string, error) {
	bz, err := ks.Key(uid)
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name: %s", uid)
	}

	return crypto.ArmorPubKeyBytes(bz.GetPubKey().Bytes(), string(bz.GetAlgo())), nil
}

func (ks keystore) ExportPubKeyArmorByAddress(address sdk.Address) (string, error) {
	info, err := ks.KeyByAddress(address)
	if err != nil {
		return "", err
	}

	return ks.ExportPubKeyArmor(info.GetName())
}

func (ks keystore) ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error) {
	priv, err := ks.ExportPrivateKeyObject(uid)
	if err != nil {
		return "", err
	}

	info, err := ks.Key(uid)
	if err != nil {
		return "", err
	}

	return crypto.EncryptArmorPrivKey(priv, encryptPassphrase, string(info.GetAlgo())), nil
}

// ExportPrivateKeyObject exports an armored private key object.
func (ks keystore) ExportPrivateKeyObject(uid string) (tmcrypto.PrivKey, error) {
	info, err := ks.Key(uid)
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

func (ks keystore) ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error) {
	byAddress, err := ks.KeyByAddress(address)
	if err != nil {
		return "", err
	}

	return ks.ExportPrivKeyArmor(byAddress.GetName(), encryptPassphrase)
}

func (ks keystore) ImportPrivKey(uid, armor, passphrase string) error {
	if _, err := ks.Key(uid); err == nil {
		return fmt.Errorf("cannot overwrite key: %s", uid)
	}

	privKey, algo, err := crypto.UnarmorDecryptPrivKey(armor, passphrase)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt private key")
	}

	_, err = ks.writeLocalKey(uid, privKey, hd.PubKeyType(algo))
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) ImportPubKey(uid string, armor string) error {
	if _, err := ks.Key(uid); err == nil {
		return fmt.Errorf("cannot overwrite key: %s", uid)
	}

	pubBytes, algo, err := crypto.UnarmorPubKeyBytes(armor)
	if err != nil {
		return err
	}

	pubKey, err := cryptoAmino.PubKeyFromBytes(pubBytes)
	if err != nil {
		return err
	}

	_, err = ks.writeOfflineKey(uid, pubKey, hd.PubKeyType(algo))
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) Sign(uid string, msg []byte) ([]byte, tmcrypto.PubKey, error) {
	info, err := ks.Key(uid)
	if err != nil {
		return nil, nil, err
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
		return SignWithLedger(info, msg)

	case offlineInfo, multiInfo:
		return nil, info.GetPubKey(), errors.New("cannot sign with offline keys")
	}

	sig, err := priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil
}

func (ks keystore) SignByAddress(address sdk.Address, msg []byte) ([]byte, tmcrypto.PubKey, error) {
	key, err := ks.KeyByAddress(address)
	if err != nil {
		return nil, nil, err
	}

	return ks.Sign(key.GetName(), msg)
}

func (ks keystore) SaveLedgerKey(uid string, algo SignatureAlgo, hrp string, coinType, account, index uint32) (Info, error) {
	if !ks.options.SupportedAlgosLedger.Contains(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	hdPath := hd.NewFundraiserParams(account, coinType, index)

	priv, _, err := crypto.NewPrivKeyLedgerSecp256k1(*hdPath, hrp)
	if err != nil {
		return nil, err
	}

	return ks.writeLedgerKey(uid, priv.PubKey(), *hdPath, algo.Name())
}

func (ks keystore) writeLedgerKey(name string, pub tmcrypto.PubKey, path hd.BIP44Params, algo hd.PubKeyType) (Info, error) {
	info := newLedgerInfo(name, pub, path, algo)
	err := ks.writeInfo(info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (ks keystore) SaveMultisig(uid string, pubkey tmcrypto.PubKey) (Info, error) {
	return ks.writeMultisigKey(uid, pubkey)
}

func (ks keystore) SavePubKey(uid string, pubkey tmcrypto.PubKey, algo hd.PubKeyType) (Info, error) {
	return ks.writeOfflineKey(uid, pubkey, algo)
}

func (ks keystore) DeleteByAddress(address sdk.Address) error {
	info, err := ks.KeyByAddress(address)
	if err != nil {
		return err
	}

	err = ks.Delete(info.GetName())
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) Delete(uid string) error {
	info, err := ks.Key(uid)
	if err != nil {
		return err
	}

	err = ks.db.Remove(addrHexKeyAsString(info.GetAddress()))
	if err != nil {
		return err
	}

	err = ks.db.Remove(string(infoKey(uid)))
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) KeyByAddress(address sdk.Address) (Info, error) {
	ik, err := ks.db.Get(addrHexKeyAsString(address))
	if err != nil {
		return nil, err
	}

	if len(ik.Data) == 0 {
		return nil, fmt.Errorf("key with address %s not found", address)
	}

	bs, err := ks.db.Get(string(ik.Data))
	if err != nil {
		return nil, err
	}

	return unmarshalInfo(bs.Data)
}

func (ks keystore) List() ([]Info, error) {
	var res []Info
	keys, err := ks.db.Keys()
	if err != nil {
		return nil, err
	}

	sort.Strings(keys)

	for _, key := range keys {
		if strings.HasSuffix(key, infoSuffix) {
			rawInfo, err := ks.db.Get(key)
			if err != nil {
				return nil, err
			}

			if len(rawInfo.Data) == 0 {
				return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, key)
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

func (ks keystore) NewMnemonic(uid string, language Language, hdPath string, algo SignatureAlgo) (Info, string, error) {
	if language != English {
		return nil, "", ErrUnsupportedLanguage
	}

	if !ks.isSupportedSigningAlgo(algo) {
		return nil, "", ErrUnsupportedSigningAlgo
	}

	// Default number of words (24): This generates a mnemonic directly from the
	// number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(defaultEntropySize)
	if err != nil {
		return nil, "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", err
	}

	info, err := ks.NewAccount(uid, mnemonic, DefaultBIP39Passphrase, hdPath, algo)
	if err != nil {
		return nil, "", err
	}

	return info, mnemonic, err
}

func (ks keystore) NewAccount(uid string, mnemonic string, bip39Passphrase string, hdPath string, algo SignatureAlgo) (Info, error) {
	if !ks.isSupportedSigningAlgo(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	// create master key and derive first key for keyring
	derivedPriv, err := algo.Derive()(mnemonic, bip39Passphrase, hdPath)
	if err != nil {
		return nil, err
	}

	privKey := algo.Generate()(derivedPriv)

	return ks.writeLocalKey(uid, privKey, algo.Name())
}

func (ks keystore) isSupportedSigningAlgo(algo SignatureAlgo) bool {
	return ks.options.SupportedAlgos.Contains(algo)
}

func (ks keystore) Key(uid string) (Info, error) {
	key := infoKey(uid)

	bs, err := ks.db.Get(string(key))
	if err != nil {
		return nil, err
	}

	if len(bs.Data) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, uid)
	}

	return unmarshalInfo(bs.Data)
}

// SignWithLedger signs a binary message with the ledger device referenced by an Info object
// and returns the signed bytes and the public key. It returns an error if the device could
// not be queried or it returned an error.
func SignWithLedger(info Info, msg []byte) (sig []byte, pub tmcrypto.PubKey, err error) {
	switch info.(type) {
	case *ledgerInfo, ledgerInfo:
	default:
		return nil, nil, errors.New("not a ledger object")
	}
	path, err := info.GetPath()
	if err != nil {
		return
	}

	priv, err := crypto.NewPrivKeyLedgerSecp256k1Unsafe(*path)
	if err != nil {
		return
	}

	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil
}

func newOSBackendKeyringConfig(appName, dir string, buf io.Reader) keyring.Config {
	return keyring.Config{
		ServiceName:      appName,
		FileDir:          dir,
		FilePasswordFunc: newRealPrompt(dir, buf),
	}
}

func newTestBackendKeyringConfig(appName, dir string) keyring.Config {
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		ServiceName:     appName,
		FileDir:         filepath.Join(dir, keyringTestDirName),
		FilePasswordFunc: func(_ string) (string, error) {
			return "test", nil
		},
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

func newPassBackendKeyringConfig(appName, _ string, _ io.Reader) keyring.Config {
	prefix := fmt.Sprintf(passKeyringPrefix, appName)
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.PassBackend},
		ServiceName:     appName,
		PassPrefix:      prefix,
	}
}

func newFileBackendKeyringConfig(name, dir string, buf io.Reader) keyring.Config {
	fileDir := filepath.Join(dir, keyringFileDirName)
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

			saltBytes := tmcrypto.CRandBytes(16)
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

func (ks keystore) writeLocalKey(name string, priv tmcrypto.PrivKey, algo hd.PubKeyType) (Info, error) {
	// encrypt private key using keyring
	pub := priv.PubKey()

	info := newLocalInfo(name, pub, string(priv.Bytes()), algo)
	err := ks.writeInfo(info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (ks keystore) writeInfo(info Info) error {
	// write the info by key
	key := infoKey(info.GetName())
	serializedInfo := marshalInfo(info)

	exists, err := ks.existsInDb(info)
	if exists {
		return fmt.Errorf("public key already exist in keybase")
	}
	if err != nil {
		return err
	}

	err = ks.db.Set(keyring.Item{
		Key:  string(key),
		Data: serializedInfo,
	})
	if err != nil {
		return err
	}

	err = ks.db.Set(keyring.Item{
		Key:  addrHexKeyAsString(info.GetAddress()),
		Data: key,
	})
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) existsInDb(info Info) (bool, error) {
	if _, err := ks.db.Get(addrHexKeyAsString(info.GetAddress())); err == nil {
		return true, nil // address lookup succeeds - info exists
	} else if err != keyring.ErrKeyNotFound {
		return false, err // received unexpected error - returns error
	}

	if _, err := ks.db.Get(string(infoKey(info.GetName()))); err == nil {
		return true, nil // uid lookup succeeds - info exists
	} else if err != keyring.ErrKeyNotFound {
		return false, err // received unexpected error - returns
	}

	// both lookups failed, info does not exist
	return false, nil
}

func (ks keystore) writeOfflineKey(name string, pub tmcrypto.PubKey, algo hd.PubKeyType) (Info, error) {
	info := newOfflineInfo(name, pub, algo)
	err := ks.writeInfo(info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (ks keystore) writeMultisigKey(name string, pub tmcrypto.PubKey) (Info, error) {
	info := NewMultiInfo(name, pub)
	err := ks.writeInfo(info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func addrHexKeyAsString(address sdk.Address) string {
	return fmt.Sprintf("%s.%s", hex.EncodeToString(address.Bytes()), addressSuffix)
}
