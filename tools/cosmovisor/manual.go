package cosmovisor

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// ReadManualUpgrades reads the manual upgrade data.
func ReadManualUpgrades(cfg *Config) (ManualUpgradeBatch, error) {
	bz, err := os.ReadFile(cfg.UpgradeInfoBatchFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var manualUpgrades ManualUpgradeBatch
	if err := json.Unmarshal(bz, &manualUpgrades); err != nil {
		return nil, err
	}

	sortUpgrades(manualUpgrades)
	return manualUpgrades, nil
}

// AddManualUpgrade adds a manual upgrade plan.
// If an upgrade with the same name already exists, it will only be overwritten if forceOverwrite is true,
// otherwise an error will be returned.
func AddManualUpgrade(cfg *Config, plan ManualUpgradePlan, forceOverwrite bool) error {
	manualUpgrades, err := ReadManualUpgrades(cfg)
	if err != nil {
		return err
	}

	var newUpgrades ManualUpgradeBatch
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

func sortUpgrades(upgrades []ManualUpgradePlan) {
	sort.Slice(upgrades, func(i, j int) bool {
		return upgrades[i].Height < upgrades[j].Height
	})
}

type ManualUpgradeBatch []ManualUpgradePlan

func (m ManualUpgradeBatch) ValidateBasic() error {
	for _, upgrade := range m {
		if err := upgrade.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid upgrade plan %s: %w", upgrade.Name, err)
		}
	}
	return nil
}

type ManualUpgradePlan struct {
	Name   string `json:"name"`
	Height int64  `json:"height"`
	Info   string `json:"info"`
}

func (m ManualUpgradePlan) ValidateBasic() error {
	if m.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if m.Height <= 0 {
		return fmt.Errorf("height must be greater than 0")
	}
	return nil
}
