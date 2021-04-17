package derive

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	bip39 "github.com/cosmos/go-bip39"
)

func mnemonicToSeed(mnemonic string) []byte {
	return bip39.NewSeed(mnemonic, "" /* Default passphrase */)
}

func Fuzz(in []byte) int {
	splits := bytes.Split(in, []byte("*"))
        if len(splits) == 1 {
            return -1
        }
	mnemonic, path := splits[0], splits[1]
	seed := mnemonicToSeed(string(mnemonic))
	master, ch := hd.ComputeMastersFromSeed(seed)
	_, err := hd.DerivePrivateKeyForPath(master, ch, string(path))
	if err == nil {
		return 1
	}
	return -1
}
