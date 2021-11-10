package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
)

// environment variable names
const (
	EnvHome                 = "DAEMON_HOME"
	EnvName                 = "DAEMON_NAME"
	EnvDownloadBin          = "DAEMON_ALLOW_DOWNLOAD_BINARIES"
	EnvRestartUpgrade       = "DAEMON_RESTART_AFTER_UPGRADE"
	EnvSkipBackup           = "UNSAFE_SKIP_BACKUP"
	EnvInterval             = "DAEMON_POLL_INTERVAL"
	EnvPreupgradeMaxRetries = "DAEMON_PREUPGRADE_MAX_RETRIES"
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
	PreupgradeMaxRetries  int

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
	return filepath.Join(cfg.BaseUpgradeDir(), safeName)
}

// BaseUpgradeDir is the directory containing the named upgrade directories.
func (cfg *Config) BaseUpgradeDir() string {
	return filepath.Join(cfg.Root(), upgradesDir)
}

// UpgradeInfoFilePath is the expected upgrade-info filename created by `x/upgrade/keeper`.
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
	var errs []error
	cfg := &Config{
		Home: os.Getenv(EnvHome),
		Name: os.Getenv(EnvName),
	}

	var err error
	if cfg.AllowDownloadBinaries, err = booleanOption(EnvDownloadBin, false); err != nil {
		errs = append(errs, err)
	}
	if cfg.RestartAfterUpgrade, err = booleanOption(EnvRestartUpgrade, true); err != nil {
		errs = append(errs, err)
	}
	if cfg.UnsafeSkipBackup, err = booleanOption(EnvSkipBackup, false); err != nil {
		errs = append(errs, err)
	}

	interval := os.Getenv(EnvInterval)
	if interval != "" {
		switch i, e := strconv.ParseUint(interval, 10, 32); {
		case e != nil:
			errs = append(errs, fmt.Errorf("invalid %s: %w", EnvInterval, err))
		case i == 0:
			errs = append(errs, fmt.Errorf("invalid %s: cannot be 0", EnvInterval))
		default:
			cfg.PollInterval = time.Millisecond * time.Duration(i)
		}
	} else {
		cfg.PollInterval = 300 * time.Millisecond
	}

	envPreupgradeMaxRetriesVal := os.Getenv(EnvPreupgradeMaxRetries)
	if cfg.PreupgradeMaxRetries, err = strconv.Atoi(envPreupgradeMaxRetriesVal); err != nil && envPreupgradeMaxRetriesVal != "" {
		errs = append(errs, fmt.Errorf("%s could not be parsed to int: %w", EnvPreupgradeMaxRetries, err))
	}

	errs = append(errs, cfg.validate()...)

	if len(errs) > 0 {
		return nil, cverrors.FlattenErrors(errs...)
	}
	return cfg, nil
}

// validate returns an error if this config is invalid.
// it enforces Home/cosmovisor is a valid directory and exists,
// and that Name is set
func (cfg *Config) validate() []error {
	var errs []error
	if cfg.Name == "" {
		errs = append(errs, errors.New(EnvName+" is not set"))
	}

	switch {
	case cfg.Home == "":
		errs = append(errs, errors.New(EnvHome+" is not set"))
	case !filepath.IsAbs(cfg.Home):
		errs = append(errs, errors.New(EnvHome+" must be an absolute path"))
	default:
		switch info, err := os.Stat(cfg.Root()); {
		case err != nil:
			errs = append(errs, fmt.Errorf("cannot stat home dir: %w", err))
		case !info.IsDir():
			errs = append(errs, fmt.Errorf("%s is not a directory", cfg.Root()))
		}
	}

	return errs
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
	if bz, err = os.ReadFile(filename); err != nil {
		goto returnError
	}
	if err = json.Unmarshal(bz, &u); err != nil {
		goto returnError
	}
	cfg.currentUpgrade = u
	return cfg.currentUpgrade

returnError:
	Logger.Error().Err(err).Str("filename", filename).Msg("failed to read")
	cfg.currentUpgrade.Name = "_"
	return cfg.currentUpgrade
}

// checks and validates env option
func booleanOption(name string, defaultVal bool) (bool, error) {
	p := strings.ToLower(os.Getenv(name))
	switch p {
	case "":
		return defaultVal, nil
	case "false":
		return false, nil
	case "true":
		return true, nil
	}
	return false, fmt.Errorf("env variable %q must have a boolean value (\"true\" or \"false\"), got %q", name, p)
}

// DetailString returns a multi-line string with details about this config.
func (cfg Config) DetailString() string {
	configEntries := []struct{ name, value string }{
		{EnvHome, cfg.Home},
		{EnvName, cfg.Name},
		{EnvDownloadBin, fmt.Sprintf("%t", cfg.AllowDownloadBinaries)},
		{EnvRestartUpgrade, fmt.Sprintf("%t", cfg.RestartAfterUpgrade)},
		{EnvInterval, fmt.Sprintf("%s", cfg.PollInterval)},
		{EnvSkipBackup, fmt.Sprintf("%t", cfg.UnsafeSkipBackup)},
		{EnvPreupgradeMaxRetries, fmt.Sprintf("%d", cfg.PreupgradeMaxRetries)},
	}
	derivedEntries := []struct{ name, value string }{
		{"Root Dir", cfg.Root()},
		{"Upgrade Dir", cfg.BaseUpgradeDir()},
		{"Genesis Bin", cfg.GenesisBin()},
		{"Monitored File", cfg.UpgradeInfoFilePath()},
	}

	var sb strings.Builder
	sb.WriteString("Configurable Values:\n")
	for _, kv := range configEntries {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", kv.name, kv.value))
	}
	sb.WriteString("Derived Values:\n")
	dnl := 0
	for _, kv := range derivedEntries {
		if len(kv.name) > dnl {
			dnl = len(kv.name)
		}
	}
	dFmt := fmt.Sprintf("  %%%ds: %%s\n", dnl)
	for _, kv := range derivedEntries {
		sb.WriteString(fmt.Sprintf(dFmt, kv.name, kv.value))
	}
	return sb.String()
}
