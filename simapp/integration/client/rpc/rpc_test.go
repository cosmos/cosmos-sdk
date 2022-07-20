//go:build norace
// +build norace

package rpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/simapp"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

type IntegrationTestSuite struct {
	suite.Suite

	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), network.DefaultConfig(simapp.NewTestNetworkFixture))
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
}

func (s *IntegrationTestSuite) TestCLIQueryConn() {
	var header metadata.MD

	testClient := testdata.NewQueryClient(s.network.Validators[0].ClientCtx)
	res, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"}, grpc.Header(&header))
	s.NoError(err)

	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	s.Require().Equal([]string{"1"}, blockHeight)

	s.Equal("hello", res.Message)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
