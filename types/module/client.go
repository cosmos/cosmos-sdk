package module

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/spf13/cobra"
)

type ClientModule interface {
	NewTxCmd(ctx context.CLIContext) *cobra.Command
	NewQueryCmd(ctx context.CLIContext) *cobra.Command
}
