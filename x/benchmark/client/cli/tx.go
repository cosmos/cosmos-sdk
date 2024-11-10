package cli

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"

	"cosmossdk.io/x/benchmark"
	gen "cosmossdk.io/x/benchmark/generator"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	var accountNum int
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

			var txCount uint64
			defer func() {
				cmd.Printf("generated %d transactions\n", txCount)
			}()
			// accNum, accSeq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, clientCtx.FromAddress)
			// if err != nil {
			// 	return err
			// }
			// cmd.Printf("account number: %d, sequence: %d\n", accNum, accSeq)
			// txf, err := clienttx.NewFactoryCLI(clientCtx, cmd.Flags())
			// if err != nil {
			// 	return err
			// }
			// txf = txf.WithSequence(accSeq).
			// 	WithAccountNumber(accNum).
			// 	WithChainID(clientCtx.ChainID)

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
				Seed:        seed,
				KeyMean:     64,
				KeyStdDev:   8,
				ValueMean:   1024,
				ValueStdDev: 256,
				BucketCount: storeKeyCount,
			})

			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					op, ski := g.Next()
					op.Actor = storeKeys[ski]
					msg := &benchmark.MsgLoadTest{
						Caller: clientCtx.FromAddress,
						Ops:    []*benchmark.Op{op},
					}
					clienttx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
					// if err := clienttx.BroadcastTx(clientCtx, txf, msg); err != nil {
					// 	return err
					// }
					// accSeq++
					// txf = txf.WithSequence(accSeq)
				}
			}
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().IntVar(&accountNum, "account-num", 1, "number of accounts to use for load testing")

	return cmd
}

type LoadTestTxFactory struct {
	account  *keyring.Record
	sequence uint64
}

func (f *LoadTestTxFactory) Generate() (*benchmark.MsgLoadTest, error) {
	return nil, nil
}

func (f *LoadTestTxFactory) Broadcast() error {
	return nil
}

func NewLoadTestTxFactories(
	kr keyring.Keyring,
	accountRetriever func(sdk.AccAddress) (uint64, error),
	seed uint64,
	count uint64,
) ([]*LoadTestTxFactory, error) {
	factories := make([]*LoadTestTxFactory, count)
	records, err := kr.List()
	if err != nil {
		return nil, err
	}
	const maxTries = 10
	picked := make(map[int]struct{})
	r := rand.New(rand.NewPCG(seed, seed+1))
	var j int
	for i := uint64(0); i < count; i++ {
		j++
		if j >= maxTries {
			return nil, fmt.Errorf("failed to pick %d unique accounts out of %d after %d tries", count, len(records), maxTries)
		}
		ri := r.IntN(len(records))
		if _, ok := picked[ri]; ok {
			i--
			continue
		}
		picked[ri] = struct{}{}
		record := records[ri]
		addr, err := record.GetAddress()
		if err != nil {
			return nil, err
		}
		seq, err := accountRetriever(addr)
		if err != nil {
			return nil, err
		}

		factories[i] = &LoadTestTxFactory{
			account:  record,
			sequence: seq,
		}
	}

	return factories, nil
}
