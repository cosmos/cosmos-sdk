package keys

import (
	"fmt"
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

// Create generates a new key and persists it to storage, encrypted
// using the passphrase.  It returns the generated seedphrase
// (mnemonic) and the key Info.  It returns an error if it fails to
// generate a key for the given algo type, or if another key is
// already stored under the same name.
func (kb dbKeybase) Create(name, passphrase string, algo CryptoAlgo) (Info, string, error) {
	// NOTE: secret is SHA256 hashed by secp256k1 and ed25519.
	// 16 byte secret corresponds to 12 BIP39 words.
	// XXX: Ledgers use 24 words now - should we ?
	secret := crypto.CRandBytes(16)
	priv, err := generate(algo, secret)
	if err != nil {
		return Info{}, "", err
	}

	// encrypt and persist the key
	info := kb.writeKey(priv, name, passphrase)

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

// Recover converts a seedphrase to a private key and persists it,
// encrypted with the given passphrase.  Functions like Create, but
// seedphrase is input not output.
func (kb dbKeybase) Recover(name, passphrase, seedphrase string) (Info, error) {
	words := strings.Split(strings.TrimSpace(seedphrase), " ")
	secret, err := kb.codec.WordsToBytes(words)
	if err != nil {
		return Info{}, err
	}

	// secret is comprised of the actual secret with the type
	// appended.
	// ie [secret] = [type] + [secret]
	typ, secret := secret[0], secret[1:]
	algo := byteToCryptoAlgo(typ)
	priv, err := generate(algo, secret)
	if err != nil {
		return Info{}, err
	}

	// encrypt and persist key.
	public := kb.writeKey(priv, name, passphrase)
	return public, err
}

// List returns the keys from storage in alphabetical order.
func (kb dbKeybase) List() ([]Info, error) {
	var res []Info
	iter := kb.db.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		// key := iter.Key()
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
	priv, err := unarmorDecryptPrivKey(info.PrivKeyArmor, passphrase)
	if err != nil {
		return
	}
	sig = priv.Sign(msg)
	pub = priv.PubKey()
	return
}

func (kb dbKeybase) Export(name string) (armor string, err error) {
	bz := kb.db.Get(infoKey(name))
	if bz == nil {
		return "", errors.New("No key to export with name " + name)
	}
	return armorInfoBytes(bz), nil
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

// Delete removes key forever, but we must present the
// proper passphrase before deleting it (for security).
func (kb dbKeybase) Delete(name, passphrase string) error {
	// verify we have the proper password before deleting
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	_, err = unarmorDecryptPrivKey(info.PrivKeyArmor, passphrase)
	if err != nil {
		return err
	}
	kb.db.DeleteSync(infoKey(name))
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
	key, err := unarmorDecryptPrivKey(info.PrivKeyArmor, oldpass)
	if err != nil {
		return err
	}

	kb.writeKey(key, name, newpass)
	return nil
}
func (kb dbKeybase) writeKey(priv crypto.PrivKey, name, passphrase string) Info {
	// generate the encrypted privkey
	privArmor := encryptArmorPrivKey(priv, passphrase)
	// make Info
	info := newInfo(name, priv.PubKey(), privArmor)

	// write them both
	kb.db.SetSync(infoKey(name), info.bytes())
	return info
}

func generate(algo CryptoAlgo, secret []byte) (crypto.PrivKey, error) {
	switch algo {
	case AlgoEd25519:
		return crypto.GenPrivKeyEd25519FromSecret(secret).Wrap(), nil
	case AlgoSecp256k1:
		return crypto.GenPrivKeySecp256k1FromSecret(secret).Wrap(), nil
	default:
		err := errors.Errorf("Cannot generate keys for algorithm: %s", algo)
		return crypto.PrivKey{}, err
	}
}

func infoKey(name string) []byte {
	return []byte(fmt.Sprintf("%s.info", name))
}
