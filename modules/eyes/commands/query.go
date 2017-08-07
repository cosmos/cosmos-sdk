package commands

import (
	"encoding/hex"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/query"
	"github.com/tendermint/basecoin/modules/eyes"
	"github.com/tendermint/basecoin/stack"
)

// EtcQueryCmd - command to query raw data
var EtcQueryCmd = &cobra.Command{
	Use:   "etc [key]",
	Short: "Get data stored under key in etc",
	RunE:  commands.RequireInit(etcQueryCmd),
}

func etcQueryCmd(cmd *cobra.Command, args []string) error {
	var res etc.Data

	arg, err := commands.GetOneArg(args, "key")
	if err != nil {
		return err
	}
	key, err := hex.DecodeString(cmn.StripHex(arg))
	if err != nil {
		return err
	}

	key = stack.PrefixedKey(etc.Name, key)
	prove := !viper.GetBool(commands.FlagTrustNode)
	height, err := query.GetParsed(key, &res, prove)
	if err != nil {
		return err
	}

	return query.OutputProof(res, height)
}
