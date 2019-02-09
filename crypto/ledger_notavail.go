// +build !cgo !ledger
// test_ledger_mock

package crypto

import (
	"github.com/pkg/errors"
)

// If ledger support (build tag) has been enabled, which implies a CGO dependency,
// set the discoverLedger function which is responsible for loading the Ledger
// device at runtime or returning an error.
func init() {
	discoverLedger = func() (LedgerSECP256K1, error) {
		return nil, errors.New("support for ledger devices is not available in this executable")
	}
}
