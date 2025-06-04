package internal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"cosmossdk.io/log"
	"github.com/otiai10/copy"

	"cosmossdk.io/tools/cosmovisor"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type UpgradeCheckResult struct {
	Upgraded   bool
	HaltHeight uint64
}

func UpgradeIfNeeded(cfg *cosmovisor.Config, logger log.Logger, knownHeight uint64) (upgraded bool, haltHeight uint64, err error) {
	logger.Info("Checking for upgrade-info.json")
	if upgradePlan, err := cfg.UpgradeInfo(); err == nil {
		upgrader := NewUpgrader(cfg, logger, upgradePlan)
		err := upgrader.DoUpgrade()
		if err != nil {
			return false, 0, err
		}
		return true, 0, nil
	}
	logger.Info("Checking for upgrade-info.json.batch")
	manualUpgradeBatch, err := cfg.ReadManualUpgrades()
	if err != nil {
		return false, 0, err
	}
	logger.Info("Checking last known height")
	lastKnownHeight := knownHeight
	if lastKnownHeight == 0 {
		lastKnownHeight = cfg.ReadLastKnownHeight()
	}
	if manualUpgrade := manualUpgradeBatch.FirstUpgrade(); manualUpgrade != nil {
		haltHeight = uint64(manualUpgrade.Height)
		if lastKnownHeight == haltHeight {
			logger.Info("At manual upgrade", "upgrade", manualUpgrade, "halt_height", haltHeight)
			upgrader := NewUpgrader(cfg, logger, *manualUpgrade)
			err := upgrader.DoUpgrade()
			if err != nil {
				return false, 0, err
			}
			err = cfg.DeleteManualUpgradeAtHeight(haltHeight)
			return true, haltHeight, err
		} else if lastKnownHeight > haltHeight {
			return false, haltHeight, fmt.Errorf("missed manual upgrade %s at height %d, last known height is %d")
		}
		logger.Info("Found pending manual upgrade", "upgrade", manualUpgrade, "halt_height", haltHeight)
		return false, haltHeight, nil
	}
	return false, 0, nil
}

type Upgrader struct {
	cfg         *cosmovisor.Config
	logger      log.Logger
	upgradePlan upgradetypes.Plan
}

func NewUpgrader(cfg *cosmovisor.Config, logger log.Logger, upgradePlan upgradetypes.Plan) *Upgrader {
	return &Upgrader{
		cfg:         cfg,
		logger:      logger,
		upgradePlan: upgradePlan,
	}
}

func (u *Upgrader) DoUpgrade() error {
	u.logger.Info("Starting upgrade process")
	u.cfg.WaitRestartDelay()

	if err := u.doBackup(); err != nil {
		return err
	}

	if err := u.doCustomPreUpgrade(); err != nil {
		return err
	}

	if err := cosmovisor.UpgradeBinary(u.logger, u.cfg, u.upgradePlan); err != nil {
		return err
	}

	if err := u.doPreUpgrade(); err != nil {
		return err
	}

	return nil
}

