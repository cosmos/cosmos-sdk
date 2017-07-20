package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	"github.com/tendermint/basecoin/modules/roles"
)

// CreateRoleTxCmd is CLI command to send tokens between basecoin accounts
var CreateRoleTxCmd = &cobra.Command{
	Use:   "create-role",
	Short: "Create a new role",
	RunE:  commands.RequireInit(createRoleTxCmd),
}

//nolint
const (
	FlagRole    = "role"
	FlagMembers = "members"
	FlagMinSigs = "min-sigs"
)

func init() {
	flags := CreateRoleTxCmd.Flags()
	flags.String(FlagRole, "", "Name of the role to create")
	flags.String(FlagMembers, "", "Set of comma-separated addresses for this role")
	flags.Int(FlagMinSigs, 0, "Minimum number of signatures needed to assume this role")
}

// createRoleTxCmd is an example of how to make a tx
func createRoleTxCmd(cmd *cobra.Command, args []string) error {
	tx, err := readCreateRoleTxFlags()
	if err != nil {
		return err
	}
	return txcmd.DoTx(tx)
}

func readCreateRoleTxFlags() (tx basecoin.Tx, err error) {
	role, err := parseRole(viper.GetString(FlagRole))
	if err != nil {
		return tx, err
	}

	sigs := viper.GetInt(FlagMinSigs)
	if sigs < 1 {
		return tx, errors.Errorf("--%s must be at least 1", FlagMinSigs)
	}

	signers, err := commands.ParseActors(viper.GetString(FlagMembers))
	if err != nil {
		return tx, err
	}
	if len(signers) == 0 {
		return tx, errors.New("must specify at least one member")
	}

	tx = roles.NewCreateRoleTx(role, uint32(sigs), signers)
	return tx, nil
}
