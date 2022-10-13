package cli

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"
)

type AppCommandOptions struct {
	Name            string
	ModuleOptions   map[string]*autocliv1.ModuleOptions
	CustomQueryCmds map[string]*cobra.Command
	CustomTxCmds    map[string]*cobra.Command
}

func (b *Builder) BuildAppCommand(options AppCommandOptions) (*cobra.Command, error) {
	cmd := topLevelCmd(options.Name, fmt.Sprintf("%s app commands", options.Name))
	err := b.AddAppCommands(cmd, options)
	return cmd, err
}

func (b *Builder) AddAppCommands(cmd *cobra.Command, options AppCommandOptions) error {
	queryCmd, err := b.BuildQueryCommand(options.ModuleOptions, options.CustomQueryCmds)
	if err != nil {
		return err
	}
	cmd.AddCommand(queryCmd)
	return nil
}
