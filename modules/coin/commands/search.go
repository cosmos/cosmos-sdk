package commands

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-wire/data"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/search"
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

	// output as json
	blob, err := data.ToJSON(all)
	if err != nil {
		return err
	}
	fmt.Println(string(blob))
	return nil
}
