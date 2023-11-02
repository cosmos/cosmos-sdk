package internal

import (
	"github.com/spf13/cobra"

	"cosmossdk.io/tools/hubl/internal/config"
)

func RootCommand() (*cobra.Command, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(configDir)
	if err != nil {
		return nil, err
	}

	cmd := &cobra.Command{
		Use:   "hubl",
		Short: "Hubl is a CLI for interacting with Cosmos SDK chains",
	}

	// add commands
	commands, err := RemoteCommand(cfg, configDir)
	if err != nil {
		return nil, err
	}
	commands = append(
		commands,
		InitCmd(cfg, configDir),
		KeyringCmd(""),
		VersionCmd(),
	)

	cmd.AddCommand(commands...)
	return cmd, nil
}
