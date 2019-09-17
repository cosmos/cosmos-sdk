package keys

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/99designs/keyring"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	"github.com/cosmos/cosmos-sdk/types"

	bip39 "github.com/cosmos/go-bip39"

	tmcrypto "github.com/tendermint/tendermint/crypto"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var _ Keybase = keyringKeybase{}

type keyringKeybase struct {
	db keyring.Keyring
}

func newKeyringKeybase(db keyring.Keyring) Keybase {
	return keyringKeybase{
		db: db,
	}
}

// CreateMnemonic generates a new key and persists it to storage, encrypted
// using the provided password.
// It returns the generated mnemonic and the key Info.
// It returns an error if it fails to
// generate a key for the given algo type, or if another key is
// already stored under the same name.
func (kb keyringKeybase) CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, mnemonic string, err error) {
	seed, fullFundraiserPath, mnemonic, err := createMnemonic(language, algo)
	if err != nil {
		return
	}
	info, err = kb.persistDerivedKey(seed, passwd, name, fullFundraiserPath)
	return
}

// CreateAccount converts a mnemonic to a private key and persists it, encrypted with the given password.
func (kb keyringKeybase) CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error) {
	coinType := types.GetConfig().GetCoinType()
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	return kb.Derive(name, mnemonic, bip39Passwd, encryptPasswd, *hdPath)
}

// Derive computes a BIP39 seed from th mnemonic and bip39Passwd.
// Derive private key from the seed using the BIP44 params.
func (kb keyringKeybase) Derive(name, mnemonic, bip39Passphrase, encryptPasswd string, params hd.BIP44Params) (info Info, err error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return
	}

	info, err = kb.persistDerivedKey(seed, encryptPasswd, name, params.String())
	return
}

// CreateLedger creates a new locally-stored reference to a Ledger keypair
// It returns the created key info and an error if the Ledger could not be queried
func (kb keyringKeybase) CreateLedger(name string, algo SigningAlgo, hrp string, account uint32, index uint32) (Info, error) {
	pub, hdPath, err := createLedger(algo, hrp, account, index)
	if err != nil {
		return nil, err
	}
	return kb.writeLedgerKey(name, pub, *hdPath), nil
}

// CreateOffline creates a new reference to an offline keypair. It returns the
// created key info.
func (kb keyringKeybase) CreateOffline(name string, pub tmcrypto.PubKey) (Info, error) {
	return kb.writeOfflineKey(name, pub), nil
}

// CreateMulti creates a new reference to a multisig (offline) keypair. It
// returns the created key info.
//CreateMulti for keyring
func (kb keyringKeybase) CreateMulti(name string, pub tmcrypto.PubKey) (Info, error) {
	return kb.writeMultisigKey(name, pub), nil
}

func (kb *keyringKeybase) persistDerivedKey(seed []byte, passwd, name, fullHdPath string) (info Info, err error) {
	// create master key and derive first key for keyring
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath)
	if err != nil {
		return
	}

	if passwd != "" {
		info = kb.writeLocalKey(name, secp256k1.PrivKeySecp256k1(derivedPriv), passwd)
	} else {
		pubk := secp256k1.PrivKeySecp256k1(derivedPriv).PubKey()
		info = kb.writeOfflineKey(name, pubk)
	}
	return
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

			info, err := readInfo(rawInfo.Data)
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
	return readInfo(bs.Data)
}

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
	return readInfo(bs.Data)
}

// Sign signs the msg with the named key.
// It returns an error if the key doesn't exist or the decryption fails.
func (kb keyringKeybase) Sign(name, passphrase string, msg []byte) (sig []byte, pub tmcrypto.PubKey, err error) {
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

		priv, err = cryptoAmino.PrivKeyFromBytes([]byte(i.PrivKeyArmor))
		if err != nil {
			return nil, nil, err
		}

	case ledgerInfo:
		return signWithLedger(info, msg)

	case offlineInfo, multiInfo:
		return decodeSignature(info, msg)
	}

	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, priv.PubKey(), nil
}

//ExportPrivateKeyObject exports an armored private key object to the terminal
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

//Export exports armored private key to the caller
func (kb keyringKeybase) Export(name string) (armor string, err error) {
	bz, err := kb.db.Get(string(infoKey(name)))
	if err != nil {
		return "", err
	}
	if bz.Data == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}

	return mintkey.ArmorInfoBytes(bz.Data), nil
}

// ExportPubKey returns public keys in ASCII armored format.
// Retrieve a Info object by its name and return the public key in
// a portable format.
func (kb keyringKeybase) ExportPubKey(name string) (armor string, err error) {
	bz, err := kb.Get(name)
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}
	return mintkey.ArmorPubKeyBytes(bz.GetPubKey().Bytes()), nil
}

