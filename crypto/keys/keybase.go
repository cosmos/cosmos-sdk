package keys

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var _ Keybase = dbKeybase{}

// Language is a language to create the BIP 39 mnemonic in.
// Currently, only english is supported though.
// Find a list of all supported languages in the BIP 39 spec (word lists).
type Language int

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
	defaultBIP39Passphrase = ""

	// bits of entropy to draw when creating a mnemonic
	defaultEntropySize = 256
)

var (
	// ErrUnsupportedSigningAlgo is raised when the caller tries to use a
	// different signing scheme than secp256k1.
	ErrUnsupportedSigningAlgo = errors.New("unsupported signing algo: only secp256k1 is supported")

	// ErrUnsupportedLanguage is raised when the caller tries to use a
	// different language than english for creating a mnemonic sentence.
	ErrUnsupportedLanguage = errors.New("unsupported language: only english is supported")
)

// dbKeybase combines encryption and storage implementation to provide
// a full-featured key manager
type dbKeybase struct {
	db dbm.DB
}

// New creates a new keybase instance using the passed DB for reading and writing keys.
func New(db dbm.DB) Keybase {
	return dbKeybase{
		db: db,
	}
}

// CreateMnemonic generates a new key and persists it to storage, encrypted
// using the provided password.
// It returns the generated mnemonic and the key Info.
// It returns an error if it fails to
// generate a key for the given algo type, or if another key is
// already stored under the same name.
func (kb dbKeybase) CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, mnemonic string, err error) {
	if language != English {
		return nil, "", ErrUnsupportedLanguage
	}
	if algo != Secp256k1 {
		err = ErrUnsupportedSigningAlgo
		return
	}

	// default number of words (24):
	// this generates a mnemonic directly from the number of words by reading system entropy.
	entropy, err := bip39.NewEntropy(defaultEntropySize)
	if err != nil {
		return
	}
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return
	}

	seed := bip39.NewSeed(mnemonic, defaultBIP39Passphrase)
	info, err = kb.persistDerivedKey(seed, passwd, name, hd.FullFundraiserPath)
	return
}

// TEMPORARY METHOD UNTIL WE FIGURE OUT USER FACING HD DERIVATION API
func (kb dbKeybase) CreateKey(name, mnemonic, passwd string) (info Info, err error) {
	words := strings.Split(mnemonic, " ")
	if len(words) != 12 && len(words) != 24 {
		err = fmt.Errorf("recovering only works with 12 word (fundraiser) or 24 word mnemonics, got: %v words", len(words))
		return
	}
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, defaultBIP39Passphrase)
	if err != nil {
		return
	}
	info, err = kb.persistDerivedKey(seed, passwd, name, hd.FullFundraiserPath)
	return
}

// CreateFundraiserKey converts a mnemonic to a private key and persists it,
// encrypted with the given password.
// TODO(ismail)
func (kb dbKeybase) CreateFundraiserKey(name, mnemonic, passwd string) (info Info, err error) {
	words := strings.Split(mnemonic, " ")
	if len(words) != 12 {
		err = fmt.Errorf("recovering only works with 12 word (fundraiser), got: %v words", len(words))
		return
	}
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, defaultBIP39Passphrase)
	if err != nil {
		return
	}
	info, err = kb.persistDerivedKey(seed, passwd, name, hd.FullFundraiserPath)
	return
}

func (kb dbKeybase) Derive(name, mnemonic, bip39Passphrase, encryptPasswd string, params hd.BIP44Params) (info Info, err error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return
	}
	info, err = kb.persistDerivedKey(seed, encryptPasswd, name, params.String())

	return
}

// CreateLedger creates a new locally-stored reference to a Ledger keypair
// It returns the created key info and an error if the Ledger could not be queried
func (kb dbKeybase) CreateLedger(name string, path crypto.DerivationPath, algo SigningAlgo) (Info, error) {
	if algo != Secp256k1 {
		return nil, ErrUnsupportedSigningAlgo
	}
	priv, err := crypto.NewPrivKeyLedgerSecp256k1(path)
	if err != nil {
		return nil, err
	}
	pub := priv.PubKey()
	return kb.writeLedgerKey(pub, path, name), nil
}

