package keys

import (
	"os"
	"path/filepath"

	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/types"
)

var _ Keybase = lazyKeybase{}

type lazyKeybase struct {
	name string
	dir  string
}

// New creates a new instance of a lazy keybase.
func New(name, dir string) Keybase {
	return lazyKeybase{name: name, dir: dir}
}

func (lkb lazyKeybase) List() ([]Info, error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).List()
}

func (lkb lazyKeybase) Get(name string) (Info, error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).Get(name)
}

func (lkb lazyKeybase) GetByAddress(address types.AccAddress) (Info, error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).GetByAddress(address)
}

func (lkb lazyKeybase) Delete(name, passphrase string, skipPass bool) error {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return newDbKeybase(db).Delete(name, passphrase, skipPass)
}

func (lkb lazyKeybase) Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	return newDbKeybase(db).Sign(name, passphrase, msg)
}

func (lkb lazyKeybase) CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, "", err
	}
	defer db.Close()

	return newDbKeybase(db).CreateMnemonic(name, language, passwd, algo)
}

func (lkb lazyKeybase) CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd, account, index)
}

func (lkb lazyKeybase) Derive(name, mnemonic, bip39Passwd, encryptPasswd string, params hd.BIP44Params) (Info, error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).Derive(name, mnemonic, bip39Passwd, encryptPasswd, params)
}

func (lkb lazyKeybase) CreateLedger(name string, algo SigningAlgo, account uint32, index uint32) (info Info, err error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).CreateLedger(name, algo, account, index)
}

func (lkb lazyKeybase) CreateOffline(name string, pubkey crypto.PubKey) (info Info, err error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).CreateOffline(name, pubkey)
}

func (lkb lazyKeybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return newDbKeybase(db).Update(name, oldpass, getNewpass)
}

func (lkb lazyKeybase) Import(name string, armor string) (err error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return newDbKeybase(db).Import(name, armor)
}

func (lkb lazyKeybase) ImportPubKey(name string, armor string) (err error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return newDbKeybase(db).ImportPubKey(name, armor)
}

func (lkb lazyKeybase) Export(name string) (armor string, err error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	return newDbKeybase(db).Export(name)
}

func (lkb lazyKeybase) ExportPubKey(name string) (armor string, err error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	return newDbKeybase(db).ExportPubKey(name)
}

func (lkb lazyKeybase) ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error) {
	db, err := lkb.newGoLevelDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).ExportPrivateKeyObject(name, passphrase)
}

func (lkb lazyKeybase) CloseDB() {}

// newGoLevelDB returns a new reference to a GoLevelDB instance with the
// appropriate file permissions overridden. An error is returned if the DB
// cannot be opened or if updating the file permissions fails. In the latter
// case, the DB is closed. If no error is returned, it is up to the caller to
// close the DB connection handle.
func (lkb lazyKeybase) newGoLevelDB() (*dbm.GoLevelDB, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}

	if err := chmodR(lkb.dir, 0700); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// chmodR recursively invokes chmod with a given permissions os.FileMode on
// path and all of its contents. An error is returned if walking the directory
// path fails or any chmod call fails.
func chmodR(path string, mode os.FileMode) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chmod(name, mode)
		}
		return err
	})
}
