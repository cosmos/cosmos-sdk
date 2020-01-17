package keys

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	"github.com/cosmos/cosmos-sdk/types"
)

var _ Keybase = dbKeybase{}

// Language is a language to create the BIP 39 mnemonic in.
// Currently, only english is supported though.
// Find a list of all supported languages in the BIP 39 spec (word lists).
type Language int

//noinspection ALL
const (
	// English is the default language to create a mnemonic.
	// It is the only supported language by this package.
	English Language = iota + 1
	// Japanese is currently not supported.
	Japanese
	// Korean is currently not supported.
	Korean
	// Spanish is currently not supported.
	Spanish
	// ChineseSimplified is currently not supported.
	ChineseSimplified
	// ChineseTraditional is currently not supported.
	ChineseTraditional
	// French is currently not supported.
	French
	// Italian is currently not supported.
	Italian
	addressSuffix = "address"
	infoSuffix    = "info"
)

const (
	// used for deriving seed from mnemonic
	DefaultBIP39Passphrase = ""

	// bits of entropy to draw when creating a mnemonic
	defaultEntropySize = 256
)

var (
	// ErrUnsupportedSigningAlgo is raised when the caller tries to use a
	// different signing scheme than secp256k1.
	ErrUnsupportedSigningAlgo = errors.New("unsupported signing algo")

	// ErrUnsupportedLanguage is raised when the caller tries to use a
	// different language than english for creating a mnemonic sentence.
	ErrUnsupportedLanguage = errors.New("unsupported language: only english is supported")
)

// dbKeybase combines encryption and storage implementation to provide a
// full-featured key manager.
//
// NOTE: dbKeybase will be deprecated in favor of keyringKeybase.
type dbKeybase struct {
	base baseKeybase
	db   dbm.DB
}

// newDBKeybase creates a new dbKeybase instance using the provided DB for
// reading and writing keys.
func newDBKeybase(db dbm.DB, opts ...KeybaseOption) Keybase {
	return dbKeybase{
		base: newBaseKeybase(opts...),
		db:   db,
	}
}

// NewInMemory creates a transient keybase on top of in-memory storage
// instance useful for testing purposes and on-the-fly key generation.
// Keybase options can be applied when generating this new Keybase.
func NewInMemory(opts ...KeybaseOption) Keybase { return newDBKeybase(dbm.NewMemDB(), opts...) }

// CreateMnemonic generates a new key and persists it to storage, encrypted
// using the provided password. It returns the generated mnemonic and the key Info.
// It returns an error if it fails to generate a key for the given key algorithm
// type, or if another key is already stored under the same name.
func (kb dbKeybase) CreateMnemonic(
	name string, language Language, passwd string, algo SigningAlgo,
) (info Info, mnemonic string, err error) {

	return kb.base.CreateMnemonic(kb, name, language, passwd, algo)
}

// CreateAccount converts a mnemonic to a private key and persists it, encrypted
// with the given password.
func (kb dbKeybase) CreateAccount(
	name, mnemonic, bip39Passwd, encryptPasswd, hdPath string, algo SigningAlgo,
) (Info, error) {

	return kb.base.CreateAccount(kb, name, mnemonic, bip39Passwd, encryptPasswd, hdPath, algo)
}

// CreateLedger creates a new locally-stored reference to a Ledger keypair.
// It returns the created key info and an error if the Ledger could not be queried.
func (kb dbKeybase) CreateLedger(
	name string, algo SigningAlgo, hrp string, account, index uint32,
) (Info, error) {

	return kb.base.CreateLedger(kb, name, algo, hrp, account, index)
}

// CreateOffline creates a new reference to an offline keypair. It returns the
// created key info.
func (kb dbKeybase) CreateOffline(name string, pub tmcrypto.PubKey, algo SigningAlgo) (Info, error) {
	return kb.base.writeOfflineKey(kb, name, pub, algo), nil
}

// CreateMulti creates a new reference to a multisig (offline) keypair. It
// returns the created key info.
func (kb dbKeybase) CreateMulti(name string, pub tmcrypto.PubKey) (Info, error) {
	return kb.base.writeMultisigKey(kb, name, pub), nil
}

