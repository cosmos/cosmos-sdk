package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/x/coin"
	"github.com/cosmos/cosmos-sdk/stack"
)

// AccountQueryCmd - command to query an account
var AccountQueryCmd = &cobra.Command{
	Use:   "account [address]",
	Short: "Get details of an account, with proof",
	RunE:  commands.RequireInit(accountQueryCmd),
}

func accountQueryCmd(cmd *cobra.Command, args []string) error {
	addr, err := commands.GetOneArg(args, "address")
	if err != nil {
		return err
	}
	act, err := commands.ParseActor(addr)
	if err != nil {
		return err
	}
	act = coin.ChainAddr(act)
	key := stack.PrefixedKey(coin.NameCoin, act.Bytes())

	acc := coin.Account{}
	prove := !viper.GetBool(commands.FlagTrustNode)
	height, err := query.GetParsed(key, &acc, query.GetHeight(), prove)
	if client.IsNoDataErr(err) {
		return errors.Errorf("Account bytes are empty for address %s ", addr)
	} else if err != nil {
		return err
	}

	return query.OutputProof(acc, height)
}
