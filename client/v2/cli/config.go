package cli

import (
	"cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/depinject"
	"github.com/spf13/cobra"
)

// ModuleConfig is
type ModuleConfig struct {
	AutoCLIOptions     *autocliv1.ModuleOptions
	CustomQueryCommand *cobra.Command
	CustomTxCommand    *cobra.Command
}

func (C ModuleConfig) IsOnePerModuleType() {}

var _ depinject.OnePerModuleType = ModuleConfig{}
