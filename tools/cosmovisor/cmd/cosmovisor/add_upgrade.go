package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"cosmossdk.io/tools/cosmovisor/v2"
)

func NewAddUpgradeCmd() *cobra.Command {
	addUpgrade := &cobra.Command{
		Use:          "add-upgrade <upgrade-name> <path to executable>",
		Short:        "Add APP upgrade binary to cosmovisor",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(2),
		RunE:         addUpgradeCmd,
	}

	addUpgrade.Flags().Bool(cosmovisor.FlagForce, false, "overwrite existing upgrade binary and plan with the same name")
	addUpgrade.Flags().Int64(cosmovisor.FlagUpgradeHeight, 0, "define a height at which to upgrade the binary automatically (without governance proposal)")

	return addUpgrade
}

// addUpgrade adds upgrade info to manifest
func addUpgrade(cfg *cosmovisor.Config, force bool, upgradeHeight int64, upgradeName, executablePath string) (*upgradetypes.Plan, error) {
	logger := cfg.Logger(os.Stdout)

	if !cfg.DisableRecase {
		upgradeName = strings.ToLower(upgradeName)
	}

	if _, err := os.Stat(executablePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("invalid executable path: %w", err)
		}

		return nil, fmt.Errorf("failed to load executable path: %w", err)
	}

	// create upgrade dir
	upgradeLocation := cfg.UpgradeDir(upgradeName)
	if err := os.MkdirAll(path.Join(upgradeLocation, "bin"), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create upgrade directory: %w", err)
	}

	// copy binary to upgrade dir
	executableData, err := os.ReadFile(executablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read binary: %w", err)
	}

	if err := saveOrAbort(cfg.UpgradeBin(upgradeName), executableData, force); err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Using %s for %s upgrade", executablePath, upgradeName))
	logger.Info(fmt.Sprintf("Upgrade binary located at %s", cfg.UpgradeBin(upgradeName)))

	var plan *upgradetypes.Plan
	if upgradeHeight > 0 {
		plan = &upgradetypes.Plan{
			Name:   upgradeName,
			Height: upgradeHeight,
		}
	}

	return plan, nil
}

// addUpgradeCmd parses input flags and adds upgrade info to manifest
func addUpgradeCmd(cmd *cobra.Command, args []string) error {
	cfg, err := getConfigFromCmd(cmd)
	if err != nil {
		return err
	}

	upgradeName, executablePath := args[0], args[1]

	force, err := cmd.Flags().GetBool(cosmovisor.FlagForce)
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
	}

	upgradeHeight, err := cmd.Flags().GetInt64(cosmovisor.FlagUpgradeHeight)
	if err != nil {
		return fmt.Errorf("failed to get upgrade-height flag: %w", err)
	}

	plan, err := addUpgrade(cfg, force, upgradeHeight, upgradeName, executablePath)
	if err != nil {
		return err
	}
	if plan == nil {
		return nil // No plan to add
	}
	return cfg.AddManualUpgrades(force, plan)
}

// saveOrAbort saves data to path or aborts if file exists and force is false
func saveOrAbort(path string, data []byte, force bool) error {
	if _, err := os.Stat(path); err == nil {
		if !force {
			return fmt.Errorf("file already exists at %s", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	//nolint:gosec // We need broader permissions to make it executable
	if err := os.WriteFile(path, data, 0o755); err != nil {
		return fmt.Errorf("failed to write binary to location: %w", err)
	}

	return nil
}
