package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	rootName        = "cosmovisor"
	genesisDir      = "genesis"
	upgradesDir     = "upgrades"
	currentLink     = "current"
	upgradeFilename = "upgrade-info.json"
)

// must be the same as x/upgrade/types.UpgradeInfoFilename
const defaultFilename = "upgrade-info.json"

// Config is the information passed in to control the daemon
type Config struct {
	Home                  string
	Name                  string
	AllowDownloadBinaries bool
	RestartAfterUpgrade   bool
	PollInterval          time.Duration
	UnsafeSkipBackup      bool

	// currently running upgrade
	currentUpgrade UpgradeInfo
}

// Root returns the root directory where all info lives
func (cfg *Config) Root() string {
	return filepath.Join(cfg.Home, rootName)
}

// GenesisBin is the path to the genesis binary - must be in place to start manager
func (cfg *Config) GenesisBin() string {
	return filepath.Join(cfg.Root(), genesisDir, "bin", cfg.Name)
}

// UpgradeBin is the path to the binary for the named upgrade
func (cfg *Config) UpgradeBin(upgradeName string) string {
	return filepath.Join(cfg.UpgradeDir(upgradeName), "bin", cfg.Name)
}

// UpgradeDir is the directory named upgrade
func (cfg *Config) UpgradeDir(upgradeName string) string {
	safeName := url.PathEscape(upgradeName)
	return filepath.Join(cfg.Home, rootName, upgradesDir, safeName)
}

// UpgradeInfoFile is the expected upgrade-info filename created by `x/upgrade/keeper`.
func (cfg *Config) UpgradeInfoFilePath() string {
	return filepath.Join(cfg.Home, "data", defaultFilename)
}

// SymLinkToGenesis creates a symbolic link from "./current" to the genesis directory.
func (cfg *Config) SymLinkToGenesis() (string, error) {
	genesis := filepath.Join(cfg.Root(), genesisDir)
	link := filepath.Join(cfg.Root(), currentLink)

	if err := os.Symlink(genesis, link); err != nil {
		return "", err
	}
	// and return the genesis binary
	return cfg.GenesisBin(), nil
}

// CurrentBin is the path to the currently selected binary (genesis if no link is set)
// This will resolve the symlink to the underlying directory to make it easier to debug
func (cfg *Config) CurrentBin() (string, error) {
	cur := filepath.Join(cfg.Root(), currentLink)
	// if nothing here, fallback to genesis
	info, err := os.Lstat(cur)
	if err != nil {
		// Create symlink to the genesis
		return cfg.SymLinkToGenesis()
	}
	// if it is there, ensure it is a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		// Create symlink to the genesis
		return cfg.SymLinkToGenesis()
	}

	// resolve it
	dest, err := os.Readlink(cur)
	if err != nil {
		// Create symlink to the genesis
		return cfg.SymLinkToGenesis()
	}

	// and return the binary
	binpath := filepath.Join(dest, "bin", cfg.Name)
	return binpath, nil
}

// GetConfigFromEnv will read the environmental variables into a config
// and then validate it is reasonable
func GetConfigFromEnv() (*Config, error) {
	cfg := &Config{
		Home: os.Getenv("DAEMON_HOME"),
		Name: os.Getenv("DAEMON_NAME"),
	}

	if os.Getenv("DAEMON_ALLOW_DOWNLOAD_BINARIES") == "true" {
		cfg.AllowDownloadBinaries = true
	}

	if os.Getenv("DAEMON_RESTART_AFTER_UPGRADE") == "true" {
		cfg.RestartAfterUpgrade = true
	}

	interval := os.Getenv("DAEMON_POLL_INTERVAL")
	if interval != "" {
		i, err := strconv.ParseUint(interval, 10, 32)
		if err != nil {
			return nil, err
		}
		cfg.PollInterval = time.Millisecond * time.Duration(i)
	} else {
		cfg.PollInterval = 300 * time.Millisecond
	}

	cfg.UnsafeSkipBackup = os.Getenv("UNSAFE_SKIP_BACKUP") == "true"

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate returns an error if this config is invalid.
// it enforces Home/cosmovisor is a valid directory and exists,
// and that Name is set
func (cfg *Config) validate() error {
	if cfg.Name == "" {
		return errors.New("DAEMON_NAME is not set")
	}

	if cfg.Home == "" {
		return errors.New("DAEMON_HOME is not set")
	}

	if !filepath.IsAbs(cfg.Home) {
		return errors.New("DAEMON_HOME must be an absolute path")
	}

	// ensure the root directory exists
	info, err := os.Stat(cfg.Root())
	if err != nil {
		return fmt.Errorf("cannot stat home dir: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", info.Name())
	}

	return nil
}

// SetCurrentUpgrade sets the named upgrade to be the current link, returns error if this binary doesn't exist
func (cfg *Config) SetCurrentUpgrade(u UpgradeInfo) error {
	// ensure named upgrade exists
	bin := cfg.UpgradeBin(u.Name)

	if err := EnsureBinary(bin); err != nil {
		return err
	}

	// set a symbolic link
	link := filepath.Join(cfg.Root(), currentLink)
	safeName := url.PathEscape(u.Name)
	upgrade := filepath.Join(cfg.Root(), upgradesDir, safeName)

	// remove link if it exists
	if _, err := os.Stat(link); err == nil {
		os.Remove(link)
	}

	// point to the new directory
	if err := os.Symlink(upgrade, link); err != nil {
		return fmt.Errorf("creating current symlink: %w", err)
	}

	cfg.currentUpgrade = u
	f, err := os.Create(filepath.Join(upgrade, upgradeFilename))
	if err != nil {
		return err
	}
	bz, err := json.Marshal(u)
	if err != nil {
		return err
	}
	if _, err := f.Write(bz); err != nil {
		return err
	}
	return f.Close()
}

func (cfg *Config) UpgradeInfo() UpgradeInfo {
	if cfg.currentUpgrade.Name != "" {
		return cfg.currentUpgrade
	}

	filename := filepath.Join(cfg.Root(), currentLink, upgradeFilename)
	_, err := os.Lstat(filename)
	var u UpgradeInfo
	var bz []byte
	if err != nil { // no current directory
		goto returnError
	}
	if bz, err = ioutil.ReadFile(filename); err != nil {
		goto returnError
	}
	if err = json.Unmarshal(bz, &u); err != nil {
		goto returnError
	}
	cfg.currentUpgrade = u
	return cfg.currentUpgrade

returnError:
	fmt.Println("[cosmovisor], error reading", filename, err)
	cfg.currentUpgrade.Name = "_"
	return cfg.currentUpgrade
}
