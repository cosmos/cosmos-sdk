package keyring

import (
	"fmt"

	tmos "github.com/tendermint/tendermint/libs/os"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewLegacy creates a new instance of a lazy keybase.
func NewLegacy(name, dir string, opts ...KeybaseOption) (LegacyKeybase, error) {
	if err := tmos.EnsureDir(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create Keybase directory: %s", err)
	}

	db, err := sdk.NewLevelDB(name, dir)
	if err != nil {
		return nil, err
	}
	return newDBKeybase(db, opts...), nil
}

// LegacyKeybase is implemented by the legacy keybase implementation.
type LegacyKeybase interface {
	List() ([]Info, error)
	Export(name string) (armor string, err error)
	ExportPrivKey(name, decryptPassphrase, encryptPassphrase string) (armor string, err error)
	ExportPubKey(name string) (armor string, err error)
	Update(name, oldpass string, getNewpass func() (string, error)) error
	Close() error
}
