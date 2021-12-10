package cosmovisor

import (
	"errors"
	"fmt"
	"os"

	upgradeplan "github.com/cosmos/cosmos-sdk/x/upgrade/plan"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// DoUpgrade will be called after the log message has been parsed and the process has terminated.
// We can now make any changes to the underlying directory without interference and leave it
// in a state, so we can make a proper restart
func DoUpgrade(cfg *Config, info upgradetypes.Plan) error {
	// Simplest case is to switch the link
	err := upgradeplan.EnsureBinary(cfg.UpgradeBin(info.Name))
	if err == nil {
		// we have the binary - do it
		return cfg.SetCurrentUpgrade(info)
	}
	// if auto-download is disabled, we fail
	if !cfg.AllowDownloadBinaries {
		return fmt.Errorf("binary not present, downloading disabled: %w", err)
	}

	// if the dir is there already, don't download either
	if _, err := os.Stat(cfg.UpgradeDir(info.Name)); !os.IsNotExist(err) {
		return errors.New("upgrade dir already exists, won't overwrite")
	}

	// If not there, then we try to download it... maybe
	Logger.Info().Msg("No upgrade binary found, beginning to download it")
	if err = DownloadBinary(cfg, info); err != nil {
		return fmt.Errorf("cannot download binary. %w", err)
	}
	Logger.Info().Msg("Downloading binary complete")

	// and then set the binary again
	if err = upgradeplan.EnsureBinary(cfg.UpgradeBin(info.Name)); err != nil {
		return fmt.Errorf("downloaded binary doesn't check out: %w", err)
	}

	return cfg.SetCurrentUpgrade(info)
}

// DownloadBinary will grab the binary and place it in the proper directory
func DownloadBinary(cfg *Config, plan upgradetypes.Plan) error {
	url, err := GetDownloadURL(cfg, plan)
	if err != nil {
		return err
	}
	return upgradeplan.DownloadUpgrade(cfg.UpgradeDir(plan.Name), url, cfg.Name)
}

// GetDownloadURL gets the url for the arch-dependant binary download.
func GetDownloadURL(cfg *Config, plan upgradetypes.Plan) (string, error) {
	info, err := upgradeplan.ParseInfo(plan.Info, false)
	if err != nil {
		return "", err
	}
	url, found := info.Binaries[cfg.OSArch]
	if !found {
		url, found = info.Binaries["any"]
		if !found {
			return "", fmt.Errorf("cannot find binary for os/arch: neither %s, nor any", cfg.OSArch)
		}
	}
	return url, nil
}
