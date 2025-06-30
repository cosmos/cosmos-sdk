package cosmovisor

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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

// AddManualUpgrades adds a manual upgrade plan.
// If an upgrade with the same name already exists, it will only be overwritten if forceOverwrite is true,
// otherwise an error will be returned.
func (cfg *Config) AddManualUpgrades(forceOverwrite bool, plans ...*upgradetypes.Plan) error {
	if len(plans) == 0 {
		return nil
	}
	existing, err := cfg.ReadManualUpgrades()
	if err != nil {
		return err
	}

	planMap := map[string]*upgradetypes.Plan{}
	for _, existingPlan := range existing {
		planMap[existingPlan.Name] = existingPlan
	}
	for _, plan := range plans {
		if err := plan.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid upgrade plan %s: %w", plan.Name, err)
		}
		if _, ok := planMap[plan.Name]; ok {
			if !forceOverwrite {
				return fmt.Errorf("upgrade with name %s already exists", plan.Name)
			}
		}
		planMap[plan.Name] = plan
	}

	var newUpgrades ManualUpgradeBatch
	for _, plan := range planMap {
		newUpgrades = append(newUpgrades, plan)
	}

	return cfg.saveManualUpgrades(newUpgrades)
}

func (cfg *Config) RemoveManualUpgrade(height int64) error {
	manualUpgrades, err := cfg.ReadManualUpgrades()
	if err != nil {
		return err
	}

	var newUpgrades ManualUpgradeBatch
	for _, existing := range manualUpgrades {
		if existing.Height == height {
			continue
		} else {
			newUpgrades = append(newUpgrades, existing)
		}
	}
	if len(newUpgrades) == len(manualUpgrades) {
		return nil
	}
	return cfg.saveManualUpgrades(newUpgrades)
}

func (cfg *Config) saveManualUpgrades(manualUpgrades ManualUpgradeBatch) error {
	sortUpgrades(manualUpgrades)

	manualUpgradesData, err := json.MarshalIndent(manualUpgrades, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.UpgradeInfoBatchFilePath(), manualUpgradesData, 0o644)
}

func sortUpgrades(upgrades ManualUpgradeBatch) {
	sort.Slice(upgrades, func(i, j int) bool {
		return upgrades[i].Height < upgrades[j].Height
	})
}

type ManualUpgradeBatch []*upgradetypes.Plan

func (m ManualUpgradeBatch) ValidateBasic() error {
	for _, upgrade := range m {
		if err := upgrade.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid upgrade plan %s: %w", upgrade.Name, err)
		}
	}
	return nil
}

func (m ManualUpgradeBatch) FirstUpgrade() *upgradetypes.Plan {
	// ensure the upgrades are sorted before searching
	sortUpgrades(m)
	if len(m) == 0 {
		return nil
	}
	return m[0]
}
