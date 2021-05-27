package keyring

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/99designs/keyring"
	bip39 "github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/tendermint/crypto/bcrypt"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/ledger"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Backend options for Keyring
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

// Keyring version
const (
	CURRENT_VERSION = 1
)

var (
	_                          Keyring = &keystore{}
	maxPassphraseEntryAttempts         = 3
	VERSION_KEY                        = string([]byte{0})
)


// Keyring exposes operations over a backend supported by github.com/99designs/keyring.
type Keyring interface {
	// List all keys.
	List() ([]*Record, error)

	// Supported signing algorithms for Keyring and Ledger respectively.
	SupportedAlgorithms() (SigningAlgoList, SigningAlgoList)

	// Key and KeyByAddress return keys by uid and address respectively.
	Key(uid string) (*Record, error)
	KeyByAddress(address sdk.Address) (*Record, error)

	// Delete and DeleteByAddress remove keys from the keyring.
	Delete(uid string) error
	DeleteByAddress(address sdk.Address) error

	// NewMnemonic generates a new mnemonic, derives a hierarchical deterministic key from it, and
	// persists the key to storage. Returns the generated mnemonic and the key Info.
	// It returns an error if it fails to generate a key for the given algo type, or if
	// another key is already stored under the same name or address.
	//
	// A passphrase set to the empty string will set the passphrase to the DefaultBIP39Passphrase value.
	NewMnemonic(uid string, language Language, hdPath, bip39Passphrase string, algo SignatureAlgo) (*Record, string, error)

	// NewAccount converts a mnemonic to a private key and BIP-39 HD Path and persists it.
	// It fails if there is an existing key Info with the same address.
	NewAccount(uid, mnemonic, bip39Passphrase, hdPath string, algo SignatureAlgo) (*Record, error)

	// SaveLedgerKey retrieves a public key reference from a Ledger device and persists it.
	SaveLedgerKey(uid string, algo SignatureAlgo, hrp string, coinType, account, index uint32) (*Record, error)

	// SavePubKey stores a public key and returns the persisted Info structure.
	// TODO it persists only Offline key, does it sound ok?
	SavePubKey(uid string, pubkey types.PubKey) (*Record, error)

	// SaveMultisig stores and returns a new multsig (offline) key reference.
	SaveMultisig(uid string, pubkey types.PubKey) (*Record, error)

	Signer

	Importer
	Exporter
}

// UnsafeKeyring exposes unsafe operations such as unsafe unarmored export in
// addition to those that are made available by the Keyring interface.
type UnsafeKeyring interface {
	Keyring
	UnsafeExporter
}

// Signer is implemented by key stores that want to provide signing capabilities.
type Signer interface {
	// Sign sign byte messages with a user key.
	Sign(uid string, msg []byte) ([]byte, types.PubKey, error)

	// SignByAddress sign byte messages with a user key providing the address.
	SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error)
}

// Importer is implemented by key stores that support import of public and private keys.
type Importer interface {
	// ImportPrivKey imports ASCII armored passphrase-encrypted private keys.
	ImportPrivKey(uid, armor, passphrase string) error

	// ImportPubKey imports ASCII armored public keys.
	ImportPubKey(uid string, armor string) error
}


// TODO try to remove to see where it is used
// LegacyInfoImporter is implemented by key stores that support import of Info types.
type LegacyInfoImporter interface {
	// ImportInfo import a keyring.Info into the current keyring.
	// It is used to migrate multisig, ledger, and public key Info structure.
	ImportInfo(oldInfo LegacyInfo) error
}

// Exporter is implemented by key stores that support export of public and private keys.
type Exporter interface {
	// Export public key
	ExportPubKeyArmor(uid string) (string, error)
	ExportPubKeyArmorByAddress(address sdk.Address) (string, error)

	// ExportPrivKeyArmor returns a private key in ASCII armored format.
	// It returns an error if the key does not exist or a wrong encryption passphrase is supplied.
	ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error)
	ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error)
}

// UnsafeExporter is implemented by key stores that support unsafe export
// of private keys' material.
type UnsafeExporter interface {
	// UnsafeExportPrivKeyHex returns a private key in unarmored hex format
	UnsafeExportPrivKeyHex(uid string) (string, error)
}

