package commands

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/client/commands"
	proofcmd "github.com/tendermint/basecoin/client/commands/proofs"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
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
	proof, err := proofcmd.GetAndParseAppProof(key, &res)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(res, proof.BlockHeight())
}
