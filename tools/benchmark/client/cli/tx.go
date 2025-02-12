package cli

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	modulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/tools/benchmark"
	gen "cosmossdk.io/tools/benchmark/generator"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
)

func NewTxCmd(params *modulev1.GeneratorParams) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "benchmark",
		Short:                      "benchmark transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewLoadTestCmd(params),
	)

	return txCmd
}

func NewLoadTestCmd(params *modulev1.GeneratorParams) *cobra.Command {
	var (
		verbose bool
		pause   int64
		numOps  uint64
	)
	cmd := &cobra.Command{
		Use: "load-test",
		RunE: func(cmd *cobra.Command, args []string) (runErr error) {
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
				successCount int
				errCount     int
				since        = time.Now()
				last         int
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
			txf = txf.WithAccountNumber(accNum).WithChainID(clientCtx.ChainID).WithGas(1_000_000_000)

			storeKeys, err := gen.StoreKeys("benchmark", params.Seed, params.BucketCount)
			if err != nil {
				return err
			}
			var seed uint64
			for _, c := range clientCtx.FromAddress {
				// root the generator seed in the account address
				seed += uint64(c)
			}
			g := gen.NewGenerator(gen.Options{
				HomeDir:         clientCtx.HomeDir,
				GeneratorParams: params,
				InsertWeight:    0.25,
				DeleteWeight:    0.05,
				UpdateWeight:    0.50,
				GetWeight:       0.20,
			},
				gen.WithGenesis(),
				gen.WithSeed(seed),
			)
			if err = g.Load(); err != nil {
				return err
			}
			defer func() {
				if err = g.Close(); err != nil {
					runErr = errors.Join(runErr, err)
				}
			}()

			begin := time.Now()
			ops := make([]*benchmark.Op, numOps)
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
				}
				if time.Since(since) > 5*time.Second {
					cmd.Printf(
						"success_tx=%d err_tx=%d seq=%d rate=%.2f/s overall=%.2f/s\n",
						successCount, errCount, accSeq,
						float64(successCount-last)/time.Since(since).Seconds(),
						float64(successCount)/time.Since(begin).Seconds(),
					)
					since = time.Now()
					last = successCount
				}

				for j := range numOps {
					bucket, op, err := g.Next()
					if err != nil {
						return err
					}
					op.Actor = storeKeys[bucket]
					ops[j] = op
				}
				msg := &benchmark.MsgLoadTest{
					Caller: clientCtx.FromAddress,
					Ops:    ops,
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
						err = clientCtx.PrintProto(res)
						if err != nil {
							return err
						}
					}
					errCount++
				} else {
					accSeq++
					successCount++
				}
				if pause > 0 {
					time.Sleep(time.Duration(pause) * time.Microsecond)
				}
			}
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "print the response")
	cmd.Flags().Uint64Var(&numOps, "ops", 1, "number of operations per transaction")
	cmd.Flags().Int64Var(&pause, "pause", 0, "pause between transactions in microseconds")

	return cmd
}