// Option overrides keyring configuration options.
type Option func(options *Options)

// Options define the options of the Keyring.
type Options struct {
	// supported signing algorithms for keyring
	SupportedAlgos SigningAlgoList
	// supported signing algorithms for Ledger
	SupportedAlgosLedger SigningAlgoList
}

// NewInMemory creates a transient keyring useful for testing
// purposes and on-the-fly key generation.
// Keybase options can be applied when generating this new Keybase.
func NewInMemory(cdc codec.Codec, opts ...Option) Keyring {
	return newKeystore(keyring.NewArrayKeyring(nil), cdc, opts...)
}

// New creates a new instance of a keyring.
// Keyring ptions can be applied when generating the new instance.
// Available backends are "os", "file", "kwallet", "memory", "pass", "test".
func New(
	appName, backend, rootDir string, userInput io.Reader, cdc codec.Codec, opts ...Option,
) (Keyring, error) {
	var (
		db  keyring.Keyring
		err error
	)

	switch backend {
	case BackendMemory:
		return NewInMemory(cdc, opts...), err
	case BackendTest:
		db, err = keyring.Open(newTestBackendKeyringConfig(appName, rootDir))
	case BackendFile:
		db, err = keyring.Open(newFileBackendKeyringConfig(appName, rootDir, userInput))
	case BackendOS:
		db, err = keyring.Open(newOSBackendKeyringConfig(appName, rootDir, userInput))
	case BackendKWallet:
		db, err = keyring.Open(NewKWalletBackendKeyringConfig(appName, rootDir, userInput))
	case BackendPass:
		db, err = keyring.Open(NewPassBackendKeyringConfig(appName, rootDir, userInput))
	default:
		return nil, fmt.Errorf("unknown keyring backend %v", backend)
	}

	if err != nil {
		return nil, err
	}

	return newKeystore(db, cdc, opts...), nil
}

type keystore struct {
	db      keyring.Keyring
	cdc     codec.Codec
	options Options
}

func newKeystore(kr keyring.Keyring, cdc codec.Codec, opts ...Option) keystore {
	// Default options for keybase
	options := Options{
		SupportedAlgos:       SigningAlgoList{hd.Secp256k1},
		SupportedAlgosLedger: SigningAlgoList{hd.Secp256k1},
	}

	for _, optionFn := range opts {
		optionFn(&options)
	}

	return keystore{kr, cdc, options}
}

func (ks keystore) ExportPubKeyArmor(uid string) (string, error) {
	bz, err := ks.Key(uid)
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name: %s", uid)
	}
	key, err := bz.GetPubKey()
	if err != nil {
		return "", err
	}
	return crypto.ArmorPubKeyBytes(legacy.Cdc.MustMarshal(key), string(bz.GetAlgo())), nil
}

func (ks keystore) ExportPubKeyArmorByAddress(address sdk.Address) (string, error) {
	re, err := ks.KeyByAddress(address)
	if err != nil {
		return "", err
	}

	return ks.ExportPubKeyArmor(re.GetName())
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
func (ks keystore) ExportPrivateKeyObject(uid string) (types.PrivKey, error) {
	re, err := ks.Key(uid)
	if err != nil {
		return nil, err
	}

	return re.extractPrivKeyFromLocalInfo()
}

func (ks keystore) ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error) {
	re, err := ks.KeyByAddress(address)
	if err != nil {
		return "", err
	}

	return ks.ExportPrivKeyArmor(re.GetName(), encryptPassphrase)
}

func (ks keystore) ImportPrivKey(uid, armor, passphrase string) error {
	if _, err := ks.Key(uid); err == nil {
		return fmt.Errorf("cannot overwrite key: %s", uid)
	}

	privKey, algo, err := crypto.UnarmorDecryptPrivKey(armor, passphrase)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt private key")
	}

	_, err = ks.writeLocalKey(uid, privKey, algo)
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) ImportPubKey(uid string, armor string) error {
	if _, err := ks.Key(uid); err == nil {
		return fmt.Errorf("cannot overwrite key: %s", uid)
	}

	pubBytes, _, err := crypto.UnarmorPubKeyBytes(armor)
	if err != nil {
		return err
	}

	pubKey, err := legacy.PubKeyFromBytes(pubBytes)
	if err != nil {
		return err
	}

	_, err = ks.writeOfflineKey(uid, pubKey)
	if err != nil {
		return err
	}

	return nil
}


