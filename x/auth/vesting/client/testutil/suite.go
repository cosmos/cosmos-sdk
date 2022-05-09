package testutil

import (
	"fmt"

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

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.network = network.New(s.T(), s.cfg)

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
		expectedCode uint32
		respType     proto.Message
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
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
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
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		"invalid address": {
			args: []string{
				sdk.AccAddress("addr4").String(),
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"4070908800",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			},
			expectErr:    true,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		"invalid coins": {
			args: []string{
				sdk.AccAddress("addr4_______________").String(),
				"fooo",
				"4070908800",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			},
			expectErr:    true,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		"invalid end time": {
			args: []string{
				sdk.AccAddress("addr4_______________").String(),
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"-4070908800",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			},
			expectErr:    true,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(bw.Bytes(), tc.respType), bw.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewMsgCreateClawbackVestingAccountCmd() {
	val := s.network.Validators[0]
	for _, tc := range []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			name: "basic",
			args: []string{
				sdk.AccAddress("addr10______________").String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=%s", cli.FlagLockup, "testdata/periods1.json"),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/periods1.json"),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		{
			name: "defaultLockup",
			args: []string{
				sdk.AccAddress("addr11______________").String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/periods1.json"),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		{
			name: "defaultVesting",
			args: []string{
				sdk.AccAddress("addr12______________").String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=%s", cli.FlagLockup, "testdata/periods1.json"),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		{
			name: "merge",
			args: []string{
				sdk.AccAddress("addr10______________").String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=%s", cli.FlagLockup, "testdata/periods1.json"),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/periods1.json"),
				fmt.Sprintf("--%s=%s", cli.FlagMerge, "true"),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		{
			name: "bad vesting addr",
			args: []string{
				"foo",
			},
			expectErr: true,
		},
		{
			name: "no files",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
			},
			expectErr: true,
		},
		{
			name: "bad lockup filename",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
				fmt.Sprintf("--%s=%s", cli.FlagLockup, "testdata/noexist"),
			},
			expectErr: true,
		},
		{
			name: "bad lockup json",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
				fmt.Sprintf("--%s=%s", cli.FlagLockup, "testdata/badjson"),
			},
			expectErr: true,
		},
		{
			name: "bad lockup periods",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
				fmt.Sprintf("--%s=%s", cli.FlagLockup, "testdata/badperiod.json"),
			},
			expectErr: true,
		},
		{
			name: "bad vesting filename",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/noexist"),
			},
			expectErr: true,
		},
		{
			name: "bad vesting json",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/badjson"),
			},
			expectErr: true,
		},
		{
			name: "bad vesting periods",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/badperiod.json"),
			},
			expectErr: true,
		},
	} {
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			bw, err := clitestutil.ExecTestCLICmd(clientCtx, cli.NewMsgCreateClawbackVestingAccountCmd(), tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(bw.Bytes(), tc.respType), bw.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewMsgClawbackCmd() {
	val := s.network.Validators[0]
	addr := sdk.AccAddress("addr30______________")

	_, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewMsgCreateClawbackVestingAccountCmd(), []string{
		addr.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
		fmt.Sprintf("--%s=%s", cli.FlagLockup, "testdata/periods1.json"),
		fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/periods1.json"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	})
	s.Require().NoError(err)

	for _, tc := range []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			name: "basic",
			args: []string{
				addr.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=%s", cli.FlagDest, sdk.AccAddress("addr32______________").String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		{
			name: "bad vesting addr",
			args: []string{
				"foo",
			},
			expectErr: true,
		},
		{
			name: "bad dest addr",
			args: []string{
				addr.String(),
				fmt.Sprintf("--%s=%s", cli.FlagDest, "bar"),
			},
			expectErr: true,
		},
		{
			name: "default dest",
			args: []string{
				addr.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
	} {
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			bw, err := clitestutil.ExecTestCLICmd(clientCtx, cli.NewMsgClawbackCmd(), tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(bw.Bytes(), tc.respType), bw.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}
