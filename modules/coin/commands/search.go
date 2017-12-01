package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/search"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/coin"
)

// SentSearchCmd - command to find sendtx
var SentSearchCmd = &cobra.Command{
	Use:   "sent [address]",
	Short: "Find all tx with this address as sender or receiver",
	RunE:  commands.RequireInit(sentSearchCmd),
}

func sentSearchCmd(cmd *cobra.Command, args []string) error {
	addr, err := commands.GetOneArg(args, "address")
	if err != nil {
		return err
	}
	act, err := commands.ParseActor(addr)
	if err != nil {
		return err
	}

	findSender := fmt.Sprintf("coin.sender='%s'", act)
	findReceiver := fmt.Sprintf("coin.receiver='%s'", act)

	prove := !viper.GetBool(commands.FlagTrustNode)
	txs, err := search.FindTx(findSender, prove)
	if err != nil {
		return err
	}
	txr, err := search.FindTx(findReceiver, prove)
	if err != nil {
		return err
	}

	// combine both lists and sort by height descending
	all := append(txs, txr...)
	sort.Slice(all, func(i, j int) bool {
		return all[i].Height > all[j].Height
	})

	// format....
	output, err := formatResults(all)
	if err != nil {
		return err
	}

	// output as json
	blob, err := data.ToJSON(output)
	if err != nil {
		return err
	}
	fmt.Println(string(blob))
	return nil
}

type ResultOutput struct {
	Tx     *CoinTx `json:"tx"`
	Height int64   `json:"height"`
}

type CoinTx struct {
	Inputs  []CoinInput  `json:"inputs"`
	Outputs []CoinOutput `json:"outputs"`
}

type CoinInput struct {
	Sender string     `json:"sender"`
	Coins  coin.Coins `json:"coins"`
}

type CoinOutput struct {
	Receiver string     `json:"receiver"`
	Coins    coin.Coins `json:"coins"`
}

func formatResults(res []*ctypes.ResultTx) ([]ResultOutput, error) {
	out := make([]ResultOutput, 0, len(res))
	for _, r := range res {
		ctx, err := extractCoinTx(r.Tx)
		if err != nil {
			return nil, err
		}
		ro := ResultOutput{
			Height: int64(r.Height),
			Tx:     ctx,
		}

		out = append(out, ro)
	}
	return out, nil
}

func extractCoinTx(data []byte) (*CoinTx, error) {
	tx, err := sdk.LoadTx(data)
	if err != nil {
		return nil, err
	}
	txl, ok := tx.Unwrap().(sdk.TxLayer)
	for ok {
		tx = txl.Next()
		txl, ok = tx.Unwrap().(sdk.TxLayer)
	}
	ctx, ok := tx.Unwrap().(coin.SendTx)
	if !ok {
		return nil, errors.ErrUnknownTxType(tx)
	}
	// now reformat this....
	return &CoinTx{
		Inputs:  transformInputs(ctx.Inputs),
		Outputs: transformOutputs(ctx.Outputs),
	}, nil
}

func transformInputs(input []coin.TxInput) []CoinInput {
	out := make([]CoinInput, 0, len(input))
	for _, in := range input {
		out = append(out, CoinInput{
			Sender: in.Address.String(),
			Coins:  in.Coins,
		})
	}
	return out
}

func transformOutputs(output []coin.TxOutput) []CoinOutput {
	out := make([]CoinOutput, 0, len(output))
	for _, val := range output {
		out = append(out, CoinOutput{
			Receiver: val.Address.String(),
			Coins:    val.Coins,
		})
	}
	return out
}
