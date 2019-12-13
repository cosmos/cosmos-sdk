package upgrade_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
)

type TestSuite struct {
	// TODO: remove this when we
	suite.Suite

	module                 module.AppModule
	keeper                 upgrade.Keeper
	querier                sdk.Querier
	handler                gov.Handler
	ctx                    sdk.Context
	FlagUnsafeSkipUpgrades string
}

func (s *TestSuite) SetupTest() {
	setupTestSuiteInPlace(s, 10, []int64{})
}

// this should be called by all functions that do not belong to the suite
func setupTest(height int64, skip []int64) TestSuite {
	var s TestSuite
	setupTestSuiteInPlace(&s, height, skip)
	return s
}

// this is a temporary way to unify TestSuite.SetupTest and setupTest
// can be merged into setupTest when TestSuite goes away
func setupTestSuiteInPlace(s *TestSuite, height int64, skip []int64) {
	// create the app with the proper skip flags set
	checkTx := false
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, skip, 0)
	simapp.SetupDeliverTx(app)

	// get info from the generic simapp
	s.keeper = app.UpgradeKeeper
	s.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: height, Time: time.Now()})
	s.FlagUnsafeSkipUpgrades = upgrade.FlagUnsafeSkipUpgrades

	// and construct a few upgrade-specific structs
	s.module = upgrade.NewAppModule(s.keeper)
	s.querier = s.module.NewQuerierHandler()
	s.handler = upgrade.NewSoftwareUpgradeProposalHandler(s.keeper)
}

func (s *TestSuite) TestRequireName() {
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestRequireFutureTime() {
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Time: s.ctx.BlockHeader().Time}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestRequireFutureBlock() {
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Height: s.ctx.BlockHeight()}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestCantSetBothTimeAndHeight() {
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Time: time.Now(), Height: s.ctx.BlockHeight() + 1}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestDoTimeUpgrade() {
	s.T().Log("Verify can schedule an upgrade")
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Time: time.Now()}})
	s.Require().Nil(err)

	s.VerifyDoUpgrade()
}

func (s *TestSuite) TestDoHeightUpgrade() {
	s.T().Log("Verify can schedule an upgrade")
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	s.Require().Nil(err)

	s.VerifyDoUpgrade()
}

func (s *TestSuite) TestCanOverwriteScheduleUpgrade() {
	s.T().Log("Can overwrite plan")
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "bad_test", Height: s.ctx.BlockHeight() + 10}})
	s.Require().Nil(err)
	err = s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	s.Require().Nil(err)

	s.VerifyDoUpgrade()
}

func (s *TestSuite) VerifyDoUpgrade() {
	s.T().Log("Verify that a panic happens at the upgrade time/height")
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	s.Require().Panics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.T().Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler("test", func(ctx sdk.Context, plan upgrade.Plan) {})
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.VerifyCleared(newCtx)
}

func (s *TestSuite) VerifyDoUpgradeWithCtx(newCtx sdk.Context, proposalName string) {
	s.T().Log("Verify that a panic happens at the upgrade time/height")
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	s.Require().Panics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.T().Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler(proposalName, func(ctx sdk.Context, plan upgrade.Plan) {})
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.VerifyCleared(newCtx)
}

func (s *TestSuite) TestHaltIfTooNew() {
	s.T().Log("Verify that we don't panic with registered plan not in database at all")
	var called int
	s.keeper.SetUpgradeHandler("future", func(ctx sdk.Context, plan upgrade.Plan) { called++ })

	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})
	s.Require().Equal(0, called)

	s.T().Log("Verify we panic if we have a registered handler ahead of time")
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "future", Height: s.ctx.BlockHeight() + 3}})
	s.Require().NoError(err)
	s.Require().Panics(func() {
		s.module.BeginBlock(newCtx, req)
	})
	s.Require().Equal(0, called)

	s.T().Log("Verify we no longer panic if the plan is on time")

	futCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 3).WithBlockTime(time.Now())
	req = abci.RequestBeginBlock{Header: futCtx.BlockHeader()}
	s.Require().NotPanics(func() {
		s.module.BeginBlock(futCtx, req)
	})
	s.Require().Equal(1, called)

	s.VerifyCleared(futCtx)
}

