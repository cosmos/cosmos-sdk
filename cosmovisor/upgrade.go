package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
)

// DoUpgrade will be called after the log message has been parsed and the process has terminated.
// We can now make any changes to the underlying directory without interference and leave it
// in a state, so we can make a proper restart
func DoUpgrade(cfg *Config, info *UpgradeInfo) error {
	// Simplest case is to switch the link
	err := EnsureBinary(cfg.UpgradeBin(info.Name))
	if err == nil {
		// we have the binary - do it
		return cfg.SetCurrentUpgrade(info.Name)
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
		return fmt.Errorf("cannot download binary: %w", err)
	}

	// and then set the binary again
	if err := EnsureBinary(cfg.UpgradeBin(info.Name)); err != nil {
		return fmt.Errorf("downloaded binary doesn't check out: %w", err)
	}

	return cfg.SetCurrentUpgrade(info.Name)
}

// DownloadBinary will grab the binary and place it in the proper directory
func DownloadBinary(cfg *Config, info *UpgradeInfo) error {
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
	}
	if err != nil {
		return err
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
	if info.Mode()&0001 == 1 {
		return nil
	}
	// now try to set all exec bits
	newMode := info.Mode().Perm() | 0111
	return os.Chmod(path, newMode)
}

// UpgradeConfig is expected format for the info field to allow auto-download
type UpgradeConfig struct {
	Binaries map[string]string `json:"binaries"`
}

// GetDownloadURL will check if there is an arch-dependent binary specified in Info
func GetDownloadURL(info *UpgradeInfo) (string, error) {
	doc := strings.TrimSpace(info.Info)
	// if this is a url, then we download that and try to get a new doc with the real info
	if _, err := url.Parse(doc); err == nil {
		tmpDir, err := ioutil.TempDir("", "upgrade-manager-reference")
		if err != nil {
			return "", fmt.Errorf("create tempdir for reference file: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		refPath := filepath.Join(tmpDir, "ref")
		if err := getter.GetFile(refPath, doc); err != nil {
			return "", fmt.Errorf("downloading reference link %s: %w", doc, err)
		}

		refBytes, err := ioutil.ReadFile(refPath)
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

// SetCurrentUpgrade sets the named upgrade to be the current link, returns error if this binary doesn't exist
func (cfg *Config) SetCurrentUpgrade(upgradeName string) error {
	// ensure named upgrade exists
	bin := cfg.UpgradeBin(upgradeName)

	if err := EnsureBinary(bin); err != nil {
		return err
	}

	// set a symbolic link
	link := filepath.Join(cfg.Root(), currentLink)
	safeName := url.PathEscape(upgradeName)
	upgrade := filepath.Join(cfg.Root(), upgradesDir, safeName)

	// remove link if it exists
	if _, err := os.Stat(link); err == nil {
		os.Remove(link)
	}

	// point to the new directory
	if err := os.Symlink(upgrade, link); err != nil {
		return fmt.Errorf("creating current symlink: %w", err)
	}

	return nil
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
	exec := info.Mode().Perm() & 0001
	if exec == 0 {
		return fmt.Errorf("%s is not world executable", info.Name())
	}

	return nil
}
