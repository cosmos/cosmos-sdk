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

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
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

var (
	_                          Keyring = &keystore{}
	maxPassphraseEntryAttempts         = 3
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

	// SaveOfflineKey stores a public key and returns the persisted Info structure.
	SaveOfflineKey(uid string, pubkey types.PubKey) (*Record, error)

	// SaveMultisig stores and returns a new multsig (offline) key reference.
	SaveMultisig(uid string, pubkey types.PubKey) (*Record, error)

	Signer

	Importer
	Exporter

	Migrator
	Setter
	Marshaler
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

type Migrator interface {
	MigrateAll() (bool, error)
	Migrate(key string) (bool, error)
}

// used in migration_test.go
type Setter interface {
	SetItem(item keyring.Item) error
}

// TODO if it is onlyfo tests it should not bei n public interface
// used in migration_test.go
type Marshaler interface {
	ProtoMarshalRecord(k *Record) ([]byte, error)
	ProtoUnmarshalRecord(bz []byte) (*Record, error)
}

// Exporter is implemented by key stores that support export of public and private keys.
// TODO decide if need exporter interface as it is used only in keyring_test.go
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
	k, err := ks.Key(uid)
	if err != nil {
		return "", err
	}

	key, err := k.GetPubKey()
	if err != nil {
		return "", err
	}

	return crypto.ArmorPubKeyBytes(legacy.Cdc.MustMarshal(key), key.Type()), nil
}

func (ks keystore) ExportPubKeyArmorByAddress(address sdk.Address) (string, error) {
	k, err := ks.KeyByAddress(address)
	if err != nil {
		return "", err
	}

	return ks.ExportPubKeyArmor(k.Name)
}


// we use ExportPrivateKeyFromLegacyInfo(info LegacyInfo) (cryptotypes.PrivKey, error) { for LegacyInfo
func (ks keystore) ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error) {
	fmt.Println("ExportPrivKeyArmor start")
	_, priv, err := ks.ExportPrivateKeyObject(uid)
	if err != nil {
		return "", err
	}
	fmt.Println("ExportPrivKeyArmor done")

	return crypto.EncryptArmorPrivKey(priv, encryptPassphrase, priv.Type()), nil
}

// ExportPrivateKeyObject exports an armored private key object.
func (ks keystore) ExportPrivateKeyObject(uid string) (*Record, types.PrivKey, error) {
	k, err := ks.Key(uid)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("ExportPrivateKeyObject ks.Key done")
	priv, err := ExtractPrivKeyFromRecord(k)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("priv", priv)

	return k, priv, err
}

func (ks keystore) ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error) {
	k, err := ks.KeyByAddress(address)
	if err != nil {
		return "", err
	}

	return ks.ExportPrivKeyArmor(k.Name, encryptPassphrase)
}

func (ks keystore) ImportPrivKey(uid, armor, passphrase string) error {
	if _, err := ks.Key(uid); err == nil {
		return fmt.Errorf("cannot overwrite key: %s", uid)
	}

	privKey, _, err := crypto.UnarmorDecryptPrivKey(armor, passphrase)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt private key")
	}

	_, err = ks.writeLocalKey(uid, privKey)
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
// TODO do i need it or not?
/*
func (ks keystore) ImportInfo(oldInfo LegacyInfo) error {
	if _, err := ks.Key(oldInfo.Name); err == nil {
		return fmt.Errorf("cannot overwrite key: %s", oldInfo.Name)
	}

	return ks.writeLegacyInfo(oldInfo)
}
*/

func (ks keystore) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	k, err := ks.Key(uid)
	if err != nil {
		return nil, nil, err
	}

	var priv types.PrivKey

	switch {
	case k.GetLocal() != nil:
		priv, err = extractPrivKeyFromLocal(k.GetLocal())
		if err != nil {
			return nil, nil, err
		}

	case k.GetLedger() != nil:
		return SignWithLedger(k, msg)

		// empty record
	default:
		pub, err := k.GetPubKey()
		if err != nil {
			return nil, nil, err
		}

		return nil, pub, errors.New("cannot sign with offline keys")
	}

	sig, err := priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil

}

func (ks keystore) SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error) {
	k, err := ks.KeyByAddress(address)
	if err != nil {
		return nil, nil, err
	}

	return ks.Sign(k.Name, msg)
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

	return ks.writeLedgerKey(uid, priv.PubKey(), hdPath)
}

func (ks keystore) writeLedgerKey(name string, pk types.PubKey, path *hd.BIP44Params) (*Record, error) {
	ledgerRecord := NewLedgerRecord(path)
	ledgerRecordItem := NewLedgerRecordItem(ledgerRecord)
	return ks.newRecord(name, pk, ledgerRecordItem)
}

func (ks keystore) SaveMultisig(uid string, pubkey types.PubKey) (*Record, error) {
	return ks.writeMultisigKey(uid, pubkey)
}

func (ks keystore) SaveOfflineKey(uid string, pubkey types.PubKey) (*Record, error) {
	return ks.writeOfflineKey(uid, pubkey)
}

