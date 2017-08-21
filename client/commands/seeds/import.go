package seeds

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/certifiers"

	"github.com/cosmos/cosmos-sdk/client/commands"
)

const (
	dryFlag = "dry-run"
)

var importCmd = &cobra.Command{
	Use:          "import <file>",
	Short:        "Imports a new seed from the given file",
	Long:         `Validate this file and update to the given seed if secure.`,
	RunE:         commands.RequireInit(importSeed),
	SilenceUsage: true,
}

func init() {
	importCmd.Flags().Bool(dryFlag, false, "Test the import fully, but do not import")
	RootCmd.AddCommand(importCmd)
}

func importSeed(cmd *cobra.Command, args []string) error {
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
	seed, err := certifiers.LoadSeed(path)
	if err != nil {
		return err
	}

	// just do simple checks in --dry-run
	if viper.GetBool(dryFlag) {
		fmt.Printf("Testing seed %d/%X\n", seed.Height(), seed.Hash())
		err = seed.ValidateBasic(cert.ChainID())
	} else {
		fmt.Printf("Importing seed %d/%X\n", seed.Height(), seed.Hash())
		err = cert.Update(seed.Checkpoint, seed.Validators)
	}
	return err
}
