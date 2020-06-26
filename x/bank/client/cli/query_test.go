package cli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	cfg.NumValidators = 1

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

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  fmt.Stringer
		expected  fmt.Stringer
	}{
		{"no address provided", nil, true, nil, nil},
		{
			"total account balance",
			[]string{val.Address.String()},
			false,
			&sdk.Coins{},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
				sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
			),
		},
		{
			"total account balance of a specific denom",
			[]string{val.Address.String(), fmt.Sprintf("--%s=%s", cli.FlagDenom, s.cfg.BondDenom)},
			false,
			&sdk.Coin{},
			sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
		},
		{
			"total account balance of a bogus denom",
			[]string{val.Address.String(), fmt.Sprintf("--%s=foobar", cli.FlagDenom)},
			false,
			&sdk.Coin{},
			sdk.NewCoin("foobar", sdk.ZeroInt()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(buf.Bytes(), tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
