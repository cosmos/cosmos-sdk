package testutil

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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

func (s *IntegrationTestSuite) TestNewMsgCreatePeriodicVestingAccountCmd() {
	val := s.network.Validators[0]
	testCases := map[string]struct {
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		"create a periodic vesting account": {
			args: []string{
				sdk.AccAddress("addr5_______________").String(),
				"testdata/periods1.json",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
		"bad to address": {
			args: []string{
				"foo",
				"testdata/periods1.json",
			},
			expectErr: true,
		},
		"bad from address": {
			args: []string{
				sdk.AccAddress("addr5_______________").String(),
				"testdata/periods1.json",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "foo"),
			},
			expectErr: true,
		},
		"bad file": {
			args: []string{
				sdk.AccAddress("addr6_______________").String(),
				"testdata/noexist",
			},
			expectErr: true,
		},
		"bad json": {
			args: []string{
				sdk.AccAddress("addr7_______________").String(),
				"testdata/badjson",
			},
			expectErr: true,
		},
		"bad periods length": {
			args: []string{
				sdk.AccAddress("addr8_______________").String(),
				"testdata/badperiod.json",
			},
			expectErr: true,
		},
		"bad periods amount": {
			args: []string{
				sdk.AccAddress("addr9_______________").String(),
				"testdata/badperiod2.json",
			},
			expectErr: true,
		},
		"merge": {
			args: []string{
				sdk.AccAddress("addr9_______________").String(),
				"testdata/periods1.json",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				"--merge",
			},
			expectErr:    false,
			expectedCode: 0,
			respType:     &sdk.TxResponse{},
		},
	}

	for name, tc := range testCases {
		tc := tc

		s.Run(name, func() {
			clientCtx := val.ClientCtx

			bw, err := clitestutil.ExecTestCLICmd(clientCtx, cli.NewMsgCreatePeriodicVestingAccountCmd(), tc.args)
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
			name: "bad vesting periods length",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/badperiod.json"),
			},
			expectErr: true,
		},
		{
			name: "bad vesting periods amount",
			args: []string{
				sdk.AccAddress("addr13______________").String(),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, "testdata/badperiod2.json"),
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

func (s *IntegrationTestSuite) TestNewMsgReturnGrantsCmd() {
	val := s.network.Validators[0]

	consPrivKey := ed25519.GenPrivKey()
	consPubKeyBz, err := s.cfg.Codec.MarshalInterfaceJSON(consPrivKey.PubKey())
	s.Require().NoError(err)
	s.Require().NotNil(consPubKeyBz)

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewClawback", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	addr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewMsgCreateClawbackVestingAccountCmd(), []string{
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr),
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

			bw, err := clitestutil.ExecTestCLICmd(clientCtx, cli.NewMsgReturnGrantsCmd(), tc.args)
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
