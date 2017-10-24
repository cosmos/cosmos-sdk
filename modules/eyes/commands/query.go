package commands

import (
	"encoding/hex"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/modules/eyes"
	"github.com/cosmos/cosmos-sdk/util"
)

// EyesQueryCmd - command to query raw data
var EyesQueryCmd = &cobra.Command{
	Use:   "eyes [key]",
	Short: "Get data stored under key in eyes",
	RunE:  commands.RequireInit(eyesQueryCmd),
}

func eyesQueryCmd(cmd *cobra.Command, args []string) error {
	var res eyes.Data

	arg, err := commands.GetOneArg(args, "key")
	if err != nil {
		return err
	}
	key, err := hex.DecodeString(cmn.StripHex(arg))
	if err != nil {
		return err
	}

	key = util.PrefixedKey(eyes.Name, key)
	prove := !viper.GetBool(commands.FlagTrustNode)
	height, err := query.GetParsed(key, &res, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(res, height)
}
