package cli

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/stretchr/testify/require"
)

func TestGetQueryCmd(t *testing.T) {
	testCases := []struct {
		name          string
		cmd           *cobra.Command
		args          []string
		expectedError bool
		expectedOut   string
	}{
		{
			name: "test get disable list command",
			cmd:  GetDisabeListCmd(),
			args: []string{
				"disabled_messages",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			expectedOut: "disabled_messages",
		},
		{
			name: "test get account command",
			cmd:  GetAccountCmd(),
			args: []string{
				"test_account",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			expectedOut: "test_account",
		},
		{
			name: "test get accounts command",
			cmd:  GetAccountsCmd(),
			args: []string{
				"accounts",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			expectedOut: "accounts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.cmd
			cmd.SetArgs(tc.args)
			require.Contains(t, fmt.Sprint(cmd), tc.expectedOut)
		})
	}
}
