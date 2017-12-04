package client

/*

import (
	"path/filepath"

	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/cryptostore"
	"github.com/tendermint/go-crypto/keys/storage/filestorage"
)

// KeySubdir is the directory name under root where we store the keys
const KeySubdir = "keys"

// GetKeyManager initializes a key manager based on the configuration
func GetKeyManager(rootDir string) keys.Manager {
	keyDir := filepath.Join(rootDir, KeySubdir)
	// TODO: smarter loading??? with language and fallback?
	codec := keys.MustLoadCodec("english")

	// and construct the key manager
	manager := cryptostore.New(
		cryptostore.SecretBox,
		filestorage.New(keyDir),
		codec,
	)
	return manager
}
*/
