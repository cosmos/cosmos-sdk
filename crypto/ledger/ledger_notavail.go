//go:build !cgo || !ledger
// +build !cgo !ledger

// test_ledger_mock

package ledger

import (
	"errors"
)

// If ledger support (build tag) has been disabled, which means no CGO dependency,
// set the discoverLedger function to return an error indicating that ledger
// device support is not available in this executable.
func init() {
	options.discoverLedger = func() (SECP256K1, error) {
		return nil, errors.New("support for ledger devices is not available in this executable")
	}

	initOptionsDefault()
}