// CreateOffline creates a new reference to an offline keypair
// It returns the created key info
func (kb dbKeybase) CreateOffline(name string, pub tmcrypto.PubKey) (Info, error) {
	return kb.writeOfflineKey(pub, name), nil
}

func (kb *dbKeybase) persistDerivedKey(seed []byte, passwd, name, fullHdPath string) (info Info, err error) {
	// create master key and derive first key:
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath)
	if err != nil {
		return
	}

	// if we have a password, use it to encrypt the private key and store it
	// else store the public key only
	if passwd != "" {
		info = kb.writeLocalKey(secp256k1.PrivKeySecp256k1(derivedPriv), name, passwd)
	} else {
		pubk := secp256k1.PrivKeySecp256k1(derivedPriv).PubKey()
		info = kb.writeOfflineKey(pubk, name)
	}
	return
}

// List returns the keys from storage in alphabetical order.
func (kb dbKeybase) List() ([]Info, error) {
	var res []Info
	iter := kb.db.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key())

		// need to include only keys in storage that have an info suffix
		if strings.HasSuffix(key, infoSuffix) {
			info, err := readInfo(iter.Value())
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
	bs := kb.db.Get(infoKey(name))
	if len(bs) == 0 {
		return nil, keyerror.NewErrKeyNotFound(name)
	}
	return readInfo(bs)
}

func (kb dbKeybase) GetByAddress(address types.AccAddress) (Info, error) {
	ik := kb.db.Get(addrKey(address))
	if len(ik) == 0 {
		return nil, fmt.Errorf("key with address %s not found", address)
	}
	bs := kb.db.Get(ik)
	return readInfo(bs)
}

// Sign signs the msg with the named key.
// It returns an error if the key doesn't exist or the decryption fails.
func (kb dbKeybase) Sign(name, passphrase string, msg []byte) (sig []byte, pub tmcrypto.PubKey, err error) {
	info, err := kb.Get(name)
	if err != nil {
		return
	}
	var priv tmcrypto.PrivKey
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return
		}
		priv, err = mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, nil, err
		}
	case ledgerInfo:
		linfo := info.(ledgerInfo)
		priv, err = crypto.NewPrivKeyLedgerSecp256k1(linfo.Path)
		if err != nil {
			return
		}
	case offlineInfo:
		linfo := info.(offlineInfo)
		_, err := fmt.Fprintf(os.Stderr, "Bytes to sign:\n%s", msg)
		if err != nil {
			return nil, nil, err
		}
		buf := bufio.NewReader(os.Stdin)
		_, err = fmt.Fprintf(os.Stderr, "\nEnter Amino-encoded signature:\n")
		if err != nil {
			return nil, nil, err
		}
		// Will block until user inputs the signature
		signed, err := buf.ReadString('\n')
		if err != nil {
			return nil, nil, err
		}
		cdc.MustUnmarshalBinaryLengthPrefixed([]byte(signed), sig)
		return sig, linfo.GetPubKey(), nil
	}
	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}
	pub = priv.PubKey()
	return sig, pub, nil
}

func (kb dbKeybase) ExportPrivateKeyObject(name string, passphrase string) (tmcrypto.PrivKey, error) {
	info, err := kb.Get(name)
	if err != nil {
		return nil, err
	}
	var priv tmcrypto.PrivKey
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return nil, err
		}
		priv, err = mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, err
		}
	case ledgerInfo:
		return nil, errors.New("Only works on local private keys")
	case offlineInfo:
		return nil, errors.New("Only works on local private keys")
	}
	return priv, nil
}

func (kb dbKeybase) Export(name string) (armor string, err error) {
	bz := kb.db.Get(infoKey(name))
	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}
	return mintkey.ArmorInfoBytes(bz), nil
}

