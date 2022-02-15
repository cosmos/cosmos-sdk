package upgrade_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type TestSuite struct {
	module  module.AppModule
	keeper  keeper.Keeper
	querier sdk.Querier
	handler govtypes.Handler
	ctx     sdk.Context
}

var s TestSuite

func setupTest(t *testing.T, height int64, skip map[int64]bool) TestSuite {

	db := memdb.NewDB()
	app := simapp.NewSimappWithCustomOptions(t, false, simapp.SetupOptions{
		Logger:             log.NewNopLogger(),
		SkipUpgradeHeights: skip,
		DB:                 db,
		InvCheckPeriod:     0,
		HomePath:           simapp.DefaultNodeHome,
		EncConfig:          simapp.MakeTestEncodingConfig(),
		AppOpts:            simapp.EmptyAppOptions{},
	})

	s.keeper = app.UpgradeKeeper
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{Height: height, Time: time.Now()})

	s.module = upgrade.NewAppModule(s.keeper)
	s.querier = s.module.LegacyQuerierHandler(app.LegacyAmino())
	s.handler = upgrade.NewSoftwareUpgradeProposalHandler(s.keeper)
	return s
}

func TestRequireName(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})

	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{}})
	require.NotNil(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestRequireFutureBlock(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight()}})
	require.NotNil(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestDoHeightUpgrade(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Verify can schedule an upgrade")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.Nil(t, err)

	VerifyDoUpgrade(t)
}

func TestCanOverwriteScheduleUpgrade(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Can overwrite plan")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "bad_test", Height: s.ctx.BlockHeight() + 10}})
	require.Nil(t, err)
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.Nil(t, err)

	VerifyDoUpgrade(t)
}

func VerifyDoUpgrade(t *testing.T) {
	t.Log("Verify that a panic happens at the upgrade height")
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler("test", func(ctx sdk.Context, plan types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyCleared(t, newCtx)
}

func VerifyDoUpgradeWithCtx(t *testing.T, newCtx sdk.Context, proposalName string) {
	t.Log("Verify that a panic happens at the upgrade height")
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler(proposalName, func(ctx sdk.Context, plan types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return vm, nil
	})
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyCleared(t, newCtx)
}

func TestHaltIfTooNew(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Verify that we don't panic with registered plan not in database at all")
	var called int
	s.keeper.SetUpgradeHandler("future", func(ctx sdk.Context, plan types.Plan, vm module.VersionMap) (module.VersionMap, error) {
		called++
		return vm, nil
	})

	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})
	require.Equal(t, 0, called)

	t.Log("Verify we panic if we have a registered handler ahead of time")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "future", Height: s.ctx.BlockHeight() + 3}})
	require.NoError(t, err)
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})
	require.Equal(t, 0, called)

	t.Log("Verify we no longer panic if the plan is on time")

	futCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 3).WithBlockTime(time.Now())
	req = abci.RequestBeginBlock{Header: futCtx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(futCtx, req)
	})
	require.Equal(t, 1, called)

	VerifyCleared(t, futCtx)
}

func VerifyCleared(t *testing.T, newCtx sdk.Context) {
	t.Log("Verify that the upgrade plan has been cleared")
	bz, err := s.querier(newCtx, []string{types.QueryCurrent}, abci.RequestQuery{})
	require.NoError(t, err)
	require.Nil(t, bz, string(bz))
}

func TestCanClear(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Verify upgrade is scheduled")
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 100}})
	require.Nil(t, err)

	err = s.handler(s.ctx, &types.CancelSoftwareUpgradeProposal{Title: "cancel"})
	require.Nil(t, err)

	VerifyCleared(t, s.ctx)
}

