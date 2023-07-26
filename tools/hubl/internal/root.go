package internal

import (
	"os"
	"path"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/hubl/internal/config"
	"cosmossdk.io/tools/hubl/internal/keyring"
)

func RootCommand() (*cobra.Command, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := path.Join(homeDir, config.DefaultConfigDirName)
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
		keyring.Cmd(),
		VersionCmd(),
	)

	cmd.AddCommand(commands...)
	return cmd, nil
}
