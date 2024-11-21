//go:build system_test

package systemtests

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"cosmossdk.io/math"
	systest "cosmossdk.io/systemtests"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSubmitProposal(t *testing.T) {
	// given a running chain

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "0.address").String()
	require.NotEmpty(t, valAddr)

	systest.Sut.StartChain(t)

	// get gov module address
	resp := cli.CustomQuery("q", "auth", "module-account", "gov")
	govAddress := gjson.Get(resp, "account.value.address").String()

	invalidProp := `{
	"title": "",
	"description": "Where is the title!?",
	"type": "Text",
	"deposit": "-324foocoin"
}`

	invalidPropFile := systest.StoreTempFile(t, []byte(invalidProp))
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
	validPropFile := systest.StoreTempFile(t, []byte(validProp))
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
				systest.RequireTxSuccess(t, txResult)
			}
		})
	}
}

func TestSubmitLegacyProposal(t *testing.T) {
	// given a running chain

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "0.address").String()
	require.NotEmpty(t, valAddr)

	systest.Sut.StartChain(t)

	invalidProp := `{
	"title": "",
		"description": "Where is the title!?",
		"type": "Text",
	"deposit": "-324foocoin"
	}`
	invalidPropFile := systest.StoreTempFile(t, []byte(invalidProp))
	defer invalidPropFile.Close()

	validProp := fmt.Sprintf(`{
		"title": "Text Proposal",
		  "description": "Hello, World!",
		  "type": "Text",
		"deposit": "%s"
	  }`, sdk.NewCoin("stake", math.NewInt(154310)))
	validPropFile := systest.StoreTempFile(t, []byte(validProp))
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
				fmt.Sprintf("--%s=%s", "proposal", invalidPropFile.Name()),
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
				fmt.Sprintf("--%s='Where is the title!?'", "description"),
				fmt.Sprintf("--%s=%s", "type", "Text"),
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
				fmt.Sprintf("--%s='Where is the title!?'", "description"),
				fmt.Sprintf("--%s=%s", "type", "Text"),
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
				systest.RequireTxSuccess(t, txResult)
			}
		})
	}
}

func TestNewCmdWeightedVote(t *testing.T) {
	// given a running chain

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "0.address").String()
	require.NotEmpty(t, valAddr)

	systest.Sut.StartChain(t)

	// Submit a new proposal for voting
	proposalArgs := []string{
		"tx", "gov", "submit-legacy-proposal",
		fmt.Sprintf("--%s='Text Proposal'", "title"),
		fmt.Sprintf("--%s='Where is the title!?'", "description"),
		fmt.Sprintf("--%s=%s", "type", "Text"),
		fmt.Sprintf("--%s=%s", "deposit", sdk.NewCoin("stake", math.NewInt(10_000_000)).String()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
	}
	rsp := cli.Run(proposalArgs...)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)

	proposalsResp := cli.CustomQuery("q", "gov", "proposals")
	proposals := gjson.Get(proposalsResp, "proposals.#.id").Array()
	require.NotEmpty(t, proposals)

	proposal1 := cli.CustomQuery("q", "gov", "proposal", "1")
	fmt.Println("first proposal", proposal1)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		broadcasted  bool
		errMsg       string
	}{
		{
			"vote for invalid proposal",
			[]string{
				"tx", "gov", "weighted-vote",
				"10",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			true, 3, true,
			"inactive proposal",
		},
		{
			"valid vote",
			[]string{
				"tx", "gov", "weighted-vote",
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			false, 0, true,
			"",
		},
		{
			"valid vote with metadata",
			[]string{
				"tx", "gov", "weighted-vote",
				"1",
				"yes",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--metadata=%s", "AQ=="),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			false, 0, true,
			"",
		},
		{
			"invalid valid split vote string",
			[]string{
				"tx", "gov", "weighted-vote",
				"1",
				"yes/0.6,no/0.3,abstain/0.05,no_with_veto/0.05",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			true, 0, false,
			"is not a valid vote option",
		},
		{
			"valid split vote",
			[]string{
				"tx", "gov", "weighted-vote",
				"1",
				"yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
			},
			false, 0, true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.broadcasted {
				assertOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					require.Contains(t, gotOutputs[0], tc.errMsg)
					return false
				}
				cli.WithRunErrorMatcher(assertOutput).Run(tc.args...)
			} else {
				rsp := cli.Run(tc.args...)
				if tc.expectErr {
					systest.RequireTxFailure(t, rsp)
				} else {
					cli.AwaitTxCommitted(rsp)
				}
			}
		})
	}
}

func TestQueryDeposit(t *testing.T) {
	// given a running chain
	systest.Sut.ResetChain(t)
	// short voting period
	// update expedited voting period to avoid validation error
	votingPeriod := 3 * time.Second
	systest.Sut.ModifyGenesisJSON(
		t,
		systest.SetGovVotingPeriod(t, votingPeriod),
		systest.SetGovExpeditedVotingPeriod(t, votingPeriod-time.Second),
	)
	systest.Sut.StartChain(t)

	// get validator address
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	valAddr := cli.GetKeyAddr("node0")

	// Submit a new proposal for voting
	proposalArgs := []string{
		"tx", "gov", "submit-legacy-proposal",
		fmt.Sprintf("--%s='Text Proposal'", "title"),
		fmt.Sprintf("--%s='Where is the title!?'", "description"),
		fmt.Sprintf("--%s=%s", "type", "Text"),
		fmt.Sprintf("--%s=%s", "deposit", sdk.NewCoin("stake", math.NewInt(10_000_000)).String()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))).String()),
	}
	rsp := cli.Run(proposalArgs...)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	systest.RequireTxSuccess(t, txResult)

	// Query initial deposit
	resp := cli.CustomQuery("q", "gov", "deposit", "1", valAddr)
	depositAmount := gjson.Get(resp, "deposit.amount.0.amount").Int()
	require.Equal(t, depositAmount, int64(10_000_000))

	resp = cli.CustomQuery("q", "gov", "deposits", "1")
	deposits := gjson.Get(resp, "deposits").Array()
	require.Equal(t, len(deposits), 1)

	assert.Eventually(t, func() bool {
		resp = cli.CustomQuery("q", "gov", "deposits", "1")
		deposits = gjson.Get(resp, "deposits").Array()
		return len(deposits) == 0
	}, votingPeriod, 100*time.Millisecond)
}
