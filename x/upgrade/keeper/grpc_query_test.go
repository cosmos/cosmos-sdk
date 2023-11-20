package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/upgrade"
	"cosmossdk.io/x/upgrade/keeper"
	"cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type UpgradeTestSuite struct {
	suite.Suite

	upgradeKeeper    *keeper.Keeper
	ctx              sdk.Context
	queryClient      types.QueryClient
	encCfg           moduletestutil.TestEncodingConfig
	encodedAuthority string
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(upgrade.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	skipUpgradeHeights := make(map[int64]bool)
	authority, err := addresscodec.NewBech32Codec("cosmos").BytesToString(authtypes.NewModuleAddress(types.GovModuleName))
	suite.Require().NoError(err)
	suite.encodedAuthority = authority
	suite.upgradeKeeper = keeper.NewKeeper(skipUpgradeHeights, storeService, suite.encCfg.Codec, suite.T().TempDir(), nil, authority)
	err = suite.upgradeKeeper.SetModuleVersionMap(suite.ctx, module.VersionMap{
		"bank": 0,
	})
	suite.Require().NoError(err)
	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, suite.encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.upgradeKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *UpgradeTestSuite) TestQueryCurrentPlan() {
	var (
		req         *types.QueryCurrentPlanRequest
		expResponse types.QueryCurrentPlanResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"without current upgrade plan",
			func() {
				req = &types.QueryCurrentPlanRequest{}
				expResponse = types.QueryCurrentPlanResponse{}
			},
			true,
		},
		{
			"with current upgrade plan",
			func() {
				plan := types.Plan{Name: "test-plan", Height: 5}
				err := suite.upgradeKeeper.ScheduleUpgrade(suite.ctx, plan)
				suite.Require().NoError(err)
				req = &types.QueryCurrentPlanRequest{}
				expResponse = types.QueryCurrentPlanResponse{Plan: &plan}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			res, err := suite.queryClient.CurrentPlan(context.Background(), req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(&expResponse, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *UpgradeTestSuite) TestAppliedCurrentPlan() {
	var (
		req       *types.QueryAppliedPlanRequest
		expHeight int64
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"with non-existent upgrade plan",
			func() {
				req = &types.QueryAppliedPlanRequest{Name: "foo"}
			},
			true,
		},
		{
			"with applied upgrade plan",
			func() {
				expHeight = 5

				planName := "test-plan"
				plan := types.Plan{Name: planName, Height: expHeight}
				err := suite.upgradeKeeper.ScheduleUpgrade(suite.ctx, plan)
				suite.Require().NoError(err)
				suite.ctx = suite.ctx.WithHeaderInfo(header.Info{Height: expHeight})
				suite.upgradeKeeper.SetUpgradeHandler(planName, func(ctx context.Context, plan types.Plan, vm module.VersionMap) (module.VersionMap, error) {
					return vm, nil
				})
				err = suite.upgradeKeeper.ApplyUpgrade(suite.ctx, plan)
				suite.Require().NoError(err)
				req = &types.QueryAppliedPlanRequest{Name: planName}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

			res, err := suite.queryClient.AppliedPlan(context.Background(), req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expHeight, res.Height)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *UpgradeTestSuite) TestModuleVersions() {
	testCases := []struct {
		msg     string
		req     types.QueryModuleVersionsRequest
		single  bool
		expPass bool
	}{
		{
			msg:     "test full query",
			req:     types.QueryModuleVersionsRequest{},
			single:  false,
			expPass: true,
		},
		{
			msg:     "test single module",
			req:     types.QueryModuleVersionsRequest{ModuleName: "bank"},
			single:  true,
			expPass: true,
		},
		{
			msg:     "test non-existent module",
			req:     types.QueryModuleVersionsRequest{ModuleName: "abcdefg"},
			single:  true,
			expPass: false,
		},
	}

	vm, err := suite.upgradeKeeper.GetModuleVersionMap(suite.ctx)
	suite.Require().NoError(err)

	mv, err := suite.upgradeKeeper.GetModuleVersions(suite.ctx)
	suite.Require().NoError(err)

	for _, tc := range testCases {
		tc := tc

		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			res, err := suite.queryClient.ModuleVersions(context.Background(), &tc.req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				if tc.single {
					// test that the single module response is valid
					suite.Require().Len(res.ModuleVersions, 1)
					// make sure we got the right values
					suite.Require().Equal(vm[tc.req.ModuleName], res.ModuleVersions[0].Version)
					suite.Require().Equal(tc.req.ModuleName, res.ModuleVersions[0].Name)
				} else {
					// check that the full response is valid
					suite.Require().NotEmpty(res.ModuleVersions)
					suite.Require().Equal(len(mv), len(res.ModuleVersions))
					for i, v := range res.ModuleVersions {
						suite.Require().Equal(mv[i].Version, v.Version)
						suite.Require().Equal(mv[i].Name, v.Name)
					}
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *UpgradeTestSuite) TestAuthority() {
	res, err := suite.queryClient.Authority(context.Background(), &types.QueryAuthorityRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.encodedAuthority, res.Address)
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}