// ImportInfo implements Importer.MigrateInfo.
// TODO do i need it or not
/*
func (ks keystore) ImportInfo(oldInfo LegacyInfo) error {
	if _, err := ks.Key(oldInfo.GetName()); err == nil {
		return fmt.Errorf("cannot overwrite key: %s", oldInfo.GetName())
	}

	return ks.writeInfo(oldInfo)
}
*/

func (ks keystore) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	ke, err := ks.Key(uid)
	if err != nil {
		return nil, nil, err
	}

	priv, err := ke.extractPrivKeyFromLocalInfo()
	if err != nil {
		return nil, nil, err
	}

	sig, err := priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil

}

func (ks keystore) SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error) {
	ke, err := ks.KeyByAddress(address)
	if err != nil {
		return nil, nil, err
	}

	return ks.Sign(ke.GetName(), msg)
}


func (ks keystore) SaveLedgerKey(uid string, algo SignatureAlgo, hrp string, coinType, account, index uint32) (*Record, error) {

	if !ks.options.SupportedAlgosLedger.Contains(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	hdPath := hd.NewFundraiserParams(account, coinType, index)

	priv, _, err := ledger.NewPrivKeySecp256k1(*hdPath, hrp)
	if err != nil {
		return nil, err
	}

	apk, err := codectypes.NewAnyWithValue(priv.PubKey())
	if err != nil {
		return nil, err
	}

	return ks.writeLedgerKey(uid, apk, hdPath)
}

func (ks keystore) writeLedgerKey(name string, pubKey *codectypes.Any, path *hd.BIP44Params,) (*Record, error) {

	ledgerInfo := newLedgerInfo(path)
	ledgerInfoItem := newLedgerInfoItem(ledgerInfo)
	re := NewRecord(name, pubKey, ledgerInfoItem)
	if err := ks.writeRecord(re); err != nil {
		return nil, err
	}

	return re, nil
}

func (ks keystore) SaveMultisig(uid string, pubkey types.PubKey) (*Record, error) {
	return ks.writeMultisigKey(uid, pubkey)
}

func (ks keystore) SavePubKey(uid string, pubkey types.PubKey) (*Record, error) {
	return ks.writeOfflineKey(uid, pubkey)
}

func (ks keystore) DeleteByAddress(address sdk.Address) error {
	ke, err := ks.KeyByAddress(address)
	if err != nil {
		return err
	}

	err = ks.Delete(ke.GetName())
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) Delete(uid string) error {
	ke, err := ks.Key(uid)
	if err != nil {
		return err
	}

	addr, err := ke.GetAddress()
	if err != nil {
		return err
	}

	err = ks.db.Remove(addrHexKeyAsString(addr))
	if err != nil {
		return err
	}

	err = ks.db.Remove(infoKey(uid))
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) KeyByAddress(address sdk.Address) (*Record, error) {
	ik, err := ks.db.Get(addrHexKeyAsString(address))
	if err != nil {
		return nil, wrapKeyNotFound(err, fmt.Sprint("key with address", address, "not found"))
	}

	if len(ik.Data) == 0 {
		return nil, wrapKeyNotFound(err, fmt.Sprint("key with address", address, "not found"))
	}
	return ks.key(string(ik.Data))
}

func wrapKeyNotFound(err error, msg string) error {
	if err == keyring.ErrKeyNotFound {
		return sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, msg)
	}
	return err
}

func (ks keystore) List() ([]*Record, error) {
	if err := ks.checkMigrate(); err != nil {
		return nil, err
	}

	var res []*Record

	keys, err := ks.db.Keys()
	if err != nil {
		return nil, err
	}

	sort.Strings(keys)
	for _, key := range keys {
		if !strings.HasSuffix(key, infoSuffix) {
			continue
		}

		item, err := ks.db.Get(key)
		if err != nil {
			return nil, err
		}

		if len(item.Data) == 0 {
			return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, key)
		}
		re := new(Record)
		if err := ks.cdc.Unmarshal(item.Data, re); err != nil {
			return nil, err
		}
		res = append(res, re)
	}

	return res, nil
}

