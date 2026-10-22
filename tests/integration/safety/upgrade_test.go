package safety_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/module"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// TestUpgradeSetsOperatorWithNoGap runs the upgrade handler a chain moving
// off contrib/x/circuit uses: RunMigrations adds x/safety with its default
// params, which have no operator, and SetParams names the operator and its
// scope right after. A MsgTrip in the upgrade block itself succeeds, so the
// chain is never without a breaker the operator can trip.
func TestUpgradeSetsOperatorWithNoGap(t *testing.T) {
	const upgradeName = "circuit-to-safety"
	f := newFixture(t, nil, nil)
	operator := f.operator.String()
	require.Empty(t, f.params().Operator)

	height := f.app.LastBlockHeight() + 10
	f.mustPass(&upgradetypes.MsgSoftwareUpgrade{
		Authority: govAddr,
		Plan:      upgradetypes.Plan{Name: upgradeName, Height: height},
	})
	for f.app.LastBlockHeight() < height-1 {
		f.finalize()
	}

	// the new binary registers the handler
	app := f.app
	app.UpgradeKeeper.SetUpgradeHandler(upgradeName, func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// The fixture starts with x/safety in genesis. Delete it from
		// fromVM so RunMigrations adds it as new, as on a live chain.
		delete(fromVM, safetytypes.ModuleName)
		vm, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		if err != nil {
			return nil, err
		}
		return vm, app.SafetyKeeper.SetParams(ctx, safetytypes.Params{
			Authority: govAddr,
			Operator:  operator,
			Scope:     []string{safetytypes.Wildcard},
		})
	})

	// the trip runs in the upgrade block, after the PreBlocker has applied
	// the upgrade
	res := f.finalize(f.encode(f.operator, safetytypes.NewMsgTrip(operator, msgSendURL)))
	require.Zero(t, res.TxResults[0].Code, res.TxResults[0].Log)
	require.Equal(t, height, f.app.LastBlockHeight())
	require.Equal(t, operator, f.params().Operator)
	require.Equal(t, []string{msgSendURL}, f.disabled())
}
