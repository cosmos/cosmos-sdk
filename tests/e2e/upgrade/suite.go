package upgrade

import (
	"fmt"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/x/upgrade/client/cli"
	"cosmossdk.io/x/upgrade/keeper"
	"cosmossdk.io/x/upgrade/types"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewE2ETestSuite(cfg network.Config, keeper *keeper.Keeper, ctx sdk.Context) *E2ETestSuite {
	return &E2ETestSuite{
		cfg:           cfg,
		upgradeKeeper: keeper,
		ctx:           ctx,
	}
}

type E2ETestSuite struct {
	suite.Suite

	upgradeKeeper *keeper.Keeper
	cfg           network.Config
	network       *network.Network
	ctx           sdk.Context
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestModuleVersionsCLI() {
	testCases := []struct {
		msg     string
		req     types.QueryModuleVersionsRequest
		single  bool
		expPass bool
	}{
		{
			msg:     "test full query",
			req:     types.QueryModuleVersionsRequest{ModuleName: ""},
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

	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	// avoid printing as yaml from CLI command
	clientCtx.OutputFormat = "JSON"

	vm := s.upgradeKeeper.GetModuleVersionMap(s.ctx)
	mv := s.upgradeKeeper.GetModuleVersions(s.ctx)
	s.Require().NotEmpty(vm)

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			expect := mv
			if tc.expPass {
				if tc.single {
					expect = []*types.ModuleVersion{{Name: tc.req.ModuleName, Version: vm[tc.req.ModuleName]}}
				}
				// setup expected response
				pm := types.QueryModuleVersionsResponse{
					ModuleVersions: expect,
				}
				jsonVM, _ := clientCtx.Codec.MarshalJSON(&pm)
				expectedRes := string(jsonVM)
				// append new line to match behavior of PrintProto
				expectedRes += "\n"

				// get actual module versions list response from cli
				cmd := cli.GetModuleVersionsCmd()
				outVM, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{tc.req.ModuleName})
				s.Require().NoError(err)

				s.Require().Equal(expectedRes, outVM.String())
			} else {
				cmd := cli.GetModuleVersionsCmd()
				_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{tc.req.ModuleName})
				s.Require().Error(err)
			}
		})
	}
}
