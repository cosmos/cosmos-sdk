package cli_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/x/params/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	lock    sync.RWMutex
	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite() *IntegrationTestSuite {
	return &IntegrationTestSuite{
		lock: sync.RWMutex{},
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.lock.Lock()
	defer s.lock.Unlock()

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
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, NewIntegrationTestSuite())
}
