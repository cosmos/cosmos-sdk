package keys

import (
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

func NewLazyKeybase(name, dir string) Keybase {
	return lazyKeybase{name: name, dir: dir}
}

func (lkb lazyKeybase) List() ([]Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return New(db).List()
}

func (lkb lazyKeybase) Get(name string) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return New(db).Get(name)
}

func (lkb lazyKeybase) GetByAddress(address types.AccAddress) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return New(db).GetByAddress(address)
}

func (lkb lazyKeybase) Delete(name, passphrase string, skipPass bool) error {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()
	return New(db).Delete(name, passphrase, skipPass)
}

func (lkb lazyKeybase) Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()
	return New(db).Sign(name, passphrase, msg)
}

func (lkb lazyKeybase) CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, "", err
	}
	defer db.Close()
	return New(db).CreateMnemonic(name, language, passwd, algo)
}

func (lkb lazyKeybase) CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return New(db).CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd, account, index)
}

func (lkb lazyKeybase) Derive(name, mnemonic, bip39Passwd, encryptPasswd string, params hd.BIP44Params) (Info, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return New(db).Derive(name, mnemonic, bip39Passwd, encryptPasswd, params)
}

func (lkb lazyKeybase) CreateLedger(name string, algo SigningAlgo, account uint32, index uint32) (info Info, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return New(db).CreateLedger(name, algo, account, index)
}

func (lkb lazyKeybase) CreateOffline(name string, pubkey crypto.PubKey) (info Info, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return New(db).CreateOffline(name, pubkey)
}

func (lkb lazyKeybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()
	return New(db).Update(name, oldpass, getNewpass)
}

func (lkb lazyKeybase) Import(name string, armor string) (err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()
	return New(db).Import(name, armor)
}

func (lkb lazyKeybase) ImportPubKey(name string, armor string) (err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()
	return New(db).ImportPubKey(name, armor)
}

func (lkb lazyKeybase) Export(name string) (armor string, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()
	return New(db).Export(name)
}

func (lkb lazyKeybase) ExportPubKey(name string) (armor string, err error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()
	return New(db).ExportPubKey(name)
}

func (lkb lazyKeybase) ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error) {
	db, err := dbm.NewGoLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return New(db).ExportPrivateKeyObject(name, passphrase)
}

func (lkb lazyKeybase) CloseDB() {}