// ExportPubKey returns public keys in ASCII armored format.
// Retrieve a Info object by its name and return the public key in
// a portable format.
func (kb dbKeybase) ExportPubKey(name string) (armor string, err error) {
	bz := kb.db.Get(infoKey(name))
	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}
	info, err := readInfo(bz)
	if err != nil {
		return
	}
	return mintkey.ArmorPubKeyBytes(info.GetPubKey().Bytes()), nil
}

func (kb dbKeybase) Import(name string, armor string) (err error) {
	bz := kb.db.Get(infoKey(name))
	if len(bz) > 0 {
		return errors.New("Cannot overwrite data for name " + name)
	}
	infoBytes, err := mintkey.UnarmorInfoBytes(armor)
	if err != nil {
		return
	}
	kb.db.Set(infoKey(name), infoBytes)
	return nil
}

// ImportPubKey imports ASCII-armored public keys.
// Store a new Info object holding a public key only, i.e. it will
// not be possible to sign with it as it lacks the secret key.
func (kb dbKeybase) ImportPubKey(name string, armor string) (err error) {
	bz := kb.db.Get(infoKey(name))
	if len(bz) > 0 {
		return errors.New("Cannot overwrite data for name " + name)
	}
	pubBytes, err := mintkey.UnarmorPubKeyBytes(armor)
	if err != nil {
		return
	}
	pubKey, err := cryptoAmino.PubKeyFromBytes(pubBytes)
	if err != nil {
		return
	}
	kb.writeOfflineKey(pubKey, name)
	return
}

// Delete removes key forever, but we must present the
// proper passphrase before deleting it (for security).
// A passphrase of 'yes' is used to delete stored
// references to offline and Ledger / HW wallet keys
func (kb dbKeybase) Delete(name, passphrase string) error {
	// verify we have the proper password before deleting
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		_, err = mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return err
		}
		kb.db.DeleteSync(addrKey(linfo.GetAddress()))
		kb.db.DeleteSync(infoKey(name))
		return nil
	case ledgerInfo:
	case offlineInfo:
		if passphrase != "yes" {
			return fmt.Errorf("enter 'yes' exactly to delete the key - this cannot be undone")
		}
		kb.db.DeleteSync(addrKey(info.GetAddress()))
		kb.db.DeleteSync(infoKey(name))
		return nil
	}

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
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		key, err := mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, oldpass)
		if err != nil {
			return err
		}
		newpass, err := getNewpass()
		if err != nil {
			return err
		}
		kb.writeLocalKey(key, name, newpass)
		return nil
	default:
		return fmt.Errorf("locally stored key required")
	}
}

// CloseDB releases the lock and closes the storage backend.
func (kb dbKeybase) CloseDB() {
	kb.db.Close()
}

func (kb dbKeybase) writeLocalKey(priv tmcrypto.PrivKey, name, passphrase string) Info {
	// encrypt private key using passphrase
	privArmor := mintkey.EncryptArmorPrivKey(priv, passphrase)
	// make Info
	pub := priv.PubKey()
	info := newLocalInfo(name, pub, privArmor)
	kb.writeInfo(info, name)
	return info
}

func (kb dbKeybase) writeLedgerKey(pub tmcrypto.PubKey, path crypto.DerivationPath, name string) Info {
	info := newLedgerInfo(name, pub, path)
	kb.writeInfo(info, name)
	return info
}

func (kb dbKeybase) writeOfflineKey(pub tmcrypto.PubKey, name string) Info {
	info := newOfflineInfo(name, pub)
	kb.writeInfo(info, name)
	return info
}

func (kb dbKeybase) writeInfo(info Info, name string) {
	// write the info by key
	key := infoKey(name)
	kb.db.SetSync(key, writeInfo(info))
	// store a pointer to the infokey by address for fast lookup
	kb.db.SetSync(addrKey(info.GetAddress()), key)
}

func addrKey(address types.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s.%s", address.String(), addressSuffix))
}

func infoKey(name string) []byte {
	return []byte(fmt.Sprintf("%s.%s", name, infoSuffix))
}
