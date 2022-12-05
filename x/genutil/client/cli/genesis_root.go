package cli

import (
	"cosmossdk.io/simapp"
	"cosmossdk.io/simapp/params"
	"github.com/cosmos/cosmos-sdk/client"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
)

func GenesisRootCmd(encodingConfig params.EncodingConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "genesis",
		Short:                      "Genesis subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	gentxModule := simapp.ModuleBasics[genutiltypes.ModuleName].(genutil.AppModuleBasic)

	cmd.AddCommand(
		GenTxCmd(simapp.ModuleBasics, encodingConfig.TxConfig,
			banktypes.GenesisBalancesIterator{}, simapp.DefaultNodeHome),
		MigrateGenesisCmd(),
		CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, simapp.DefaultNodeHome,
			gentxModule.GenTxValidator),
		ValidateGenesisCmd(simapp.ModuleBasics),
		AddGenesisAccountCmd(simapp.DefaultNodeHome),
	)

	return cmd
}
