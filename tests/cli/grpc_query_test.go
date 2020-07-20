// +build cli_test

package cli_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func TestCliQueryConn(t *testing.T) {
	// TODO use in process tests for this
	t.Parallel()
	f := cli.NewFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	ctx := client.NewContext()
	testClient := testdata.NewTestServiceClient(ctx)
	res, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", res.Message)
}

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.network = network.New(s.T(), network.DefaultConfig())
	s.Require().NotNil(s.network)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestGRPC() {
	val := s.network.Validators[0]
	conn, err := grpc.Dial(
		val.AppConfig.GRPC.Address,
		grpc.WithInsecure(), // Or else we get "no transport security set"
	)
	s.Require().NoError(err)
	testClient := testdata.NewTestServiceClient(conn)
	res, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("a", "b")
	s.Require().Equal("hello", res.Message)
}

func TestCLIIntegrationTestSuite(t *testing.T) {
	suite.Run(t, &IntegrationTestSuite{})
}
