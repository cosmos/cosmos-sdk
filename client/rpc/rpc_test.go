package rpc_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/rpc"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/version"
)

type IntegrationTestSuite struct {
	suite.Suite

	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), network.DefaultConfig())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestStatusCommand() {
	val0 := s.network.Validators[0]
	cmd := rpc.StatusCommand()

	out, err := clitestutil.ExecTestCLICmd(val0.ClientCtx, cmd, []string{})
	s.Require().NoError(err)

	// Make sure the output has the validator moniker.
	s.Require().Contains(out.String(), fmt.Sprintf("\"moniker\":\"%s\"", val0.Moniker))

	// Make sure the output has the binary name.
	s.Require().Contains(out.String(), fmt.Sprintf("\"binary_version\":\"%s\"", version.NewInfo().Version))
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
