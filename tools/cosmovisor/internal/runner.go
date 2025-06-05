package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"cosmossdk.io/log"

	"cosmossdk.io/tools/cosmovisor"
	"cosmossdk.io/tools/cosmovisor/internal/watchers"
	"github.com/cosmos/cosmos-sdk/x/upgrade/plan"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type Runner struct {
	runCfg      RunConfig
	cfg         *cosmovisor.Config
	logger      log.Logger
	knownHeight uint64
}

// NewRunner creates a new Runner instance with the provided configuration and logger.
func NewRunner(cfg *cosmovisor.Config, runCfg RunConfig, logger log.Logger) Runner {
	return Runner{
		runCfg: runCfg,
		cfg:    cfg,
		logger: logger,
	}
}

func (r Runner) Start(ctx context.Context, args []string) error {
	// TODO handle cases where daemon shuts down without an upgrade or change to halt height, either have a retry count of fail in that case ideally backoff retry
	for {
		if testCallback := GetTestCallback(ctx); testCallback != nil {
			testCallback()
		}
		upgraded, haltHeight, err := UpgradeIfNeeded(r.cfg, r.logger, r.knownHeight)
		if err != nil {
			return err
		}
		if upgraded {
			r.logger.Info("Upgrade completed, restarting process")
			if !r.cfg.RestartAfterUpgrade {
				r.logger.Info("DAEMON_RESTART_AFTER_UPGRADE is disabled, exiting process")
				return nil
			}
		}
		err = r.RunOnce(ctx, args, haltHeight)
		if err != nil {
			var restartNeeded ErrRestartNeeded
			if ok := errors.As(err, &restartNeeded); ok {
				r.logger.Info("Restart needed")
			} else if errors.Is(err, errDone) {
				return nil
			} else {
				return err
			}

		}
	}
}

var errDone = errors.New("done")

