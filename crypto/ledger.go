// +build cgo,ledger

package crypto

import (
	ledger "github.com/zondax/ledger-goclient"
)

// If ledger support (build tag) has been enabled, which implies a CGO dependency,
// set the discoverLedger function which is responsible for loading the Ledger
// device at runtime or returning an error.
func init() {
	discoverLedger = func() (LedgerSECP256K1, error) {
		device, err := ledger.FindLedger()
		if err != nil {
			return nil, err
		}

		return device, nil
	}
}
