package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

func NewBatchAddUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-batch-upgrade [flags]",
		Short: "Add multiple upgrade binaries at specified heights to cosmovisor",
		Long: `This command allows you to specify multiple upgrades at once at specific heights, copying or creating a batch upgrade file that's actively watched during 'cosmovisor run'.
You can provide upgrades in two ways:

1. Using --upgrade-file: Specify a path to a batch upgrade file ie. a JSON array of upgrade-info objects.
   The file is validated before it's copied over to the upgrade directory.

2. Using --upgrade-list: Provide a comma-separated list of upgrades.
   Each upgrade is defined by three colon-separated values:
   a. upgrade-name: A unique identifier for the upgrade
   b. path-to-exec: The file path to the upgrade's executable binary
   c. upgrade-height: The block height at which the upgrade should occur
   This creates a batch upgrade JSON file with the upgrade-info objects in the upgrade directory.

Note: You must provide either --upgrade-file or --upgrade-list.`,
		Example: `cosmovisor add-batch-upgrade --upgrade-list upgrade_v2:/path/to/v2/binary:1000000,upgrade_v3:/path/to/v3/binary:2000000

cosmovisor add-batch-upgrade --upgrade-file /path/to/batch_upgrade.json`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE:         addBatchUpgrade,
	}

	cmd.Flags().String("upgrade-file", "", "Path to a batch upgrade file which is a JSON array of upgrade-info objects")
	cmd.Flags().StringSlice("upgrade-list", []string{}, "List of comma-separated upgrades in the format 'name:path/to/binary:height'")

	return cmd
}

// addBatchUpgrade takes in multiple specified upgrades and creates a single
// batch upgrade file out of them
func addBatchUpgrade(cmd *cobra.Command, args []string) error {
	cfg, err := getConfigFromCmd(cmd)
	if err != nil {
		return err
	}
	upgradeFile, err := cmd.Flags().GetString("upgrade-file")
	if err == nil && upgradeFile != "" {
		info, err := os.Stat(upgradeFile)
		if err != nil {
			return fmt.Errorf("error getting reading upgrade file %s: %w", upgradeFile, err)
		}
		if info.IsDir() {
			return fmt.Errorf("upgrade file %s is a directory", upgradeFile)
		}
		return processUpgradeFile(cfg, upgradeFile)
	}
	upgradeList, err := cmd.Flags().GetStringSlice("upgrade-list")
	if err != nil || len(upgradeList) == 0 {
		return fmt.Errorf("either --upgrade-file or --upgrade-list must be provided")
	}
	return processUpgradeList(cfg, upgradeList)
}

// processUpgradeList takes in a list of upgrades and creates a batch upgrade file
func processUpgradeList(cfg *cosmovisor.Config, upgradeList []string) error {
	upgradeInfoPaths := []string{}
	for i, as := range upgradeList {
		a := strings.Split(as, ":")
		if len(a) != 3 {
			return fmt.Errorf("argument at position %d (%s) is invalid", i, as)
		}
		upgradeName := filepath.Base(a[0])
		upgradePath := a[1]
		upgradeHeight, err := strconv.ParseInt(a[2], 10, 64)
		if err != nil {
			return fmt.Errorf("upgrade height at position %d (%s) is invalid", i, a[2])
		}
		upgradeInfoPath := cfg.UpgradeInfoFilePath() + "." + upgradeName
		upgradeInfoPaths = append(upgradeInfoPaths, upgradeInfoPath)
		if err := addUpgrade(cfg, true, upgradeHeight, upgradeName, upgradePath, upgradeInfoPath); err != nil {
			return err
		}
	}

	var allData []json.RawMessage
	for _, uip := range upgradeInfoPaths {
		fileData, err := os.ReadFile(uip)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", uip, err)
		}

		// Verify it's valid JSON
		var jsonData json.RawMessage
		if err := json.Unmarshal(fileData, &jsonData); err != nil {
			return fmt.Errorf("error parsing JSON from file %s: %w", uip, err)
		}

		// Add to our slice
		allData = append(allData, jsonData)
	}

	// Marshal the combined data
	batchData, err := json.MarshalIndent(allData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling combined JSON: %w", err)
	}

	// Write to output file
	err = os.WriteFile(cfg.UpgradeInfoBatchFilePath(), batchData, 0o600)
	if err != nil {
		return fmt.Errorf("error writing combined JSON to file: %w", err)
	}

	return nil
}

// processUpgradeFile takes in a batch upgrade file, validates it and copies it to the upgrade directory
func processUpgradeFile(cfg *cosmovisor.Config, upgradeFile string) error {
	b, err := os.ReadFile(upgradeFile)
	if err != nil {
		return fmt.Errorf("error reading upgrade file %s: %w", upgradeFile, err)
	}
	var batch []upgradetypes.Plan
	if err := json.Unmarshal(b, &batch); err != nil {
		return fmt.Errorf("error unmarshalling upgrade file %s: %w", upgradeFile, err)
	}
	return copyFile(upgradeFile, cfg.UpgradeInfoBatchFilePath())
}