func (ks keystore) DeleteByAddress(address sdk.Address) error {
	k, err := ks.KeyByAddress(address)
	if err != nil {
		return err
	}

	err = ks.Delete(k.Name)
	if err != nil {
		return err
	}

	return nil
}

func (ks keystore) Delete(uid string) error {
	k, err := ks.Key(uid)
	if err != nil {
		return err
	}

	addr, err := k.GetAddress()
	if err != nil {
		return err
	}

	err = ks.db.Remove(addrHexKeyAsString(addr))
	if err != nil {
		return err
	}

	err = ks.db.Remove(uid)
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

	return ks.Key(string(ik.Data))
}

func wrapKeyNotFound(err error, msg string) error {
	if err == keyring.ErrKeyNotFound {
		return sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, msg)
	}
	return err
}

func (ks keystore) List() ([]*Record, error) {
	if _, err := ks.MigrateAll(); err != nil {
		return nil, err
	}

	keys, err := ks.db.Keys()
	if err != nil {
		return nil, err
	}

	var res []*Record
	sort.Strings(keys)
	for _, key := range keys {
		if strings.Contains(key, addressSuffix) {
			continue
		}

		item, err := ks.db.Get(key)
		if err != nil {
			return nil, err
		}

		if len(item.Data) == 0 {
			return nil, sdkerrors.ErrKeyNotFound.Wrap(key)
		}

		k, err := ks.ProtoUnmarshalRecord(item.Data)
		if err != nil {
			return nil, err
		}

		res = append(res, k)
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

	k, err := ks.NewAccount(uid, mnemonic, bip39Passphrase, hdPath, algo)
	if err != nil {
		return nil, "", err
	}

	return k, mnemonic, nil
}

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
		return nil, errors.New("duplicated address created")
	}

	return ks.writeLocalKey(name, privKey)
}

func (ks keystore) isSupportedSigningAlgo(algo SignatureAlgo) bool {
	return ks.options.SupportedAlgos.Contains(algo)
}

func (ks keystore) Key(uid string) (*Record, error) {
	if _, err := ks.Migrate(uid); err != nil {
		return nil, err
	}

	item, err := ks.db.Get(uid)
	if err != nil {
		return nil, wrapKeyNotFound(err, uid)
	}

	if len(item.Data) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, uid)
	}

	return ks.ProtoUnmarshalRecord(item.Data)
}

// SupportedAlgorithms returns the keystore Options' supported signing algorithm.
// for the keyring and Ledger.
func (ks keystore) SupportedAlgorithms() (SigningAlgoList, SigningAlgoList) {
	return ks.options.SupportedAlgos, ks.options.SupportedAlgosLedger
}

