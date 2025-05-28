package cosmovisor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// doCustomPreUpgrade executes the custom preupgrade script if provided.
func (l Launcher) doCustomPreUpgrade() error {
	if l.cfg.CustomPreUpgrade == "" {
		return nil
	}

	// check if preupgradeFile is executable file
	preupgradeFile := filepath.Join(l.cfg.Home, "cosmovisor", l.cfg.CustomPreUpgrade)
	l.logger.Info("looking for COSMOVISOR_CUSTOM_PREUPGRADE file", "file", preupgradeFile)
	info, err := os.Stat(preupgradeFile)
	if err != nil {
		l.logger.Error("COSMOVISOR_CUSTOM_PREUPGRADE file missing", "file", preupgradeFile)
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
			l.logger.Info("COSMOVISOR_CUSTOM_PREUPGRADE could not add execute permission")
			return errors.New("COSMOVISOR_CUSTOM_PREUPGRADE could not add execute permission")
		}
	}

	// Run preupgradeFile
	cmd := exec.Command(preupgradeFile, l.upgradePlan.Name, fmt.Sprintf("%d", l.upgradePlan.Height))
	cmd.Dir = l.cfg.Home
	result, err := cmd.Output()
	if err != nil {
		return err
	}

	l.logger.Info("COSMOVISOR_CUSTOM_PREUPGRADE result", "command", preupgradeFile, "argv1", l.upgradePlan.Name, "argv2", fmt.Sprintf("%d", l.upgradePlan.Height), "result", result)

	return nil
}

// doPreUpgrade runs the pre-upgrade command defined by the application and handles respective error codes.
// cfg contains the cosmovisor config from env var.
// doPreUpgrade runs the new APP binary in order to process the upgrade (post-upgrade for cosmovisor).
func (l *Launcher) doPreUpgrade() error {
	counter := 0
	for {
		if counter > l.cfg.PreUpgradeMaxRetries {
			return fmt.Errorf("pre-upgrade command failed. reached max attempt of retries - %d", l.cfg.PreUpgradeMaxRetries)
		}

		if err := l.executePreUpgradeCmd(); err != nil {
			counter++

			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				switch exitErr.ExitCode() {
				case 1:
					l.logger.Info("pre-upgrade command does not exist. continuing the upgrade.")
					return nil
				case 30:
					return fmt.Errorf("pre-upgrade command failed : %w", err)
				case 31:
					l.logger.Error("pre-upgrade command failed. retrying", "error", err, "attempt", counter)
					continue
				}
			}
		}

		l.logger.Info("pre-upgrade successful. continuing the upgrade.")
		return nil
	}
}

// executePreUpgradeCmd runs the pre-upgrade command defined by the application
// cfg contains the cosmovisor config from the env vars
func (l *Launcher) executePreUpgradeCmd() error {
	bin, err := l.cfg.CurrentBin()
	if err != nil {
		return fmt.Errorf("error while getting current binary path: %w", err)
	}

	result, err := exec.Command(bin, "pre-upgrade").Output()
	if err != nil {
		return err
	}

	l.logger.Info("pre-upgrade result", "result", result)
	return nil
}
