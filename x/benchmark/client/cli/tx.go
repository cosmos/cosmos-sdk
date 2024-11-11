package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	modulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/x/benchmark"
	gen "cosmossdk.io/x/benchmark/generator"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
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
	var verbose bool
	cmd := &cobra.Command{
		Use: "load-test",
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

			var (
				successCount uint64
				errCount     uint64
			)
			defer func() {
				cmd.Printf("done! success_tx=%d err_tx=%d\n", successCount, errCount)
			}()
			accNum, accSeq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, clientCtx.FromAddress)
			if err != nil {
				return err
			}
			txf, err := clienttx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithAccountNumber(accNum).WithChainID(clientCtx.ChainID)

			// TODO: fetch or share state from genesis
			seed := uint64(34)
			storeKeyCount := uint64(10)
			storeKeys, err := gen.StoreKeys("benchmark", seed, storeKeyCount)
			if err != nil {
				return err
			}
			for _, c := range clientCtx.FromAddress {
				seed += uint64(c)
			}
			g := gen.NewGenerator(gen.Options{
				GeneratorParams: &modulev1.GeneratorParams{
					Seed:         seed,
					KeyMean:      64,
					KeyStdDev:    8,
					ValueMean:    1024,
					ValueStdDev:  256,
					BucketCount:  storeKeyCount,
					GenesisCount: 500_000,
				},
				InsertWeight: 0.25,
				DeleteWeight: 0.05,
				UpdateWeight: 0.50,
				GetWeight:    0.20,
			})
			g.Load()

			i := 0
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					if i != 0 && i%1000 == 0 {
						cmd.Printf("success_tx=%d err_tx=%d seq=%d\n", successCount, errCount, accSeq)
					}
					bucket, op, err := g.Next()
					if err != nil {
						return err
					}
					op.Actor = storeKeys[bucket]
					msg := &benchmark.MsgLoadTest{
						Caller: clientCtx.FromAddress,
						Ops:    []*benchmark.Op{op},
					}
					txf = txf.WithSequence(accSeq)
					tx, err := txf.BuildUnsignedTx(msg)
					if err != nil {
						return err
					}
					err = clienttx.Sign(clientCtx, txf, clientCtx.From, tx, true)
					if err != nil {
						return err
					}
					txBytes, err := clientCtx.TxConfig.TxEncoder()(tx.GetTx())
					if err != nil {
						return err
					}
					res, err := clientCtx.BroadcastTxAsync(txBytes)
					if err != nil {
						return err
					}
					if res.Code != 0 {
						if verbose {
							clientCtx.PrintProto(res)
						}
						errCount++
					} else {
						accSeq++
						successCount++
					}
					i++
				}
			}
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().BoolVar(&verbose, "verbose", false, "print the response")

	return cmd
}
