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
	retryMgr := NewRetryBackoffManager()
	for {
		// First we check if we need to upgrade and if we do we perform the upgrade
		upgraded, err := UpgradeIfNeeded(r.cfg, r.logger, r.knownHeight)
		if err != nil {
			return err
		}
		// If we upgraded, we need to restart the process, but some configurations do not allow automatic restarts
		if upgraded {
			r.logger.Info("Upgrade completed, restarting process")
			if !r.cfg.RestartAfterUpgrade {
				r.logger.Info("DAEMON_RESTART_AFTER_UPGRADE is disabled, exiting process")
				return nil
			}
		}
		// Now we compute the command to run and figure out the halt height if needed
		cmd, haltHeight, err := r.ComputeRunPlan(args)
		if err != nil {
			return err
		}

		// Usually restarts should be due to either:
		// 1. an upgrade that requires a restart
		// 2. a change in the halt height due to a new manual upgrade plan
		// There are also cases where an app could just shut down due to some error.
		// If we're in that sort of situation, we want to retry running the command, but
		// we apply a backoff strategy to avoid hammering the process in case of repeated failures.
		// We pass the current command and args to the retry manager so it can check whether
		// the command or its arguments have changed (e.g. if the binary was updated or the halt height changed),
		// or if we're just in some sort of error restart loop.
		// TODO add tests for this behavior
		if err := retryMgr.BeforeRun(cmd.Path, cmd.Args); err != nil {
			return err
		}

		// In order to make in process testing feasible, we allow a test callback to be set
		// and we call it here right before running the process.
		// Without this it would be much harder to test the cosmovisor runner in a controlled but realistic scenario.
		if testCallback := GetTestCallback(ctx); testCallback != nil {
			testCallback()
		}

		// Now we actually run the process
		err = r.RunProcess(ctx, cmd, haltHeight)
		if err != nil {
			// There are three types of errors we're checking for here:
			// 1. ErrRestartNeeded: this is a custom error that is returned whenever the run loop detects that a restart is needed.
			// 2. errDone: this is a sentinel error that indicates that the cosmovisor process itself should be stopped gracefully.
			// 3. Any other error: this is an unexpected error that should be logged and returned, causing cosmovisor to exit
			// TODO is it right for cosmovisor to exit on any other error (basically a non-zero return code)? Maybe we should just log it and continue?
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

func (r Runner) ComputeRunPlan(args []string) (cmd *exec.Cmd, haltHeight uint64, err error) {
	bin, err := r.cfg.CurrentBin()
	if err != nil {
		return nil, 0, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := plan.EnsureBinary(bin); err != nil {
		return nil, 0, fmt.Errorf("current binary is invalid: %w", err)
	}

	cmd = exec.Command(bin, args...)
	cmd.Stdin = r.runCfg.StdIn
	cmd.Stdout = r.runCfg.StdOut
	cmd.Stderr = r.runCfg.StdErr
	r.logger.Info("Checking for upgrade-info.json.batch")
	manualUpgradeBatch, err := r.cfg.ReadManualUpgrades()
	if err != nil {
		return nil, 0, err
	}
	manualUpgrade := manualUpgradeBatch.FirstUpgrade()
	if manualUpgrade != nil {
		haltHeight = uint64(manualUpgrade.Height)
		r.logger.Info("Setting --halt-height flag for manual upgrade", "halt_height", haltHeight)
		cmd.Args = append(cmd.Args, fmt.Sprintf("--halt-height=%d", haltHeight))
	}
	return
}

func (r Runner) RunProcess(ctx context.Context, cmd *exec.Cmd, haltHeight uint64) error {
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

	r.logger.Info("Starting process %s with args %v", cmd.Path, cmd.Args)
	processRunner := RunProcess(cmd)
	defer func() {
		// TODO always check for the latest block height before shutting down
		_, _ = heightChecker.GetLatestBlockHeight()
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

// RunConfig defines the configuration for running a command
type RunConfig struct {
	StdIn  io.Reader
	StdOut io.Writer
	StdErr io.Writer
}
