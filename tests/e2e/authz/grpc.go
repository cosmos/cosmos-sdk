package authz

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	authzclitestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"gotest.tools/v3/assert"
)

func TestQueryGrantGRPC(t *testing.T) {
	f := initFixture(t)

	val := f.network.Validators[0]
	grantee := f.grantee[1]
	grantsURL := val.APIAddress + "/cosmos/authz/v1beta1/grants?granter=%s&grantee=%s&msg_type_url=%s"
	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errorMsg  string
	}{
		{
			"fail invalid granter address",
			fmt.Sprintf(grantsURL, "invalid_granter", grantee.String(), typeMsgSend),
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
			fmt.Sprintf(grantsURL, "", grantee.String(), typeMsgSend),
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
			fmt.Sprintf(grantsURL, val.Address.String(), grantee.String(), "invalidMsg"),
			true,
			"authorization not found for invalidMsg type",
		},
		{
			"valid query",
			fmt.Sprintf(grantsURL, val.Address.String(), grantee.String(), typeMsgSend),
			false,
			"",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := testutil.GetRequest(tc.url)
			// require := f.Require()
			if tc.expectErr {
				assert.Equal(t, strings.Contains(string(resp), tc.errorMsg), true)
			} else {
				var g authz.QueryGrantsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &g)
				assert.NilError(t, err, "")
				assert.Equal(t, len(g.Grants), 1)
				g.Grants[0].UnpackInterfaces(val.ClientCtx.InterfaceRegistry)
				auth, err := g.Grants[0].GetAuthorization()
				assert.NilError(t, err, "")
				assert.Equal(t, auth.MsgTypeURL(), banktypes.SendAuthorization{}.MsgTypeURL())
			}
		})
	}
}

func TestQueryGrantsGRPC(t *testing.T) {
	f := initFixture(t)
	val := f.network.Validators[0]
	grantee := f.grantee[1]
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
			fmt.Sprintf(grantsURL, val.Address.String(), grantee.String()),
			false,
			"",
			func() {},
			func(g *authz.QueryGrantsResponse) {
				assert.Equal(t, len(g.Grants), 1)
			},
		},
		{
			"valid query: expect two grants",
			fmt.Sprintf(grantsURL, val.Address.String(), grantee.String()),
			false,
			"",
			func() {
				_, err := authzclitestutil.CreateGrant(val.ClientCtx, []string{
					grantee.String(),
					"generic",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
					fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
					fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
					fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
					fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
					fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
				})
				assert.NilError(t, err) //TODO
				assert.NilError(t, f.network.WaitForNextBlock())
			},

			func(g *authz.QueryGrantsResponse) {
				assert.Equal(t, len(g.Grants), 2)
			},
		},
		{
			"valid query: expect single grant with pagination",
			fmt.Sprintf(grantsURL+"&pagination.limit=1", val.Address.String(), grantee.String()),
			false,
			"",
			func() {},
			func(g *authz.QueryGrantsResponse) {
				assert.Equal(t, len(g.Grants), 1)

			},
		},
		{
			"valid query: expect two grants with pagination",
			fmt.Sprintf(grantsURL+"&pagination.limit=2", val.Address.String(), grantee.String()),
			false,
			"",
			func() {},
			func(g *authz.QueryGrantsResponse) {
				assert.Equal(t, len(g.Grants), 2)

			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.preRun()
			resp, err := testutil.GetRequest(tc.url)
			assert.NilError(t, err)

			if tc.expectErr {
				assert.Equal(t, strings.Contains(string(resp), tc.errMsg), true)
			} else {
				var authorizations authz.QueryGrantsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &authorizations)
				assert.NilError(t, err)
				tc.postRun(&authorizations)
			}
		})
	}
}

func TestQueryGranterGrantsGRPC(t *testing.T) {
	f := initFixture(t)
	val := f.network.Validators[0]
	grantee := f.grantee[1]
	// require := s.Require()

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errMsg    string
		numItems  int
	}{
		{
			"invalid account address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/grants/granter/%s", val.APIAddress, "invalid address"),
			true,
			"decoding bech32 failed",
			0,
		},
		{
			"no authorizations found",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/grants/granter/%s", val.APIAddress, grantee.String()),
			false,
			"",
			0,
		},
		{
			"valid query",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/grants/granter/%s", val.APIAddress, val.Address.String()),
			false,
			"",
			8,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequest(tc.url)
			assert.NilError(t, err)

			if tc.expectErr {
				assert.Equal(t, strings.Contains(string(resp), tc.errMsg), true)
			} else {
				var authorizations authz.QueryGranterGrantsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &authorizations)
				assert.NilError(t, err)
				assert.Equal(t, len(authorizations.Grants), tc.numItems)
			}
		})
	}
}

func TestQueryGranteeGrantsGRPC(t *testing.T) {
	f := initFixture(t)
	val := f.network.Validators[0]
	grantee := f.grantee[1]

	testCases := []struct {
		name      string
		url       string
		expectErr bool
		errMsg    string
		numItems  int
	}{
		{
			"invalid account address",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/grants/grantee/%s", val.APIAddress, "invalid address"),
			true,
			"decoding bech32 failed",
			0,
		},
		{
			"no authorizations found",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/grants/grantee/%s", val.APIAddress, val.Address.String()),
			false,
			"",
			0,
		},
		{
			"valid query",
			fmt.Sprintf("%s/cosmos/authz/v1beta1/grants/grantee/%s", val.APIAddress, grantee.String()),
			false,
			"",
			1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequest(tc.url)
			assert.NilError(t, err)

			if tc.expectErr {
				assert.Equal(t, strings.Contains(string(resp), tc.errMsg), true)
			} else {
				var authorizations authz.QueryGranteeGrantsResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(resp, &authorizations)
				assert.NilError(t, err)
				assert.Equal(t, len(authorizations.Grants), tc.numItems)
			}
		})
	}
}
