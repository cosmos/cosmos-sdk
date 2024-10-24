package main

import (
	"errors"
	"fmt"
	"os"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/simapp/v2/simdv2/cmd"
)

func main() {
	// reproduce default cobra behavior so that eager parsing of flags is possible.
	// see: https://github.com/spf13/cobra/blob/e94f6d0dd9a5e5738dca6bce03c4b1207ffbc0ec/command.go#L1082
	args := os.Args[1:]
	rootCmd, err := cmd.NewRootCmd[transaction.Tx](args...)
	if err != nil {
		if _, pErr := fmt.Fprintln(os.Stderr, err); pErr != nil {
			panic(errors.Join(err, pErr))
		}
		os.Exit(1)
	}
	if err = rootCmd.Execute(); err != nil {
		if _, pErr := fmt.Fprintln(rootCmd.OutOrStderr(), err); pErr != nil {
			panic(errors.Join(err, pErr))
		}
		os.Exit(1)
	}
}
