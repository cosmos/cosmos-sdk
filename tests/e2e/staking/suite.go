package testutil

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
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

	if testing.Short() {
		s.T().Skip("skipping test in unit-tests mode.")
	}

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	unbond, err := sdk.ParseCoinNormalized("10stake")
	s.Require().NoError(err)

	val := s.network.Validators[0]
	val2 := s.network.Validators[1]

	// redelegate
	out, err := MsgRedelegateExec(
		val.ClientCtx,
		val.Address,
		val.ValAddress,
		val2.ValAddress,
		unbond,
		fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().Equal(uint32(0), txRes.Code)
	s.Require().NoError(s.network.WaitForNextBlock())

	unbondingAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(5))

	// unbonding the amount
	out, err = MsgUnbondExec(val.ClientCtx, val.Address, val.ValAddress, unbondingAmount)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().Equal(uint32(0), txRes.Code)
	s.Require().NoError(s.network.WaitForNextBlock())

	// unbonding the amount
	out, err = MsgUnbondExec(val.ClientCtx, val.Address, val.ValAddress, unbondingAmount)
	s.Require().NoError(err)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().Equal(uint32(0), txRes.Code)
	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestGetCmdQueryValidator() {
	val := s.network.Validators[0]
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid address ",
			[]string{"somethinginvalidaddress", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
		},
		{
			"with valid and not existing address",
			[]string{"cosmosvaloper15jkng8hytwt22lllv6mw4k89qkqehtahd84ptu", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
		},
		{
			"happy case",
			[]string{val.ValAddress.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidator()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var result types.Validator
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
				s.Require().Equal(val.ValAddress.String(), result.OperatorAddress)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryValidators() {
	val := s.network.Validators[0]

	testCases := []struct {
		name              string
		args              []string
		minValidatorCount int
	}{
		{
			"one validator case",
			[]string{
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=1", flags.FlagLimit),
			},
			1,
		},
		{
			"multi validator case",
			[]string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			len(s.network.Validators),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidators()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.QueryValidatorsResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.minValidatorCount, len(result.Validators))
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryDelegation() {
	val := s.network.Validators[0]
	val2 := s.network.Validators[1]

	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"with wrong delegator address",
			[]string{
				"wrongDelAddr",
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true, nil, nil,
		},
		{
			"with wrong validator address",
			[]string{
				val.Address.String(),
				"wrongValAddr",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true, nil, nil,
		},
		{
			"with json output",
			[]string{
				val.Address.String(),
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			&types.DelegationResponse{},
			&types.DelegationResponse{
				Delegation: types.Delegation{
					DelegatorAddress: val.Address.String(),
					ValidatorAddress: val2.ValAddress.String(),
					Shares:           math.LegacyNewDec(10),
				},
				Balance: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDelegation(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryDelegations() {
	val := s.network.Validators[0]

	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"with no delegator address",
			[]string{},
			true, nil, nil,
		},
		{
			"with wrong delegator address",
			[]string{"wrongDelAddr"},
			true, nil, nil,
		},
		{
			"valid request (height specific)",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&types.QueryDelegatorDelegationsResponse{},
			&types.QueryDelegatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{
					types.NewDelegationResp(val.Address, val.ValAddress, sdk.NewDecFromInt(cli.DefaultTokens), sdk.NewCoin(sdk.DefaultBondDenom, cli.DefaultTokens)),
				},
				Pagination: &query.PageResponse{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDelegations(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryValidatorDelegations() {
	val := s.network.Validators[0]

	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"with no validator address",
			[]string{},
			true, nil, nil,
		},
		{
			"wrong validator address",
			[]string{"wrongValAddr"},
			true, nil, nil,
		},
		{
			"valid request(height specific)",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&types.QueryValidatorDelegationsResponse{},
			&types.QueryValidatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{
					types.NewDelegationResp(val.Address, val.ValAddress, sdk.NewDecFromInt(cli.DefaultTokens), sdk.NewCoin(sdk.DefaultBondDenom, cli.DefaultTokens)),
				},
				Pagination: &query.PageResponse{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDelegations(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryUnbondingDelegations() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			[]string{
				"wrongDelAddr",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryUnbondingDelegations(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubds types.QueryDelegatorUnbondingDelegationsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &ubds)

				s.Require().NoError(err)
				s.Require().Len(ubds.UnbondingResponses, 1)
				s.Require().Equal(ubds.UnbondingResponses[0].DelegatorAddress, val.Address.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryUnbondingDelegation() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			[]string{
				"wrongDelAddr",
				val.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"wrong validator address",
			[]string{
				val.Address.String(),
				"wrongValAddr",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.Address.String(),
				val.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryUnbondingDelegation(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubd types.UnbondingDelegation

				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &ubd)
				s.Require().NoError(err)
				s.Require().Equal(ubd.DelegatorAddress, val.Address.String())
				s.Require().Equal(ubd.ValidatorAddress, val.ValAddress.String())
				s.Require().Len(ubd.Entries, 2)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryValidatorUnbondingDelegations() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong validator address",
			[]string{
				"wrongValAddr",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidatorUnbondingDelegations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubds types.QueryValidatorUnbondingDelegationsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &ubds)

				s.Require().NoError(err)
				s.Require().Len(ubds.UnbondingResponses, 1)
				s.Require().Equal(ubds.UnbondingResponses[0].DelegatorAddress, val.Address.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryRedelegations() {
	val := s.network.Validators[0]
	val2 := s.network.Validators[1]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			[]string{
				"wrongdeladdr",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryRedelegations(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var redelegations types.QueryRedelegationsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &redelegations)

				s.Require().NoError(err)

				s.Require().Len(redelegations.RedelegationResponses, 1)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.DelegatorAddress, val.Address.String())
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorSrcAddress, val.ValAddress.String())
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorDstAddress, val2.ValAddress.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryRedelegation() {
	val := s.network.Validators[0]
	val2 := s.network.Validators[1]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			[]string{
				"wrongdeladdr",
				val.ValAddress.String(),
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"wrong source validator address address",
			[]string{
				val.Address.String(),
				"wrongSrcValAddress",
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"wrong destination validator address address",
			[]string{
				val.Address.String(),
				val.ValAddress.String(),
				"wrongDestValAddress",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.Address.String(),
				val.ValAddress.String(),
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryRedelegation(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var redelegations types.QueryRedelegationsResponse

				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &redelegations)
				s.Require().NoError(err)

				s.Require().Len(redelegations.RedelegationResponses, 1)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.DelegatorAddress, val.Address.String())
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorSrcAddress, val.ValAddress.String())
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorDstAddress, val2.ValAddress.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryValidatorRedelegations() {
	val := s.network.Validators[0]
	val2 := s.network.Validators[1]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong validator address",
			[]string{
				"wrongValAddr",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.ValAddress.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidatorRedelegations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var redelegations types.QueryRedelegationsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &redelegations)

				s.Require().NoError(err)

				s.Require().Len(redelegations.RedelegationResponses, 1)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.DelegatorAddress, val.Address.String())
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorSrcAddress, val.ValAddress.String())
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorDstAddress, val2.ValAddress.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryHistoricalInfo() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		error bool
	}{
		{
			"wrong height",
			[]string{
				"-1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
		},
		{
			"valid request",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryHistoricalInfo()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.error {
				s.Require().Error(err)
			} else {
				var historicalInfo types.HistoricalInfo

				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &historicalInfo)
				s.Require().NoError(err)
				s.Require().NotNil(historicalInfo)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryParams() {
	val := s.network.Validators[0]
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"with text output",
			[]string{fmt.Sprintf("--%s=text", flags.FlagOutput)},
			`bond_denom: stake
historical_entries: 10000
max_entries: 7
max_validators: 100
min_commission_rate: "0.000000000000000000"
unbonding_time: 1814400s`,
		},
		{
			"with json output",
			[]string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			`{"unbonding_time":"1814400s","max_validators":100,"max_entries":7,"historical_entries":10000,"bond_denom":"stake","min_commission_rate":"0.000000000000000000"}`,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParams()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryPool() {
	val := s.network.Validators[0]
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"with text",
			[]string{
				fmt.Sprintf("--%s=text", flags.FlagOutput),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			fmt.Sprintf(`bonded_tokens: "%s"
not_bonded_tokens: "0"`, cli.DefaultTokens.Mul(sdk.NewInt(2)).String()),
		},
		{
			"with json",
			[]string{
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			fmt.Sprintf(`{"not_bonded_tokens":"0","bonded_tokens":"%s"}`, cli.DefaultTokens.Mul(sdk.NewInt(2)).String()),
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPool()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

// TestBlockResults tests that the validator updates correctly show when
// calling the /block_results RPC endpoint.
// ref: https://github.com/cosmos/cosmos-sdk/issues/7401.
func (s *E2ETestSuite) TestBlockResults() {
	require := s.Require()
	val := s.network.Validators[0]

	// Create new account in the keyring.
	k, _, err := val.ClientCtx.Keyring.NewMnemonic("NewDelegator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)
	pub, err := k.GetPubKey()
	require.NoError(err)
	newAddr := sdk.AccAddress(pub.Address())

	// Send some funds to the new account.
	_, err = clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	// Use CLI to create a delegation from the new account to validator `val`.
	cmd := cli.NewDelegateCmd()
	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, []string{
		val.ValAddress.String(),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(150)).String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	})
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	// Create a HTTP rpc client.
	rpcClient, err := http.New(val.RPCAddress, "/websocket")
	require.NoError(err)

	// Loop until we find a block result with the correct validator updates.
	// By experience, it happens around 2 blocks after `delHeight`.
	s.network.RetryForBlocks(func() error {
		latestHeight, err := s.network.LatestHeight()
		require.NoError(err)
		res, err := rpcClient.BlockResults(context.Background(), &latestHeight)
		if err != nil {
			return err
		}

		if len(res.ValidatorUpdates) == 0 {
			return errors.New("validator update not found yet")
		}

		valUpdate := res.ValidatorUpdates[0]
		require.Equal(
			valUpdate.GetPubKey().Sum.(*crypto.PublicKey_Ed25519).Ed25519,
			val.PubKey.Bytes(),
		)

		return nil
	}, 10)
}

// https://github.com/cosmos/cosmos-sdk/issues/10660
func (s *E2ETestSuite) TestEditValidatorMoniker() {
	val := s.network.Validators[0]
	require := s.Require()

	txCmd := cli.NewEditValidatorCmd()
	moniker := "testing"
	_, err := clitestutil.ExecTestCLICmd(val.ClientCtx, txCmd, []string{
		val.ValAddress.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", cli.FlagEditMoniker, moniker),
		fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	})
	require.NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	queryCmd := cli.GetCmdQueryValidator()
	res, err := clitestutil.ExecTestCLICmd(
		val.ClientCtx, queryCmd,
		[]string{val.ValAddress.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
	)
	require.NoError(err)
	var result types.Validator
	require.NoError(val.ClientCtx.Codec.UnmarshalJSON(res.Bytes(), &result))
	require.Equal(result.GetMoniker(), moniker)

	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, txCmd, []string{
		val.ValAddress.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	})
	require.NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	res, err = clitestutil.ExecTestCLICmd(
		val.ClientCtx, queryCmd,
		[]string{val.ValAddress.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
	)
	require.NoError(err)
	require.NoError(val.ClientCtx.Codec.UnmarshalJSON(res.Bytes(), &result))
	require.Equal(result.GetMoniker(), moniker)
}
