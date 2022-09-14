package runtime

import (
	"cosmossdk.io/depinject"
	"github.com/spf13/cobra"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

// CLIConfig is
type CLIConfig struct {
	// AutoCLIOptions
	AutoCLIOptions     *autocliv1.ModuleOptions
	CustomQueryCommand *cobra.Command
	CustomTxCommand    *cobra.Command
}

func (C CLIConfig) IsOnePerModuleType() {}

var _ depinject.OnePerModuleType = CLIConfig{}
