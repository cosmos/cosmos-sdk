package distribution

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

// SetupSuite creates a new network for _each_ e2e test. We create a new
// network for each test because there are some state modifications that are
// needed to be made in order to make useful queries. However, we don't want
// these state changes to be present in other tests.
func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	s.cfg = cfg

	genesisState := s.cfg.GenesisState
	var mintData minttypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	inflation := math.LegacyMustNewDecFromStr("1.0")
	mintData.Minter.Inflation = inflation
	mintData.Params.InflationMin = inflation
	mintData.Params.InflationMax = inflation

	mintDataBz, err := s.cfg.Codec.MarshalJSON(&mintData)
	s.Require().NoError(err)
	genesisState[minttypes.ModuleName] = mintDataBz
	s.cfg.GenesisState = genesisState

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

// TearDownSuite cleans up the curret test network after _each_ test.
func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite1")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestNewWithdrawRewardsCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name                 string
		valAddr              fmt.Stringer
		args                 []string
		expectErr            bool
		expectedCode         uint32
		respType             proto.Message
		expectedResponseType []string
	}{
		{
			"invalid validator address",
			val.Address,
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			true, 0, nil,
			[]string{},
		},
		{
			"valid transaction",
			sdk.ValAddress(val.Address),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
			[]string{
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse",
			},
		},
		{
			"valid transaction (with commission)",
			sdk.ValAddress(val.Address),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=true", cli.FlagCommission),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
			[]string{
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse",
				"/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			args := append([]string{tc.valAddr.String()}, tc.args...)

			_, _ = s.network.WaitForHeightWithTimeout(10, time.Minute)

			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd := cli.NewWithdrawRewardsCmd(address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
			cmd.SetContext(ctx)
			cmd.SetArgs(args)
			s.Require().NoError(client.SetCmdClientContextHandler(clientCtx, cmd))

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().NoError(s.network.WaitForNextBlock())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedCode, txResp.Code)

				data, err := hex.DecodeString(txResp.Data)
				s.Require().NoError(err)

				txMsgData := sdk.TxMsgData{}
				err = s.cfg.Codec.Unmarshal(data, &txMsgData)
				s.Require().NoError(err)
				for responseIdx, msgResponse := range txMsgData.MsgResponses {
					s.Require().Equal(tc.expectedResponseType[responseIdx], msgResponse.TypeUrl)
					switch msgResponse.TypeUrl {
					case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse":
						var resp distrtypes.MsgWithdrawDelegatorRewardResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse":
						var resp distrtypes.MsgWithdrawValidatorCommissionResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					}
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestNewWithdrawAllRewardsCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name                 string
		args                 []string
		expectErr            bool
		expectedCode         uint32
		respType             proto.Message
		expectedResponseType []string
	}{
		{
			"valid transaction (offline)",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			true, 0, nil,
			[]string{},
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
			[]string{
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewWithdrawAllRewardsCmd(address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			_, _ = s.network.WaitForHeightWithTimeout(10, time.Minute)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().NoError(s.network.WaitForNextBlock())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedCode, txResp.Code)

				data, err := hex.DecodeString(txResp.Data)
				s.Require().NoError(err)

				txMsgData := sdk.TxMsgData{}
				err = s.cfg.Codec.Unmarshal(data, &txMsgData)
				s.Require().NoError(err)
				for responseIdx, msgResponse := range txMsgData.MsgResponses {
					s.Require().Equal(tc.expectedResponseType[responseIdx], msgResponse.TypeUrl)
					switch msgResponse.TypeUrl {
					case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse":
						var resp distrtypes.MsgWithdrawDelegatorRewardResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse":
						var resp distrtypes.MsgWithdrawValidatorCommissionResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					}
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestNewSetWithdrawAddrCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid withdraw address",
			[]string{
				"foo",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			true, 0, nil,
		},
		{
			"valid transaction",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSetWithdrawAddrCmd(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestNewFundCommunityPoolCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid funding amount",
			[]string{
				"-43foocoin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			true, 0, nil,
		},
		{
			"valid transaction",
			[]string{
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(5431))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewFundCommunityPoolCmd(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}
