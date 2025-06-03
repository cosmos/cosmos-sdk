package main

import (
	"os"

	"github.com/tajeor/chain/cmd/tajeord/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
