package cosmovisor

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"cosmossdk.io/log"
	"cosmossdk.io/x/upgrade/plan"
	upgradetypes "cosmossdk.io/x/upgrade/types"
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

	if err := upgradeInfo.ValidateFull(cfg.Name); err != nil {
		return fmt.Errorf("invalid binaries: %w", err)
	}

	url, err := GetBinaryURL(upgradeInfo.Binaries)
	if err != nil {
		return err
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
