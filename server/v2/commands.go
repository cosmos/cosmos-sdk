package serverv2

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

type NewCometBFTServerFunc func(*viper.Viper, log.Logger, transaction.Codec[transaction.Tx]) ServerModule


func Commands(rootCmd *cobra.Command, codec transaction.Codec[transaction.Tx], logger log.Logger, homePath string, newCometFunc NewCometBFTServerFunc, modules ...ServerModule,) (CLIConfig, error) {
	// if len(modules) == 0 {
	// 	// TODO figure if we should define default modules
	// 	// and if so it should be done here to avoid uncessary dependencies
	// 	return CLIConfig{}, errors.New("no modules provided")
	// }

	


	server := NewServer(logger, modules...)
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Run the application",
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := ReadConfig(filepath.Join(homePath, "config"))
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}
			if err := v.BindPFlags(cmd.Flags()); err != nil { // the server modules are already instantiated here, so binding the flags is useless.
				return err
			}

			// Init CometBFTServer when server start
			cometServer := newCometFunc(v, logger, codec)
			server.modules = append(server.modules, cometServer)

			srvConfig := Config{StartBlock: true}
			ctx := cmd.Context()
			ctx = context.WithValue(ctx, ServerContextKey, srvConfig)
			SetCmdServerContext(cmd, v, logger)
			ctx, cancelFn := context.WithCancel(ctx)
			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
				sig := <-sigCh
				cancelFn()
				cmd.Printf("caught %s signal\n", sig.String())

				if err := server.Stop(ctx); err != nil {
					cmd.PrintErrln("failed to stop servers:", err)
				}
			}()

			if err := server.Start(ctx); err != nil {
				return fmt.Errorf("failed to start servers: %w", err)
			}

			return nil
		},
	}

	cmds := server.CLICommands()
	cmds.Commands = append(cmds.Commands, startCmd)

	return cmds, nil
}

func AddCommands(rootCmd *cobra.Command, codec transaction.Codec[transaction.Tx], logger log.Logger, homePath string, newCometFunc NewCometBFTServerFunc, modules ...ServerModule) error {
	cmds, err := Commands(rootCmd, codec, logger, homePath, newCometFunc, modules...)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cmds.Commands...)
	return nil
}
