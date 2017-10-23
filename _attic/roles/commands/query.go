package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/cosmos/cosmos-sdk/stack"
)

// RoleQueryCmd - command to query a role
var RoleQueryCmd = &cobra.Command{
	Use:   "role [name]",
	Short: "Get details of a role, with proof",
	RunE:  commands.RequireInit(roleQueryCmd),
}

func roleQueryCmd(cmd *cobra.Command, args []string) error {
	arg, err := commands.GetOneArg(args, "name")
	if err != nil {
		return err
	}
	role, err := parseRole(arg)
	if err != nil {
		return err
	}

	var res roles.Role
	key := stack.PrefixedKey(roles.NameRole, role)
	prove := !viper.GetBool(commands.FlagTrustNode)
	height, err := query.GetParsed(key, &res, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(res, height)
}