// List returns the keys from storage in alphabetical order.
func (kb dbKeybase) List() ([]Info, error) {
	var res []Info

	iter, err := kb.db.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key())

		// need to include only keys in storage that have an info suffix
		if strings.HasSuffix(key, infoSuffix) {
			info, err := unmarshalInfo(iter.Value())
			if err != nil {
				return nil, err
			}

			res = append(res, info)
		}
	}

	return res, nil
}

// Get returns the public information about one key.
func (kb dbKeybase) Get(name string) (Info, error) {
	bs, err := kb.db.Get(infoKey(name))
	if err != nil {
		return nil, err
	}

	if len(bs) == 0 {
		return nil, keyerror.NewErrKeyNotFound(name)
	}

	return unmarshalInfo(bs)
}

// GetByAddress returns Info based on a provided AccAddress. An error is returned
// if the address does not exist.
func (kb dbKeybase) GetByAddress(address types.AccAddress) (Info, error) {
	ik, err := kb.db.Get(addrKey(address))
	if err != nil {
		return nil, err
	}

	if len(ik) == 0 {
		return nil, fmt.Errorf("key with address %s not found", address)
	}

	bs, err := kb.db.Get(ik)
	if err != nil {
		return nil, err
	}

	return unmarshalInfo(bs)
}

// Sign signs the msg with the named key. It returns an error if the key doesn't
// exist or the decryption fails.
func (kb dbKeybase) Sign(name, passphrase string, msg []byte) (sig []byte, pub tmcrypto.PubKey, err error) {
	info, err := kb.Get(name)
	if err != nil {
		return
	}

	var priv tmcrypto.PrivKey

	switch i := info.(type) {
	case localInfo:
		if i.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return
		}

		priv, _, err = mintkey.UnarmorDecryptPrivKey(i.PrivKeyArmor, passphrase)
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

// ExportPrivateKeyObject returns a PrivKey object given the key name and
// passphrase. An error is returned if the key does not exist or if the Info for
// the key is invalid.
func (kb dbKeybase) ExportPrivateKeyObject(name string, passphrase string) (tmcrypto.PrivKey, error) {
	info, err := kb.Get(name)
	if err != nil {
		return nil, err
	}

	var priv tmcrypto.PrivKey

	switch i := info.(type) {
	case localInfo:
		linfo := i
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return nil, err
		}

		priv, _, err = mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, err
		}

	case ledgerInfo, offlineInfo, multiInfo:
		return nil, errors.New("only works on local private keys")
	}

	return priv, nil
}

func (kb dbKeybase) Export(name string) (armor string, err error) {
	bz, err := kb.db.Get(infoKey(name))
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}

	return mintkey.ArmorInfoBytes(bz), nil
}

// ExportPubKey returns public keys in ASCII armored format. It retrieves a Info
// object by its name and return the public key in a portable format.
func (kb dbKeybase) ExportPubKey(name string) (armor string, err error) {
	bz, err := kb.db.Get(infoKey(name))
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}

	info, err := unmarshalInfo(bz)
	if err != nil {
		return
	}

	return mintkey.ArmorPubKeyBytes(info.GetPubKey().Bytes(), string(info.GetAlgo())), nil
}