//Import imports armored private key
func (kb keyringKeybase) Import(name string, armor string) (err error) {
	//err is purposefully discarded
	bz, _ := kb.Get(name)

	if bz != nil {
		pubkey := bz.GetPubKey()

		if len(pubkey.Bytes()) > 0 {
			return errors.New("Cannot overwrite data for name " + name)
		}
	}

	infoBytes, err := mintkey.UnarmorInfoBytes(armor)
	if err != nil {
		return
	}
	info, err := readInfo(infoBytes)

	if err != nil {
		return
	}

	kb.writeInfo(name, info)

	err = kb.db.Set(keyring.Item{
		Key:  string(addrKey(info.GetAddress())),
		Data: infoKey(name),
	})

	if err != nil {
		return
	}

	return nil
}

// ExportPrivKey returns a private key in ASCII armored format.
// It returns an error if the key does not exist or a wrong encryption passphrase is supplied.
func (kb keyringKeybase) ExportPrivKey(name string, decryptPassphrase string,
	encryptPassphrase string) (armor string, err error) {
	priv, err := kb.ExportPrivateKeyObject(name, decryptPassphrase)
	if err != nil {
		return "", err
	}

	return mintkey.EncryptArmorPrivKey(priv, encryptPassphrase), nil
}

// ImportPrivKey imports a private key in ASCII armor format.
// It returns an error if a key with the same name exists or a wrong encryption passphrase is
// supplied.
func (kb keyringKeybase) ImportPrivKey(name string, armor string, passphrase string) error {
	if kb.HasKey(name) {
		return errors.New("cannot overwrite key " + name)
	}

	privKey, err := mintkey.UnarmorDecryptPrivKey(armor, passphrase)
	if err != nil {
		return errors.Wrap(err, "couldn't import private key")
	}

	kb.writeLocalKey(name, privKey, passphrase)
	return nil
}

//HasKey is checking if the key exists in the keyring
func (kb keyringKeybase) HasKey(name string) bool {
	bz, _ := kb.Get(name)
	if bz != nil {
		return true
	}
	return false
}

// ImportPubKey imports ASCII-armored public keys.
// Store a new Info object holding a public key only, i.e. it will
// not be possible to sign with it as it lacks the secret key.
//ImportPubKey for keyring
func (kb keyringKeybase) ImportPubKey(name string, armor string) (err error) {
	bz, _ := kb.Get(name)
	if bz != nil {
		pubkey := bz.GetPubKey()

		if len(pubkey.Bytes()) > 0 {
			return errors.New("cannot overwrite data for name " + name)
		}

	}
	pubBytes, err := mintkey.UnarmorPubKeyBytes(armor)
	if err != nil {
		return
	}
	pubKey, err := cryptoAmino.PubKeyFromBytes(pubBytes)
	if err != nil {
		return
	}
	kb.writeOfflineKey(name, pubKey)
	return
}

// Delete removes key forever, but we must present the
// proper passphrase before deleting it (for security).
// It returns an error if the key doesn't exist or
// passphrases don't match.
// Passphrase is ignored when deleting references to
// offline and Ledger / HW wallet keys.
func (kb keyringKeybase) Delete(name, passphrase string, skipPass bool) error {
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

// Update changes the passphrase with which an already stored key is
// encrypted.
//
// oldpass must be the current passphrase used for encryption,
// getNewpass is a function to get the passphrase to permanently replace
// the current passphrase
func (kb keyringKeybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	switch linfo := info.(type) {
	case localInfo:
		key, err := mintkey.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, oldpass)
		if err != nil {
			return err
		}
		newpass, err := getNewpass()
		if err != nil {
			return err
		}
		kb.writeLocalKey(name, key, newpass)
		return nil
	default:
		return fmt.Errorf("locally stored key required. Received: %v", reflect.TypeOf(info).String())
	}
}

// CloseDB releases the lock and closes the storage backend.
func (kb keyringKeybase) CloseDB() {
}

func (kb keyringKeybase) writeLocalKey(name string, priv tmcrypto.PrivKey, passphrase string) Info {
	//encrypt private key using keyring
	pub := priv.PubKey()
	info := newLocalInfo(name, pub, string(priv.Bytes()))
	kb.writeInfo(name, info)
	return info

}

func (kb keyringKeybase) writeLedgerKey(name string, pub tmcrypto.PubKey, path hd.BIP44Params) Info {
	info := newLedgerInfo(name, pub, path)
	kb.writeInfo(name, info)
	return info
}

func (kb keyringKeybase) writeOfflineKey(name string, pub tmcrypto.PubKey) Info {
	info := newOfflineInfo(name, pub)
	kb.writeInfo(name, info)
	return info
}

func (kb keyringKeybase) writeMultisigKey(name string, pub tmcrypto.PubKey) Info {
	info := NewMultiInfo(name, pub)
	kb.writeInfo(name, info)
	return info
}

func (kb keyringKeybase) writeInfo(name string, info Info) {
	//write the info by key
	key := infoKey(name)
	serializedInfo := writeInfo(info)
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
