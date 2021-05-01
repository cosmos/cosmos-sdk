// +build norace

package rest_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	authztestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type IntegrationTestSuite struct {
	suite.Suite
	cfg     network.Config
	network *network.Network
	grantee sdk.AccAddress
}

var typeMsgSend = banktypes.SendAuthorization{}.MethodName()
var typeMsgVote = sdk.MsgTypeURL(&govtypes.MsgVote{})

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
	out, err := banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)
	s.Require().Contains(out.String(), `"code":0`)

	// grant authorization
	out, err = authztestutil.ExecGrantAuthorization(val, []string{
		newAddr.String(),
		"send",
		fmt.Sprintf("--%s=100steak", cli.FlagSpendLimit),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
	})
	s.Require().NoError(err)
	s.Require().Contains(out.String(), `"code":0`)

	s.grantee = newAddr
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryAuthorizationGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
	}{
		{
			"fail invalid granter address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?method_name=%s", baseURL, "invalid_granter", s.grantee.String(), typeMsgSend),
			true,
			"decoding bech32 failed: invalid index of 1: invalid request",
		},
		{
			"fail invalid grantee address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?method_name=%s", baseURL, val.Address.String(), "invalid_grantee", typeMsgSend),
			true,
			"decoding bech32 failed: invalid index of 1: invalid request",
		},
		{
			"fail with empty granter",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?method_name=%s", baseURL, "", s.grantee.String(), typeMsgSend),
			true,
			"Not Implemented",
		},
		{
			"fail with empty grantee",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?method_name=%s", baseURL, val.Address.String(), "", typeMsgSend),
			true,
			"Not Implemented",
		},
		{
			"fail invalid msg-type",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?method_name=%s", baseURL, val.Address.String(), s.grantee.String(), "invalidMsg"),
			true,
			"rpc error: code = NotFound desc = no authorization found for invalidMsg type: key not found",
		},
		{
			"valid query",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grant?method_name=%s", baseURL, val.Address.String(), s.grantee.String(), typeMsgSend),
			false,
			"",
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, _ := rest.GetRequest(tc.url)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errorMsg)
			} else {
				var authorization authz.QueryAuthorizationResponse
				err := val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &authorization)
				s.Require().NoError(err)
				authorization.Authorization.UnpackInterfaces(val.ClientCtx.InterfaceRegistry)
				auth := authorization.Authorization.GetAuthorizationGrant()
				s.Require().Equal(auth.MethodName(), banktypes.SendAuthorization{}.MethodName())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryAuthorizationsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errMsg    string
		preRun    func()
		postRun   func(*authz.QueryAuthorizationsResponse)
	}{
		{
			"fail invalid granter address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants", baseURL, "invalid_granter", s.grantee.String()),
			true,
			"decoding bech32 failed: invalid index of 1: invalid request",
			func() {},
			func(_ *authz.QueryAuthorizationsResponse) {},
		},
		{
			"fail invalid grantee address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants", baseURL, val.Address.String(), "invalid_grantee"),
			true,
			"decoding bech32 failed: invalid index of 1: invalid request",
			func() {},
			func(_ *authz.QueryAuthorizationsResponse) {},
		},
		{
			"fail empty grantee address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants", baseURL, "", "invalid_grantee"),
			true,
			"Not Implemented",
			func() {},
			func(_ *authz.QueryAuthorizationsResponse) {},
		},
		{
			"valid query: expect single grant",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants", baseURL, val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {},
			func(authorizations *authz.QueryAuthorizationsResponse) {
				s.Require().Len(authorizations.Authorizations), 1)
			},
		},
		{
			"valid query: expect two grants",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants", baseURL, val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {
				_, err := authztestutil.ExecGrantAuthorization(val, []string{
					s.grantee.String(),
					"generic",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
					fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
					fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
					fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
					fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
					fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
				})
				s.Require().NoError(err)
			},
			func(authorizations *authz.QueryAuthorizationsResponse) {
				s.Require().Equal(len(authorizations.Authorizations), 2)
			},
		},
		{
			"valid query: expect single grant with pagination",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants?pagination.limit=1", baseURL, val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {},
			func(authorizations *authz.QueryAuthorizationsResponse) {
				s.Require().Equal(len(authorizations.Authorizations), 1)
			},
		},
		{
			"valid query: expect two grants with pagination",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/granters/%s/grantees/%s/grants?pagination.limit=2", baseURL, val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {},
			func(authorizations *authz.QueryAuthorizationsResponse) {
				s.Require().Equal(len(authorizations.Authorizations), 2)
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			tc.preRun()
			resp, _ := rest.GetRequest(tc.url)
			if tc.expectErr {
				s.Require().Contains(string(resp), tc.errMsg)
			} else {
				var authorizations authz.QueryAuthorizationsResponse
				err := val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &authorizations)
				s.Require().NoError(err)
				tc.postRun(&authorizations)
			}

		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
