package keys

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys/words"
	dbm "github.com/tendermint/tmlibs/db"
)

// dbKeybase combines encyption and storage implementation to provide
// a full-featured key manager
type dbKeybase struct {
	db    dbm.DB
	codec words.Codec
}

func New(db dbm.DB, codec words.Codec) dbKeybase {
	return dbKeybase{
		db:    db,
		codec: codec,
	}
}

var _ Keybase = dbKeybase{}

// CreateMnemonic generates a new key and persists it to storage, encrypted
// using the passphrase.  It returns the generated seedphrase
// (mnemonic) and the key Info.  It returns an error if it fails to
// generate a key for the given algo type, or if another key is
// already stored under the same name.
func (kb dbKeybase) CreateMnemonic(name, passphrase string, algo SignAlgo) (Info, string, error) {
	// NOTE: secret is SHA256 hashed by secp256k1 and ed25519.
	// 16 byte secret corresponds to 12 BIP39 words.
	// XXX: Ledgers use 24 words now - should we ?
	secret := crypto.CRandBytes(16)
	priv, err := generate(algo, secret)
	if err != nil {
		return nil, "", err
	}

	// encrypt and persist the key
	info := kb.writeLocalKey(priv, name, passphrase)

	// we append the type byte to the serialized secret to help with
	// recovery
	// ie [secret] = [type] + [secret]
	typ := cryptoAlgoToByte(algo)
	secret = append([]byte{typ}, secret...)

	// return the mnemonic phrase
	words, err := kb.codec.BytesToWords(secret)
	seed := strings.Join(words, " ")
	return info, seed, err
}

// CreateLedger creates a new locally-stored reference to a Ledger keypair
// It returns the created key info and an error if the Ledger could not be queried
func (kb dbKeybase) CreateLedger(name string, path crypto.DerivationPath, algo SignAlgo) (Info, error) {
	if algo != AlgoSecp256k1 {
		return nil, fmt.Errorf("Only secp256k1 is supported for Ledger devices")
	}
	priv, err := crypto.NewPrivKeyLedgerSecp256k1(path)
	if err != nil {
		return nil, err
	}
	pub, err := priv.PubKey()
	if err != nil {
		return nil, err
	}
	return kb.writeLedgerKey(pub, path, name), nil
}

// CreateOffline creates a new reference to an offline keypair
// It returns the created key info
func (kb dbKeybase) CreateOffline(name string, pub crypto.PubKey) (Info, error) {
	return kb.writeOfflineKey(pub, name), nil
}

// Recover converts a seedphrase to a private key and persists it,
// encrypted with the given passphrase.  Functions like Create, but
// seedphrase is input not output.
func (kb dbKeybase) Recover(name, passphrase, seedphrase string) (Info, error) {
	words := strings.Split(strings.TrimSpace(seedphrase), " ")
	secret, err := kb.codec.WordsToBytes(words)
	if err != nil {
		return nil, err
	}

	// secret is comprised of the actual secret with the type
	// appended.
	// ie [secret] = [type] + [secret]
	typ, secret := secret[0], secret[1:]
	algo := byteToSignAlgo(typ)
	priv, err := generate(algo, secret)
	if err != nil {
		return nil, err
	}

	// encrypt and persist key.
	public := kb.writeLocalKey(priv, name, passphrase)
	return public, nil
}

// List returns the keys from storage in alphabetical order.
func (kb dbKeybase) List() ([]Info, error) {
	var res []Info
	iter := kb.db.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		info, err := readInfo(iter.Value())
		if err != nil {
			return nil, err
		}
		res = append(res, info)
	}
	return res, nil
}

// Get returns the public information about one key.
func (kb dbKeybase) Get(name string) (Info, error) {
	bs := kb.db.Get(infoKey(name))
	return readInfo(bs)
}

