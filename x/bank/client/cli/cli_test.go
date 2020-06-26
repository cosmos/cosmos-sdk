package cli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutil.Config
	network *testutil.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultConfig()
	cfg.NumValidators = 2

	s.cfg = cfg
	s.network = testutil.NewTestNetwork(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetBalancesCmd() {
	buf := new(bytes.Buffer)
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx.WithOutput(buf)

	cmd := cli.GetBalancesCmd(clientCtx)
	cmd.SetErr(buf)
	cmd.SetOut(buf)
	s.Require().Error(cmd.Execute())

	buf.Reset()

	cmd.SetArgs([]string{val.Address.String()})
	s.Require().NoError(cmd.Execute())
	fmt.Println(string(buf.Bytes()))
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
