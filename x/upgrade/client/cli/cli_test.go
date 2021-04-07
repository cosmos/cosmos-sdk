package cli_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/upgrade/client/cli"
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

func (s *IntegrationTestSuite) TestVersionMapCMD() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// avoid printing as yaml
	clientCtx.OutputFormat = "JSON"

	vm := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)
	s.Require().NotEmpty(vm)
	pm := types.QueryVersionMapResponse{
		VersionMap: vm,
	}
	jsonVM, err := clientCtx.JSONMarshaler.MarshalJSON(&pm)
	expectedVM := string(jsonVM)
	// append new line to match behaviour of PrintProto
	expectedVM += "\n"

	// get full versionmap from cli
	cmd := cli.GetVersionMapCmd()
	outVM, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{})
	s.Require().NoError(err)

	s.Require().Equal(expectedVM, outVM.String())
}

func (s *IntegrationTestSuite) TestSingleVersionMapCMD() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// avoid printing as yaml
	clientCtx.OutputFormat = "JSON"

	// create verisonmap response with only bank module
	testModule := "bank"
	vm := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)
	s.Require().NotEmpty(vm)
	singleVM := make(module.VersionMap)
	singleVM[testModule] = vm[testModule]
	pm := types.QueryVersionMapResponse{
		VersionMap: singleVM,
	}
	jsonVM, err := clientCtx.JSONMarshaler.MarshalJSON(&pm)
	expectedVM := string(jsonVM)
	// append new line to match behaviour of PrintProto
	expectedVM += "\n"

	// get versionmap from cli requesting the bank module
	cmd := cli.GetVersionMapCmd()
	outVM, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{testModule})
	s.Require().NoError(err)

	s.Require().Equal(expectedVM, outVM.String())
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