func (s *TestSuite) VerifyCleared(newCtx sdk.Context) {
	s.T().Log("Verify that the upgrade plan has been cleared")
	bz, err := s.querier(newCtx, []string{upgrade.QueryCurrent}, abci.RequestQuery{})
	s.Require().NoError(err)
	s.Require().Nil(bz)
}

func (s *TestSuite) TestCanClear() {
	s.T().Log("Verify upgrade is scheduled")
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Time: time.Now()}})
	s.Require().Nil(err)

	err = s.handler(s.ctx, upgrade.CancelSoftwareUpgradeProposal{Title: "cancel"})
	s.Require().Nil(err)

	s.VerifyCleared(s.ctx)
}

func (s *TestSuite) TestCantApplySameUpgradeTwice() {
	s.TestDoTimeUpgrade()
	s.T().Log("Verify an upgrade named \"test\" can't be scheduled twice")
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Time: time.Now()}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestNoSpuriousUpgrades() {
	s.T().Log("Verify that no upgrade panic is triggered in the BeginBlocker when we haven't scheduled an upgrade")
	req := abci.RequestBeginBlock{Header: s.ctx.BlockHeader()}
	s.Require().NotPanics(func() {
		s.module.BeginBlock(s.ctx, req)
	})
}

func (s *TestSuite) TestPlanStringer() {
	t, err := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	s.Require().Nil(err)
	s.Require().Equal(`Upgrade Plan
  Name: test
  Time: 2020-01-01T00:00:00Z
  Info: `, upgrade.Plan{Name: "test", Time: t}.String())
	s.Require().Equal(`Upgrade Plan
  Name: test
  Height: 100
  Info: `, upgrade.Plan{Name: "test", Height: 100}.String())
}

func (s *TestSuite) VerifyNotDone(newCtx sdk.Context, name string) {
	s.T().Log("Verify that upgrade was not done")
	height := s.keeper.GetDoneHeight(newCtx, name)
	s.Require().Zero(height)
}

func (s *TestSuite) VerifyDone(newCtx sdk.Context, name string) {
	s.T().Log("Verify that the upgrade plan has been executed")
	height := s.keeper.GetDoneHeight(newCtx, name)
	s.Require().NotZero(height)
}

func (s *TestSuite) VerifySet(skipUpgradeHeights []int64) {
	s.T().Log("Verify if the skip upgrade has been set")
	s.Require().Equal(s.keeper.GetSkipUpgradeHeights(), skipUpgradeHeights)
}

func (s *TestSuite) VerifyConversion(skipUpgrade []int) {
	skipUpgradeHeights := upgrade.ConvertIntArrayToInt64(skipUpgrade)
	s.Require().Equal(reflect.TypeOf(skipUpgradeHeights).Elem().Kind(), reflect.Int64)
}

func TestContains(t *testing.T) {
	var (
		skipOne int64 = 11
	)
	s := setupTest(10, []int64{skipOne})

	s.Suite.SetT(t)
	s.VerifySet([]int64{skipOne})
	s.T().Log("case where array contains the element")
	present := upgrade.Contains(s.keeper.GetSkipUpgradeHeights(), 11)
	s.Require().True(present)

	s.T().Log("case where array doesn't contain the element")
	present = upgrade.Contains(s.keeper.GetSkipUpgradeHeights(), 4)
	s.Require().False(present)
}

