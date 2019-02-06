// +build cgo,ledger
// +build !test_ledger_mock

package crypto

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	ledger "github.com/zondax/ledger-cosmos-go"
)

// warnIfErrors wraps a function and writes a warning to stderr. This is required
// to avoid ignoring errors when defer is used. Using defer may result in linter warnings.
func panicIfErrors(f func() error) {
	if err := f(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, "received error when closing ledger connection", err)
	}
}

func TestDiscoverDevice(t *testing.T) {
	device, err := discoverLedger()
	require.NoError(t, err)
	require.NotNil(t, device)
	defer panicIfErrors(device.Close)
}

func TestDiscoverDeviceTwice(t *testing.T) {
	// We expect the second call not to find a device
	device, err := discoverLedger()
	require.NoError(t, err)
	require.NotNil(t, device)
	defer panicIfErrors(device.Close)

	device2, err := discoverLedger()
	require.Error(t, err)
	require.Equal(t, "no ledger connected", err.Error())
	require.Nil(t, device2)
}

func TestDiscoverDeviceTwiceClosing(t *testing.T) {
	{
		device, err := ledger.FindLedgerCosmosUserApp()
		require.NoError(t, err)
		require.NotNil(t, device)
		require.NoError(t, device.Close())
	}

	device2, err := discoverLedger()
	require.NoError(t, err)
	require.NotNil(t, device2)
	require.NoError(t, device2.Close())
}