// doCustomPreUpgrade executes the custom preupgrade script if provided.
func (u *Upgrader) doCustomPreUpgrade() error {
	if u.cfg.CustomPreUpgrade == "" {
		return nil
	}

	u.logger.Info("Running custom pre-upgrade script", "script", u.cfg.CustomPreUpgrade)

	// check if preupgradeFile is executable file
	preupgradeFile := filepath.Join(u.cfg.Home, "cosmovisor", u.cfg.CustomPreUpgrade)
	u.logger.Info("looking for COSMOVISOR_CUSTOM_PREUPGRADE file", "file", preupgradeFile)
	info, err := os.Stat(preupgradeFile)
	if err != nil {
		u.logger.Error("COSMOVISOR_CUSTOM_PREUPGRADE file missing", "file", preupgradeFile)
		return err
	}
	if !info.Mode().IsRegular() {
		_, f := filepath.Split(preupgradeFile)
		return fmt.Errorf("COSMOVISOR_CUSTOM_PREUPGRADE: %s is not a regular file", f)
	}

	// Set the execute bit for only the current user
	// Given:  Current user - Group - Everyone
	//       0o     RWX     - RWX   - RWX
	oldMode := info.Mode().Perm()
	newMode := oldMode | 0o100
	if oldMode != newMode {
		if err := os.Chmod(preupgradeFile, newMode); err != nil {
			u.logger.Info("COSMOVISOR_CUSTOM_PREUPGRADE could not add execute permission")
			return errors.New("COSMOVISOR_CUSTOM_PREUPGRADE could not add execute permission")
		}
	}

	// Run preupgradeFile
	cmd := exec.Command(preupgradeFile, u.upgradePlan.Name, fmt.Sprintf("%d", u.upgradePlan.Height))
	cmd.Dir = u.cfg.Home
	result, err := cmd.Output()
	if err != nil {
		return err
	}

	u.logger.Info("COSMOVISOR_CUSTOM_PREUPGRADE result", "command", preupgradeFile, "argv1", u.upgradePlan.Name, "argv2", fmt.Sprintf("%d", u.upgradePlan.Height), "result", result)

	return nil
}

// doPreUpgrade runs the pre-upgrade command defined by the application and handles respective error codes.
// cfg contains the cosmovisor config from env var.
// doPreUpgrade runs the new APP binary in order to process the upgrade (post-upgrade for cosmovisor).
func (u *Upgrader) doPreUpgrade() error {
	counter := 0
	for {
		if counter > u.cfg.PreUpgradeMaxRetries {
			return fmt.Errorf("pre-upgrade command failed. reached max attempt of retries - %d", u.cfg.PreUpgradeMaxRetries)
		}

		if err := u.executePreUpgradeCmd(); err != nil {
			counter++

			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				switch exitErr.ExitCode() {
				case 1:
					u.logger.Info("pre-upgrade command does not exist. continuing the upgrade.")
					return nil
				case 30:
					return fmt.Errorf("pre-upgrade command failed : %w", err)
				case 31:
					u.logger.Error("pre-upgrade command failed. retrying", "error", err, "attempt", counter)
					continue
				}
			}
		}

		u.logger.Info("pre-upgrade successful. continuing the upgrade.")
		return nil
	}
}

// executePreUpgradeCmd runs the pre-upgrade command defined by the application
// cfg contains the cosmovisor config from the env vars
func (u *Upgrader) executePreUpgradeCmd() error {
	bin, err := u.cfg.CurrentBin()
	if err != nil {
		return fmt.Errorf("error while getting current binary path: %w", err)
	}

	result, err := exec.Command(bin, "pre-upgrade").Output()
	if err != nil {
		return err
	}

	u.logger.Info("pre-upgrade result", "result", result)
	return nil
}

func (u *Upgrader) doBackup() error {
	// take backup if `UNSAFE_SKIP_BACKUP` is not set.
	if u.cfg.UnsafeSkipBackup {
		return nil
	}

	// a destination directory, Format YYYY-MM-DD
	st := time.Now()
	ymd := fmt.Sprintf("%d-%d-%d", st.Year(), st.Month(), st.Day())
	dst := filepath.Join(u.cfg.DataBackupPath, fmt.Sprintf("data"+"-backup-%s", ymd))

	u.logger.Info("Taking backup of data directory", "backup_path", dst)

	// copy the $DAEMON_HOME/data to a backup dir
	if err := copy.Copy(filepath.Join(u.cfg.Home, "data"), dst); err != nil {
		return fmt.Errorf("error while taking data backup: %w", err)
	}

	// backup is done, lets check endtime to calculate total time taken for backup process
	et := time.Now()
	u.logger.Info("Backup completed", "backup_path", dst, "completion_time", et, "duration", et.Sub(st))

	return nil
}
