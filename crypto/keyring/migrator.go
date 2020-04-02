package keyring

import (
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/crypto"
)

type Migrator struct {
	kr keystore
}

func NewMigrator(
	appName, backend, rootDir string, userInput io.Reader, opts ...AltKeyringOption,
) (Migrator, error) {
	keyring, err := New(appName, backend, rootDir, userInput, opts...)
	if err != nil {
		return Migrator{}, err
	}
	kr := keyring.(keystore)
	return Migrator{kr}, nil
}

func (m Migrator) Import(uid string, armor string) error {
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
