package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version string

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version number of Cosmovisor and the application",
	Long:  `Prints the version number of Comovisor and the underlying application`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Cosmovisor Version: ", Version)
	},
}