func TestSkipUpgradeSkippingAll(t *testing.T) {
	var (
		skipOne int64 = 11
		skipTwo int64 = 20
	)
	s := setupTest(10, []int64{skipOne, skipTwo})

	s.Suite.SetT(t)
	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Height: skipOne}})
	require.NoError(t, err)

	t.Log("Verify if skip upgrade flag clears upgrade plan in both cases")
	s.VerifySet([]int64{skipOne, skipTwo})

	newCtx = newCtx.WithBlockHeight(skipOne)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	t.Log("Verify a second proposal also is being cleared")
	err = s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop2", Plan: upgrade.Plan{Name: "test2", Height: skipTwo}})
	require.NoError(t, err)

	newCtx = newCtx.WithBlockHeight(skipTwo)
	require.NotPanics(t, func() {
		s.module.BeginBlock(newCtx, req)
	})

	//To ensure verification is being done only after both upgrades are cleared
	t.Log("Verify if both proposals are cleared")
	s.VerifyCleared(s.ctx)
	s.VerifyNotDone(s.ctx, "test")
	s.VerifyNotDone(s.ctx, "test2")
}

func TestUpgradeSkippingOne(t *testing.T) {
	var (
		skipOne int64 = 11
		skipTwo int64 = 20
	)
	s := setupTest(10, []int64{skipOne})

	s.Suite.SetT(t)
	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Height: skipOne}})
	require.Nil(t, err)

	s.T().Log("Verify if skip upgrade flag clears upgrade plan in one case and does upgrade on another")
	s.VerifySet([]int64{skipOne})

	//Setting block height of proposal test
	newCtx = newCtx.WithBlockHeight(skipOne)
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.T().Log("Verify the second proposal is not skipped")
	err = s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop2", Plan: upgrade.Plan{Name: "test2", Height: skipTwo}})
	s.Require().Nil(err)
	//Setting block height of proposal test2
	newCtx = newCtx.WithBlockHeight(skipTwo)
	s.VerifyDoUpgradeWithCtx(newCtx, "test2")

	s.T().Log("Verify first proposal is cleared and second is done")
	s.VerifyNotDone(s.ctx, "test")
	s.VerifyDone(s.ctx, "test2")
}

func TestUpgradeSkippingOnlyTwo(t *testing.T) {
	var (
		skipOne   int64 = 11
		skipTwo   int64 = 20
		skipThree int64 = 25
	)
	s := setupTest(10, []int64{skipOne, skipTwo})

	s.Suite.SetT(t)
	newCtx := s.ctx

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Height: skipOne}})
	s.Require().Nil(err)

	s.T().Log("Verify if skip upgrade flag clears upgrade plan in both cases and does third upgrade")
	s.VerifySet([]int64{skipOne, skipTwo})

	s.VerifyConversion(viper.GetIntSlice(s.FlagUnsafeSkipUpgrades))

	//Setting block height of proposal test
	newCtx = newCtx.WithBlockHeight(skipOne)
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	//A new proposal with height in skipUpgradeHeights
	err = s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop2", Plan: upgrade.Plan{Name: "test2", Height: skipTwo}})
	s.Require().Nil(err)
	//Setting block height of proposal test2
	newCtx = newCtx.WithBlockHeight(skipTwo)
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.T().Log("Verify a new proposal is not skipped")
	err = s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop3", Plan: upgrade.Plan{Name: "test3", Height: skipThree}})
	s.Require().Nil(err)
	newCtx = newCtx.WithBlockHeight(skipThree)
	s.VerifyDoUpgradeWithCtx(newCtx, "test3")

	s.T().Log("Verify two proposals are cleared and third is done")
	s.VerifyNotDone(s.ctx, "test")
	s.VerifyNotDone(s.ctx, "test2")
	s.VerifyDone(s.ctx, "test3")
}

func (s *TestSuite) TestUpgradeWithoutSkip() {
	newCtx := s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Now())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	s.Require().Nil(err)
	s.T().Log("Verify if upgrade happens without skip upgrade")
	s.Require().Panics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.VerifyDoUpgrade()
	s.VerifyDone(s.ctx, "test")
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
