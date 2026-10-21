package safety_test

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	upgrademodulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/msgguard"
	safetytestutil "github.com/cosmos/cosmos-sdk/x/safety/testutil"
)

// TestDepinjectMsgServersCheckCircuitBreakerFirst builds a depinject app with
// every in-repo module that has a msg server, enterprise/group included, and
// requires each msg server method it registers to check the circuit breaker
// before anything else when called directly. Only the msgs x/safety always
// allows may skip the check.
func TestDepinjectMsgServersCheckCircuitBreakerFirst(t *testing.T) {
	var (
		appBuilder *runtime.AppBuilder
		cdc        codec.Codec
		registry   codectypes.InterfaceRegistry
	)
	require.NoError(t, depinject.Inject(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
				configurator.SlashingModule(),
				configurator.GenutilModule(),
				configurator.DistributionModule(),
				configurator.FeegrantModule(),
				configurator.VestingModule(),
				configurator.GovModule(),
				configurator.ConsensusModule(),
				configurator.MintModule(),
				configurator.EvidenceModule(),
				configurator.AuthzModule(),
				configurator.TxModule(),
				groupModule(),
				func(config *configurator.Config) {
					config.ModuleConfigs["upgrade"] = &appv1alpha1.ModuleConfig{
						Name:   "upgrade",
						Config: appconfig.WrapAny(&upgrademodulev1.Module{}),
					}
				},
				safetytestutil.SafetyModule(""),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&appBuilder, &cdc, &registry,
	))
	app := appBuilder.Build(dbm.NewMemDB())

	methods := msgguard.RequireAppGuarded(t, cdc, app.ModuleManager, registry, app.MsgServiceRouter())

	modules := map[string]bool{}
	for _, m := range methods {
		modules[m.Service] = true
	}
	for _, svc := range []string{
		"cosmos.auth.v1beta1.Msg", "cosmos.bank.v1beta1.Msg", "cosmos.staking.v1beta1.Msg",
		"cosmos.slashing.v1beta1.Msg", "cosmos.distribution.v1beta1.Msg", "cosmos.feegrant.v1beta1.Msg",
		"cosmos.vesting.v1beta1.Msg", "cosmos.gov.v1.Msg", "cosmos.consensus.v1.Msg", "cosmos.mint.v1beta1.Msg",
		"cosmos.evidence.v1beta1.Msg", "cosmos.authz.v1beta1.Msg", "cosmos.group.v1.Msg",
		"cosmos.upgrade.v1beta1.Msg", "cosmos.safety.v1.Msg",
	} {
		require.True(t, modules[svc], "the depinject app registers no %s service", svc)
	}
}
