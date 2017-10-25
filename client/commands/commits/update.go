package commits

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/certifiers"

	"github.com/cosmos/cosmos-sdk/client/commands"
)

var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Update commit to current height if possible",
	RunE:         commands.RequireInit(updateCommit),
	SilenceUsage: true,
}

func init() {
	updateCmd.Flags().Int(heightFlag, 0, "Update to this height, not latest")
	RootCmd.AddCommand(updateCmd)
}

func updateCommit(cmd *cobra.Command, args []string) error {
	cert, err := commands.GetCertifier()
	if err != nil {
		return err
	}

	h := viper.GetInt(heightFlag)
	var fc certifiers.FullCommit
	if h <= 0 {
		// get the lastest from our source
		fc, err = cert.Source.LatestCommit()
	} else {
		fc, err = cert.Source.GetByHeight(h)
	}
	if err != nil {
		return err
	}

	// let the certifier do it's magic to update....
	fmt.Printf("Trying to update to height: %d...\n", fc.Height())
	err = cert.Update(fc)
	if err != nil {
		return err
	}
	fmt.Println("Success!")
	return nil
}
