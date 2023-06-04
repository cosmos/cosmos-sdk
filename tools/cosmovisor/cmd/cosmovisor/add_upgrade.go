package main

import (
	"fmt"
	"os"

	"cosmossdk.io/log"
	"cosmossdk.io/tools/cosmovisor"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var addUpgradeCmd = &cobra.Command{
	Use:          "add-upgrade [upgrade-name] [path to executable]",
	Short:        "Manually add upgrade binary to Cosmovisor",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(2),
	RunE:         AddUpgrade,
}

// AddUpgrade adds upgrade info to manifest
func AddUpgrade(cmd *cobra.Command, args []string) error {
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}

	logger := cmd.Context().Value(log.ContextKey).(log.Logger)
	if cfg.DisableLogs {
		logger = log.NewCustomLogger(zerolog.Nop())
	}

	upgradeName := args[0]
	if len(upgradeName) == 0 {
		return fmt.Errorf("upgrade name cannot be empty")
	}

	executablePath := args[1]
	if _, err := os.Stat(executablePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("invalid executable path: %w", err)
		}

		return fmt.Errorf("failed to load executable path: %w", err)
	}

	// create upgrade dir
	upgradeLocation := cfg.UpgradeDir(upgradeName)
	if err := os.MkdirAll(upgradeLocation, 0o644); err != nil {
		return fmt.Errorf("failed to create upgrade directory: %w", err)
	}

	// copy binary to upgrade dir
	executableData, err := os.ReadFile(executablePath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}

	if err := os.WriteFile(upgradeLocation, executableData, 0o600); err != nil {
		return fmt.Errorf("failed to write binary to location: %w", err)
	}

	logger.Info(fmt.Sprintf("Using %s for %s upgrade (copied to %s)", executablePath, upgradeName, upgradeLocation))

	return nil
}
