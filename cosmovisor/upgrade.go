package cosmovisor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
)

// DoUpgrade will be called after the log message has been parsed and the process has terminated.
// We can now make any changes to the underlying directory without interference and leave it
// in a state, so we can make a proper restart
func DoUpgrade(cfg *Config, info *UpgradeInfo) error {
	err := EnsureBinary(cfg.UpgradeBin(info.Name))

	// Simplest case is to switch the link
	if err == nil {
		// we have the binary - do it
		return cfg.SetCurrentUpgrade(info.Name)
	}

	// if auto-download is disabled, we fail
	if !cfg.AllowDownloadBinaries {
		return errors.Wrap(err, "binary not present, downloading disabled")
	}
	// if the dir is there already, don't download either
	_, err = os.Stat(cfg.UpgradeDir(info.Name))
	if !os.IsNotExist(err) {
		return errors.Errorf("upgrade dir already exists, won't overwrite")
	}

	// If not there, then we try to download it... maybe
	if err := DownloadBinary(cfg, info); err != nil {
		return errors.Wrap(err, "cannot download binary")
	}

	// and then set the binary again
	err = EnsureBinary(cfg.UpgradeBin(info.Name))
	if err != nil {
		return errors.Wrap(err, "downloaded binary doesn't check out")
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
		return errors.Wrap(err, "stating binary")
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
			return "", errors.Wrap(err, "create tempdir for reference file")
		}
		defer os.RemoveAll(tmpDir)
		refPath := filepath.Join(tmpDir, "ref")
		err = getter.GetFile(refPath, doc)
		if err != nil {
			return "", errors.Wrapf(err, "downloading reference link %s", doc)
		}
		refBytes, err := ioutil.ReadFile(refPath)
		if err != nil {
			return "", errors.Wrap(err, "reading downloaded reference")
		}
		// if download worked properly, then we use this new file as the binary map to parse
		doc = string(refBytes)
	}

	// check if it is the upgrade config
	var config UpgradeConfig
	err := json.Unmarshal([]byte(doc), &config)
	if err == nil {
		url, ok := config.Binaries[osArch()]
		if !ok {
			return "", errors.Errorf("cannot find binary for os/arch: %s", osArch())
		}
		return url, nil
	}

	return "", errors.New("upgrade info doesn't contain binary map")
}

func osArch() string {
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
		return errors.Wrap(err, "creating current symlink")
	}
	return nil
}

// EnsureBinary ensures the file exists and is executable, or returns an error
func EnsureBinary(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return errors.Wrap(err, "cannot stat home dir")
	}
	if !info.Mode().IsRegular() {
		return errors.Errorf("%s is not a regular file", info.Name())
	}
	// this checks if the world-executable bit is set (we cannot check owner easily)
	exec := info.Mode().Perm() & 0001
	if exec == 0 {
		return errors.Errorf("%s is not world executable", info.Name())
	}
	return nil
}
