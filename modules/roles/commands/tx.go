package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/roles"
)

// CreateRoleTxCmd is CLI command to create a new role
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

// createRoleTxCmd creates a basic role tx and then wraps, signs, and posts it
func createRoleTxCmd(cmd *cobra.Command, args []string) error {
	tx, err := readCreateRoleTxFlags()
	if err != nil {
		return err
	}
	return txcmd.DoTx(tx)
}

func readCreateRoleTxFlags() (tx sdk.Tx, err error) {
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