func TestCantApplySameUpgradeTwice(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	height := s.ctx.BlockHeader().Height + 1
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: height}})
	require.Nil(t, err)
	VerifyDoUpgrade(t)
	t.Log("Verify an executed upgrade \"test\" can't be rescheduled")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: height}})
	require.NotNil(t, err)
	require.True(t, errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func TestNoSpuriousUpgrades(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	t.Log("Verify that no upgrade panic is triggered in the BeginBlocker when we haven't scheduled an upgrade")
	req := abci.RequestBeginBlock{Header: s.ctx.BlockHeader()}
	require.NotPanics(t, func() {
		s.module.BeginBlock(s.ctx, req)
	})
}

func TestPlanStringer(t *testing.T) {
	require.Equal(t, `Upgrade Plan
  Name: test
  height: 100
  Info: .`, types.Plan{Name: "test", Height: 100, Info: ""}.String())

	require.Equal(t, fmt.Sprintf(`Upgrade Plan
  Name: test
  height: 100
  Info: .`), types.Plan{Name: "test", Height: 100, Info: ""}.String())
}

func VerifyNotDone(t *testing.T, newCtx sdk.Context, name string) {
	t.Log("Verify that upgrade was not done")
	height := s.keeper.GetDoneHeight(newCtx, name)
	require.Zero(t, height)
}

func VerifyDone(t *testing.T, newCtx sdk.Context, name string) {
	t.Log("Verify that the upgrade plan has been executed")
	height := s.keeper.GetDoneHeight(newCtx, name)
	require.NotZero(t, height)
}

func VerifySet(t *testing.T, skipUpgradeHeights map[int64]bool) {
	t.Log("Verify if the skip upgrade has been set")

	for k := range skipUpgradeHeights {
		require.True(t, s.keeper.IsSkipHeight(k))
	}
}

func TestContains(t *testing.T) {
	var (
		skipOne int64 = 11
	)
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

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.NoError(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases")
	VerifySet(t, map[int64]bool{skipOne: true, skipTwo: true})

	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify a second proposal also is being cleared")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.NoError(t, err)

	newCtx = newCtx.WithBlockHeight(skipTwo)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

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

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.Nil(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in one case and does upgrade on another")
	VerifySet(t, map[int64]bool{skipOne: true})

	// Setting block height of proposal test
	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify the second proposal is not skipped")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.Nil(t, err)
	// Setting block height of proposal test2
	newCtx = newCtx.WithBlockHeight(skipTwo)
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

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: skipOne}})
	require.Nil(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases and does third upgrade")
	VerifySet(t, map[int64]bool{skipOne: true, skipTwo: true})

	// Setting block height of proposal test
	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	// A new proposal with height in skipUpgradeHeights
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop2", Plan: types.Plan{Name: "test2", Height: skipTwo}})
	require.Nil(t, err)
	// Setting block height of proposal test2
	newCtx = newCtx.WithBlockHeight(skipTwo)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify a new proposal is not skipped")
	err = s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop3", Plan: types.Plan{Name: "test3", Height: skipThree}})
	require.Nil(t, err)
	newCtx = newCtx.WithBlockHeight(skipThree)
	VerifyDoUpgradeWithCtx(t, newCtx, "test3")

	t.Log("Verify two proposals are cleared and third is done")
	VerifyNotDone(t, s.ctx, "test")
	VerifyNotDone(t, s.ctx, "test2")
	VerifyDone(t, s.ctx, "test3")
}

func TestUpgradeWithoutSkip(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "prop", Plan: types.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	require.Nil(t, err)
	t.Log("Verify if upgrade happens without skip upgrade")
	require.Panics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	VerifyDoUpgrade(t)
	VerifyDone(t, s.ctx, "test")
}

func TestDumpUpgradeInfoToFile(t *testing.T) {
	s := setupTest(t, 10, map[int64]bool{})
	require := require.New(t)

	// require no error when the upgrade info file does not exist
	_, err := s.keeper.ReadUpgradeInfoFromDisk()
	require.NoError(err)

	planHeight := s.ctx.BlockHeight() + 1
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
		preRun      func() (sdk.Context, abci.RequestBeginBlock)
		expectPanic bool
	}{
		{
			"test not panic: no scheduled upgrade or applied upgrade is present",
			func() (sdk.Context, abci.RequestBeginBlock) {
				req := abci.RequestBeginBlock{Header: s.ctx.BlockHeader()}
				return s.ctx, req
			},
			false,
		},
		{
			"test not panic: upgrade handler is present for last applied upgrade",
			func() (sdk.Context, abci.RequestBeginBlock) {
				s.keeper.SetUpgradeHandler("test0", func(_ sdk.Context, _ types.Plan, vm module.VersionMap) (module.VersionMap, error) {
					return vm, nil
				})

				err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "Upgrade test", Plan: types.Plan{Name: "test0", Height: s.ctx.BlockHeight() + 2}})
				require.Nil(t, err)

				newCtx := s.ctx.WithBlockHeight(12)
				s.keeper.ApplyUpgrade(newCtx, types.Plan{
					Name:   "test0",
					Height: 12,
				})

				req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
				return newCtx, req
			},
			false,
		},
		{
			"test panic: upgrade needed",
			func() (sdk.Context, abci.RequestBeginBlock) {
				err := s.handler(s.ctx, &types.SoftwareUpgradeProposal{Title: "Upgrade test", Plan: types.Plan{Name: "test2", Height: 13}})
				require.Nil(t, err)

				newCtx := s.ctx.WithBlockHeight(13)
				req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
				return newCtx, req
			},
			true,
		},
	}

	for _, tc := range testCases {
		ctx, req := tc.preRun()
		if tc.expectPanic {
			require.Panics(t, func() {
				s.module.BeginBlock(ctx, req)
			})
		} else {
			require.NotPanics(t, func() {
				s.module.BeginBlock(ctx, req)
			})
		}
	}
}
