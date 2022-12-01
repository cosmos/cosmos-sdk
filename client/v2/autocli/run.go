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

func Run(appOptions AppOptions) error {
	cmd, err := RootCmd(appOptions)
	if err != nil {
		return err
	}

	return cmd.Execute()
}

func RunFromAppConfig(appConfig depinject.Config) error {
	var appOptions AppOptions
	err := depinject.Inject(appConfig, &appOptions)
	if err != nil {
		return err
	}

	return Run(appOptions)
}

func RootCmd(appOptions AppOptions) (*cobra.Command, error) {
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

	queryCmd, err := builder.BuildQueryCommand(moduleOptions, customQueryCmds)
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(queryCmd)
	return rootCmd, nil
}
