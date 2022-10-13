package cli

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
)

type HasAutoCLIConfig interface {
	appmodule.AppModule

	AutoCLIOptions() *autocliv1.ModuleOptions
}

type HasCustomQueryCommand interface {
	appmodule.AppModule

	GetQueryCmd() *cobra.Command
}

type HasCustomTxCommand interface {
	appmodule.AppModule

	GetTxCmd() *cobra.Command
}
