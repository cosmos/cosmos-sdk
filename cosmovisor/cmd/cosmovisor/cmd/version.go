package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Cosmovisor and the application",
	Long:  `Prints the version number of Comovisor and the underlying application`,
	Run: func(cmd *cobra.Command, args []string) {

		// TODO - get actual version number
		fmt.Println("1.2.3.4")
	},
}
