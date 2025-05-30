package cosmovisor

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// ReadManualUpgrades reads the manual upgrade data.
func (cfg *Config) ReadManualUpgrades() (ManualUpgradeBatch, error) {
	bz, err := os.ReadFile(cfg.UpgradeInfoBatchFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return cfg.ParseManualUpgrades(bz)
}

func (cfg *Config) ParseManualUpgrades(bz []byte) (ManualUpgradeBatch, error) {
	var manualUpgrades ManualUpgradeBatch
	if err := json.Unmarshal(bz, &manualUpgrades); err != nil {
		return nil, err
	}

	sortUpgrades(manualUpgrades)

	if err := manualUpgrades.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid manual upgrade batch: %w", err)
	}

	return manualUpgrades, nil
}

// AddManualUpgrade adds a manual upgrade plan.
// If an upgrade with the same name already exists, it will only be overwritten if forceOverwrite is true,
// otherwise an error will be returned.
func AddManualUpgrade(cfg *Config, plan *ManualUpgradePlan, forceOverwrite bool) error {
	// TODO only allow plans that are AFTER the last known height
	manualUpgrades, err := cfg.ReadManualUpgrades()
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

	// TODO we should not write the file every time we add an upgrade, but only once per command otherwise we can trigger spurious
	manualUpgradesData, err := json.MarshalIndent(manualUpgrades, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.UpgradeInfoBatchFilePath(), manualUpgradesData, 0644)
}

func sortUpgrades(upgrades ManualUpgradeBatch) {
	sort.Slice(upgrades, func(i, j int) bool {
		return upgrades[i].Height < upgrades[j].Height
	})
}

type ManualUpgradeBatch []*ManualUpgradePlan

func (m ManualUpgradeBatch) ValidateBasic() error {
	for _, upgrade := range m {
		if err := upgrade.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid upgrade plan %s: %w", upgrade.Name, err)
		}
	}
	return nil
}

func (m ManualUpgradeBatch) FirstUpgrade() *ManualUpgradePlan {
	// ensure the upgrades are sorted before searching
	sortUpgrades(m)
	if len(m) == 0 {
		return nil
	}
	return m[0]
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
