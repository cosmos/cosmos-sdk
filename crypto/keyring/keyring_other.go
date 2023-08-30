//go:build !linux
// +build !linux

package keyring

import (
	"io"

	"github.com/cosmos/cosmos-sdk/codec"
)

func New(
	appName, backend, rootDir string, userInput io.Reader, cdc codec.Codec, opts ...Option,
) (Keyring, error) {
	return newKeyringGeneric(appName, backend, rootDir, userInput, cdc, opts...)
}
