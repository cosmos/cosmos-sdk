package keys

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	tmos "github.com/tendermint/tendermint/libs/os"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Keybase = lazyKeybase{}

// NOTE: lazyKeybase will be deprecated in favor of lazyKeybaseKeyring.
type lazyKeybase struct {
	name    string
	dir     string
	options []KeybaseOption
}

// New creates a new instance of a lazy keybase.
func New(name, dir string, opts ...KeybaseOption) Keybase {
	if err := tmos.EnsureDir(dir, 0700); err != nil {
		panic(fmt.Sprintf("failed to create Keybase directory: %s", err))
	}

	return lazyKeybase{name: name, dir: dir, options: opts}
}

func (lkb lazyKeybase) List() ([]Info, error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).List()
}

func (lkb lazyKeybase) Get(name string) (Info, error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).Get(name)
}

func (lkb lazyKeybase) GetByAddress(address sdk.AccAddress) (Info, error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).GetByAddress(address)
}

func (lkb lazyKeybase) Delete(name, passphrase string, skipPass bool) error {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).Delete(name, passphrase, skipPass)
}

func (lkb lazyKeybase) Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).Sign(name, passphrase, msg)
}

func (lkb lazyKeybase) CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo, mnemonicInput string) (info Info, seed string, err error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, "", err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).CreateMnemonic(name, language, passwd, algo, mnemonicInput)
}

func (lkb lazyKeybase) CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd, hdPath string, algo SigningAlgo) (Info, error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDBKeybase(db,
		lkb.options...).CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd, hdPath, algo)
}

func (lkb lazyKeybase) CreateLedger(name string, algo SigningAlgo, hrp string, account, index uint32) (info Info, err error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).CreateLedger(name, algo, hrp, account, index)
}

func (lkb lazyKeybase) CreateOffline(name string, pubkey crypto.PubKey, algo SigningAlgo) (info Info, err error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).CreateOffline(name, pubkey, algo)
}

func (lkb lazyKeybase) CreateMulti(name string, pubkey crypto.PubKey) (info Info, err error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).CreateMulti(name, pubkey)
}

func (lkb lazyKeybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).Update(name, oldpass, getNewpass)
}

func (lkb lazyKeybase) Import(name string, armor string) (err error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).Import(name, armor)
}

func (lkb lazyKeybase) ImportPrivKey(name string, armor string, passphrase string) error {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).ImportPrivKey(name, armor, passphrase)
}

func (lkb lazyKeybase) ImportPubKey(name string, armor string) (err error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).ImportPubKey(name, armor)
}

func (lkb lazyKeybase) Export(name string) (armor string, err error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).Export(name)
}

func (lkb lazyKeybase) ExportPubKey(name string) (armor string, err error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).ExportPubKey(name)
}

func (lkb lazyKeybase) ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).ExportPrivateKeyObject(name, passphrase)
}

func (lkb lazyKeybase) ExportPrivKey(name string, decryptPassphrase string,
	encryptPassphrase string) (armor string, err error) {

	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()

	return newDBKeybase(db, lkb.options...).ExportPrivKey(name, decryptPassphrase, encryptPassphrase)
}

// SupportedAlgos returns a list of supported signing algorithms.
func (lkb lazyKeybase) SupportedAlgos() []SigningAlgo {
	return newBaseKeybase(lkb.options...).SupportedAlgos()
}

// SupportedAlgosLedger returns a list of supported ledger signing algorithms.
func (lkb lazyKeybase) SupportedAlgosLedger() []SigningAlgo {
	return newBaseKeybase(lkb.options...).SupportedAlgosLedger()
}

func (lkb lazyKeybase) CloseDB() {}
