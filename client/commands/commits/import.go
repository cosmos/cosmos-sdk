package commits

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/certifiers/files"

	"github.com/cosmos/cosmos-sdk/client/commands"
)

const (
	dryFlag = "dry-run"
)

var importCmd = &cobra.Command{
	Use:          "import <file>",
	Short:        "Imports a new commit from the given file",
	Long:         `Validate this file and update to the given commit if secure.`,
	RunE:         commands.RequireInit(importCommit),
	SilenceUsage: true,
}

func init() {
	importCmd.Flags().Bool(dryFlag, false, "Test the import fully, but do not import")
	RootCmd.AddCommand(importCmd)
}

func importCommit(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide an input file")
	}

	// prepare the certifier
	cert, err := commands.GetCertifier()
	if err != nil {
		return err
	}

	// parse the input file
	path := args[0]
	fc, err := files.LoadFullCommitJSON(path)
	if err != nil {
		return err
	}

	// just do simple checks in --dry-run
	if viper.GetBool(dryFlag) {
		fmt.Printf("Testing commit %d/%X\n", fc.Height(), fc.ValidatorsHash())
		err = fc.ValidateBasic(cert.ChainID())
	} else {
		fmt.Printf("Importing commit %d/%X\n", fc.Height(), fc.ValidatorsHash())
		err = cert.Update(fc)
	}
	return err
}