// Sign signs the msg with the named key.
// It returns an error if the key doesn't exist or the decryption fails.
func (kb dbKeybase) Sign(name, passphrase string, msg []byte) (sig crypto.Signature, pub crypto.PubKey, err error) {
	info, err := kb.Get(name)
	if err != nil {
		return
	}
	var priv crypto.PrivKey
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return
		}
		priv, err = unarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
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
		fmt.Printf("Bytes to sign:\n%s", msg)
		buf := bufio.NewReader(os.Stdin)
		fmt.Printf("\nEnter Amino-encoded signature:\n")
		// Will block until user inputs the signature
		signed, err := buf.ReadString('\n')
		if err != nil {
			return nil, nil, err
		}
		cdc.MustUnmarshalBinary([]byte(signed), sig)
		return sig, linfo.GetPubKey(), nil
	}
	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}
	pub, err = priv.PubKey()
	if err != nil {
		return nil, nil, err
	}
	return sig, pub, nil
}

func (kb dbKeybase) Export(name string) (armor string, err error) {
	bz := kb.db.Get(infoKey(name))
	if bz == nil {
		return "", errors.New("No key to export with name " + name)
	}
	return armorInfoBytes(bz), nil
}

// ExportPubKey returns public keys in ASCII armored format.
// Retrieve a Info object by its name and return the public key in
// a portable format.
func (kb dbKeybase) ExportPubKey(name string) (armor string, err error) {
	bz := kb.db.Get(infoKey(name))
	if bz == nil {
		return "", errors.New("No key to export with name " + name)
	}
	info, err := readInfo(bz)
	if err != nil {
		return
	}
	return armorPubKeyBytes(info.GetPubKey().Bytes()), nil
}

func (kb dbKeybase) Import(name string, armor string) (err error) {
	bz := kb.db.Get(infoKey(name))
	if len(bz) > 0 {
		return errors.New("Cannot overwrite data for name " + name)
	}
	infoBytes, err := unarmorInfoBytes(armor)
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
	pubBytes, err := unarmorPubKeyBytes(armor)
	if err != nil {
		return
	}
	pubKey, err := crypto.PubKeyFromBytes(pubBytes)
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
		_, err = unarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return err
		}
		kb.db.DeleteSync(infoKey(name))
		return nil
	case ledgerInfo:
	case offlineInfo:
		if passphrase != "yes" {
			return fmt.Errorf("enter exactly 'yes' to delete the key")
		}
		kb.db.DeleteSync(infoKey(name))
		return nil
	}
	return nil
}

// Update changes the passphrase with which an already stored key is
// encrypted.
//
// oldpass must be the current passphrase used for encryption,
// newpass will be the only valid passphrase from this time forward.
func (kb dbKeybase) Update(name, oldpass, newpass string) error {
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		key, err := unarmorDecryptPrivKey(linfo.PrivKeyArmor, oldpass)
		if err != nil {
			return err
		}
		kb.writeLocalKey(key, name, newpass)
		return nil
	default:
		return fmt.Errorf("Locally stored key required")
	}
}

func (kb dbKeybase) writeLocalKey(priv crypto.PrivKey, name, passphrase string) Info {
	// encrypt private key using passphrase
	privArmor := encryptArmorPrivKey(priv, passphrase)
	// make Info
	pub, err := priv.PubKey()
	if err != nil {
		panic(err)
	}
	info := newLocalInfo(name, pub, privArmor)
	kb.writeInfo(info, name)
	return info
}

func (kb dbKeybase) writeLedgerKey(pub crypto.PubKey, path crypto.DerivationPath, name string) Info {
	info := newLedgerInfo(name, pub, path)
	kb.writeInfo(info, name)
	return info
}

func (kb dbKeybase) writeOfflineKey(pub crypto.PubKey, name string) Info {
	info := newOfflineInfo(name, pub)
	kb.writeInfo(info, name)
	return info
}

func (kb dbKeybase) writeInfo(info Info, name string) {
	// write the info by key
	kb.db.SetSync(infoKey(name), writeInfo(info))
}

func generate(algo SignAlgo, secret []byte) (crypto.PrivKey, error) {
	switch algo {
	case AlgoEd25519:
		return crypto.GenPrivKeyEd25519FromSecret(secret), nil
	case AlgoSecp256k1:
		return crypto.GenPrivKeySecp256k1FromSecret(secret), nil
	default:
		err := errors.Errorf("Cannot generate keys for algorithm: %s", algo)
		return nil, err
	}
}

func infoKey(name string) []byte {
	return []byte(fmt.Sprintf("%s.info", name))
}
