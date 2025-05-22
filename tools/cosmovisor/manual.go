package cosmovisor

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// readManualUpgrades reads the manual upgrade data.
func readManualUpgrades(cfg *Config) ([]upgradetypes.Plan, error) {
	bz, err := os.ReadFile(cfg.UpgradeInfoBatchFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var manualUpgrades []upgradetypes.Plan
	if err := json.Unmarshal(bz, &manualUpgrades); err != nil {
		return nil, err
	}

	sortUpgrades(manualUpgrades)
	return manualUpgrades, nil
}

// addManualUpgrade adds a manual upgrade plan.
// If an upgrade with the same name already exists, it will only be overwritten if forceOverwrite is true,
// otherwise an error will be returned.
func addManualUpgrade(cfg *Config, plan upgradetypes.Plan, forceOverwrite bool) error {
	manualUpgrades, err := readManualUpgrades(cfg)
	if err != nil {
		return err
	}

	var newUpgrades []upgradetypes.Plan
	for _, existing := range manualUpgrades {
		if existing.Name == plan.Name {
			if !forceOverwrite {
				return fmt.Errorf("upgrade with name %s already exists", plan.Name)
			}
			newUpgrades = append(newUpgrades, plan)
		} else {
			newUpgrades = append(newUpgrades, existing)
		}
	}

	sortUpgrades(manualUpgrades)

	manualUpgradesData, err := json.MarshalIndent(manualUpgrades, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.UpgradeInfoBatchFilePath(), manualUpgradesData, 0644)
}

func sortUpgrades(upgrades []upgradetypes.Plan) {
	sort.Slice(upgrades, func(i, j int) bool {
		return upgrades[i].Height < upgrades[j].Height
	})
}
