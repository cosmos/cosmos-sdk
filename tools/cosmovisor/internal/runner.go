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

type TestCallback func()

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
	// TODO handle cases where daemon shuts down without an upgrade, either have a retry count of fail in that case ideally backoff retry
	startsWithoutUpgrade := 0
	for {
		upgraded, haltHeight, err := UpgradeIfNeeded(r.cfg, r.logger, r.knownHeight)
		if err != nil {
			return err
		}
		if upgraded {
			r.logger.Info("Upgrade completed, restarting process")
			if !r.cfg.RestartAfterUpgrade {
				r.logger.Info("DAEMON_RESTART_AFTER_UPGRADE is disabled, exiting process")
			}
			startsWithoutUpgrade = 0
		} else {
			if startsWithoutUpgrade >= 5 {
				return fmt.Errorf("process restarted %d times without an upgrade, exiting", startsWithoutUpgrade)
			}
			startsWithoutUpgrade++
		}
		err = r.RunOnce(ctx, args, haltHeight)
		if err != nil {
			var upgradeNeeded ErrUpgradeNeeded
			if ok := errors.As(err, &upgradeNeeded); ok {
				r.logger.Info("Upgrade needed")
			} else {
				return err
			}
		}
	}
}

func (r Runner) RunOnce(ctx context.Context, args []string, haltHeight uint64) error {
	dirWatcher, err := watchers.NewFSNotifyWatcher(ctx, r.cfg.UpgradeInfoDir(), []string{
		r.cfg.UpgradeInfoFilePath(),
		r.cfg.UpgradeInfoBatchFilePath(),
	})
	if err != nil {
		r.logger.Warn("failed to intialize fsnotify, it's probably not available on this platform, using polling only", "error", err)
	}

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
		// TODO start height watcher
		args = append(args, fmt.Sprintf("--halt-height=%d", haltHeight))
	}
	//// TODO start process runner
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
		case _, ok := <-upgradePlanWatcher.Updated():
			// TODO check skip upgrade heights?? (although not sure why we need this as the node should not emit an upgrade plan if skip heights is enabled)
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
