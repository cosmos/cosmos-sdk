package authz

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	authzclitestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
	"gotest.tools/v3/assert"
)

func TestQueryAuthorizations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]

	grantee := f.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		expErrMsg string
	}{
		{
			"Error: Invalid grantee",
			[]string{
				val.Address.String(),
				"invalid grantee",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"decoding bech32 failed: invalid character in string: ' '",
		},
		{
			"Error: Invalid granter",
			[]string{
				"invalid granter",
				grantee.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"decoding bech32 failed: invalid character in string: ' '",
		},
		{
			"Valid txn (json)",
			[]string{
				val.Address.String(),
				grantee.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			``,
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetCmdQueryGrants()
			clientCtx := val.ClientCtx
			resp, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Equal(t, strings.Contains(string(resp.Bytes()), tc.expErrMsg), true)
			} else {
				assert.NilError(t, err)
				var grants authz.QueryGrantsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &grants)
				assert.NilError(t, err)
			}
		})
	}
}

func TestQueryAuthorization(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]

	grantee := f.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	assert.NilError(t, err)
	assert.NilError(t, f.network.WaitForNextBlock())

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"Error: Invalid grantee",
			[]string{
				val.Address.String(),
				"invalid grantee",
				typeMsgSend,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"",
		},
		{
			"Error: Invalid granter",
			[]string{
				"invalid granter",
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"",
		},
		{
			"no authorization found",
			[]string{
				val.Address.String(),
				grantee.String(),
				"typeMsgSend",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"",
		},
		{
			"Valid query (json)",
			[]string{
				val.Address.String(),
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			`{"@type":"/cosmos.bank.v1beta1.SendAuthorization","spend_limit":[{"denom":"stake","amount":"100"}],"allow_list":[]}`,
		},
		{
			"Valid query with allowed list (json)",
			[]string{
				val.Address.String(),
				f.grantee[3].String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.SendAuthorization","spend_limit":[{"denom":"stake","amount":"100"}],"allow_list":["%s"]}`, f.grantee[4]),
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetCmdQueryGrants()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, "")
			} else {
				assert.NilError(t, err)
				fmt.Println("out", strings.TrimSpace(out.String()), tc.expectedOutput)
				assert.Equal(t, strings.Contains(strings.TrimSpace(out.String()), tc.expectedOutput), true)
			}
		})
	}
}

func TestQueryGranterGrants(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.TearDownSuite(t)
	val := f.network.Validators[0]
	grantee := f.grantee[0]

	testCases := []struct {
		name        string
		args        []string
		expectErr   bool
		expectedErr string
		expItems    int
	}{
		{
			"invalid address",
			[]string{
				"invalid-address",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"decoding bech32 failed",
			0,
		},
		{
			"no authorization found",
			[]string{
				grantee.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			"",
			0,
		},
		{
			"valid case",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			"",
			3,
		},
		{
			"valid case with pagination",
			[]string{
				val.Address.String(),
				"--limit=2",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			"",
			2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetQueryGranterGrants()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Equal(t, strings.Contains(out.String(), tc.expectedErr), true)
			} else {
				assert.NilError(t, err)
				var grants authz.QueryGranterGrantsResponse
				assert.NilError(t, val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &grants))
				assert.Equal(t, len(grants.Grants), tc.expItems)
			}
		})
	}
}
