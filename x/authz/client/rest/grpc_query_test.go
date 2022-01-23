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

var typeMsgSend = banktypes.SendAuthorization{}.MsgTypeURL()
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
	out, err = authztestutil.ExecGrant(val, []string{
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

func (s *IntegrationTestSuite) TestQueryGrantGRPC() {
	val := s.network.Validators[0]
	grantsURL := val.APIAddress + "/cosmos/authz/v1beta1/grants?granter=%s&grantee=%s&msg_type_url=%s"
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
	}{
		{
			"fail invalid granter address",
			fmt.Sprintf(grantsURL, "invalid_granter", s.grantee.String(), typeMsgSend),
			true,
			"decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			"fail invalid grantee address",
			fmt.Sprintf(grantsURL, val.Address.String(), "invalid_grantee", typeMsgSend),
			true,
			"decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			"fail with empty granter",
			fmt.Sprintf(grantsURL, "", s.grantee.String(), typeMsgSend),
			true,
			"empty address string is not allowed: invalid request",
		},
		{
			"fail with empty grantee",
			fmt.Sprintf(grantsURL, val.Address.String(), "", typeMsgSend),
			true,
			"empty address string is not allowed: invalid request",
		},
		{
			"fail invalid msg-type",
			fmt.Sprintf(grantsURL, val.Address.String(), s.grantee.String(), "invalidMsg"),
			true,
			"rpc error: code = NotFound desc = no authorization found for invalidMsg type: key not found",
		},
		{
			"valid query",
			fmt.Sprintf(grantsURL, val.Address.String(), s.grantee.String(), typeMsgSend),
			false,
			"",
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, _ := rest.GetRequest(tc.url)
			require := s.Require()
			if tc.expectErr {
				require.Contains(string(resp), tc.errorMsg)
			} else {
				var g authz.QueryGrantsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &g)
				require.NoError(err)
				require.Len(g.Grants, 1)
				g.Grants[0].UnpackInterfaces(val.ClientCtx.InterfaceRegistry)
				auth := g.Grants[0].GetAuthorization()
				require.Equal(auth.MsgTypeURL(), banktypes.SendAuthorization{}.MsgTypeURL())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGrantsGRPC() {
	val := s.network.Validators[0]
	grantsURL := val.APIAddress + "/cosmos/authz/v1beta1/grants?granter=%s&grantee=%s"
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errMsg    string
		preRun    func()
		postRun   func(*authz.QueryGrantsResponse)
	}{
		{
			"valid query: expect single grant",
			fmt.Sprintf(grantsURL, val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {},
			func(g *authz.QueryGrantsResponse) {
				s.Require().Len(g.Grants, 1)
			},
		},
		{
			"valid query: expect two grants",
			fmt.Sprintf(grantsURL, val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {
				_, err := authztestutil.ExecGrant(val, []string{
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
			func(g *authz.QueryGrantsResponse) {
				s.Require().Len(g.Grants, 2)
			},
		},
		{
			"valid query: expect single grant with pagination",
			fmt.Sprintf(grantsURL+"&pagination.limit=1", val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {},
			func(g *authz.QueryGrantsResponse) {
				s.Require().Len(g.Grants, 1)
			},
		},
		{
			"valid query: expect two grants with pagination",
			fmt.Sprintf(grantsURL+"&pagination.limit=2", val.Address.String(), s.grantee.String()),
			false,
			"",
			func() {},
			func(g *authz.QueryGrantsResponse) {
				s.Require().Len(g.Grants, 2)
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
				var authorizations authz.QueryGrantsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &authorizations)
				s.Require().NoError(err)
				tc.postRun(&authorizations)
			}

		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
