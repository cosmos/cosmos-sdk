package seeds

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tendermint/light-client/certifiers"

	"github.com/tendermint/basecoin/client/commands"
)

var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Update seed to current chain state if possible",
	RunE:         commands.RequireInit(updateSeed),
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(updateCmd)
}

func updateSeed(cmd *cobra.Command, args []string) error {
	cert, err := commands.GetCertifier()
	if err != nil {
		return err
	}

	// get the lastest from our source
	seed, err := certifiers.LatestSeed(cert.SeedSource)
	if err != nil {
		return err
	}
	fmt.Printf("Trying to update to height: %d...\n", seed.Height())

	// let the certifier do it's magic to update....
	err = cert.Update(seed.Checkpoint, seed.Validators)
	if err != nil {
		return err
	}
	fmt.Println("Success!")
	return nil
}
