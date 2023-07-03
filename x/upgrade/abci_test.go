package upgrade_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/upgrade"
	"cosmossdk.io/x/upgrade/keeper"
	"cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type TestSuite struct {
	suite.Suite

	module  appmodule.HasBeginBlocker
	keeper  *keeper.Keeper
	handler govtypesv1beta1.Handler
	ctx     sdk.Context
	baseApp *baseapp.BaseApp
	encCfg  moduletestutil.TestEncodingConfig
}

var s TestSuite

func setupTest(t *testing.T, height int64, skip map[int64]bool) *TestSuite {
	t.Helper()
	s.encCfg = moduletestutil.MakeTestEncodingConfig(upgrade.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))

	s.baseApp = baseapp.NewBaseApp(
		"upgrade",
		log.NewNopLogger(),
		testCtx.DB,
		s.encCfg.TxConfig.TxDecoder(),
	)

	s.keeper = keeper.NewKeeper(skip, storeService, s.encCfg.Codec, t.TempDir(), nil, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	s.keeper.SetVersionSetter(s.baseApp)

	s.ctx = testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now(), Height: height})

	s.module = upgrade.NewAppModule(s.keeper, addresscodec.NewBech32Codec("cosmos"))
	s.handler = upgrade.NewSoftwareUpgradeProposalHandler(s.keeper)
	return &s
}

func TestRequireName(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})

	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{}}) //nolint:staticcheck // we're testing deprecated code
	require.Error(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestRequireFutureBlock(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.HeaderInfo().Height - 1}}) //nolint:staticcheck // we're testing deprecated code
	require.Error(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestDoHeightUpgrade(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Verify can schedule an upgrade")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.HeaderInfo().Height + 1}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	VerifyDoUpgrade(t)
}

func TestCanOverwriteScheduleUpgrade(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Can overwrite plan")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "bad_test", Height: s.ctx.HeaderInfo().Height + 10}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.HeaderInfo().Height + 1}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	VerifyDoUpgrade(t)
}

func VerifyDoUpgrade(t *testing.T) {
	t.Helper()
	t.Log("Verify that a panic happens at the upgrade height")
	newCtx := s.ctx.WithHeaderInfo(header.Info{Height: s.ctx.HeaderInfo().Height + 1, Time: time.Now()})

	err := s.module.BeginBlock(newCtx)
	require.ErrorContains(t, err, "UPGRADE \"test\" NEEDED at height: 11: ")

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler("test", func(ctx context.Context, plan types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})

	err = s.module.BeginBlock(newCtx)
	require.NoError(t, err)

	VerifyCleared(t, newCtx)
}

func VerifyDoUpgradeWithCtx(t *testing.T, newCtx sdk.Context, proposalName string) {
	t.Helper()
	t.Log("Verify that a panic happens at the upgrade height")

	err := s.module.BeginBlock(newCtx)
	require.ErrorContains(t, err, "UPGRADE \""+proposalName+"\" NEEDED at height: ")

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler(proposalName, func(ctx context.Context, plan types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})

	err = s.module.BeginBlock(newCtx)
	require.NoError(t, err)

	VerifyCleared(t, newCtx)
}

func TestHaltIfTooNew(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Verify that we don't panic with registered plan not in database at all")
	var called int
	s.keeper.SetUpgradeHandler("future", func(_ context.Context, _ types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		called++
		return vm, nil
	})

	newCtx := s.ctx.WithHeaderInfo(header.Info{Height: s.ctx.HeaderInfo().Height + 1, Time: time.Now()})

	err := s.module.BeginBlock(newCtx)
	require.NoError(t, err)
	require.Equal(t, 0, called)

	t.Log("Verify we error if we have a registered handler ahead of time")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "future", Height: s.ctx.HeaderInfo().Height + 3}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	err = s.module.BeginBlock(newCtx)
	require.EqualError(t, err, "BINARY UPDATED BEFORE TRIGGER! UPGRADE \"future\" - in binary but not executed on chain. Downgrade your binary")
	require.Equal(t, 0, called)

	t.Log("Verify we no longer panic if the plan is on time")

	futCtx := s.ctx.WithHeaderInfo(header.Info{Height: s.ctx.HeaderInfo().Height + 3, Time: time.Now()})
	err = s.module.BeginBlock(futCtx)
	require.NoError(t, err)
	require.Equal(t, 1, called)

	VerifyCleared(t, futCtx)
}

