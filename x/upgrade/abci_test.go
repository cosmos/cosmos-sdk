package upgrade

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

type TestSuite struct {
	suite.Suite
	keeper  Keeper
	querier sdk.Querier
	handler gov.Handler
	module  module.AppModule
	ctx     sdk.Context
	cms     store.CommitMultiStore
	FlagUnsafeSkipUpgrade string
}

func (s *TestSuite) SetupTest() {
	db := dbm.NewMemDB()
	s.cms = store.NewCommitMultiStore(db)
	key := sdk.NewKVStoreKey("upgrade")
	cdc := codec.New()
	RegisterCodec(cdc)
	s.keeper = NewKeeper(key, cdc)
	s.handler = NewSoftwareUpgradeProposalHandler(s.keeper)
	s.querier = NewQuerier(s.keeper)
	s.module = NewAppModule(s.keeper)
	s.cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	_ = s.cms.LoadLatestVersion()
	s.ctx = sdk.NewContext(s.cms, abci.Header{Height: 10, Time: time.Now()}, false, log.NewNopLogger())
	s.FlagUnsafeSkipUpgrade = FlagUnsafeSkipUpgrade
}

func (s *TestSuite) TestRequireName() {
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestRequireFutureTime() {
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Time: s.ctx.BlockHeader().Time}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestRequireFutureBlock() {
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Height: s.ctx.BlockHeight()}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestCantSetBothTimeAndHeight() {
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Time: time.Now(), Height: s.ctx.BlockHeight() + 1}})
	s.Require().NotNil(err)
	s.Require().Equal(sdk.CodeUnknownRequest, err.Code())
}

func (s *TestSuite) TestDoTimeUpgrade() {
	s.T().Log("Verify can schedule an upgrade")
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Time: time.Now()}})
	s.Require().Nil(err)

	s.VerifyDoUpgrade()
}

func (s *TestSuite) TestDoHeightUpgrade() {
	s.T().Log("Verify can schedule an upgrade")
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	s.Require().Nil(err)

	s.VerifyDoUpgrade()
}

func (s *TestSuite) TestCanOverwriteScheduleUpgrade() {
	s.T().Log("Can overwrite plan")
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "bad_test", Height: s.ctx.BlockHeight() + 10}})
	s.Require().Nil(err)
	err = s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	s.Require().Nil(err)

	s.VerifyDoUpgrade()
}

func (s *TestSuite) VerifyDoUpgrade() {
	s.T().Log("Verify that a panic happens at the upgrade time/height")
	newCtx := sdk.NewContext(s.cms, abci.Header{Height: s.ctx.BlockHeight() + 1, Time: time.Now()}, false, log.NewNopLogger())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	viper.Set(s.FlagUnsafeSkipUpgrade, false )
	s.Require().Panics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.T().Log("Verify that the upgrade can be successfully applied with a handler")
	s.keeper.SetUpgradeHandler("test", func(ctx sdk.Context, plan Plan) {})
	viper.Set(s.FlagUnsafeSkipUpgrade, false )
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.VerifyCleared(newCtx)
}

func (s *TestSuite) TestHaltIfTooNew() {
	s.T().Log("Verify that we don't panic with registered plan not in database at all")
	var called int
	s.keeper.SetUpgradeHandler("future", func(ctx sdk.Context, plan Plan) { called++ })

	newCtx := sdk.NewContext(s.cms, abci.Header{Height: s.ctx.BlockHeight() + 1, Time: time.Now()}, false, log.NewNopLogger())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})
	s.Require().Equal(0, called)

	s.T().Log("Verify we panic if we have a registered handler ahead of time")
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "future", Height: s.ctx.BlockHeight() + 3}})
	s.Require().NoError(err)
	s.Require().Panics(func() {
		s.module.BeginBlock(newCtx, req)
	})
	s.Require().Equal(0, called)

	s.T().Log("Verify we no longer panic if the plan is on time")

	futCtx := sdk.NewContext(s.cms, abci.Header{Height: s.ctx.BlockHeight() + 3, Time: time.Now()}, false, log.NewNopLogger())
	req = abci.RequestBeginBlock{Header: futCtx.BlockHeader()}
	s.Require().NotPanics(func() {
		s.module.BeginBlock(futCtx, req)
	})
	s.Require().Equal(1, called)

	s.VerifyCleared(futCtx)
}

func (s *TestSuite) VerifyCleared(newCtx sdk.Context) {
	s.T().Log("Verify that the upgrade plan has been cleared")
	bz, err := s.querier(newCtx, []string{QueryCurrent}, abci.RequestQuery{})
	s.Require().NoError(err)
	s.Require().Nil(bz)
}

func (s *TestSuite) TestCanClear() {
	s.T().Log("Verify upgrade is scheduled")
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Time: time.Now()}})
	s.Require().Nil(err)

	s.handler(s.ctx, CancelSoftwareUpgradeProposal{Title: "cancel"})

	s.VerifyCleared(s.ctx)
}

func (s *TestSuite) TestCantApplySameUpgradeTwice() {
	s.TestDoTimeUpgrade()
	s.T().Log("Verify an upgrade named \"test\" can't be scheduled twice")
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Time: time.Now()}})
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
  Info: `, Plan{Name: "test", Time: t}.String())
	s.Require().Equal(`Upgrade Plan
  Name: test
  Height: 100
  Info: `, Plan{Name: "test", Height: 100}.String())
}

func (s *TestSuite) TestSkipUpgrade()  {
	newCtx := sdk.NewContext(s.cms, abci.Header{Height: s.ctx.BlockHeight() + 1, Time: time.Now()}, false, log.NewNopLogger())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop", Plan: Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	s.Require().Nil(err)

	s.T().Log("Verify if skip upgrade flag clears upgrade plan")
	viper.Set(s.FlagUnsafeSkipUpgrade, true )
	s.Require().NotPanics(func() {
		s.module.BeginBlock(newCtx, req)
	})
	s.VerifyCleared(s.ctx)

}

func (s *TestSuite) TestUpgradeWithoutSkip() {
	newCtx := sdk.NewContext(s.cms, abci.Header{Height: s.ctx.BlockHeight() + 1, Time: time.Now()}, false, log.NewNopLogger())
	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	err := s.handler(s.ctx, SoftwareUpgradeProposal{Title: "prop1", Plan: Plan{Name: "test", Height: s.ctx.BlockHeight() + 1}})
	s.Require().Nil(err)
	s.T().Log("Verify if upgrade happens without skip upgrade")
	viper.Set(s.FlagUnsafeSkipUpgrade, false )
	s.Require().Panics(func() {
		s.module.BeginBlock(newCtx, req)
	})

	s.VerifyDoUpgrade()
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
