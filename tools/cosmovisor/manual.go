package cosmovisor

import (
	"encoding/json"
	"fmt"
	"os"
)

// ManualUpgradesFilename is the file to store manual upgrade information
const ManualUpgradesFilename = "manual-upgrades.json"

// ManualUpgradePlan is the plan for a manual upgrade.
type ManualUpgradePlan struct {
	// Name is the name of the upgrade.
	Name string `json:"name"`
	// Height is the block height that the node will halt at (by setting the --halt-height flag).
	Height int64 `json:"height"`
	// Info is any additional information about the upgrade.
	Info string `json:"info"`
}

func (p ManualUpgradePlan) ValidateBasic() error {
	if len(p.Name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if p.Height <= 0 {
		return fmt.Errorf("height must be greater than 0")
	}

	return nil
}

// ManualUpgradeData is the data structure for manual upgrades.
type ManualUpgradeData struct {
	// Upgrades is a map of manual upgrade names to their plans.
	Upgrades map[string]ManualUpgradePlan `json:"upgrades"`
}

func (m ManualUpgradeData) NextPlanAfter(height int64) *ManualUpgradePlan {
	if m.Upgrades == nil {
		return nil
	}

	// sort the upgrades by height
	var nextPlan *ManualUpgradePlan
	for _, plan := range m.Upgrades {
		if plan.Height > height {
			if nextPlan == nil || plan.Height < nextPlan.Height {
				nextPlan = &plan
			}
		}
	}
	return nextPlan
}

// ReadManualUpgrades reads the manual upgrade data.
func ReadManualUpgrades(cfg *Config) (ManualUpgradeData, error) {
	bz, err := os.ReadFile(cfg.ManualUpgradesFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return ManualUpgradeData{}, nil
		}
		return ManualUpgradeData{}, err
	}
	manualUpgrades := ManualUpgradeData{}
	if err := json.Unmarshal(bz, &manualUpgrades); err != nil {
		return ManualUpgradeData{}, err
	}
	return manualUpgrades, nil
}

// AddManualUpgrade adds a manual upgrade plan.
// If an upgrade with the same name already exists, it will only be overwritten if forceOverwrite is true,
// otherwise os.ErrExist will be returned.
func AddManualUpgrade(cfg *Config, plan ManualUpgradePlan, forceOverwrite bool) error {
	manualUpgrades, err := ReadManualUpgrades(cfg)
	if err != nil {
		return err
	}

	if manualUpgrades.Upgrades == nil {
		manualUpgrades.Upgrades = make(map[string]ManualUpgradePlan)
	}

	if _, exists := manualUpgrades.Upgrades[plan.Name]; exists && !forceOverwrite {
		return os.ErrExist
	}

	manualUpgrades.Upgrades[plan.Name] = plan
	manualUpgradesData, err := json.MarshalIndent(manualUpgrades, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.ManualUpgradesFilePath(), manualUpgradesData, 0644)
}
