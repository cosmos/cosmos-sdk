package cli_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/x/params/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNewQuerySubspaceParamsCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{
				"staking", "MaxValidators",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			`{"subspace":"staking","key":"MaxValidators","value":"100"}`,
		},
		{
			"text output",
			[]string{
				"staking", "MaxValidators",
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
			},
			`key: MaxValidators
subspace: staking
value: "100"`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewQuerySubspaceParamsCmd()
			_, out := testutil.ApplyMockIO(cmd)

			clientCtx := val.ClientCtx.WithOutput(out)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

			out.Reset()
			cmd.SetArgs(tc.args)

			s.Require().NoError(cmd.ExecuteContext(ctx))
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
