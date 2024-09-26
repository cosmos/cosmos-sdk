//go:build system_test

package systemtests

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSubmitProposal(t *testing.T) {
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "0.address").String()
	require.NotEmpty(t, valAddr)

	sut.StartChain(t)

	// get gov module address
	resp := cli.CustomQuery("q", "auth", "module-account", "gov")
	govAddress := gjson.Get(resp, "account.value.address").String()

	invalidProp := `{
	"title": "",
	"description": "Where is the title!?",
	"type": "Text",
	"deposit": "-324foocoin"
}`
	invalidPropFile := testutil.WriteToNewTempFile(t, invalidProp)
	os.WriteFile("test", []byte(invalidProp), 0o600)
	defer invalidPropFile.Close()

	// Create a valid new proposal JSON.
	propMetadata := []byte{42}
	validProp := fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.gov.v1.MsgExecLegacyContent",
			"authority": "%s",
			"content": {
				"@type": "/cosmos.gov.v1beta1.TextProposal",
				"title": "My awesome title",
				"description": "My awesome description"
			}
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"metadata": "%s",
	"deposit": "%s"
}`, govAddress, base64.StdEncoding.EncodeToString(propMetadata), sdk.NewCoin("stake", math.NewInt(100000)))
	validPropFile := testutil.WriteToNewTempFile(t, validProp)
	defer validPropFile.Close()

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			"invalid proposal",
			[]string{
				"tx", "gov", "submit-proposal",
				invalidPropFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			true,
			"invalid character in coin string",
		},
		{
			"valid proposal",
			[]string{
				"tx", "gov", "submit-proposal",
				validPropFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectErr {
				assertOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					require.Contains(t, gotOutputs[0], tc.errMsg)
					return false
				}

				cli.WithRunErrorMatcher(assertOutput).Run(tc.args...)
			} else {
				rsp := cli.Run(tc.args...)
				txResult, found := cli.AwaitTxCommitted(rsp)
				require.True(t, found)
				RequireTxSuccess(t, txResult)
			}
		})
	}
}

func TestSubmitLegacyProposal(t *testing.T) {
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "0.address").String()
	require.NotEmpty(t, valAddr)

	sut.StartChain(t)

	// get gov module address
	// resp := cli.CustomQuery("q", "auth", "module-account", "gov")
	// govAddress := gjson.Get(resp, "account.value.address").String()

	invalidProp := `{
	"title": "",
		"description": "Where is the title!?",
		"type": "Text",
	"deposit": "-324foocoin"
	}`
	invalidPropFile := testutil.WriteToNewTempFile(t, invalidProp)
	defer invalidPropFile.Close()

	validProp := fmt.Sprintf(`{
		"title": "Text Proposal",
		  "description": "Hello, World!",
		  "type": "Text",
		"deposit": "%s"
	  }`, sdk.NewCoin("stake", math.NewInt(154310)))
	validPropFile := testutil.WriteToNewTempFile(t, validProp)
	defer validPropFile.Close()

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			"invalid proposal (file)",
			[]string{
				"tx", "gov", "submit-legacy-proposal",
				fmt.Sprintf("--%s=%s", "proposal", invalidPropFile.Name()), //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			true,
			"proposal title is required",
		},
		{
			"invalid proposal",
			[]string{
				"tx", "gov", "submit-legacy-proposal",
				fmt.Sprintf("--%s='Where is the title!?'", "description"), //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", "type", "Text"),                    //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", "deposit", sdk.NewCoin("stake", math.NewInt(10000)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			true,
			"proposal title is required",
		},
		{
			"valid transaction (file)",
			[]string{
				"tx", "gov", "submit-legacy-proposal",
				fmt.Sprintf("--%s=%s", "proposal", validPropFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			false,
			"",
		},
		{
			"valid transaction",
			[]string{
				"tx", "gov", "submit-legacy-proposal",
				fmt.Sprintf("--%s='Text Proposal'", "title"),
				fmt.Sprintf("--%s='Where is the title!?'", "description"), //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", "type", "Text"),                    //nolint:staticcheck // we are intentionally using a deprecated flag here.
				fmt.Sprintf("--%s=%s", "deposit", sdk.NewCoin("stake", math.NewInt(100000)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectErr {
				assertOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					fmt.Println("gotOut", gotOutputs)
					require.Contains(t, gotOutputs[0], tc.errMsg)
					return false
				}

				cli.WithRunErrorMatcher(assertOutput).Run(tc.args...)
			} else {
				rsp := cli.Run(tc.args...)
				txResult, found := cli.AwaitTxCommitted(rsp)
				require.True(t, found)
				RequireTxSuccess(t, txResult)
			}
		})
	}
}
