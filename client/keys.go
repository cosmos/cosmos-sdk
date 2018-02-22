package client

import (
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/words"
	dbm "github.com/tendermint/tmlibs/db"
)

// KeyDBName is the directory under root where we store the keys
const KeyDBName = "keys"

// GetKeyBase initializes a keybase based on the configuration
func GetKeyBase(rootDir string) (keys.Keybase, error) {
	db, err := dbm.NewGoLevelDB(KeyDBName, rootDir)
	if err != nil {
		return nil, err
	}
	keybase := keys.New(
		db,
		words.MustLoadCodec("english"),
	)
	return keybase, nil
}

// MockKeyBase generates an in-memory keybase that will be discarded
// useful for --dry-run to generate a seed phrase without
// storing the key
func MockKeyBase() keys.Keybase {
	return keys.New(
		dbm.NewMemDB(),
		words.MustLoadCodec("english"),
	)
}