func VerifyCleared(t *testing.T, newCtx sdk.Context) {
	t.Helper()
	t.Log("Verify that the upgrade plan has been cleared")
	_, err := s.keeper.GetUpgradePlan(newCtx)
	require.ErrorIs(t, err, types.ErrNoUpgradePlanFound)
}

func TestCanClear(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Verify upgrade is scheduled")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.HeaderInfo().Height + 100}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	err = s.handler(s.ctx, &types.CancelSoftwareUpgradeProposal{Title: "cancel"}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	VerifyCleared(t, s.ctx)
}

func TestCantApplySameUpgradeTwice(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	height := s.ctx.HeaderInfo().Height + 1
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: height}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)
	VerifyDoUpgrade(t)
	t.Log("Verify an executed upgrade \"test\" can't be rescheduled")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: height}}) //nolint:staticcheck // we're testing deprecated code
	require.Error(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestNoSpuriousUpgrades(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Verify that no upgrade panic is triggered in the BeginBlocker when we haven't scheduled an upgrade")
	err := s.module.BeginBlock(s.ctx)
	require.NoError(t, err)
}

func TestPlanStringer(t *testing.T) {
	require.Equal(t, "name:\"test\" time:<seconds:-62135596800 > height:100 ", (&types.Plan{Name: "test", Height: 100, Info: ""}).String())
	require.Equal(t, `name:"test" time:<seconds:-62135596800 > height:100 `, (&types.Plan{Name: "test", Height: 100, Info: ""}).String())
}

func VerifyNotDone(t *testing.T, newCtx sdk.Context, name string) {
	t.Helper()
	t.Log("Verify that upgrade was not done")
	height, err := s.keeper.GetDoneHeight(newCtx, name)
	require.Zero(t, height)
	require.NoError(t, err)
}

func VerifyDone(t *testing.T, newCtx sdk.Context, name string) {
	t.Helper()
	t.Log("Verify that the upgrade plan has been executed")
	height, err := s.keeper.GetDoneHeight(newCtx, name)
	require.NotZero(t, height)
	require.NoError(t, err)
}

func VerifySet(t *testing.T, skipUpgradeHeights map[int64]bool) {
	t.Helper()
	t.Log("Verify if the skip upgrade has been set")

	for k := range skipUpgradeHeights {
		require.True(t, s.keeper.IsSkipHeight(k))
	}
}

func TestContains(t *testing.T) {
	var skipOne int64 = 11
	s := setupTest(t, 10, map[int64]bool{skipOne: true})

	VerifySet(t, map[int64]bool{skipOne: true})
	t.Log("case where array contains the element")
	require.True(t, s.keeper.IsSkipHeight(11))

	t.Log("case where array doesn't contain the element")
	require.False(t, s.keeper.IsSkipHeight(4))
}

func TestSkipUpgradeSkippingAll(t *testing.T) {
	var (
		skipOne int64 = 11
		skipTwo int64 = 20
	)
	s := setupTest(t, 10, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx := s.ctx

	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases")
	VerifySet(t, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx = newCtx.WithHeaderInfo(header.Info{Height: skipOne})
	err = s.module.BeginBlock(newCtx)
	require.NoError(t, err)

	t.Log("Verify a second proposal also is being cleared")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	newCtx = newCtx.WithHeaderInfo(header.Info{Height: skipTwo})
	err = s.module.BeginBlock(newCtx)
	require.NoError(t, err)

	// To ensure verification is being done only after both upgrades are cleared
	t.Log("Verify if both proposals are cleared")
	VerifyCleared(t, s.ctx)
	VerifyNotDone(t, s.ctx, "test")
	VerifyNotDone(t, s.ctx, "test2")
}

func TestUpgradeSkippingOne(t *testing.T) {
	var (
		skipOne int64 = 11
		skipTwo int64 = 20
	)
	s := setupTest(t, 10, map[int64]bool{skipOne: true})

	newCtx := s.ctx

	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in one case and does upgrade on another")
	VerifySet(t, map[int64]bool{skipOne: true})

	// Setting block height of proposal test
	newCtx = newCtx.WithHeaderInfo(header.Info{Height: skipOne})
	err = s.module.BeginBlock(newCtx)
	require.NoError(t, err)

	t.Log("Verify the second proposal is not skipped")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)
	// Setting block height of proposal test2
	newCtx = newCtx.WithHeaderInfo(header.Info{Height: skipTwo})
	VerifyDoUpgradeWithCtx(t, newCtx, "test2")

	t.Log("Verify first proposal is cleared and second is done")
	VerifyNotDone(t, s.ctx, "test")
	VerifyDone(t, s.ctx, "test2")
}

func TestUpgradeSkippingOnlyTwo(t *testing.T) {
	var (
		skipOne   int64 = 11
		skipTwo   int64 = 20
		skipThree int64 = 25
	)
	s := setupTest(t, 10, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx := s.ctx

	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases and does third upgrade")
	VerifySet(t, map[int64]bool{skipOne: true, skipTwo: true})

	// Setting block height of proposal test
	newCtx = newCtx.WithHeaderInfo(header.Info{Height: skipOne})
	err = s.module.BeginBlock(newCtx)
	require.NoError(t, err)

	// A new proposal with height in skipUpgradeHeights
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)
	// Setting block height of proposal test2
	newCtx = newCtx.WithHeaderInfo(header.Info{Height: skipTwo})
	err = s.module.BeginBlock(newCtx)
	require.NoError(t, err)

	t.Log("Verify a new proposal is not skipped")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop3", Plan: types.Plan{Name: "test3", Height: skipThree}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)
	newCtx = newCtx.WithHeaderInfo(header.Info{Height: skipThree})
	VerifyDoUpgradeWithCtx(t, newCtx, "test3")

	t.Log("Verify two proposals are cleared and third is done")
	VerifyNotDone(t, s.ctx, "test")
	VerifyNotDone(t, s.ctx, "test2")
	VerifyDone(t, s.ctx, "test3")
}

func TestUpgradeWithoutSkip(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	newCtx := s.ctx.WithHeaderInfo(header.Info{Height: s.ctx.HeaderInfo().Height + 1, Time: time.Now()})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.HeaderInfo().Height + 1}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)
	t.Log("Verify if upgrade happens without skip upgrade")

	err = s.module.BeginBlock(newCtx)
	require.ErrorContains(t, err, "UPGRADE \"test\" NEEDED at height:")

	VerifyDoUpgrade(t)
	VerifyDone(t, s.ctx, "test")
}

