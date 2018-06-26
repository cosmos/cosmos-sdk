package crypto

import (
	ledger "github.com/zondax/ledger-goclient"
)

var device *ledger.Ledger

// Ledger derivation path
type DerivationPath = []uint32

// getLedger gets a copy of the device, and caches it
func getLedger() (*ledger.Ledger, error) {
	var err error
	if device == nil {
		device, err = ledger.FindLedger()
	}
	return device, err
}
