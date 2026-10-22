package simapp

import (
	"testing"

	"github.com/stretchr/testify/require"

	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// x/safety is new in this version, so the upgrade from v0.54 must add its
// store, and every added store must be one the app mounts.
func TestUpgradeStoreUpgrades(t *testing.T) {
	app := Setup(t, false)
	upgrades := upgradeStoreUpgrades()

	require.Contains(t, upgrades.Added, safetytypes.StoreKey)
	for _, name := range upgrades.Added {
		require.NotNil(t, app.GetKey(name), "added store %s is not mounted", name)
	}
	for _, name := range upgrades.Deleted {
		require.Nil(t, app.GetKey(name), "deleted store %s is still mounted", name)
	}
}
