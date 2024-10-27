package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/hashicorp/go-getter"
	"github.com/otiai10/copy"
	"github.com/rs/zerolog"
)

// UpgradeBinary will be called after the log message has been parsed and the process has terminated.
// We can now make any changes to the underlying directory without interference and leave it
// in a state, so we can make a proper restart
func UpgradeBinary(logger *zerolog.Logger, cfg *Config, info upgradetypes.Plan) error {
	// simplest case is to switch the link
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
	switch fi, err := os.Stat(cfg.UpgradeDir(info.Name)); {
	case fi != nil: // The directory exists, do not overwrite.
		return errors.New("upgrade dir already exists, won't overwrite")

	case os.IsNotExist(err): // In this case the directory doesn't exist, continue below.
		// Do nothing and we shall download the binary down below.

	default: // Otherwise an unexpected error
		return fmt.Errorf("unhandled error: %w", err)
	}

	// If not there, then we try to download it... maybe
	logger.Info().Msg("no upgrade binary found, beginning to download it")
	if err := DownloadBinary(cfg, info); err != nil {
		return fmt.Errorf("cannot download binary. %w", err)
	}
	logger.Info().Msg("downloading binary complete")

	// and then set the binary again
	if err := EnsureBinary(cfg.UpgradeBin(info.Name)); err != nil {
		return fmt.Errorf("downloaded binary doesn't check out: %w", err)
	}

	return cfg.SetCurrentUpgrade(info)
}

// DownloadBinary will grab the binary and place it in the proper directory
func DownloadBinary(cfg *Config, info upgradetypes.Plan) error {
	url, err := GetDownloadURL(info)
	if err != nil {
		return err
	}

	// download into the bin dir (works for one file)
	binPath := cfg.UpgradeBin(info.Name)
	err = getter.GetFile(binPath, url)

	// if this fails, let's see if it is a zipped directory
	if err != nil {
		dirPath := cfg.UpgradeDir(info.Name)
		err = getter.Get(dirPath, url)
		if err != nil {
			return err
		}
		err = EnsureBinary(binPath)
		// copy binary to binPath from dirPath if zipped directory don't contain bin directory to wrap the binary
		if err != nil {
			err = copy.Copy(filepath.Join(dirPath, cfg.Name), binPath)
			if err != nil {
				return err
			}
		}
	}

	// if it is successful, let's ensure the binary is executable
	return MarkExecutable(binPath)
}

// MarkExecutable will try to set the executable bits if not already set
// Fails if file doesn't exist or we cannot set those bits
func MarkExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stating binary: %w", err)
	}
	// end early if world exec already set
	if info.Mode()&0o001 == 1 {
		return nil
	}
	// now try to set all exec bits
	newMode := info.Mode().Perm() | 0o111
	return os.Chmod(path, newMode)
}

// UpgradeConfig is expected format for the info field to allow auto-download
type UpgradeConfig struct {
	Binaries map[string]string `json:"binaries"`
}

// GetDownloadURL will check if there is an arch-dependent binary specified in Info
func GetDownloadURL(info upgradetypes.Plan) (string, error) {
	doc := strings.TrimSpace(info.Info)
	// if this is a url, then we download that and try to get a new doc with the real info
	if _, err := url.Parse(doc); err == nil {
		tmpDir, err := os.MkdirTemp("", "upgrade-manager-reference")
		if err != nil {
			return "", fmt.Errorf("create tempdir for reference file: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		refPath := filepath.Join(tmpDir, "ref")
		if err := getter.GetFile(refPath, doc); err != nil {
			return "", fmt.Errorf("downloading reference link %s: %w", doc, err)
		}

		refBytes, err := os.ReadFile(refPath)
		if err != nil {
			return "", fmt.Errorf("reading downloaded reference: %w", err)
		}
		// if download worked properly, then we use this new file as the binary map to parse
		doc = string(refBytes)
	}

	// check if it is the upgrade config
	var config UpgradeConfig

	if err := json.Unmarshal([]byte(doc), &config); err == nil {
		url, ok := config.Binaries[OSArch()]
		if !ok {
			url, ok = config.Binaries["any"]
		}
		if !ok {
			return "", fmt.Errorf("cannot find binary for os/arch: neither %s, nor any", OSArch())
		}

		return url, nil
	}

	return "", errors.New("upgrade info doesn't contain binary map")
}

func OSArch() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

// EnsureBinary ensures the file exists and is executable, or returns an error
func EnsureBinary(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat dir %s: %w", path, err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", info.Name())
	}

	// this checks if the world-executable bit is set (we cannot check owner easily)
	exec := info.Mode().Perm() & 0o001
	if exec == 0 {
		return fmt.Errorf("%s is not world executable", info.Name())
	}

	return nil
}
