package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

func NewAddUpgradeCmd() *cobra.Command {
	addUpgrade := &cobra.Command{
		Use:          "add-upgrade [upgrade-name] [path to executable]",
		Short:        "Add APP upgrade binary to cosmovisor",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(2),
		RunE:         AddUpgrade,
	}

	addUpgrade.Flags().Bool(cosmovisor.FlagForce, false, "overwrite existing upgrade binary / upgrade-info.json file")
	addUpgrade.Flags().Int64(cosmovisor.FlagUpgradeHeight, 0, "define a height at which to upgrade the binary automatically (without governance proposal)")

	return addUpgrade
}

// AddUpgrade adds upgrade info to manifest
func AddUpgrade(cmd *cobra.Command, args []string) error {
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}

	logger := cfg.Logger(os.Stdout)

	upgradeName := args[0]
	if !cfg.DisableRecase {
		upgradeName = strings.ToLower(args[0])
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
	if err := os.MkdirAll(path.Join(upgradeLocation, "bin"), 0o750); err != nil {
		return fmt.Errorf("failed to create upgrade directory: %w", err)
	}

	// copy binary to upgrade dir
	executableData, err := os.ReadFile(executablePath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}

	force, err := cmd.Flags().GetBool(cosmovisor.FlagForce)
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
	}

	if err := saveOrAbort(cfg.UpgradeBin(upgradeName), executableData, force); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Using %s for %s upgrade", executablePath, upgradeName))
	logger.Info(fmt.Sprintf("Upgrade binary located at %s", cfg.UpgradeBin(upgradeName)))

	if upgradeHeight, err := cmd.Flags().GetInt64(cosmovisor.FlagUpgradeHeight); err != nil {
		return fmt.Errorf("failed to get upgrade-height flag: %w", err)
	} else if upgradeHeight > 0 {
		plan := upgradetypes.Plan{Name: upgradeName, Height: upgradeHeight}
		if err := plan.ValidateBasic(); err != nil {
			panic(fmt.Errorf("something is wrong with cosmovisor: %w", err))
		}

		// create upgrade-info.json file
		planData, err := json.Marshal(plan)
		if err != nil {
			return fmt.Errorf("failed to marshal upgrade plan: %w", err)
		}

		if err := saveOrAbort(cfg.UpgradeInfoFilePath(), planData, force); err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("%s created, %s upgrade binary will switch at height %d", filepath.Join(cfg.UpgradeInfoFilePath(), upgradetypes.UpgradeInfoFilename), upgradeName, upgradeHeight))
	}

	return nil
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

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write binary to location: %w", err)
	}

	return nil
}
