package internal

import (
	"cosmossdk.io/log"

	"cosmossdk.io/tools/cosmovisor"
	"cosmossdk.io/tools/cosmovisor/internal/checkers"
	"cosmossdk.io/tools/cosmovisor/internal/watchers"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type BasicRunner struct {
	logger log.Logger
	cfg    *cosmovisor.Config
	// upgradePlanWatcher watches for data in an upgrade-info.json created by the running node
	upgradePlanWatcher watchers.Watcher[upgradetypes.Plan]
	// manualUpgradesWatcher watchers for data in an upgrade-info.json.batch created by the node operator
	manualUpgradesWatcher watchers.Watcher[cosmovisor.ManualUpgradeBatch]
	actualHeightWatcher   watchers.Watcher[uint64]
	heightChecker         checkers.HeightChecker
	runner                ProcessRunner
}

func (r *BasicRunner) ComputePlan() error {
	// TODO check for upgrade-info.json
	if _, err := r.cfg.UpgradeInfo(); err == nil {

	}
	// TODO check for upgrade-info.json.batch
	return nil
}

func (r *BasicRunner) DoUpgrade(plan upgradetypes.Plan) error {
	return nil
}

func (r *BasicRunner) Run(haltHeight uint64) error {
	correctHeightConfirmed := false
	for {
		select {
		case <-r.upgradePlanWatcher.Updated():
			// TODO shutdown
		case <-r.manualUpgradesWatcher.Updated():
			if haltHeight == 0 {
				// TODO shutdown, no halt height set
			} else {
				// TODO check if this would change the halt height
			}
		case <-r.runner.Done():
			// TODO handle process exit
		case actualHeight := <-r.actualHeightWatcher.Updated():
			if !correctHeightConfirmed {
				// TODO read manual upgrade batch and check if we'd still be at the correct halt height
				correctHeightConfirmed = true
			}
			if actualHeight >= haltHeight {
				// TODO shutdown
			}
		}
	}
}
