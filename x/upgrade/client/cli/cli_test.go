package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/upgrade/client/cli"
	xp "github.com/cosmos/cosmos-sdk/x/upgrade/exported"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app     *simapp.SimApp
	cfg     network.Config
	network *network.Network
	ctx     sdk.Context
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	app := simapp.Setup(false)
	s.app = app
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{})

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestModuleVersionsCLI() {
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

	vm := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)
	s.Require().NotEmpty(vm)

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {

			if tc.expPass {
				// setup the expected map
				var reqVM module.VersionMap
				if tc.single {
					reqVM = make(module.VersionMap)
					reqVM[tc.req.ModuleName] = vm[tc.req.ModuleName]
				} else {
					reqVM = vm
				}

				// setup expected response
				req := make([]*types.ModuleVersion, 0)
				for m, v := range reqVM {
					req = append(req, &types.ModuleVersion{Name: m, Version: v})
				}
				req = xp.Sort(req)
				pm := types.QueryModuleVersionsResponse{
					ModuleVersions: req,
				}
				jsonVM, _ := clientCtx.JSONCodec.MarshalJSON(&pm)
				expectedVM := string(jsonVM)
				// append new line to match behaviour of PrintProto
				expectedVM += "\n"

				// get actual module versions list response from cli
				cmd := cli.GetModuleVersionsCmd()
				outVM, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{tc.req.ModuleName})
				s.Require().NoError(err)

				s.Require().Equal(expectedVM, outVM.String())
			} else {
				cmd := cli.GetModuleVersionsCmd()
				_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{tc.req.ModuleName})
				s.Require().Error(err)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
