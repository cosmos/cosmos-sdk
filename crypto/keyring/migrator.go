package keyring

import (
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/crypto"
)

// InfoImporter is implemented by those types that want to provide functions necessary
// to migrate keys from LegacyKeybase types to Keyring types.
type InfoImporter interface {
	Import(uid string, armor string) error
}

type keyringMigrator struct {
	kr keystore
}

func NewMigrator(
	appName, backend, rootDir string, userInput io.Reader, opts ...AltKeyringOption,
) (InfoImporter, error) {
	keyring, err := New(appName, backend, rootDir, userInput, opts...)
	if err != nil {
		return keyringMigrator{}, err
	}
	kr := keyring.(keystore)
	return keyringMigrator{kr}, nil
}

func (m keyringMigrator) Import(uid string, armor string) error {
	_, err := m.kr.Key(uid)
	if err == nil {
		return fmt.Errorf("cannot overwrite key %q", uid)
	}

	infoBytes, err := crypto.UnarmorInfoBytes(armor)
	if err != nil {
		return err
	}

	info, err := unmarshalInfo(infoBytes)
	if err != nil {
		return err
	}

	return m.kr.writeInfo(uid, info)
}
