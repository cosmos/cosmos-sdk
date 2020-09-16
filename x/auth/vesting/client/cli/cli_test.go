// +build norace

package cli_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
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

func (s *IntegrationTestSuite) TestNewMsgCreateVestingAccountCmd() {
	val := s.network.Validators[0]

	testCases := map[string]struct {
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		"create a continuous vesting account": {
			args: []string{
				sdk.AccAddress("addr2_______________").String(),
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"4070908800",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		"create a delayed vesting account": {
			args: []string{
				sdk.AccAddress("addr3_______________").String(),
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"4070908800",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=true", cli.FlagDelayed),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		"invalid address": {
			args: []string{
				sdk.AccAddress("addr4").String(),
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"4070908800",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		"invalid coins": {
			args: []string{
				sdk.AccAddress("addr4_______________").String(),
				"fooo",
				"4070908800",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		"invalid end time": {
			args: []string{
				sdk.AccAddress("addr4_______________").String(),
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"-4070908800",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
	}

	for name, tc := range testCases {
		tc := tc

		s.Run(name, func() {
			clientCtx := val.ClientCtx

			bw, err := clitestutil.ExecTestCLICmd(clientCtx, cli.NewMsgCreateVestingAccountCmd(), tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(bw.Bytes(), tc.respType), bw.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
