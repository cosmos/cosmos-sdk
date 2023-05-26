package circuit

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	"cosmossdk.io/x/circuit/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

type GRPCQueryTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *GRPCQueryTestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	s.cfg = cfg

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *GRPCQueryTestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite1")
	s.network.Cleanup()
}

func (s GRPCQueryTestSuite) TestQueryAccount() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		args     []string
		respType proto.Message
		expected proto.Message
	}{
		{
			name:     "disable-list",
			url:      baseURL + "/cosmos/circuit/v1/disable_list",
			args:     []string{},
			respType: &types.DisabledListResponse{},
			expected: &types.DisabledListResponse{[]string{}},
		},
		{
			name:     "account",
			url:      baseURL + "/cosmos/circuit/v1/accounts/%s",
			args:     []string{"cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9"},
			respType: &types.AccountResponse{},
			expected: &types.AccountResponse{
				Permission: &types.Permissions{
					Level:         types.Permissions_Level(1),
					LimitTypeUrls: []string{},
				},
			},
		},
		{
			name:     "accounts",
			url:      baseURL + "/cosmos/circuit/v1/accounts",
			args:     []string{},
			respType: &types.AccountsResponse{},
			expected: &types.AccountsResponse{
				Accounts: []*types.GenesisAccountPermissions{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			url := tc.url
			if args := tc.args; len(args) > 0 {
				url = fmt.Sprintf(url, tc.args[0])
			}
			resp, err := sdktestutil.GetRequest(url)
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
		})
	}

}
