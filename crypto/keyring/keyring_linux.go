//go:build linux
// +build linux

package keyring

import (
	"io"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/codec"
)

const BackendKeyctl = "keyctl"

func newKeyctlBackendConfig(appName, _ string, inpt io.Reader) keyring.Config {
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.KeyCtlBackend},
		ServiceName:     appName,
		KeyCtlScope:     "user",
		KeyCtlPerm:      0x3f3f0000,
	}
}

func New(
	appName, backend, rootDir string, userInput io.Reader, cdc codec.Codec, opts ...Option,
) (Keyring, error) {
	return newSupportedKeyring(appName, backend, rootDir, userInput, cdc, opts...)
}

// New creates a new instance of a keyring.
// Keyring options can be applied when generating the new instance.
// Available backends are "os", "file", "kwallet", "memory", "pass", "test".
func newSupportedKeyring(
	appName, backend, rootDir string, userInput io.Reader, cdc codec.Codec, opts ...Option,
) (Keyring, error) {
	var (
		db  keyring.Keyring
		err error
	)

	if backend != BackendKeyctl {
		return newKeyringGeneric(appName, backend, rootDir, userInput, cdc, opts...)
	}

	db, err = keyring.Open(newKeyctlBackendConfig(appName, "", userInput))
	if err != nil {
		return nil, err
	}

	return newKeystore(db, cdc, backend, opts...), nil
}
