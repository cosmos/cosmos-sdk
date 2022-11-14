package autocli

import (
	"context"
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type AppConfig struct {
	depinject.In

	Modules       map[string]appmodule.AppModule
	ModuleOptions map[string]*autocliv1.ModuleOptions `optional:"true"`
}

func Run(cfg AppConfig) error {
	cmd, err := RootCmd(cfg)
	if err != nil {
		return err
	}

	return cmd.Execute()
}

type contextKey string

const grpcEndpointContextKey = contextKey("grpc")

func RootCmd(cfg AppConfig) (*cobra.Command, error) {
	builder := &Builder{
		GetClientConn: func(ctx context.Context) (grpc.ClientConnInterface, error) {
			grpcEndpoint, ok := ctx.Value(grpcEndpointContextKey).(string)
			if !ok || grpcEndpoint == "" {
				return nil, fmt.Errorf("no gRPC endpoint configured")
			}

			return grpc.Dial(grpcEndpoint)
		},
	}

	moduleOptions := cfg.ModuleOptions
	if moduleOptions == nil {
		moduleOptions = map[string]*autocliv1.ModuleOptions{}

		for name, module := range cfg.Modules {
			if module, ok := module.(HasAutoCLIConfig); ok {
				moduleOptions[name] = module.AutoCLIOptions()
			}
		}
	}

	customQueryCmds := map[string]*cobra.Command{}
	for name, module := range cfg.Modules {
		if module, ok := module.(HasCustomQueryCommand); ok {
			customQueryCmds[name] = module.GetQueryCmd()
		}
	}

	queryCmd, err := builder.BuildQueryCommand(moduleOptions, customQueryCmds)
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{}
	var grpcEndpoint string
	rootCmd.PersistentFlags().StringVar(&grpcEndpoint, "grpc", "", "the gRPC endpoint of the node to connect with")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.SetContext(context.WithValue(cmd.Context(), grpcEndpointContextKey, grpcEndpointContextKey))
		return nil
	}
	rootCmd.AddCommand(queryCmd)
	return rootCmd, nil
}
