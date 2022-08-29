package cli_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/client/cli"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	oneYear  = 365 * 24 * 60 * 60
	tenHours = 10 * 60 * 60
	oneHour  = 60 * 60
)

type CLITestSuite struct {
	suite.Suite

	addedGranter sdk.AccAddress
	addedGrantee sdk.AccAddress
	addedGrant   feegrant.Grant

	kr      keyring.Keyring
	baseCtx client.Context
	encCfg  testutilmod.TestEncodingConfig

	accounts []sdk.AccAddress
}

var _ client.TendermintRPC = (*mockTendermintRPC)(nil)

type mockTendermintRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func (m mockTendermintRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string, _ bytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}

func (m mockTendermintRPC) BroadcastTxSync(context.Context, tmtypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return &coretypes.ResultBroadcastTx{Code: 0}, nil
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.encCfg = testutilmod.MakeTestEncodingConfig(module.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(mockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	if testing.Short() {
		s.T().Skip("skipping test in unit-tests mode.")
	}

	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 2)

	granter := accounts[0].Address
	grantee := accounts[1].Address

	s.createGrant(granter, grantee)

	grant, err := feegrant.NewGrant(granter, grantee, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
	})
	s.Require().NoError(err)

	s.addedGrant = grant
	s.addedGranter = granter
	s.addedGrantee = grantee
	for _, v := range accounts {
		s.accounts = append(s.accounts, v.Address)
	}
	s.accounts[1] = accounts[1].Address
}

// createGrant creates a new basic allowance fee grant from granter to grantee.
func (s *CLITestSuite) createGrant(granter, grantee sdk.Address) {
	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))).String()),
	}

	fee := sdk.NewCoin("stake", sdk.NewInt(100))

	args := append(
		[]string{
			granter.String(),
			grantee.String(),
			fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, fee.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
			fmt.Sprintf("--%s=%s", cli.FlagExpiration, getFormattedExpiration(oneYear)),
		},
		commonFlags...,
	)

	ctx := svrcmd.CreateExecuteContext(context.Background())

	cmd := cli.NewCmdFeeGrant()
	cmd.SetOutput(io.Discard)
	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
	err := cmd.Execute()
	s.Require().NoError(err)
}

// func (s *CLITestSuite) TearDownSuite() {
// 	s.T().Log("tearing down integration test suite")
// 	s.network.Cleanup()
// }