func (ks keystore) NewMnemonic(uid string, language Language, hdPath, bip39Passphrase string, algo SignatureAlgo) (*Record, string, error) {
	if language != English {
		return nil, "", ErrUnsupportedLanguage
	}

	if !ks.isSupportedSigningAlgo(algo) {
		return nil, "", ErrUnsupportedSigningAlgo
	}

	// Default number of words (24): This generates a mnemonic directly from the
	// number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(DefaultEntropySize)
	if err != nil {
		return nil, "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", err
	}

	if bip39Passphrase == "" {
		bip39Passphrase = DefaultBIP39Passphrase
	}

	ke, err := ks.NewAccount(uid, mnemonic, bip39Passphrase, hdPath, algo)
	if err != nil {
		return nil, "", err
	}

	return ke, mnemonic, nil
}
// TODO keep or remove SignatureAlgo???
func (ks keystore) NewAccount(name string, mnemonic string, bip39Passphrase string, hdPath string, algo SignatureAlgo) (*Record, error) {
	if !ks.isSupportedSigningAlgo(algo) {
		return nil, ErrUnsupportedSigningAlgo
	}

	// create master key and derive first key for keyring
	derivedPriv, err := algo.Derive()(mnemonic, bip39Passphrase, hdPath)
	if err != nil {
		return nil, err
	}

	privKey := algo.Generate()(derivedPriv)

	// check if the a key already exists with the same address and return an error
	// if found
	address := sdk.AccAddress(privKey.PubKey().Address())
	if _, err := ks.KeyByAddress(address); err == nil {
		return nil, fmt.Errorf("account with address %s already exists in keyring, delete the key first if you want to recreate it", address)
	}
	
	return ks.writeLocalKey(name, privKey, string(algo.Name()))
}

func (ks keystore) isSupportedSigningAlgo(algo SignatureAlgo) bool {
	return ks.options.SupportedAlgos.Contains(algo)
}

func (ks keystore) key(infoKey string) (*Record, error) {
	if err := ks.checkMigrate(); err != nil {
		return nil, err
	}
	bs, err := ks.db.Get(infoKey)
	if err != nil {
		return nil, wrapKeyNotFound(err, infoKey)
	}
	if len(bs.Data) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, infoKey)
	}

	re := new(Record)
	if err := ks.cdc.Unmarshal(bs.Data, re); err != nil {
		return nil, err
	}
	return re, nil
	//	return protoUnmarshalInfo(bs.Data, ks.cdc)
}

func (ks keystore) Key(uid string) (*Record, error) {
	return ks.key(infoKey(uid))
}

// SupportedAlgorithms returns the keystore Options' supported signing algorithm.
// for the keyring and Ledger.
func (ks keystore) SupportedAlgorithms() (SigningAlgoList, SigningAlgoList) {
	return ks.options.SupportedAlgos, ks.options.SupportedAlgosLedger
}

// SignWithLedger signs a binary message with the ledger device referenced by an Info object
// and returns the signed bytes and the public key. It returns an error if the device could
// not be queried or it returned an error.
func SignWithLedger(re *Record, msg []byte) (sig []byte, pub types.PubKey, err error) {

	ledgerInfo := re.GetLedger()
	if ledgerInfo == nil {
		return nil, nil, errors.New("not a ledger object")
	}
	// TODO remove GetPath from keyringEntry, cause it relates to Ledger only
	path := ledgerInfo.GetPath()

	// TODO should I fix replace type from hd.BIP44Params to keyring.BUP44Params in NewPrivKeySecp256k1Unsafe
	priv, err := ledger.NewPrivKeySecp256k1Unsafe(*path)
	if err != nil {
		return
	}

	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil

	/*
		switch info.(type) {
		case *LedgerInfo, LedgerInfo:
		default:
			return nil, nil, errors.New("not a ledger object")
		}

		path, err := info.GetPath()
		if err != nil {
			return err
		}

		priv, err := ledger.NewPrivKeySecp256k1Unsafe(*path)
		if err != nil {
			return
		}

		sig, err = priv.Sign(msg)
		if err != nil {
			return nil, nil, err
		}

		return sig, priv.PubKey(), nil
	*/
}