func TestDumpUpgradeInfoToFile(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	require := require.New(t)

	// require no error when the upgrade info file does not exist
	_, err := s.keeper.ReadUpgradeInfoFromDisk()
	require.NoError(err)

	planHeight := s.ctx.HeaderInfo().Height + 1
	plan := types.Plan{
		Name:   "test",
		Height: 0, // this should be overwritten by DumpUpgradeInfoToFile
	}
	t.Log("verify if upgrade height is dumped to file")
	err = s.keeper.DumpUpgradeInfoToDisk(planHeight, plan)
	require.Nil(err)

	upgradeInfo, err := s.keeper.ReadUpgradeInfoFromDisk()
	require.NoError(err)

	t.Log("Verify upgrade height from file matches ")
	require.Equal(upgradeInfo.Height, planHeight)
	require.Equal(upgradeInfo.Name, plan.Name)

	// clear the test file
	upgradeInfoFilePath, err := s.keeper.GetUpgradeInfoPath()
	require.Nil(err)
	err = os.Remove(upgradeInfoFilePath)
	require.Nil(err)
}

// TODO: add testcase to for `no upgrade handler is present for last applied upgrade`.
func TestBinaryVersion(t *testing.T) {
	var skipHeight int64 = 15
	s := setupTest(t, 10, map[int64]bool{skipHeight: true})

	testCases := []struct {
		name        string
		preRun      func() sdk.Context
		expectError bool
	}{
		{
			"test not panic: no scheduled upgrade or applied upgrade is present",
			func() sdk.Context {
				return s.ctx
			},
			false,
		},
		{
			"test not panic: upgrade handler is present for last applied upgrade",
			func() sdk.Context {
				s.keeper.SetUpgradeHandler("test0", func(_ context.Context, _ types.Plan, vm module.VersionMap) (module.VersionMap, error) {
					return vm, nil
				})

				err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "Upgrade test", Plan: types.Plan{Name: "test0", Height: s.ctx.HeaderInfo().Height + 2}}) //nolint:staticcheck // we're testing deprecated code
				require.NoError(t, err)

				newCtx := s.ctx.WithHeaderInfo(header.Info{Height: 12})
				s.keeper.ApplyUpgrade(newCtx, types.Plan{
					Name:   "test0",
					Height: 12,
				})

				return newCtx
			},
			false,
		},
		{
			"test panic: upgrade needed",
			func() sdk.Context {
				err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "Upgrade test", Plan: types.Plan{Name: "test2", Height: 13}}) //nolint:staticcheck // we're testing deprecated code
				require.NoError(t, err)

				newCtx := s.ctx.WithHeaderInfo(header.Info{Height: 13})
				return newCtx
			},
			true,
		},
	}

	for _, tc := range testCases {
		ctx := tc.preRun()
		err := s.module.BeginBlock(ctx)
		if tc.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestDowngradeVerification(t *testing.T) {
	// could not use setupTest() here, because we have to use the same key
	// for the two keepers.
	encCfg := moduletestutil.MakeTestEncodingConfig(upgrade.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now(), Height: 10})

	skip := map[int64]bool{}
	tempDir := t.TempDir()
	k := keeper.NewKeeper(skip, storeService, encCfg.Codec, tempDir, nil, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	m := upgrade.NewAppModule(k, addresscodec.NewBech32Codec("cosmos"))
	handler := upgrade.NewSoftwareUpgradeProposalHandler(k)

	// submit a plan.
	planName := "downgrade"
	err := handler(ctx, &types.SoftwareUpgradeProposal{Title: "test", Plan: types.Plan{Name: planName, Height: ctx.HeaderInfo().Height + 1}}) //nolint:staticcheck // we're testing deprecated code
	require.NoError(t, err)
	ctx = ctx.WithHeaderInfo(header.Info{Height: ctx.HeaderInfo().Height + 1})

	// set the handler.
	k.SetUpgradeHandler(planName, func(_ context.Context, _ types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})

	// successful upgrade.
	require.NoError(t, m.BeginBlock(ctx))
	ctx = ctx.WithHeaderInfo(header.Info{Height: ctx.HeaderInfo().Height + 1})

	testCases := map[string]struct {
		preRun      func(*keeper.Keeper, sdk.Context, string)
		expectError bool
	}{
		"valid binary": {
			preRun: func(k *keeper.Keeper, ctx sdk.Context, name string) {
				k.SetUpgradeHandler(planName, func(ctx context.Context, plan types.Plan, vm module.VersionMap) (module.VersionMap, error) {
					return vm, nil
				})
			},
		},
		"downgrade with an active plan": {
			preRun: func(k *keeper.Keeper, ctx sdk.Context, name string) {
				handler := upgrade.NewSoftwareUpgradeProposalHandler(k)
				err := handler(ctx, &types.SoftwareUpgradeProposal{Title: "test", Plan: types.Plan{Name: "another" + planName, Height: ctx.HeaderInfo().Height + 1}}) //nolint:staticcheck // we're testing deprecated code
				require.NoError(t, err, name)
			},
			expectError: true,
		},
		"downgrade without any active plan": {
			expectError: true,
		},
	}

	for name, tc := range testCases {
		ctx, _ := ctx.CacheContext()

		// downgrade. now keeper does not have the handler.
		k := keeper.NewKeeper(skip, storeService, encCfg.Codec, tempDir, nil, authtypes.NewModuleAddress(govtypes.ModuleName).String())
		m := upgrade.NewAppModule(k, addresscodec.NewBech32Codec("cosmos"))

		// assertions
		lastAppliedPlan, _, err := k.GetLastCompletedUpgrade(ctx)
		require.NoError(t, err)
		require.Equal(t, planName, lastAppliedPlan)
		require.False(t, k.HasHandler(planName))
		require.False(t, k.DowngradeVerified())
		_, err = k.GetUpgradePlan(ctx)
		require.ErrorIs(t, err, types.ErrNoUpgradePlanFound)

		if tc.preRun != nil {
			tc.preRun(k, ctx, name)
		}

		err = m.BeginBlock(ctx)
		if tc.expectError {
			require.Error(t, err, name)
		} else {
			require.NoError(t, err, name)
		}
	}
}