func (r Runner) RunOnce(ctx context.Context, args []string, haltHeight uint64) error {
	dirWatcher, err := watchers.NewFSNotifyWatcher(ctx, r.cfg.UpgradeInfoDir(), []string{
		r.cfg.UpgradeInfoFilePath(),
		r.cfg.UpgradeInfoBatchFilePath(),
	})
	if err != nil {
		r.logger.Warn("failed to intialize fsnotify, it's probably not available on this platform, using polling only", "error", err)
	}

	// keep the original context for cancellation detection
	parentCtx := ctx
	// create child context for controlling watchers
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	upgradePlanWatcher := watchers.InitWatcher[upgradetypes.Plan](ctx, r.cfg.PollInterval, dirWatcher, r.cfg.UpgradeInfoFilePath(), r.cfg.ParseUpgradeInfo)
	manualUpgradesWatcher := watchers.InitWatcher[cosmovisor.ManualUpgradeBatch](ctx, r.cfg.PollInterval, dirWatcher, r.cfg.UpgradeInfoBatchFilePath(), r.cfg.ParseManualUpgrades)
	heightChecker := watchers.NewHTTPRPCBLockChecker("http://localhost:8080/block")
	heightWatcher := watchers.NewHeightWatcher(ctx, heightChecker, r.cfg.PollInterval, func(height uint64) error {
		r.knownHeight = height
		return r.cfg.WriteLastKnownHeight(height)
	})

	if haltHeight > 0 {
		r.logger.Info("Setting --halt-height flag for manual upgrade", "halt_height", haltHeight)
		args = append(args, fmt.Sprintf("--halt-height=%d", haltHeight))
	}
	cmd, err := r.createCmd(args)
	if err != nil {
		return err
	}
	processRunner := RunProcess(cmd)
	defer func() {
		// TODO always check height before shutting down
		//_, _ = heightChecker.ReadNow()
		_ = processRunner.Shutdown(r.cfg.ShutdownGrace)
	}()

	correctHeightConfirmed := false
	for {
		select {
		// listen to the parent context's cancellation
		case <-parentCtx.Done():
			r.logger.Info("Parent context cancelled, shutting down")
			return errDone
		case _, ok := <-upgradePlanWatcher.Updated():
			// TODO check skip upgrade heights?? (although not sure why we need this as the node should not emit an upgrade plan if skip heights is enabled)
			if !ok {
				return nil
			}
			r.logger.Info("Received upgrade-info.json")
			return ErrRestartNeeded{}
		case manualUpgrades, ok := <-manualUpgradesWatcher.Updated():
			if !ok {
				return nil
			}
			r.logger.Info("Received updates to upgrade-info.json.batch")
			if haltHeight == 0 && len(manualUpgrades) > 0 {
				// shutdown, no halt height set
				r.logger.Info("No halt height set, but manual upgrades found, restarting process")
				return ErrRestartNeeded{}
			} else {
				// restart if we need to change the halt height based on the upgrade
				firstUpgrade := manualUpgrades.FirstUpgrade()
				if firstUpgrade == nil {
					// if we have no longer have an upgrade then we need to remove halt height
					r.logger.Info("No upgrade found, removing halt height")
					return ErrRestartNeeded{}
				}
				if uint64(firstUpgrade.Height) < haltHeight {
					// if we have an earlier halt height then we need to change the halt height
					r.logger.Info("Earlier manual upgrade found, changing halt height", "current_halt_height", haltHeight, "needed_halt_height", firstUpgrade.Height)
					return ErrRestartNeeded{}
				}
			}
		case err := <-processRunner.Done():
			// TODO handle process exit
			r.logger.Warn("Process exited unexpectedly", "error", err)
			return err
		// TODO:
		case actualHeight := <-heightWatcher.Updated():
			r.logger.Warn("Got height update from watcher", "height", actualHeight)
			if haltHeight == 0 {
				// we don't have a halt height, so we don't care to check anything about the actual height
				continue
			}
			if !correctHeightConfirmed {
				// read manual upgrade batch and check if we'd still be at the correct halt height
				manualUpgrades, err := r.cfg.ReadManualUpgrades()
				if err != nil {
					r.logger.Warn("Failed to read manual upgrades", "error", err)
					continue
				}
				firstUpgrade := manualUpgrades.FirstUpgrade()
				if firstUpgrade == nil {
					// no upgrade found, so we shouldn't have a halt height
					r.logger.Warn("No upgrade found, but halt height is set, removing halt height. This is unexpected because we didn't receive an update to upgrade-info.json.batch")
					return ErrRestartNeeded{}
				}
				if uint64(firstUpgrade.Height) == haltHeight {
					correctHeightConfirmed = true
				} else {
					// we're at the wrong halt height so we need to restart
					r.logger.Info("We're at a different height expected, so we need to set a different halt height", "current_halt_height", haltHeight, "needed_halt_height", firstUpgrade.Height)
					return ErrRestartNeeded{}
				}
			}
			// signal a restart if we're at or past the halt height
			if actualHeight >= haltHeight {
				r.logger.Info("Reached halt height, restarting process for upgrade")
				return ErrRestartNeeded{}
			}
			// TODO error channels
		}
	}
}

func (r Runner) createCmd(args []string) (*exec.Cmd, error) {
	bin, err := r.cfg.CurrentBin()
	if err != nil {
		return nil, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := plan.EnsureBinary(bin); err != nil {
		return nil, fmt.Errorf("current binary is invalid: %w", err)
	}

	r.logger.Info("running app", "path", bin, "args", args)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = r.runCfg.StdIn
	cmd.Stdout = r.runCfg.StdOut
	cmd.Stderr = r.runCfg.StdErr
	return cmd, nil
}

// RunConfig defines the configuration for running a command
type RunConfig struct {
	StdIn  io.Reader
	StdOut io.Writer
	StdErr io.Writer
}
