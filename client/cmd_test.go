package client_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
)

func TestValidateCmd(t *testing.T) {
	// setup root and subcommands
	rootCmd := &cobra.Command{
		Use: "root",
	}
	queryCmd := &cobra.Command{
		Use: "query",
	}
	rootCmd.AddCommand(queryCmd)

	// command being tested
	distCmd := &cobra.Command{
		Use:                        "distr",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}
	queryCmd.AddCommand(distCmd)

	commissionCmd := &cobra.Command{
		Use: "commission",
	}
	distCmd.AddCommand(commissionCmd)

	tests := []struct {
		reason  string
		args    []string
		wantErr bool
	}{
		{"misspelled command", []string{"COMMISSION"}, true},
		{"no command provided", []string{}, false},
		{"help flag", []string{"COMMISSION", "--help"}, false},
		{"shorthand help flag", []string{"COMMISSION", "-h"}, false},
		{"flag only, no command provided", []string{"--gas", "1000atom"}, false},
		{"flag and misspelled command", []string{"--gas", "1000atom", "COMMISSION"}, true},
	}

	for _, tt := range tests {
		err := client.ValidateCmd(distCmd, tt.args)
		require.Equal(t, tt.wantErr, err != nil, tt.reason)
	}
}
