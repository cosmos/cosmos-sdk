package systemtests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// const (
// 	msgVoteTypeURL:= `/cosmos.gov.v1.MsgVote`
// )

func TestAuthzGrantTxCmd(t *testing.T) {
	// scenario: test authz grant command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, valAddr)

	// add one more key
	account1Addr := cli.AddKey("account1")
	sut.StartChain(t)

	grantCmdArgs := []string{"tx", "authz", "grant", account1Addr, "--from", valAddr}

	// test grant command
	testCases := []struct {
		name      string
		cmdArgs   []string
		expectErr bool
		expErrMsg string
	}{
		{
			"not enough arguments",
			[]string{},
			true,
			"accepts 2 arg(s), received 1",
		},
		{
			"invalid authorization type",
			[]string{"spend"},
			true,
			"invalid authorization type",
		},
		{
			"send authorization without spend-limit",
			[]string{"send"},
			true,
			"spend-limit should be greater than zero",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := append(grantCmdArgs, tc.cmdArgs...)
			if tc.expectErr {
				assertErr := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
					require.Len(t, gotOutputs, 1)
					output := gotOutputs[0].(string)
					require.Contains(t, output, tc.expErrMsg)
					return false
				}
				_ = cli.WithRunErrorMatcher(assertErr).Run(cmd...)
			} else {
				rsp := cli.RunAndWait(cmd...)
				RequireTxSuccess(t, rsp)
			}
		})
	}
}