// SignWithLedger signs a binary message with the ledger device referenced by an Info object
// and returns the signed bytes and the public key. It returns an error if the device could
// not be queried or it returned an error.
func SignWithLedger(k *Record, msg []byte) (sig []byte, pub types.PubKey, err error) {
	ledgerInfo := k.GetLedger()
	if ledgerInfo == nil {
		return nil, nil, errors.New("not a ledger object")
	}

	path := ledgerInfo.GetPath()

	priv, err := ledger.NewPrivKeySecp256k1Unsafe(*path)
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

func (ks keystore) writeLocalKey(name string, privKey types.PrivKey) (*Record, error) {
	localRecord, err := NewLocalRecord(privKey)
	if err != nil {
		return nil, err
	}

	localRecordItem := NewLocalRecordItem(localRecord)
	return ks.newRecord(name, privKey.PubKey(), localRecordItem)
}

func (ks keystore) writeRecord(k *Record) error {
	addr, err := k.GetAddress()
	if err != nil {
		return err
	}

	key := k.Name

	exists, err := ks.existsInDb(addr, string(key))
	if err != nil {
		return err
	}
	if exists {
		return errors.New("public key already exists in keybase")
	}

	serializedRecord, err := ks.cdc.Marshal(k)
	if err != nil {
		return fmt.Errorf("Unable to serialize record, err - %s", err)
	}

	item := keyring.Item{
		Key:  key,
		Data: serializedRecord,
	}

	if err := ks.db.Set(item); err != nil {
		return err
	}

	item = keyring.Item{
		Key:  addrHexKeyAsString(addr),
		Data: []byte(key),
	}

	if err := ks.db.Set(item); err != nil {
		return err
	}

	return nil
}

// existsInDb returns true if key is in DB. Error is returned only when we have error
// different thant ErrKeyNotFound
func (ks keystore) existsInDb(addr sdk.Address, name string) (bool, error) {

	if _, err := ks.db.Get(addrHexKeyAsString(addr)); err == nil {
		return true, nil // address lookup succeeds - info exists
	} else if err != keyring.ErrKeyNotFound {
		return false, err // received unexpected error - returns error
	}

	if _, err := ks.db.Get(name); err == nil {
		return true, nil // uid lookup succeeds - info exists
	} else if err != keyring.ErrKeyNotFound {
		return false, err // received unexpected error - returns
	}

	// both lookups failed, info does not exist
	return false, nil
}

func (ks keystore) writeOfflineKey(name string, pk types.PubKey) (*Record, error) {
	emptyRecord := NewOfflineRecord()
	emptyRecordItem := NewOfflineRecordItem(emptyRecord)
	return ks.newRecord(name, pk, emptyRecordItem)
}

// writeMultisigKey investigate where thisf function is called maybe remove it
func (ks keystore) writeMultisigKey(name string, pk types.PubKey) (*Record, error) {
	emptyRecord := NewMultiRecord()
	emptyRecordItem := NewMultiRecordItem(emptyRecord)
	return ks.newRecord(name, pk, emptyRecordItem)
}

func (ks keystore) newRecord(name string, pk types.PubKey, item isRecord_Item) (*Record, error) {
	k, err := NewRecord(name, pk, item)
	if err != nil {
		return nil, err
	}
	return k, ks.writeRecord(k)
}

func (ks keystore) MigrateAll() (bool, error) {
	var migrated bool
	keys, err := ks.db.Keys()
	if err != nil {
		return migrated, fmt.Errorf("Keys() error, err: %s", err)
	}

	if len(keys) == 0 {
		fmt.Println("no keys available for migration")
		return migrated, nil
	}

	for _, key := range keys {
		if strings.Contains(key, addressSuffix) {
			continue
		}

		migrated, err = ks.Migrate(key)
		if err != nil {
			fmt.Printf("migrate err: %q", err)
			continue
		}
	}

	return migrated, nil
}

// for one key
func (ks keystore) Migrate(key string) (bool, error) {
	var migrated bool
	item, err := ks.db.Get(key)
	if err != nil {
		return migrated, wrapKeyNotFound(err, key)
	}

	if len(item.Data) == 0 {
		return migrated, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, key)
	}

	// 2.try to deserialize using proto, if good then continue, otherwise try to deserialize using amino
	if _, err := ks.ProtoUnmarshalRecord(item.Data); err == nil {
		return migrated, nil
	}

	legacyInfo, err := unMarshalLegacyInfo(item.Data)
	if err != nil {
		return migrated, fmt.Errorf("unable to unmarshal item.Data, err: %w", err)
	}

	// 4.serialize info using proto
	k, err := ks.convertFromLegacyInfo(legacyInfo)
	if err != nil {
		return migrated, fmt.Errorf("convertFromLegacyInfo, err - %s", err)
	}

	serializedRecord, err := ks.cdc.Marshal(k)
	if err != nil {
		return migrated, fmt.Errorf("unable to serialize record, err - %w", err)
	}

	// 5.overwrite the keyring entry with
	if err := ks.db.Set(keyring.Item{
		Key:         key,
		Data:        serializedRecord,
		Description: "SDK kerying version",
	}); err != nil {
		return migrated, fmt.Errorf("unable to set keyring.Item, err: %w", err)
	}

	return !migrated, nil
}

func (ks keystore) ProtoUnmarshalRecord(bz []byte) (*Record, error) {
	k := new(Record)
	if err := ks.cdc.Unmarshal(bz, k); err != nil {
		return nil, err
	}

	return k, nil
}

func (ks keystore) ProtoMarshalRecord(k *Record) ([]byte, error) {
	return ks.cdc.Marshal(k)
}

func (ks keystore) MarshalPrivKey(privKey types.PrivKey) ([]byte, error) {
	return ks.cdc.MarshalInterface(privKey)
}

func (ks keystore) convertFromLegacyInfo(info LegacyInfo) (*Record, error) {
	if info == nil {
		return nil, errors.New("unable to convert LegacyInfo to Record cause info is nil")
	}

	var item isRecord_Item

	switch info.GetType() {
	case TypeLocal:
		priv, err := exportPrivateKeyFromLegacyInfo(info)
		if err != nil {
			return nil, err
		}

		localRecord, err := NewLocalRecord(priv)
		if err != nil {
			return nil, err
		}
		item = NewLocalRecordItem(localRecord)

	case TypeOffline:
		offlineRecord := NewOfflineRecord()
		item = NewOfflineRecordItem(offlineRecord)
	case TypeMulti:
		multiRecord := NewMultiRecord()
		item = NewMultiRecordItem(multiRecord)
	case TypeLedger:
		path, err := info.GetPath()
		if err != nil {
			return nil, err
		}
		ledgerRecord := NewLedgerRecord(path)
		item = NewLedgerRecordItem(ledgerRecord)
	}

	name := info.GetName()
	pk := info.GetPubKey()
	return NewRecord(name, pk, item)
}

func (ks keystore) SetItem(item keyring.Item) error {
	return ks.db.Set(item)
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
	_, priv, err := ks.ExportPrivateKeyObject(uid)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(priv.Bytes()), nil
}

func addrHexKeyAsString(address sdk.Address) string {
	return fmt.Sprintf("%s.%s", hex.EncodeToString(address.Bytes()), addressSuffix)
}