// ExportPrivKey returns a private key in ASCII armored format.
// It returns an error if the key does not exist or a wrong encryption passphrase
// is supplied.
func (kb dbKeybase) ExportPrivKey(name string, decryptPassphrase string,
	encryptPassphrase string) (armor string, err error) {
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

// ImportPrivKey imports a private key in ASCII armor format. It returns an
// error if a key with the same name exists or a wrong encryption passphrase is
// supplied.
func (kb dbKeybase) ImportPrivKey(name string, armor string, passphrase string) error {
	if _, err := kb.Get(name); err == nil {
		return errors.New("Cannot overwrite key " + name)
	}

	privKey, algo, err := mintkey.UnarmorDecryptPrivKey(armor, passphrase)
	if err != nil {
		return errors.Wrap(err, "couldn't import private key")
	}

	kb.writeLocalKey(name, privKey, passphrase, SigningAlgo(algo))
	return nil
}

func (kb dbKeybase) Import(name string, armor string) (err error) {
	bz, err := kb.db.Get(infoKey(name))
	if err != nil {
		return err
	}

	if len(bz) > 0 {
		return errors.New("cannot overwrite data for name " + name)
	}

	infoBytes, err := mintkey.UnarmorInfoBytes(armor)
	if err != nil {
		return
	}

	kb.db.Set(infoKey(name), infoBytes)
	return nil
}

// ImportPubKey imports ASCII-armored public keys. Store a new Info object holding
// a public key only, i.e. it will not be possible to sign with it as it lacks the
// secret key.
func (kb dbKeybase) ImportPubKey(name string, armor string) (err error) {
	bz, err := kb.db.Get(infoKey(name))
	if err != nil {
		return err
	}

	if len(bz) > 0 {
		return errors.New("cannot overwrite data for name " + name)
	}

	pubBytes, algo, err := mintkey.UnarmorPubKeyBytes(armor)
	if err != nil {
		return
	}

	pubKey, err := cryptoAmino.PubKeyFromBytes(pubBytes)
	if err != nil {
		return
	}

	kb.base.writeOfflineKey(kb, name, pubKey, SigningAlgo(algo))
	return
}

// Delete removes key forever, but we must present the proper passphrase before
// deleting it (for security). It returns an error if the key doesn't exist or
// passphrases don't match. Passphrase is ignored when deleting references to
// offline and Ledger / HW wallet keys.
func (kb dbKeybase) Delete(name, passphrase string, skipPass bool) error {
	// verify we have the proper password before deleting
	info, err := kb.Get(name)
	if err != nil {
		return err
	}

	if linfo, ok := info.(localInfo); ok && !skipPass {
		if _, _, err = mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase); err != nil {
			return err
		}
	}

	kb.db.DeleteSync(addrKey(info.GetAddress()))
	kb.db.DeleteSync(infoKey(name))

	return nil
}

// Update changes the passphrase with which an already stored key is
// encrypted.
//
// oldpass must be the current passphrase used for encryption,
// getNewpass is a function to get the passphrase to permanently replace
// the current passphrase
func (kb dbKeybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	info, err := kb.Get(name)
	if err != nil {
		return err
	}

	switch i := info.(type) {
	case localInfo:
		linfo := i

		key, _, err := mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, oldpass)
		if err != nil {
			return err
		}

		newpass, err := getNewpass()
		if err != nil {
			return err
		}

		kb.writeLocalKey(name, key, newpass, i.GetAlgo())
		return nil

	default:
		return fmt.Errorf("locally stored key required. Received: %v", reflect.TypeOf(info).String())
	}
}

// CloseDB releases the lock and closes the storage backend.
func (kb dbKeybase) CloseDB() {
	kb.db.Close()
}

// SupportedAlgos returns a list of supported signing algorithms.
func (kb dbKeybase) SupportedAlgos() []SigningAlgo {
	return kb.base.SupportedAlgos()
}

// SupportedAlgosLedger returns a list of supported ledger signing algorithms.
func (kb dbKeybase) SupportedAlgosLedger() []SigningAlgo {
	return kb.base.SupportedAlgosLedger()
}

func (kb dbKeybase) writeLocalKey(name string, priv tmcrypto.PrivKey, passphrase string, algo SigningAlgo) Info {
	// encrypt private key using passphrase
	privArmor := mintkey.EncryptArmorPrivKey(priv, passphrase, string(algo))

	// make Info
	pub := priv.PubKey()
	info := newLocalInfo(name, pub, privArmor, algo)

	kb.writeInfo(name, info)
	return info
}

func (kb dbKeybase) writeInfo(name string, info Info) {
	// write the info by key
	key := infoKey(name)
	serializedInfo := marshalInfo(info)

	kb.db.SetSync(key, serializedInfo)

	// store a pointer to the infokey by address for fast lookup
	kb.db.SetSync(addrKey(info.GetAddress()), key)
}

func addrKey(address types.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s.%s", address.String(), addressSuffix))
}

func infoKey(name string) []byte {
	return []byte(fmt.Sprintf("%s.%s", name, infoSuffix))
}
