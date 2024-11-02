package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

func NewBatchAddUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-batch-upgrade [flags]",
		Short: "Add multiple upgrade binaries at specified heights to cosmovisor",
		Long: `This command allows you to specify multiple upgrades at once at specific heights, copying or creating a batch upgrade file that's actively watched during 'cosmovisor run'.
You can provide upgrades in two ways:

1. Using --upgrade-file: Specify a path to a headerless CSV batch upgrade file in the format:
   upgrade-name,path-to-exec,upgrade-height

2. Using --upgrade-list: Provide a comma-separated list of upgrades.
   Each upgrade is defined by three colon-separated values:
   a. upgrade-name: A unique identifier for the upgrade
   b. path-to-exec: The file path to the upgrade's executable binary
   c. upgrade-height: The block height at which the upgrade should occur
   This creates a batch upgrade JSON file with the upgrade-info objects in the upgrade directory.

Note: You must provide either --upgrade-file or --upgrade-list.`,
		Example: `cosmovisor add-batch-upgrade --upgrade-list upgrade_v2:/path/to/v2/binary:1000000,upgrade_v3:/path/to/v3/binary:2000000

cosmovisor add-batch-upgrade --upgrade-file /path/to/batch_upgrade.csv`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE:         addBatchUpgrade,
	}

	cmd.Flags().String("upgrade-file", "", "Path to a batch upgrade file which is a JSON array of upgrade-info objects")
	cmd.Flags().StringSlice("upgrade-list", []string{}, "List of comma-separated upgrades in the format 'name:path/to/binary:height'")
	cmd.MarkFlagsMutuallyExclusive("upgrade-file", "upgrade-list")

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
		return processUpgradeFile(cfg, upgradeFile)
	}
	upgradeList, err := cmd.Flags().GetStringSlice("upgrade-list")
	if err != nil || len(upgradeList) == 0 {
		return fmt.Errorf("either --upgrade-file or --upgrade-list must be provided")
	}
	var splitUpgrades [][]string
	for _, upgrade := range upgradeList {
		splitUpgrades = append(splitUpgrades, strings.Split(upgrade, ":"))
	}
	return processUpgradeList(cfg, splitUpgrades)
}

// processUpgradeList takes in a list of upgrades and creates a batch upgrade file
func processUpgradeList(cfg *cosmovisor.Config, upgradeList [][]string) error {
	upgradeInfoPaths := []string{}
	for i, upgrade := range upgradeList {
		if len(upgrade) != 3 {
			return fmt.Errorf("argument at position %d (%s) is invalid", i, upgrade)
		}
		upgradeName := filepath.Base(upgrade[0])
		upgradePath := upgrade[1]
		upgradeHeight, err := strconv.ParseInt(upgrade[2], 10, 64)
		if err != nil {
			return fmt.Errorf("upgrade height at position %d (%s) is invalid", i, upgrade[2])
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

// processUpgradeFile takes in a CSV batch upgrade file, parses it and calls processUpgradeList
func processUpgradeFile(cfg *cosmovisor.Config, upgradeFile string) error {
	file, err := os.Open(upgradeFile)
	if err != nil {
		return fmt.Errorf("error opening upgrade CSV file %s: %w", upgradeFile, err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = 3
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("error parsing upgrade CSV file %s: %w", upgradeFile, err)
	}
	if err := processUpgradeList(cfg, records); err != nil {
		return err
	}
	return nil
}
