package cosmovisor

import (
	"errors"
	"fmt"
	"os"
)

// DoUpgrade will be called after the log message has been parsed and the process has terminated.
// We can now make any changes to the underlying directory without interference and leave it
// in a state, so we can make a proper restart
func DoUpgrade(cfg *Config, info Plan) error {
	// Simplest case is to switch the link
	err := EnsureBinary(cfg.UpgradeBin(info.Name))
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
	if err := DownloadBinary(cfg, info); err != nil {
		return fmt.Errorf("cannot download binary. %w", err)
	}

	// and then set the binary again
	if err := EnsureBinary(cfg.UpgradeBin(info.Name)); err != nil {
		return fmt.Errorf("downloaded binary doesn't check out: %w", err)
	}

	return cfg.SetCurrentUpgrade(info)
}

// DownloadBinary will grab the binary and place it in the proper directory
func DownloadBinary(cfg *Config, info Plan) error {
	url, err := GetBinaryDownloadURL(info)
	if err != nil {
		return err
	}

	return DownloadFile(cfg, info.Name, url)
}