func newOSBackendKeyringConfig(appName, dir string, buf io.Reader) keyring.Config {
	return keyring.Config{
		ServiceName:              appName,
		FileDir:                  dir,
		KeychainTrustApplication: true,
		FilePasswordFunc:         newRealPrompt(dir, buf),
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

func NewKWalletBackendKeyringConfig(appName, _ string, _ io.Reader) keyring.Config {
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.KWalletBackend},
		ServiceName:     "kdewallet",
		KWalletAppID:    appName,
		KWalletFolder:   "",
	}
}

func NewPassBackendKeyringConfig(appName, _ string, _ io.Reader) keyring.Config {
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
				// NOTE: LGTM.io reports a false positive alert that states we are printing the password,
				// but we only log the error.
				//
				// lgtm [go/clear-text-logging]
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
				// NOTE: LGTM.io reports a false positive alert that states we are printing the password,
				// but we only log the error.
				//
				// lgtm [go/clear-text-logging]
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

func (ks keystore) writeLocalKey(name string, priv types.PrivKey, pubKeyType string) (*Record, error) {
	// encrypt private key using keyring

	apk, err := codectypes.NewAnyWithValue(priv)
	if err != nil {
		return nil, err
	}
	localInfo := newLocalInfo(apk, pubKeyType)
	localInfoItem := newLocalInfoItem(localInfo)
	re := NewRecord(name, apk, localInfoItem)
	if err := ks.writeRecord(re); err != nil {
		return nil, err
	}
	/*
		pub := priv.PubKey()
		info := newLocalInfo(name, pub, string(legacy.Cdc.MustMarshal(priv)), algo)
		if err := ks.writeRecord(info); err != nil {
			return nil, err
		}
		return info, nil
	*/

	return re, nil
}

// declare writeRecord(re Record)
func (ks keystore) writeRecord(re *Record) error {
	key := infoKeyBz(re.GetName())

	serializedRecord, err := ks.cdc.Marshal(re)
	if err != nil {
		return err
	}

	exists, err := ks.existsInDb(re)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("public key already exist in keybase")
	}

	err = ks.db.Set(keyring.Item{
		Key:  string(key),
		Data: serializedRecord,
	})
	if err != nil {
		return err
	}

	addr, err := re.GetAddress()
	if err != nil {
		return err
	}

	err = ks.db.Set(keyring.Item{
		Key:  addrHexKeyAsString(addr),
		Data: key,
	})
	if err != nil {
		return err
	}

	return nil
}

// existsInDb returns true if key is in DB. Error is returned only when we have error
// different thant ErrKeyNotFound
func (ks keystore) existsInDb(re *Record) (bool, error) {
	addr, err := re.GetAddress()
	if err != nil {
		return false, err
	}
	if _, err := ks.db.Get(addrHexKeyAsString(addr)); err == nil {
		return true, nil // address lookup succeeds - info exists
	} else if err != keyring.ErrKeyNotFound {
		return false, err // received unexpected error - returns error
	}

	if _, err := ks.db.Get(infoKey(re.GetName())); err == nil {
		return true, nil // uid lookup succeeds - info exists
	} else if err != keyring.ErrKeyNotFound {
		return false, err // received unexpected error - returns
	}

	// both lookups failed, info does not exist
	return false, nil
}

func (ks keystore) writeOfflineKey(name string, pub types.PubKey) (*Record, error) {
	apk, err := codectypes.NewAnyWithValue(pub)
	if err != nil {
		return nil, err
	}
	offlineInfo := newOfflineInfo()
	offlineInfoItem := newOfflineInfoItem(offlineInfo)
	ke := NewRecord(name, apk, offlineInfoItem)

	if err := ks.writeRecord(ke); err != nil {
		return nil, err
	}

	return ke, nil

	/*
		info := newOfflineInfo(name, pub, algo)
		err := ks.writeRecord(info)
		if err != nil {
			return nil, err
		}

		return info, nil
	*/
}

