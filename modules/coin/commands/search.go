package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/search"
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
	all, err := search.FindAnyTx(prove, findSender, findReceiver)
	if err != nil {
		return err
	}

	// format....
	output, err := search.FormatSearch(all, coin.ExtractCoinTx)
	if err != nil {
		return err
	}

	return search.Output(output)
}
