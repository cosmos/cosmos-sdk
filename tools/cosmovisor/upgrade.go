package cosmovisor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/x/upgrade/plan"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// UpgradeBinary will be called after the log message has been parsed and the process has terminated.
// We can now make any changes to the underlying directory without interference and leave it
// in a state, so we can make a proper restart
func UpgradeBinary(logger log.Logger, cfg *Config, p upgradetypes.Plan) error {
	// simplest case is to switch the link
	err := plan.EnsureBinary(cfg.UpgradeBin(p.Name))
	if err == nil {
		// we have the binary - do it
		return cfg.SetCurrentUpgrade(p)
	}

	// if auto-download is disabled, we fail
	if !cfg.AllowDownloadBinaries {
		return fmt.Errorf("binary not present, downloading disabled: %w", err)
	}

	// if the dir is there already, don't download either
	switch fi, err := os.Stat(cfg.UpgradeDir(p.Name)); {
	case fi != nil: // The directory exists, do not overwrite.
		return errors.New("upgrade dir already exists, won't overwrite")

	case os.IsNotExist(err): // In this case the directory doesn't exist, continue below.
		// Do nothing and we shall download the binary down below.

	default: // Otherwise an unexpected error
		return fmt.Errorf("unhandled error: %w", err)
	}

	upgradeInfo, err := plan.ParseInfo(p.Info, plan.ParseOptionEnforceChecksum(cfg.DownloadMustHaveChecksum))
	if err != nil {
		return fmt.Errorf("cannot parse upgrade info: %w", err)
	}

	url, err := GetBinaryURL(upgradeInfo.Binaries)
	if err != nil {
		return err
	}

	// NEW: Download PreUpgradeScript if required
	if upgradeInfo.PreUpgradeScript != "" {
		logger.Info("preUpgradeURL script found, downloading & saving to current upgrade folder")
		if err := plan.DownloadPreUpgradeScript(cfg.UpgradeDir(p.Name), upgradeInfo.PreUpgradeScript); err != nil {
			return fmt.Errorf("cannot download preUpgradeScript. %w", err)
		}
		logger.Info("downloading preUpgradeScript complete")

		// Run preupgradeFile
		preupgradeFile := filepath.Join(cfg.UpgradeDir(p.Name), "preupgrade.sh")

		info, err := os.Stat(preupgradeFile)
		if err != nil {
			logger.Error("planned preupgrade file missing", "file", preupgradeFile)
			return err
		}
		if !info.Mode().IsRegular() {
			_, f := filepath.Split(preupgradeFile)
			return fmt.Errorf("planned preupgrade file: %s is not a regular file", f)
		}

		// Set the execute bit for only the current user
		// Given:  Current user - Group - Everyone
		//       0o     RWX     - RWX   - RWX
		oldMode := info.Mode().Perm()
		newMode := oldMode | 0o100
		if oldMode != newMode {
			if err := os.Chmod(preupgradeFile, newMode); err != nil {
				logger.Info("planned preupgrade file: could not add execute permission")
				return errors.New("planned preupgrade file: could not add execute permission")
			}
		}

		cmd := exec.Command(preupgradeFile, p.Name, fmt.Sprintf("%d", p.Height))
		cmd.Dir = cfg.Home
		result, err := cmd.Output()
		if err != nil {
			return err
		}

		logger.Info("Verified pre-upgrade script result", "command", preupgradeFile, "argv1", p.Name, "argv2", fmt.Sprintf("%d", p.Height), "result", result)
	}

	// If not there, then we try to download it... maybe
	logger.Info("no upgrade binary found, beginning to download it")
	if err := plan.DownloadUpgrade(cfg.UpgradeDir(p.Name), url, cfg.Name); err != nil {
		return fmt.Errorf("cannot download binary. %w", err)
	}
	logger.Info("downloading binary complete")

	// and then set the binary again
	if err := plan.EnsureBinary(cfg.UpgradeBin(p.Name)); err != nil {
		return fmt.Errorf("downloaded binary doesn't check out: %w", err)
	}

	return cfg.SetCurrentUpgrade(p)
}

func GetBinaryURL(binaries plan.BinaryDownloadURLMap) (string, error) {
	url, ok := binaries[OSArch()]
	if !ok {
		url, ok = binaries["any"]
	}
	if !ok {
		return "", fmt.Errorf("cannot find binary for os/arch: neither %s, nor any", OSArch())
	}

	return url, nil
}

func OSArch() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}
