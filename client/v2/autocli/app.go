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

// AppOptions are autocli options for an app. These options can be built via depinject based on an app config. Ex:
// Ex:
//
//	var autoCliOpts autocli.AppOptions
//	err := depinject.Inject(appConfig, &encodingConfig.InterfaceRegistry, &autoCliOpts)
//
// If depinject isn't used, options can be provided manually or extracted from modules. One method for extracting autocli
// options is via the github.com/cosmos/cosmos-sdk/runtime/services.ExtractAutoCLIOptions function.
type AppOptions struct {
	depinject.In

	// Modules are the AppModule implementations for the modules in the app.
	Modules map[string]appmodule.AppModule

	// ModuleOptions are autocli options to be used for modules instead of what
	// is specified on the module's AppModule implementation. This allows an
	// app to override module options if they are either not provided by a
	// module or need to be improved.
	ModuleOptions map[string]*autocliv1.ModuleOptions `optional:"true"`
}

// RootCmd generates a root command for an app based on the AppOptions. This
// command currently only includes query commands but will be enhanced over
// time to cover the full scope of an app CLI.
func (appOptions AppOptions) RootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{}
	err := appOptions.EnhanceRootCommand(rootCmd)
	return rootCmd, err
}

// EnhanceRootCommand enhances the provided root command with autocli AppOptions,
// only adding missing query commands and doesn't override commands already
// in the root command. This allows for the graceful integration of autocli with
// existing app CLI commands where autocli simply automatically adds things that
// weren't manually provided. It does take into account custom query commands
// provided by modules with the HasCustomQueryCommand extension interface.
// Example Usage:
//
//	var autoCliOpts autocli.AppOptions
//	err := depinject.Inject(appConfig, &autoCliOpts)
//	if err != nil {
//		panic(err)
//	}
//	rootCmd := initRootCmd()
//	err = autoCliOpts.EnhanceRootCommand(rootCmd)
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
		if err := builder.EnhanceQueryCommand(queryCmd, moduleOptions, customQueryCmds); err != nil {
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
