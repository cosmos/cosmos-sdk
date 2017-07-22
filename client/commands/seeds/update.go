package seeds

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/certifiers"

	"github.com/tendermint/basecoin/client/commands"
)

var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Update seed to current height if possible",
	RunE:         commands.RequireInit(updateSeed),
	SilenceUsage: true,
}

func init() {
	updateCmd.Flags().Int(heightFlag, 0, "Update to this height, not latest")
	RootCmd.AddCommand(updateCmd)
}

func updateSeed(cmd *cobra.Command, args []string) error {
	cert, err := commands.GetCertifier()
	if err != nil {
		return err
	}

	h := viper.GetInt(heightFlag)
	var seed certifiers.Seed
	if h <= 0 {
		// get the lastest from our source
		seed, err = certifiers.LatestSeed(cert.SeedSource)
	} else {
		seed, err = cert.SeedSource.GetByHeight(h)
	}
	if err != nil {
		return err
	}

	// let the certifier do it's magic to update....
	fmt.Printf("Trying to update to height: %d...\n", seed.Height())
	err = cert.Update(seed.Checkpoint, seed.Validators)
	if err != nil {
		return err
	}
	fmt.Println("Success!")
	return nil
}