func (s *CLITestSuite) TestCmdGetFeeGrant() {
	granter := s.addedGranter
	grantee := s.addedGrantee

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectErr    bool
		respType     *feegrant.Grant
		resp         *feegrant.Grant
	}{
		{
			"wrong granter",
			[]string{
				"wrong_granter",
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"decoding bech32 failed",
			true, nil, nil,
		},
		{
			"wrong grantee",
			[]string{
				granter.String(),
				"wrong_grantee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"decoding bech32 failed",
			true, nil, nil,
		},
		{
			"non existed grant",
			[]string{
				"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"fee-grant not found",
			true, nil, nil,
		},
		{
			"valid req",
			[]string{
				granter.String(),
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"",
			false,
			&feegrant.Grant{},
			&s.addedGrant,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd := cli.GetCmdQueryFeeGrant()
			cmd.SetOutput(io.Discard)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
			err := cmd.Execute()
			// ctx := svrcmd.CreateExecuteContext(context.Background())
			// out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
				// s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				// s.Require().Equal(tc.respType.Grantee, tc.respType.Grantee)
				// s.Require().Equal(tc.respType.Granter, tc.respType.Granter)
				// grant, err := tc.respType.GetGrant()
				// s.Require().NoError(err)
				// grant1, err1 := tc.resp.GetGrant()
				// s.Require().NoError(err1)
				// s.Require().Equal(
				// 	grant.(*feegrant.BasicAllowance).SpendLimit,
				// 	grant1.(*feegrant.BasicAllowance).SpendLimit,
				// )
			}
		})
	}
}

func (s *CLITestSuite) TestCmdGetFeeGrantsByGrantee() {
	// val := s.addedGranter
	grantee := s.addedGrantee
	// clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		resp         *feegrant.QueryAllowancesResponse
		expectLength int
	}{
		{
			"wrong grantee",
			[]string{
				"wrong_grantee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, nil, 0,
		},
		{
			"non existent grantee",
			[]string{
				"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false, &feegrant.QueryAllowancesResponse{}, 0,
		},
		{
			"valid req",
			[]string{
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false, &feegrant.QueryAllowancesResponse{}, 1,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GetCmdQueryFeeGrantsByGrantee()
			cmd.SetOutput(io.Discard)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
			err := cmd.Execute()

			// out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.resp), out.String())
				// s.Require().Len(tc.resp.Allowances, tc.expectLength)
			}
		})
	}
}

func (s *CLITestSuite) TestCmdGetFeeGrantsByGranter() {
	// val := s.network.Validators[0]
	granter := s.addedGranter
	// clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		resp         *feegrant.QueryAllowancesByGranterResponse
		expectLength int
	}{
		{
			"wrong grantee",
			[]string{
				"wrong_grantee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, nil, 0,
		},
		{
			"non existent grantee",
			[]string{
				"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false, &feegrant.QueryAllowancesByGranterResponse{}, 0,
		},
		{
			"valid req",
			[]string{
				granter.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false, &feegrant.QueryAllowancesByGranterResponse{}, 1,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd := cli.GetCmdQueryFeeGrantsByGranter()
			cmd.SetOutput(io.Discard)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
			err := cmd.Execute()
			// out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.resp), out.String())
				// s.Require().Len(tc.resp.Allowances, tc.expectLength)
			}
		})
	}
}

func (s *CLITestSuite) TestNewCmdFeeGrant() {
	// val := s.network.Validators[0]
	granter := s.accounts[0]
	alreadyExistedGrantee := s.addedGrantee
	// clientCtx := val.ClientCtx

	fromAddr, fromName, _, err := client.GetFromFields(s.baseCtx, s.kr, granter.String())
	s.Require().Equal(fromAddr, granter)
	s.Require().NoError(err)

	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"wrong granter address",
			append(
				[]string{
					"wrong_granter",
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
		{
			"wrong grantee address",
			append(
				[]string{
					granter.String(),
					"wrong_grantee",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
		{
			"wrong granter key name",
			append(
				[]string{
					"invalid_granter",
					"cosmos16dun6ehcc86e03wreqqww89ey569wuj4em572w",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
		{
			"valid basic fee grant",
			append(
				[]string{
					granter.String(),
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant with granter key name",
			append(
				[]string{
					fromName,
					"cosmos16dun6ehcc86e03wreqqww89ey569wuj4em572w",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, fromName),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant with amino",
			append(
				[]string{
					granter.String(),
					"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant without spend limit",
			append(
				[]string{
					granter.String(),
					"cosmos17h5lzptx3ghvsuhk7wx4c4hnl7rsswxjer97em",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant without expiration",
			append(
				[]string{
					granter.String(),
					"cosmos16dlc38dcqt0uralyd8hksxyrny6kaeqfjvjwp5",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant without spend-limit and expiration",
			append(
				[]string{
					granter.String(),
					"cosmos1ku40qup9vwag4wtf8cls9mkszxfthaklxkp3c8",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"try to add existed grant",
			append(
				[]string{
					granter.String(),
					alreadyExistedGrantee.String(),
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 18, &sdk.TxResponse{},
		},
		{
			"invalid number of args(periodic fee grant)",
			append(
				[]string{
					granter.String(),
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%s", cli.FlagExpiration, getFormattedExpiration(tenHours)),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
		{
			"period mentioned and period limit omitted, invalid periodic grant",
			append(
				[]string{
					granter.String(),
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, tenHours),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%s", cli.FlagExpiration, getFormattedExpiration(oneHour)),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
		{
			"period cannot be greater than the actual expiration(periodic fee grant)",
			append(
				[]string{
					granter.String(),
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, tenHours),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%s", cli.FlagExpiration, getFormattedExpiration(oneHour)),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
		{
			"valid periodic fee grant",
			append(
				[]string{
					granter.String(),
					"cosmos1w55kgcf3ltaqdy4ww49nge3klxmrdavrr6frmp",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, oneHour),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%s", cli.FlagExpiration, getFormattedExpiration(tenHours)),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid periodic fee grant without spend-limit",
			append(
				[]string{
					granter.String(),
					"cosmos1vevyks8pthkscvgazc97qyfjt40m6g9xe85ry8",
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, oneHour),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%s", cli.FlagExpiration, getFormattedExpiration(tenHours)),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid periodic fee grant without expiration",
			append(
				[]string{
					granter.String(),
					"cosmos14cm33pvnrv2497tyt8sp9yavhmw83nwej3m0e8",
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, oneHour),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid periodic fee grant without spend-limit and expiration",
			append(
				[]string{
					granter.String(),
					"cosmos12nyk4pcf4arshznkpz882e4l4ts0lt0ap8ce54",
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, oneHour),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"invalid expiration",
			append(
				[]string{
					granter.String(),
					"cosmos1vevyks8pthkscvgazc97qyfjt40m6g9xe85ry8",
					fmt.Sprintf("--%s=%d", cli.FlagPeriod, oneHour),
					fmt.Sprintf("--%s=%s", cli.FlagPeriodLimit, "10stake"),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%s", cli.FlagExpiration, "invalid"),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdFeeGrant()
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetOutput(io.Discard)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
			err := cmd.Execute()

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				// txResp := tc.respType.(*sdk.TxResponse)
				// s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewCmdRevokeFeegrant() {
	// val := s.network.Validators[0]
	granter := s.addedGranter
	grantee := s.addedGrantee
	// clientCtx := val.ClientCtx

	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))).String()),
	}

	// Create new fee grant specifically to test amino.
	aminoGrantee, err := sdk.AccAddressFromBech32("cosmos16ydaqh0fcnh4qt7a3jme4mmztm2qel5axcpw00")
	s.Require().NoError(err)
	s.createGrant(granter, aminoGrantee)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid grantee",
			append(
				[]string{
					"wrong_granter",
					grantee.String(),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
		{
			"invalid grantee",
			append(
				[]string{
					granter.String(),
					"wrong_grantee",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, 0, nil,
		},
		{
			"Non existed grant",
			append(
				[]string{
					granter.String(),
					"cosmos1aeuqja06474dfrj7uqsvukm6rael982kk89mqr",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 38, &sdk.TxResponse{},
		},
		{
			"Valid revoke",
			append(
				[]string{
					granter.String(),
					grantee.String(),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
		{
			"Valid revoke with amino",
			append(
				[]string{
					granter.String(),
					aminoGrantee.String(),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdRevokeFeegrant()
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetOutput(io.Discard)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
			err := cmd.Execute()
			// out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				// txResp := tc.respType.(*sdk.TxResponse)
				// s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestTxWithFeeGrant() {
	// s.T().Skip() // TODO to re-enable in #12274

	// val := s.addedGranter
	// clientCtx := val.ClientCtx
	granter := s.addedGranter

	// creating an account manually (This account won't be exist in state)
	k, _, err := s.baseCtx.Keyring.NewMnemonic("grantee", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	pub, err := k.GetPubKey()
	s.Require().NoError(err)
	grantee := sdk.AccAddress(pub.Address())

	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))).String()),
	}

	fee := sdk.NewCoin("stake", sdk.NewInt(100))

	args := append(
		[]string{
			granter.String(),
			grantee.String(),
			fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, fee.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
			fmt.Sprintf("--%s=%s", cli.FlagExpiration, getFormattedExpiration(oneYear)),
		},
		commonFlags...,
	)

	cmd := cli.NewCmdFeeGrant()
	ctx := svrcmd.CreateExecuteContext(context.Background())

	cmd.SetOutput(io.Discard)
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
	err = cmd.Execute()
	s.Require().NoError(err)

	// _, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	// s.Require().NoError(err)
	// _, err = s.network.WaitForHeight(1)
	// s.Require().NoError(err)

	testcases := []struct {
		name       string
		from       string
		flags      []string
		expErrCode uint32
	}{
		{
			name:  "granted fee allowance for an account which is not in state and creating any tx with it by using --fee-granter shouldn't fail",
			from:  grantee.String(),
			flags: []string{fmt.Sprintf("--%s=%s", flags.FlagFeeGranter, granter.String())},
		},
		{
			name:       "--fee-payer should also sign the tx (direct)",
			from:       grantee.String(),
			flags:      []string{fmt.Sprintf("--%s=%s", flags.FlagFeePayer, granter.String())},
			expErrCode: 4,
		},
		{
			name: "--fee-payer should also sign the tx (amino-json)",
			from: grantee.String(),
			flags: []string{
				fmt.Sprintf("--%s=%s", flags.FlagFeePayer, granter.String()),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
			expErrCode: 4,
		},
		{
			name: "use --fee-payer and --fee-granter together works",
			from: grantee.String(),
			flags: []string{
				fmt.Sprintf("--%s=%s", flags.FlagFeePayer, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFeeGranter, granter.String()),
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			err := s.msgSubmitLegacyProposal(s.baseCtx, tc.from,
				"Text Proposal", "No desc", govv1beta1.ProposalTypeText,
				tc.flags...,
			)
			s.Require().NoError(err)

			// var resp sdk.TxResponse
			// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			// s.Require().Equal(tc.expErrCode, resp.Code, resp)
		})
	}
}

func (s *CLITestSuite) msgSubmitLegacyProposal(clientCtx client.Context, from, title, description, proposalType string, extraArgs ...string) error {

	var commonArgs = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
	}

	args := append([]string{
		fmt.Sprintf("--%s=%s", govcli.FlagTitle, title),
		fmt.Sprintf("--%s=%s", govcli.FlagDescription, description),
		fmt.Sprintf("--%s=%s", govcli.FlagProposalType, proposalType),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, extraArgs...)

	ctx := svrcmd.CreateExecuteContext(context.Background())

	cmd := govcli.NewCmdSubmitLegacyProposal()
	cmd.SetOutput(io.Discard)
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
	err := cmd.Execute()

	return err
}

func (s *CLITestSuite) TestFilteredFeeAllowance() {
	// s.T().Skip() // TODO to re-enable in #12274

	// val := s.network.Validators[0]

	granter := s.addedGranter
	k, _, err := s.baseCtx.Keyring.NewMnemonic("grantee1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	pub, err := k.GetPubKey()
	s.Require().NoError(err)
	grantee := sdk.AccAddress(pub.Address())

	// clientCtx := val.ClientCtx

	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))).String()),
	}
	spendLimit := sdk.NewCoin("stake", sdk.NewInt(1000))

	allowMsgs := strings.Join([]string{sdk.MsgTypeURL(&govv1beta1.MsgSubmitProposal{}), sdk.MsgTypeURL(&govv1.MsgVoteWeighted{})}, ",")

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid granter address",
			append(
				[]string{
					"not an address",
					"cosmos1nph3cfzk6trsmfxkeu943nvach5qw4vwstnvkl",
					fmt.Sprintf("--%s=%s", cli.FlagAllowedMsgs, allowMsgs),
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, spendLimit.String()),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid grantee address",
			append(
				[]string{
					granter.String(),
					"not an address",
					fmt.Sprintf("--%s=%s", cli.FlagAllowedMsgs, allowMsgs),
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, spendLimit.String()),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			true, &sdk.TxResponse{}, 0,
		},
		{
			"valid filter fee grant",
			append(
				[]string{
					granter.String(),
					grantee.String(),
					fmt.Sprintf("--%s=%s", cli.FlagAllowedMsgs, allowMsgs),
					fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, spendLimit.String()),
					fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
				},
				commonFlags...,
			),
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCmdFeeGrant()
			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd.SetOutput(io.Discard)
			cmd.SetArgs(tc.args)
			cmd.SetContext(ctx)
			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
			err := cmd.Execute()
			// out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				// txResp := tc.respType.(*sdk.TxResponse)
				// s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}

	args := []string{
		granter.String(),
		grantee.String(),
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	// get filtered fee allowance and check info
	cmd := cli.GetCmdQueryFeeGrant()
	ctx := svrcmd.CreateExecuteContext(context.Background())
	cmd.SetOutput(io.Discard)
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
	err = cmd.Execute()
	// out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	// resp := &feegrant.Grant{}

	// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp), out.String())
	// s.Require().Equal(resp.Grantee, resp.Grantee)
	// s.Require().Equal(resp.Granter, resp.Granter)

	// grant := s.addedGrant
	// s.Require().NoError(err)

	// filteredFeeGrant, err := grant.(*feegrant.AllowedMsgAllowance).GetAllowance()
	// s.Require().NoError(err)

	// s.Require().Equal(
	// filteredFeeGrant.(*feegrant.BasicAllowance).SpendLimit.String(),
	// spendLimit.String(),
	// )

	// exec filtered fee allowance
	cases := []struct {
		name         string
		malleate     func() error
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"valid proposal tx",
			func() error {
				return s.msgSubmitLegacyProposal(s.baseCtx, grantee.String(),
					"Text Proposal", "No desc", govv1beta1.ProposalTypeText,
					fmt.Sprintf("--%s=%s", flags.FlagFeeGranter, granter.String()),
					fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))).String()),
				)
			},
			&sdk.TxResponse{},
			0,
		},
		{
			"valid weighted_vote tx",
			func() error {
				return s.msgVote(s.baseCtx, grantee.String(), "0", "yes",
					fmt.Sprintf("--%s=%s", flags.FlagFeeGranter, granter.String()),
					fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))).String()),
				)
			},
			&sdk.TxResponse{},
			2,
		},
		{
			"should fail with unauthorized msgs",
			func() error {
				args := append(
					[]string{
						grantee.String(),
						"cosmos14cm33pvnrv2497tyt8sp9yavhmw83nwej3m0e8",
						fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, "100stake"),
						fmt.Sprintf("--%s=%s", flags.FlagFeeGranter, granter),
					},
					commonFlags...,
				)

				cmd := cli.NewCmdFeeGrant()
				ctx := svrcmd.CreateExecuteContext(context.Background())
				cmd.SetOutput(io.Discard)
				cmd.SetArgs(args)
				cmd.SetContext(ctx)
				s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
				return cmd.Execute()
			},
			&sdk.TxResponse{},
			7,
		},
	}

	for _, tc := range cases {
		tc := tc

		s.Run(tc.name, func() {
			err := tc.malleate()
			s.Require().NoError(err)
			// s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			// txResp := tc.respType.(*sdk.TxResponse)
			// s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
		})
	}
}

// msgVote votes for a proposal
func (s *CLITestSuite) msgVote(clientCtx client.Context, from, id, vote string, extraArgs ...string) error {
	var commonArgs = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
	}
	args := append([]string{
		id,
		vote,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, extraArgs...)

	ctx := svrcmd.CreateExecuteContext(context.Background())

	cmd := govcli.NewCmdWeightedVote()
	cmd.SetOutput(io.Discard)
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
	return cmd.Execute()
}

func getFormattedExpiration(duration int64) string {
	return time.Now().Add(time.Duration(duration) * time.Second).Format(time.RFC3339)
}
