package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

type DefaultHome string

func Run(options ...container.Option) {
	options = append(options, Provider)
	err := container.Run(DefaultRunner, options...)
	if err != nil {
		panic(err)
	}
}

func DefaultRunner(rootCmd *cobra.Command, home DefaultHome) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	defaultHome := filepath.Join(userHomeDir, string(home))

	if err := svrcmd.Execute(rootCmd, defaultHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}
