package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	lc "github.com/tendermint/light-client"

	"github.com/tendermint/basecoin/client/commands"
	proofcmd "github.com/tendermint/basecoin/client/commands/proofs"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/stack"
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
	height, err := proofcmd.GetParsed(key, &acc, prove)
	if lc.IsNoDataErr(err) {
		return errors.Errorf("Account bytes are empty for address %s ", addr)
	} else if err != nil {
		return err
	}

	return proofcmd.OutputProof(acc, height)
}
