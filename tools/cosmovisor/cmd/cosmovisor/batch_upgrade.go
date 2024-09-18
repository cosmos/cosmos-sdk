package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func NewBatchAddUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "add-batch-upgrade <upgrade1-name>:<path-to-exec1>:<upgrade1-height> <upgrade2-name>:<path-to-exec2>:<upgrade2-height> .. <upgradeN-name>:<path-to-execN>:<upgradeN-height>",
		Short:        "Add APP upgrades binary to cosmovisor",
		SilenceUsage: true,
		Args:         cobra.ArbitraryArgs,
		RunE:         AddBatchUpgrade,
	}
}

func AddBatchUpgrade(cmd *cobra.Command, args []string) error {
	cfg, err := GetConfig(cmd)
	if err != nil {
		return err
	}
	upgradeInfoPaths := []string{}
	for i, as := range args {
		a := strings.Split(as, ":")
		if len(a) != 3 {
			return fmt.Errorf("argument at position %d (%s) is invalid", i, as)
		}
		upgradeName := a[0]
		upgradePath := a[1]
		upgradeHeight, err := strconv.ParseInt(a[2], 10, 64)
		if err != nil {
			return fmt.Errorf("upgrade height at position %d (%s) is invalid", i, a[2])
		}
		upgradeInfoPath := cfg.UpgradeInfoFilePath() + upgradeName
		upgradeInfoPaths = append(upgradeInfoPaths, upgradeInfoPath)
		if err := AddUpgrade(cfg, true, upgradeHeight, upgradeName, upgradePath, upgradeInfoPath); err != nil {
			return err
		}
	}

	var allData []json.RawMessage
	for _, uip := range upgradeInfoPaths {
		fileData, err := os.ReadFile(uip)
		if err != nil {
			return fmt.Errorf("Error reading file %s: %v", uip, err)
		}

		// Verify it's valid JSON
		var jsonData json.RawMessage
		if err := json.Unmarshal(fileData, &jsonData); err != nil {
			return fmt.Errorf("Error parsing JSON from file %s: %v", uip, err)
		}

		// Add to our slice
		allData = append(allData, jsonData)
	}

	// Marshal the combined data
	batchData, err := json.MarshalIndent(allData, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshaling combined JSON: %v", err)
	}

	// Write to output file
	err = os.WriteFile(cfg.UpgradeInfoBatchFilePath(), batchData, 0644)
	if err != nil {
		return fmt.Errorf("Error writing combined JSON to file: %v", err)
	}

	return nil
}
