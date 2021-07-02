package keyring

import (
	"fmt"
	"strings"

	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)
// TODO delete legacy.go entirely?
// LegacyKeybase is implemented by the legacy keybase implementation.
type LegacyKeybase interface {
	List() ([]*Record, error)
	Export(name string) (armor string, err error)
	ExportPrivKey(name, decryptPassphrase, encryptPassphrase string) (armor string, err error)
	ExportPubKey(name string) (armor string, err error)
	Close() error
}

// NewLegacy creates a new instance of a legacy keybase.
func NewLegacy(name, dir string, cdc codec.Codec, opts ...KeybaseOption) (LegacyKeybase, error) {
	if err := tmos.EnsureDir(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create Keybase directory: %s", err)
	}

	db, err := sdk.NewLevelDB(name, dir)
	if err != nil {
		return nil, err
	}

	return newDBKeybase(db, cdc), nil
}

var _ LegacyKeybase = dbKeybase{}

// dbKeybase combines encryption and storage implementation to provide a
// full-featured key manager.
//
// Deprecated: dbKeybase will be removed in favor of keyringKeybase.
type dbKeybase struct {
	db  dbm.DB
	cdc codec.Codec
}

// newDBKeybase creates a new dbKeybase instance using the provided DB for
// reading and writing keys.
func newDBKeybase(db dbm.DB, cdc codec.Codec) dbKeybase {
	return dbKeybase{
		db:  db,
		cdc: cdc,
	}
}

// List returns the keys from storage in alphabetical order.
func (kb dbKeybase) List() ([]*Record, error) {
	var res []*Record

	iter, err := kb.db.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key())

		// need to include only keys in storage that have an info suffix
		if !strings.HasSuffix(key, infoSuffix) {
			continue
		}

		ke := new(Record)
		if err := kb.cdc.Unmarshal(iter.Value(), ke); err != nil {
			return nil, err
		}

		res = append(res, ke)
	}

	return res, nil
}

// Get returns the public information about one key.
func (kb dbKeybase) Get(name string) (*Record, error) {
	bs, err := kb.db.Get(infoKeyBz(name))
	if err != nil {
		return nil, err
	}

	if len(bs) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, name)
	}
	ke := new(Record)
	//return protoUnMarshalInfo(bs, kb.cdc)
	if err := kb.cdc.Unmarshal(bs, ke); err != nil {
		return nil, err
	}

	return ke, nil
}

// ExportPrivateKeyObject returns a PrivKey object given the key name and
// passphrase. An error is returned if the key does not exist or if the Info for
// the key is invalid.
func (kb dbKeybase) ExportPrivateKeyObject(name string, passphrase string) (types.PrivKey, error) {
	k, err := kb.Get(name)
	if err != nil {
		return nil, err
	}

	return ExtractPrivKeyFromRecord(kb.cdc, k)
}

func (kb dbKeybase) Export(name string) (armor string, err error) {
	bz, err := kb.db.Get(infoKeyBz(name))
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}

	return crypto.ArmorInfoBytes(bz), nil
}

// ExportPubKey returns public keys in ASCII armored format. It retrieves a Info
// object by its name and return the public key in a portable format.
func (kb dbKeybase) ExportPubKey(name string) (armor string, err error) {
	bz, err := kb.db.Get(infoKeyBz(name))
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}

	//	info, err := protoUnMarshalInfo(bz, kb.cdc)
	ke := Record{}
	if err := kb.cdc.Unmarshal(bz, &ke); err != nil {
		return "", err
	}

	key, err := ke.GetPubKey()
	if err != nil {
		return "", err
	}
	// TODO should I refactor ArmorPubKeyBytes
	return crypto.ArmorPubKeyBytes(key.Bytes(), string(ke.GetAlgo())), nil
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

	return crypto.EncryptArmorPrivKey(priv, encryptPassphrase, string(info.GetAlgo())), nil
}

// Close the underlying storage.
func (kb dbKeybase) Close() error { return kb.db.Close() }

// TODO remove InfoKey it is legacy stuff
func InfoKey(name string) string   { return fmt.Sprintf("%s.%s", name, infoSuffix) }
func infoKeyBz(name string) []byte { return []byte(InfoKey(name)) }

// KeybaseOption overrides options for the db.
type KeybaseOption func(*kbOptions)

type kbOptions struct {
}
