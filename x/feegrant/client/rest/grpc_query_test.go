// +build norace

package rest_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant/client/cli"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

type IntegrationTestSuite struct {
	suite.Suite
	cfg     network.Config
	network *network.Network
	grantee sdk.AccAddress
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()

	cfg.NumValidators = 1
	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	val := s.network.Validators[0]
	// Create new account in the keyring.
	info, _, err := val.ClientCtx.Keyring.NewMnemonic("grantee", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	// Send some funds to the new account.
	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	s.grantee = newAddr
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryFeeAllowance() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func()
		postRun   func(_ types.QueryFeeAllowanceResponse)
	}{
		{
			"fail: invalid granter",
			fmt.Sprintf("%s/cosmos/feegrant/v1beta1/fee_allowance/%s/%s", baseURL, "invalid_granter", s.grantee.String()),
			true,
			"decoding bech32 failed: invalid index of 1: invalid request",
			func() {},
			func(types.QueryFeeAllowanceResponse) {},
		},
		{
			"fail: invalid grantee",
			fmt.Sprintf("%s/cosmos/feegrant/v1beta1/fee_allowance/%s/%s", baseURL, val.Address.String(), "invalid_grantee"),
			true,
			"decoding bech32 failed: invalid index of 1: invalid request",
			func() {},
			func(types.QueryFeeAllowanceResponse) {},
		},
		{
			"fail: no grants",
			fmt.Sprintf("%s/cosmos/feegrant/v1beta1/fee_allowance/%s/%s", baseURL, val.Address.String(), s.grantee.String()),
			true,
			"fee-grant not found",
			func() {},
			func(types.QueryFeeAllowanceResponse) {},
		},
		{
			"valid query: expect single grant",
			fmt.Sprintf("%s/cosmos/feegrant/v1beta1/fee_allowance/%s/%s", baseURL, val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {
				execFeeAllowance(val, s)
			},
			func(allowance types.QueryFeeAllowanceResponse) {
				s.Require().Equal(allowance.FeeAllowance.Granter, val.Address.String())
				s.Require().Equal(allowance.FeeAllowance.Grantee, s.grantee.String())
			},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.preRun()
			resp, _ := rest.GetRequest(tc.url)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				var allowance types.QueryFeeAllowanceResponse
				err := val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &allowance)
				s.Require().NoError(err)
				tc.postRun(allowance)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGranteeAllowances() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
		preRun    func()
		postRun   func(_ types.QueryFeeAllowancesResponse)
	}{
		{
			"fail: invalid grantee",
			fmt.Sprintf("%s/cosmos/feegrant/v1beta1/fee_allowances/%s", baseURL, "invalid_grantee"),
			true,
			"decoding bech32 failed: invalid index of 1: invalid request",
			func() {},
			func(types.QueryFeeAllowancesResponse) {},
		},
		{
			"success: no grants",
			fmt.Sprintf("%s/cosmos/feegrant/v1beta1/fee_allowances/%s?pagination.offset=1", baseURL, s.grantee.String()),
			false,
			"",
			func() {},
			func(allowances types.QueryFeeAllowancesResponse) {
				s.Require().Equal(len(allowances.FeeAllowances), 0)
			},
		},
		{
			"valid query: expect single grant",
			fmt.Sprintf("%s/cosmos/feegrant/v1beta1/fee_allowances/%s", baseURL, s.grantee.String()),
			false,
			"",
			func() {
				execFeeAllowance(val, s)
			},
			func(allowances types.QueryFeeAllowancesResponse) {
				s.Require().Equal(len(allowances.FeeAllowances), 1)
				s.Require().Equal(allowances.FeeAllowances[0].Granter, val.Address.String())
				s.Require().Equal(allowances.FeeAllowances[0].Grantee, s.grantee.String())
			},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.preRun()
			resp, _ := rest.GetRequest(tc.url)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				var allowance types.QueryFeeAllowancesResponse
				err := val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &allowance)
				s.Require().NoError(err)
				tc.postRun(allowance)
			}
		})
	}
}

func execFeeAllowance(val *network.Validator, s *IntegrationTestSuite) {
	fee := sdk.NewCoin("steak", sdk.NewInt(100))
	duration := 365 * 24 * 60 * 60
	args := []string{
		val.Address.String(),
		s.grantee.String(),
		fmt.Sprintf("--%s=%s", cli.FlagSpendLimit, fee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=%v", cli.FlagExpiration, duration),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	cmd := cli.NewCmdFeeGrant()
	_, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
