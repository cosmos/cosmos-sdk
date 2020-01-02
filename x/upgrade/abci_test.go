package upgrade_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

type TestSuite struct {
	suite.Suite

	module  module.AppModule
	keeper  upgrade.Keeper
	querier sdk.Querier
	handler gov.Handler
	ctx     sdk.Context
}

func (s *TestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	s.keeper = app.UpgradeKeeper
	s.handler = upgrade.NewSoftwareUpgradeProposalHandler(s.keeper)
	s.querier = upgrade.NewQuerier(s.keeper)
	s.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 10, Time: time.Now()})
	s.module = upgrade.NewAppModule(s.keeper)
}

func (s *TestSuite) TestRequireName() {
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{}})
	s.Require().NotNil(err)
	s.Require().True(errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func (s *TestSuite) TestRequireFutureTime() {
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Time: s.ctx.BlockHeader().Time}})
	s.Require().NotNil(err)
	s.Require().True(errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func (s *TestSuite) TestRequireFutureBlock() {
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Height: s.ctx.BlockHeight()}})
	s.Require().NotNil(err)
	s.Require().True(errors.Is(sdkerrors.ErrInvalidRequest, err), err)
}

func (s *TestSuite) TestCantSetBothTimeAndHeight() {
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Time: time.Now(), Height: s.ctx.BlockHeight() + 1}})
	s.Require().NotNil(err)
	s.Require().True(errors.Is(sdkerrors.ErrInvalidRequest, err), err)
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

	s.handler(s.ctx, upgrade.CancelSoftwareUpgradeProposal{Title: "cancel"})

	s.VerifyCleared(s.ctx)
}

func (s *TestSuite) TestCantApplySameUpgradeTwice() {
	s.TestDoTimeUpgrade()
	s.T().Log("Verify an upgrade named \"test\" can't be scheduled twice")
	err := s.handler(s.ctx, upgrade.SoftwareUpgradeProposal{Title: "prop", Plan: upgrade.Plan{Name: "test", Time: time.Now()}})
	s.Require().NotNil(err)
	s.Require().True(errors.Is(sdkerrors.ErrInvalidRequest, err), err)
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

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
