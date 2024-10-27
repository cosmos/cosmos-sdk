package client

import (
	"fmt"
	"io"
	"os"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	genesisState := s.cfg.GenesisState
	var bankGenesis types.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[types.ModuleName], &bankGenesis))

	bankGenesis.DenomMetadata = []types.Metadata{
		{
			Name:        "Cosmos Hub Atom",
			Symbol:      "ATOM",
			Description: "The native staking token of the Cosmos Hub.",
			DenomUnits: []*types.DenomUnit{
				{
					Denom:    "uatom",
					Exponent: 0,
					Aliases:  []string{"microatom"},
				},
				{
					Denom:    "atom",
					Exponent: 6,
					Aliases:  []string{"ATOM"},
				},
			},
			Base:    "uatom",
			Display: "atom",
		},
		{
			Name:        "Ethereum",
			Symbol:      "ETH",
			Description: "Ethereum mainnet token",
			DenomUnits: []*types.DenomUnit{
				{
					Denom:    "wei",
					Exponent: 0,
				},
				{
					Denom:    "eth",
					Exponent: 6,
					Aliases:  []string{"ETH"},
				},
			},
			Base:    "wei",
			Display: "eth",
		},
	}

	bankGenesisBz, err := s.cfg.Codec.MarshalJSON(&bankGenesis)
	s.Require().NoError(err)
	genesisState[types.ModuleName] = bankGenesisBz
	s.cfg.GenesisState = genesisState

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestGetBalancesCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
		expected  proto.Message
	}{
		{"no address provided", []string{}, true, nil, nil},
		{
			"total account balance",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&types.QueryAllBalancesResponse{},
			&types.QueryAllBalancesResponse{
				Balances: sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
					sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
				),
				Pagination: &query.PageResponse{},
			},
		},
		{
			"total account balance of a specific denom",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=%s", cli.FlagDenom, s.cfg.BondDenom),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&sdk.Coin{},
			NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
		},
		{
			"total account balance of a bogus denom",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=foobar", cli.FlagDenom),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			&sdk.Coin{},
			NewCoin("foobar", math.ZeroInt()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetBalancesCmd()
			out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryTotalSupply() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
		expected  proto.Message
	}{
		{
			name: "total supply",
			args: []string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			respType: &types.QueryTotalSupplyResponse{},
			expected: &types.QueryTotalSupplyResponse{
				Supply: sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
					sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
				),
				Pagination: &query.PageResponse{Total: 0},
			},
		},
		{
			name: "total supply of a specific denomination",
			args: []string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=%s", cli.FlagDenom, s.cfg.BondDenom),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			respType: &sdk.Coin{},
			expected: &sdk.Coin{
				Denom:  s.cfg.BondDenom,
				Amount: s.cfg.StakingTokens.Add(sdk.NewInt(10)),
			},
		},
		{
			name: "total supply of a bogus denom",
			args: []string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=foobar", cli.FlagDenom),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			respType: &sdk.Coin{},
			expected: &sdk.Coin{
				Denom:  "foobar",
				Amount: math.ZeroInt(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryTotalSupply()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType))
				s.Require().Equal(tc.expected, tc.respType)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryDenomsMetadata() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
		expected  proto.Message
	}{
		{
			name: "all denoms client metadata",
			args: []string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			respType: &types.QueryDenomsMetadataResponse{},
			expected: &types.QueryDenomsMetadataResponse{
				Metadatas: []types.Metadata{
					{
						Name:        "Cosmos Hub Atom",
						Symbol:      "ATOM",
						Description: "The native staking token of the Cosmos Hub.",
						DenomUnits: []*types.DenomUnit{
							{
								Denom:    "uatom",
								Exponent: 0,
								Aliases:  []string{"microatom"},
							},
							{
								Denom:    "atom",
								Exponent: 6,
								Aliases:  []string{"ATOM"},
							},
						},
						Base:    "uatom",
						Display: "atom",
					},
					{
						Name:        "Ethereum",
						Symbol:      "ETH",
						Description: "Ethereum mainnet token",
						DenomUnits: []*types.DenomUnit{
							{
								Denom:    "wei",
								Exponent: 0,
								Aliases:  []string{},
							},
							{
								Denom:    "eth",
								Exponent: 6,
								Aliases:  []string{"ETH"},
							},
						},
						Base:    "wei",
						Display: "eth",
					},
				},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			name: "client metadata of a specific denomination",
			args: []string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=%s", cli.FlagDenom, "uatom"),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			respType: &types.QueryDenomMetadataResponse{},
			expected: &types.QueryDenomMetadataResponse{
				Metadata: types.Metadata{
					Name:        "Cosmos Hub Atom",
					Symbol:      "ATOM",
					Description: "The native staking token of the Cosmos Hub.",
					DenomUnits: []*types.DenomUnit{
						{
							Denom:    "uatom",
							Exponent: 0,
							Aliases:  []string{"microatom"},
						},
						{
							Denom:    "atom",
							Exponent: 6,
							Aliases:  []string{"ATOM"},
						},
					},
					Base:    "uatom",
					Display: "atom",
				},
			},
		},
		{
			name: "client metadata of a bogus denom",
			args: []string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=foobar", cli.FlagDenom),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			expectErr: true,
			respType:  &types.QueryDenomMetadataResponse{},
			expected: &types.QueryDenomMetadataResponse{
				Metadata: types.Metadata{
					DenomUnits: []*types.DenomUnit{},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdDenomsMetadata()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType))
				s.Require().Equal(tc.expected, tc.respType)
			}
		})
	}
}

