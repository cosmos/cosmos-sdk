package client

import (
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

// MockKeyBase generates an in-memory keybase that will be discarded
// useful for --dry-run to generate a seed phrase without
// storing the key
func MockKeyBase() keys.Keybase {
	return keys.New(dbm.NewMemDB())
}
