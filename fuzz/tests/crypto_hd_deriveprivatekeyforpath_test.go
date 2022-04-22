//go:build gofuzz || go1.18

package tests

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	bip39 "github.com/cosmos/go-bip39"
)

func mnemonicToSeed(mnemonic string) []byte {
	return bip39.NewSeed(mnemonic, "" /* Default passphrase */)
}

func FuzzCryptoHDDerivePrivateKeyForPath(f *testing.F) {
	f.Fuzz(func(t *testing.T, in []byte) {
		splits := bytes.Split(in, []byte("*"))
		if len(splits) == 1 {
			return
		}
		mnemonic, path := splits[0], splits[1]
		seed := mnemonicToSeed(string(mnemonic))
		master, ch := hd.ComputeMastersFromSeed(seed)
		hd.DerivePrivateKeyForPath(master, ch, string(path))
	})
}
