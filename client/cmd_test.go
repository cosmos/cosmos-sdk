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
		{"misspelled command", []string{"commission"}, true}, // nolint: misspell
		{"no command provided", []string{}, false},
		{"help flag", []string{"commission", "--help"}, false},       // nolint: misspell
		{"shorthand help flag", []string{"commission", "-h"}, false}, // nolint: misspell
	}

	for _, tt := range tests {
		err := client.ValidateCmd(distCmd, tt.args)
		require.Equal(t, tt.wantErr, err != nil, tt.reason)
	}
}
