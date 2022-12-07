package autocli

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type AppOptions struct {
	depinject.In

	Modules       map[string]appmodule.AppModule
	ModuleOptions map[string]*autocliv1.ModuleOptions `optional:"true"`
}

func (appOptions AppOptions) Run() error {
	cmd, err := appOptions.RootCmd()
	if err != nil {
		return err
	}

	return cmd.Execute()
}

func (appOptions AppOptions) RootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{}
	err := appOptions.EnhanceRootCommand(rootCmd)
	return rootCmd, err
}

func (appOptions AppOptions) EnhanceRootCommand(rootCmd *cobra.Command) error {
	builder := &Builder{
		GetClientConn: func(cmd *cobra.Command) (grpc.ClientConnInterface, error) {
			return client.GetClientQueryContext(cmd)
		},
		AddQueryConnFlags: flags.AddQueryFlagsToCmd,
	}

	moduleOptions := appOptions.ModuleOptions
	if moduleOptions == nil {
		moduleOptions = map[string]*autocliv1.ModuleOptions{}

		for name, module := range appOptions.Modules {
			if module, ok := module.(HasAutoCLIConfig); ok {
				moduleOptions[name] = module.AutoCLIOptions()
			}
		}
	}

	customQueryCmds := map[string]*cobra.Command{}
	for name, module := range appOptions.Modules {
		if module, ok := module.(HasCustomQueryCommand); ok {
			cmd := module.GetQueryCmd()
			// filter any nil commands
			if cmd != nil {
				customQueryCmds[name] = cmd
			}
		}
	}

	// if we have an existing query command, enhance it or build a custom one
	if queryCmd := findSubCommand(rootCmd, "query"); queryCmd != nil {
		err := builder.EnhanceQueryCommand(queryCmd, moduleOptions, customQueryCmds)
		if err != nil {
			return err
		}
	} else {
		queryCmd, err := builder.BuildQueryCommand(moduleOptions, customQueryCmds)
		if err != nil {
			return err
		}

		rootCmd.AddCommand(queryCmd)
	}

	return nil
}
