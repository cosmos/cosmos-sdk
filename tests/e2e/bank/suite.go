package client

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	"cosmossdk.io/x/bank/client/cli"
	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	ac      address.Codec
	network network.NetworkI
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
	s.ac = addresscodec.NewBech32Codec("cosmos")
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestNewSendTxCmdGenOnly() {
	val := s.network.GetValidators()[0]

	from := val.GetAddress()
	to := val.GetAddress()
	amount := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
	)
	fromStr, err := s.ac.BytesToString(from)
	s.Require().NoError(err)
	toStr, err := s.ac.BytesToString(to)
	s.Require().NoError(err)
	msgSend := &types.MsgSend{
		FromAddress: fromStr,
		ToAddress:   toStr,
		Amount:      amount,
	}

	bz, err := clitestutil.SubmitTestTx(
		val.GetClientCtx(),
		msgSend,
		from,
		clitestutil.TestTxConfig{
			GenOnly: true,
		},
	)
	s.Require().NoError(err)

	tx, err := s.cfg.TxConfig.TxJSONDecoder()(bz.Bytes())
	s.Require().NoError(err)
	s.Require().Equal([]sdk.Msg{types.NewMsgSend(fromStr, toStr, amount)}, tx.GetMsgs())
}

func (s *E2ETestSuite) TestNewSendTxCmdDryRun() {
	val := s.network.GetValidators()[0]

	from := val.GetAddress()
	to := val.GetAddress()
	amount := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
	)

	msgSend := &types.MsgSend{
		FromAddress: from.String(),
		ToAddress:   to.String(),
		Amount:      amount,
	}

	out, err := clitestutil.SubmitTestTx(
		val.GetClientCtx(),
		msgSend,
		from,
		clitestutil.TestTxConfig{
			Simulate: true,
		},
	)
	s.Require().NoError(err)
	s.Require().Regexp("\"gas_info\"", out.String())
	s.Require().Regexp("\"gas_used\":\"[0-9]+\"", out.String())
}

func (s *E2ETestSuite) TestNewSendTxCmd() {
	val := s.network.GetValidators()[0]

	testCases := []struct {
		name         string
		from, to     sdk.AccAddress
		amount       sdk.Coins
		config       clitestutil.TestTxConfig
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"valid transaction",
			val.GetAddress(),
			val.GetAddress(),
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			clitestutil.TestTxConfig{
				Fee: sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"not enough fees",
			val.GetAddress(),
			val.GetAddress(),
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			clitestutil.TestTxConfig{
				Fee: sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1))),
			},

			false,
			sdkerrors.ErrInsufficientFee.ABCICode(),
			&sdk.TxResponse{},
		},
		{
			"not enough gas",
			val.GetAddress(),
			val.GetAddress(),
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			clitestutil.TestTxConfig{
				Gas: 10,
				Fee: sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))),
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
			clientCtx := val.GetClientCtx()

			msgSend := types.MsgSend{
				FromAddress: tc.from.String(),
				ToAddress:   tc.to.String(),
				Amount:      tc.amount,
			}
			bz, err := clitestutil.SubmitTestTx(val.GetClientCtx(), &msgSend, tc.from, tc.config)
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
	val := s.network.GetValidators()[0]
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
			val.GetAddress(),
			[]sdk.AccAddress{val.GetAddress(), testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid split transaction",
			val.GetAddress(),
			[]sdk.AccAddress{val.GetAddress(), testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", cli.FlagSplit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"not enough arguments",
			val.GetAddress(),
			[]sdk.AccAddress{val.GetAddress()},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"chain-id shouldn't be used with offline and generate-only flags",
			val.GetAddress(),
			[]sdk.AccAddress{val.GetAddress(), testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"not enough fees",
			val.GetAddress(),
			[]sdk.AccAddress{val.GetAddress(), testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1))).String()),
			},
			false,
			sdkerrors.ErrInsufficientFee.ABCICode(),
			&sdk.TxResponse{},
		},
		{
			"not enough gas",
			val.GetAddress(),
			[]sdk.AccAddress{val.GetAddress(), testAddr},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), math.NewInt(10)),
				sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
			),
			[]string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
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
			clientCtx := val.GetClientCtx()

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
