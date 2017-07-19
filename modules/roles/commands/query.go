package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	lcmd "github.com/tendermint/basecoin/client/commands"
	proofcmd "github.com/tendermint/basecoin/client/commands/proofs"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
)

// RoleQueryCmd - command to query a role
var RoleQueryCmd = &cobra.Command{
	Use:   "role [name]",
	Short: "Get details of a role, with proof",
	RunE:  lcmd.RequireInit(roleQueryCmd),
}

func roleQueryCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("Missing required argument [name]")
	} else if len(args) > 1 {
		return errors.New("Command only supports one name")
	}

	role, err := parseRole(args[0])
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
