package cli

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// InitFixtures is called at the beginning of a test  and initializes a chain
// with 1 validator.
func InitFixtures(t *testing.T) (f *Fixtures) {
	f = NewFixtures(t)

	// reset test state
	f.UnsafeResetAll()

	// ensure keystore has foo and bar keys
	f.KeysDelete(KeyFoo)
	f.KeysDelete(KeyBar)
	f.KeysDelete(KeyBar)
	f.KeysDelete(KeyFooBarBaz)
	f.KeysAdd(KeyFoo)
	f.KeysAdd(KeyBar)
	f.KeysAdd(KeyBaz)
	f.KeysAdd(KeyVesting)
	f.KeysAdd(KeyFooBarBaz, "--multisig-threshold=2", fmt.Sprintf("--multisig=%s,%s,%s", KeyFoo, KeyBar, KeyBaz))

	// NOTE: SDInit sets the ChainID
	f.SDInit(KeyFoo)

	// start an account with tokens
	f.AddGenesisAccount(f.KeyAddress(KeyFoo), StartCoins)
	f.AddGenesisAccount(
		f.KeyAddress(KeyVesting), StartCoins,
		fmt.Sprintf("--vesting-amount=%s", VestingCoins),
		fmt.Sprintf("--vesting-start-time=%d", time.Now().UTC().UnixNano()),
		fmt.Sprintf("--vesting-end-time=%d", time.Now().Add(60*time.Second).UTC().UnixNano()),
	)

	f.GenTx(KeyFoo)
	f.CollectGenTxs()

	return f
}

// Cleanup is meant to be run at the end of a test to clean up an remaining test state
func (f *Fixtures) Cleanup(dirs ...string) {
	clean := append(dirs, f.RootDir)
	for _, d := range clean {
		require.NoError(f.T, os.RemoveAll(d))
	}
}

// Flags returns the flags necessary for making most CLI calls
func (f *Fixtures) Flags() string {
	return fmt.Sprintf("--home=%s --node=%s --chain-id=%s --output=json", f.SimcliHome, f.RPCAddr, f.ChainID)
}
