package client

import (
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/words"
	dbm "github.com/tendermint/tmlibs/db"
)

// TODO explain what the keybase is used for
// GetKeyBase initializes a keybase based on the configuration
func GetKeyBase(db dbm.DB) keys.Keybase {
	keybase := keys.New(
		db,
		words.MustLoadCodec("english"),
	)
	return keybase
}

// MockKeyBase generates an in-memory keybase that will be discarded
// useful for --dry-run to generate a seed phrase without
// storing the key
func MockKeyBase() keys.Keybase {
	return GetKeyBase(dbm.NewMemDB())
}
