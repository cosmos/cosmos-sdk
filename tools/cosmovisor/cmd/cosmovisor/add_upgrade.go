package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func NewAddUpgradeCmd() *cobra.Command {
	addUpgrade := &cobra.Command{
		Use:          "add-upgrade <upgrade-name> <path to executable>",
		Short:        "Add APP upgrade binary to cosmovisor",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(2),
		RunE:         addUpgradeCmd,
	}

	addUpgrade.Flags().Bool(cosmovisor.FlagForce, false, "overwrite existing upgrade binary / upgrade-info.json file")
	addUpgrade.Flags().Int64(cosmovisor.FlagUpgradeHeight, 0, "define a height at which to upgrade the binary automatically (without governance proposal)")

	return addUpgrade
}

// addUpgrade adds upgrade info to manifest
func addUpgrade(cfg *cosmovisor.Config, force bool, upgradeHeight int64, upgradeName, executablePath, upgradeInfoPath string) error {
	logger := cfg.Logger(os.Stdout)

	if !cfg.DisableRecase {
		upgradeName = strings.ToLower(upgradeName)
	}

	if _, err := os.Stat(executablePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("invalid executable path: %w", err)
		}

		return fmt.Errorf("failed to load executable path: %w", err)
	}

	// create upgrade dir
	upgradeLocation := cfg.UpgradeDir(upgradeName)
	if err := os.MkdirAll(path.Join(upgradeLocation, "bin"), 0o755); err != nil {
		return fmt.Errorf("failed to create upgrade directory: %w", err)
	}

	// copy binary to upgrade dir
	executableData, err := os.ReadFile(executablePath)
	if err != nil {
		return fmt.Errorf("failed to read binary: %w", err)
	}

	if err := saveOrAbort(cfg.UpgradeBin(upgradeName), executableData, force); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Using %s for %s upgrade", executablePath, upgradeName))
	logger.Info(fmt.Sprintf("Upgrade binary located at %s", cfg.UpgradeBin(upgradeName)))

	if upgradeHeight > 0 {
		plan := upgradetypes.Plan{Name: upgradeName, Height: upgradeHeight}
		if err := plan.ValidateBasic(); err != nil {
			panic(fmt.Errorf("something is wrong with cosmovisor: %w", err))
		}

		// create upgrade-info.json file
		planData, err := json.Marshal(plan)
		if err != nil {
			return fmt.Errorf("failed to marshal upgrade plan: %w", err)
		}

		if err := saveOrAbort(upgradeInfoPath, planData, force); err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("%s created, %s upgrade binary will switch at height %d", upgradeInfoPath, upgradeName, upgradeHeight))
	}

	return nil
}

// GetConfig returns a Config using passed-in flag
func getConfigFromCmd(cmd *cobra.Command) (*cosmovisor.Config, error) {
	configPath, err := cmd.Flags().GetString(cosmovisor.FlagCosmovisorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get config flag: %w", err)
	}

	cfg, err := cosmovisor.GetConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
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

	return addUpgrade(cfg, force, upgradeHeight, upgradeName, executablePath, cfg.UpgradeInfoFilePath())
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