// TODO writeRecord
// writeMultisigKey investigate where thisf function is called maybe remove it
func (ks keystore) writeMultisigKey(name string, pub types.PubKey) (*Record, error) {
	apk, err := codectypes.NewAnyWithValue(pub)
	if err != nil {
		return nil, err
	}

	multiInfo := newMultiInfo()
	multiInfoItem := newMultiInfoItem(multiInfo)
	re := NewRecord(name, apk, multiInfoItem)
	if err = ks.writeRecord(re); err != nil {
		return nil, err
	}

	return re, nil

	/*
		info, err := NewMultiInfo(name, pub)
		if err != nil {
			return nil, err
		}
		if err = ks.writeRecord(info); err != nil {
			return nil, err
		}

		return info, nil
	*/
}

func (ks keystore) checkMigrate() error {
	var version uint32 = 0
	// 1.Get a key data
	item, err := ks.db.Get(VERSION_KEY)
	if err != nil {
		if err != keyring.ErrKeyNotFound {
			return err
		}
		// key not found, all good: assume version = 0
	} else {
		if len(item.Data) != 4 {
			return sdkerrors.ErrInvalidVersion.Wrapf(
				"Can't migrate the keyring - the stored version is malformed: [%v]: %v",
				item.Description, string(item.Data))
		}
		version = binary.LittleEndian.Uint32(item.Data)
	}
	return ks.migrate(version, item)
}

func (ks keystore) migrate(version uint32, i keyring.Item) error {
	if version == CURRENT_VERSION {
		return nil
	}
	if version > CURRENT_VERSION {
		return sdkerrors.ErrInvalidVersion.Wrapf(
			"Can't migrate the keyring - wrong keyring version: [%v]: %v, expected version to be max %d",
			i.Description, string(i.Data), CURRENT_VERSION)
	}
	keys, err := ks.db.Keys()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if !strings.HasSuffix(key, infoSuffix) {
			continue
		}
		item, err := ks.db.Get(key)
		if err != nil {
			return err
		}

		if len(item.Data) == 0 {
			return sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, key)
		}

		//2.try to deserialize using amino not proto
	    //3.if any above will fail - then we will need to log an error and continue, there can be other reasons, or a key could be already migrated to proto.
	 	_, err = unmarshalInfo(item.Data)
		if err != nil {
			// should i return or or justp rint an eror in this case
		}

		//4.serialize info using proto
	    // are you sure that i have to serialize in this case? not deserrialize using proto
		var re Record
		if err := ks.cdc.Unmarshal(item.Data, &re); err != nil {
			return err
		}

		var versionBytes = make([]byte, 4)
		binary.LittleEndian.PutUint32(versionBytes, CURRENT_VERSION)
		key, err := re.GetPubKey()
		if err != nil {
			return err
		}

		//5.overwrite the keyring entry with
		ks.db.Set(keyring.Item{
			Key:         key.String(),
			Data:        versionBytes,
			Description: "SDK kerying version",
		})
		/*
			var versionBytes = make([]byte, 4)
			binary.LittleEndian.PutUint32(versionBytes, CURRENT_VERSION)
			ks.db.Set(keyring.Item{
				Key:         info.GetKey().String(),
				Data:        versionBytes,
				Description: "SDK kerying version",
			})
		*/
	}
    // 6. at the end of the loop update version
	var versionBytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(versionBytes, CURRENT_VERSION)
	ks.db.Set(keyring.Item{
		Key:         "migration",
		Data:        versionBytes,
		Description: "SDK kerying version",
	})

	return nil
}

type unsafeKeystore struct {
	keystore
}

// NewUnsafe returns a new keyring that provides support for unsafe operations.
func NewUnsafe(kr Keyring) UnsafeKeyring {
	// The type assertion is against the only keystore
	// implementation that is currently provided.
	ks := kr.(keystore)

	return unsafeKeystore{ks}
}

// UnsafeExportPrivKeyHex exports private keys in unarmored hexadecimal format.
func (ks unsafeKeystore) UnsafeExportPrivKeyHex(uid string) (privkey string, err error) {
	priv, err := ks.ExportPrivateKeyObject(uid)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(priv.Bytes()), nil
}

func addrHexKeyAsString(address sdk.Address) string {
	return fmt.Sprintf("%s.%s", hex.EncodeToString(address.Bytes()), addressSuffix)
}