func (s *E2ETestSuite) TestNewSendTxCmdGenOnly() {
	val := s.network.Validators[0]

	clientCtx := val.ClientCtx

	from := val.Address
	to := val.Address
	amount := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
	)
	args := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	}

	bz, err := clitestutil.MsgSendExec(clientCtx, from, to, amount, args...)
	s.Require().NoError(err)
	tx, err := s.cfg.TxConfig.TxJSONDecoder()(bz.Bytes())
	s.Require().NoError(err)
	s.Require().Equal([]sdk.Msg{types.NewMsgSend(from, to, amount)}, tx.GetMsgs())
}

func (s *E2ETestSuite) TestNewSendTxCmdDryRun() {
	val := s.network.Validators[0]

	clientCtx := val.ClientCtx

	from := val.Address
	to := val.Address
	amount := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
	)
	args := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagDryRun),
	}

	oldSterr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	_, err := clitestutil.MsgSendExec(clientCtx, from, to, amount, args...)
	s.Require().NoError(err)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = oldSterr

	s.Require().Regexp("gas estimate: [0-9]+", string(out))
}

func (s *E2ETestSuite) TestNewSendTxCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		from, to     sdk.AccAddress
		amount       sdk.Coins
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"valid transaction",
			val.Address,
			val.Address,
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"chain-id shouldn't be used with offline and generate-only flags",
			val.Address,
			val.Address,
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"not enough fees",
			val.Address,
			val.Address,
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(1))).String()),
			},
			false,
			sdkerrors.ErrInsufficientFee.ABCICode(),
			&sdk.TxResponse{},
		},
		{
			"not enough gas",
			val.Address,
			val.Address,
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				"--gas=10",
			},
			false,
			sdkerrors.ErrOutOfGas.ABCICode(),
			&sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Require().NoError(s.network.WaitForNextBlock())
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			bz, err := clitestutil.MsgSendExec(clientCtx, tc.from, tc.to, tc.amount, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(bz.Bytes(), tc.respType), bz.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *E2ETestSuite) TestNewMultiSendTxCmd() {
	val := s.network.Validators[0]
	testAddr := sdk.AccAddress("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")

	testCases := []struct {
		name         string
		from         sdk.AccAddress
		to           []sdk.AccAddress
		amount       sdk.Coins
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"valid transaction",
			val.Address,
			[]sdk.AccAddress{val.Address, testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid split transaction",
			val.Address,
			[]sdk.AccAddress{val.Address, testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", cli.FlagSplit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"not enough arguments",
			val.Address,
			[]sdk.AccAddress{val.Address},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"chain-id shouldn't be used with offline and generate-only flags",
			val.Address,
			[]sdk.AccAddress{val.Address, testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"not enough fees",
			val.Address,
			[]sdk.AccAddress{val.Address, testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(1))).String()),
			},
			false,
			sdkerrors.ErrInsufficientFee.ABCICode(),
			&sdk.TxResponse{},
		},
		{
			"not enough gas",
			val.Address,
			[]sdk.AccAddress{val.Address, testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				"--gas=10",
			},
			false,
			sdkerrors.ErrOutOfGas.ABCICode(),
			&sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Require().NoError(s.network.WaitForNextBlock())
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			bz, err := MsgMultiSendExec(clientCtx, tc.from, tc.to, tc.amount, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(bz.Bytes(), tc.respType), bz.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func NewCoin(denom string, amount math.Int) *sdk.Coin {
	coin := sdk.NewCoin(denom, amount)
	return &coin
}

func MsgMultiSendExec(clientCtx client.Context, from sdk.AccAddress, to []sdk.AccAddress, amount fmt.Stringer, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{from.String()}
	for _, addr := range to {
		args = append(args, addr.String())
	}

	args = append(args, amount.String())
	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewMultiSendTxCmd(), args)
}
