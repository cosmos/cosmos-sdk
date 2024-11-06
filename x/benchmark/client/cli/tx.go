package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"cosmossdk.io/x/benchmark"
	gen "cosmossdk.io/x/benchmark/generator"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "benchmark",
		Short:                      "benchmark transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewLoadTestCmd(),
	)

	return txCmd
}

func NewLoadTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "load-test [from_key_or_address]",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ctx, cancelFn := context.WithCancel(cmd.Context())
			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
				select {
				case sig := <-sigCh:
					cancelFn()
					cmd.Printf("caught %s signal\n", sig.String())
				case <-ctx.Done():
					cancelFn()
				}
			}()

			// TODO: fetch or share state from genesis
			seed := uint64(34)
			storeKeyCount := uint64(10)
			storeKeys, err := gen.StoreKeys("benchmark", seed, storeKeyCount)
			if err != nil {
				return err
			}

			g := gen.NewGenerator(gen.Options{
				Seed:        34,
				KeyMean:     64,
				KeyStdDev:   8,
				ValueMean:   1024,
				ValueStdDev: 256,
				BucketCount: storeKeyCount,
			})
			var txCount uint64
			defer func() {
				cmd.Printf("generated %d transactions\n", txCount)
			}()
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
				}
				op, ski := g.Next()
				op.Actor = storeKeys[ski]
				msg := &benchmark.MsgLoadTest{
					Caller: clientCtx.FromAddress,
					Ops:    []*benchmark.Op{op},
				}
				err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
				if err != nil {
					return err
				}
				txCount++
			}
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
