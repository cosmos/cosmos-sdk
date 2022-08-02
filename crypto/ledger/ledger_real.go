//go:build cgo && ledger && !test_ledger_mock
// +build cgo,ledger,!test_ledger_mock

package ledger

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	ledger "github.com/cosmos/ledger-cosmos-go"
)

// If ledger support (build tag) has been enabled, which implies a CGO dependency,
// set the discoverLedger function which is responsible for loading the Ledger
// device at runtime or returning an error.
func init() {
	options.discoverLedger = func() (SECP256K1, error) {
		device, err := ledger.FindLedgerCosmosUserApp()
		if err != nil {
			return nil, err
		}

		return device, nil
	}

	// Set default values for Cosmos Ledger instance. These can be updated
	// by setting fields in the Keyring Options.
	options.createPubkey = func(key []byte) types.PubKey {
		return &secp256k1.PubKey{Key: key}
	}
	options.appName = "Cosmos"
	options.skipDERConversion = false
}
