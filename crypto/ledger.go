// +build cgo,ledger

package crypto

import (
	ledger "github.com/zondax/ledger-goclient"
)

// If ledger support (build tag) has been enabled, automically attempt to load
// and set the ledger device, ledgerDevice, if it has not already been set.
func init() {
	device, err := ledger.FindLedger()
	if err != nil {
		ledgerDeviceErr = err
	} else {
		ledgerDevice = device
	}
}
