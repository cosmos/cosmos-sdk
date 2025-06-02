package internal

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"cosmossdk.io/log"

	"cosmossdk.io/tools/cosmovisor"
	"cosmossdk.io/tools/cosmovisor/internal/watchers"
	"github.com/cosmos/cosmos-sdk/x/upgrade/plan"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

//type BasicRunner struct {
//	logger log.Logger
//	cfg    *cosmovisor.Config
//	// upgradePlanWatcher watches for data in an upgrade-info.json created by the running node
//	upgradePlanWatcher watchers.Watcher[upgradetypes.Plan]
//	// manualUpgradesWatcher watchers for data in an upgrade-info.json.batch created by the node operator
//	manualUpgradesWatcher watchers.Watcher[cosmovisor.ManualUpgradeBatch]
//	actualHeightWatcher   watchers.Watcher[uint64]
//	heightChecker         checkers.HeightChecker
//	runner                ProcessRunner
//}
//
//func (r *BasicRunner) ComputePlan() error {
//	// TODO check for upgrade-info.json
//	if _, err := r.cfg.UpgradeInfo(); err == nil {
//
//	}
//	// TODO check for upgrade-info.json.batch
//	return nil
//}
//
//func (r *BasicRunner) DoUpgrade(plan upgradetypes.Plan) error {
//	return nil
//}
//
//func (r *BasicRunner) Run(haltHeight uint64) error {
//	correctHeightConfirmed := false
//	for {
//		select {
//		case <-r.upgradePlanWatcher.Updated():
//			// TODO shutdown
//		case <-r.manualUpgradesWatcher.Updated():
//			if haltHeight == 0 {
//				// TODO shutdown, no halt height set
//			} else {
//				// TODO check if this would change the halt height
//			}
//		case <-r.runner.Done():
//			// TODO handle process exit
//		case actualHeight := <-r.actualHeightWatcher.Updated():
//			if !correctHeightConfirmed {
//				// TODO read manual upgrade batch and check if we'd still be at the correct halt height
//				correctHeightConfirmed = true
//			}
//			if actualHeight >= haltHeight {
//				// TODO shutdown
//			}
//		}
//	}
//}

func Run(ctx context.Context, cfg *cosmovisor.Config, runCfg RunConfig, args []string, logger log.Logger) error {
	return RunOnce(ctx, cfg, runCfg, args, logger)
}

func RunOnce(ctx context.Context, cfg *cosmovisor.Config, runCfg RunConfig, args []string, logger log.Logger) error {
	logger.Info("Checking for upgrade-info.json")
	if _, err := cfg.UpgradeInfo(); err == nil {
		return ErrUpgradeNeeded{}
	}
	logger.Info("Checking for upgrade-info.json.batch")
	manualUpgradeBatch, err := cfg.ReadManualUpgrades()
	if err != nil {
		return err
	}
	logger.Info("Checking last known height")
	lastKnownHeight := cfg.ReadLastKnownHeight()
	haltHeight := uint64(0)
	if manualUpgrade := manualUpgradeBatch.FirstUpgrade(); manualUpgrade != nil {
		if lastKnownHeight > uint64(manualUpgrade.Height) {
			return fmt.Errorf("missed manual upgrade %s at height %d, last known height is %d")
		}
		haltHeight = uint64(manualUpgrade.Height)
		logger.Info("Found manual upgrade", "upgrade", manualUpgrade, "halt_height", haltHeight)
	}

	// TODO initialize watchers and checkers

	// create directory
	dirWatcher, err := watchers.NewFSNotifyWatcher(ctx, cfg.UpgradeInfoDir(), []string{
		cfg.UpgradeInfoFilePath(),
		cfg.UpgradeInfoBatchFilePath(),
	})
	if err != nil {
		logger.Warn("failed to intialize fsnotify, it's probably not available on this platform, using polling only", "error", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	upgradePlanWatcher := watchers.InitWatcher[upgradetypes.Plan](ctx, cfg.PollInterval, dirWatcher, cfg.UpgradeInfoFilePath(), cfg.ParseUpgradeInfo)
	manualUpgradesWatcher := watchers.InitWatcher[cosmovisor.ManualUpgradeBatch](ctx, cfg.PollInterval, dirWatcher, cfg.UpgradeInfoBatchFilePath(), cfg.ParseManualUpgrades)
	heightChecker := watchers.NewHTTPRPCBLockChecker("http://localhost:8080/block")
	heightWatcher := watchers.NewHeightWatcher(ctx, heightChecker, cfg.PollInterval, func(height uint64) error {
		return cfg.WriteLastKnownHeight(height)
	})

	if haltHeight > 0 {
		// TODO start height watcher
		args = append(args, fmt.Sprintf("--halt-height=%d", haltHeight))
	}
	//// TODO start process runner
	cmd, err := createCmd(cfg, runCfg, args, logger)
	if err != nil {
		return err
	}
	processRunner := RunProcess(cmd)
	defer func() {
		// TODO always check height before shutting down
		//_, _ = heightChecker.ReadNow()
		_ = processRunner.Shutdown(cfg.ShutdownGrace)
	}()

	correctHeightConfirmed := false
	for {
		select {
		case _, ok := <-upgradePlanWatcher.Updated():
			if !ok {
				return nil
			}
			return ErrUpgradeNeeded{}
		case _, ok := <-manualUpgradesWatcher.Updated():
			if !ok {
				return nil
			}
			if haltHeight == 0 {
				// TODO shutdown, no halt height set
				return ErrUpgradeNeeded{}
			} else {
				// TODO check if this would change the halt height
			}
		case err := <-processRunner.Done():
			// TODO handle process exit
			return err
		// TODO:
		case actualHeight := <-heightWatcher.Updated():
			if !correctHeightConfirmed {
				// TODO read manual upgrade batch and check if we'd still be at the correct halt height
				correctHeightConfirmed = true
			}
			if actualHeight >= haltHeight {
				return ErrUpgradeNeeded{}
			}
			// TODO error channels
		}
	}
}

func createCmd(cfg *cosmovisor.Config, runCfg RunConfig, args []string, logger log.Logger) (*exec.Cmd, error) {
	bin, err := cfg.CurrentBin()
	if err != nil {
		return nil, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := plan.EnsureBinary(bin); err != nil {
		return nil, fmt.Errorf("current binary is invalid: %w", err)
	}

	logger.Info("running app", "path", bin, "args", args)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = runCfg.StdIn
	cmd.Stdout = runCfg.StdOut
	cmd.Stderr = runCfg.StdErr
	return cmd, nil
}

// RunConfig defines the configuration for running a command
type RunConfig struct {
	StdIn  io.Reader
	StdOut io.Writer
	StdErr io.Writer
}
