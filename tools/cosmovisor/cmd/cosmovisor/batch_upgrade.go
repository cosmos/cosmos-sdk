package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func NewBatchAddUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "add-batch-upgrade <upgrade1-name>:<path-to-exec1>:<upgrade1-height> .. <upgradeN-name>:<path-to-execN>:<upgradeN-height>",
		Short:        "Add APP upgrades binary to cosmovisor",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE:         AddBatchUpgrade,
	}
}

// AddBatchUpgrade takes in multiple specified upgrades and creates a single
// batch upgrade file out of them
func AddBatchUpgrade(cmd *cobra.Command, args []string) error {
	cfg, err := getConfigFromCmd(cmd)
	if err != nil {
		return err
	}
	upgradeInfoPaths := []string{}
	for i, as := range args {
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
