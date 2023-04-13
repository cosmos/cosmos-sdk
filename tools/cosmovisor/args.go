package cosmovisor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/log"
	cverrors "cosmossdk.io/tools/cosmovisor/errors"
	"cosmossdk.io/x/upgrade/plan"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// environment variable names
const (
	EnvHome                 = "DAEMON_HOME"
	EnvName                 = "DAEMON_NAME"
	EnvDownloadBin          = "DAEMON_ALLOW_DOWNLOAD_BINARIES"
	EnvRestartUpgrade       = "DAEMON_RESTART_AFTER_UPGRADE"
	EnvRestartDelay         = "DAEMON_RESTART_DELAY"
	EnvSkipBackup           = "UNSAFE_SKIP_BACKUP"
	EnvDataBackupPath       = "DAEMON_DATA_BACKUP_DIR"
	EnvInterval             = "DAEMON_POLL_INTERVAL"
	EnvPreupgradeMaxRetries = "DAEMON_PREUPGRADE_MAX_RETRIES"
	EnvDisableLogs          = "COSMOVISOR_DISABLE_LOGS"
)

const (
	rootName    = "cosmovisor"
	genesisDir  = "genesis"
	upgradesDir = "upgrades"
	currentLink = "current"
)

// must be the same as x/upgrade/types.UpgradeInfoFilename
const defaultFilename = "upgrade-info.json"

// Config is the information passed in to control the daemon
type Config struct {
	Home                  string
	Name                  string
	AllowDownloadBinaries bool
	RestartAfterUpgrade   bool
	RestartDelay          time.Duration
	PollInterval          time.Duration
	UnsafeSkipBackup      bool
	DataBackupPath        string
	PreupgradeMaxRetries  int
	DisableLogs           bool

	// currently running upgrade
	currentUpgrade upgradetypes.Plan
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

// WaitRestartDelay will block and wait until the RestartDelay has elapsed.
func (cfg *Config) WaitRestartDelay() {
	if cfg.RestartDelay > 0 {
		time.Sleep(cfg.RestartDelay)
	}
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
		Home:           os.Getenv(EnvHome),
		Name:           os.Getenv(EnvName),
		DataBackupPath: os.Getenv(EnvDataBackupPath),
	}

	if cfg.DataBackupPath == "" {
		cfg.DataBackupPath = cfg.Home
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
	if cfg.DisableLogs, err = booleanOption(EnvDisableLogs, false); err != nil {
		errs = append(errs, err)
	}

	interval := os.Getenv(EnvInterval)
	if interval != "" {
		val, err := parseEnvDuration(interval)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid: %s: %w", EnvInterval, err))
		} else {
			cfg.PollInterval = val
		}
	} else {
		cfg.PollInterval = 300 * time.Millisecond
	}

	cfg.RestartDelay = 0 // default value but makes it explicit
	restartDelay := os.Getenv(EnvRestartDelay)
	if restartDelay != "" {
		val, err := parseEnvDuration(restartDelay)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid: %s: %w", EnvRestartDelay, err))
		} else {
			cfg.RestartDelay = val
		}
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

func parseEnvDuration(input string) (time.Duration, error) {
	duration, err := time.ParseDuration(input)
	if err != nil {
		return 0, fmt.Errorf("could not parse '%s' into a duration: %w", input, err)
	}

	if duration <= 0 {
		return 0, fmt.Errorf("must be greater than 0")
	}

	return duration, nil
}

// LogConfigOrError logs either the config details or the error.
func LogConfigOrError(logger log.Logger, cfg *Config, err error) {
	if cfg == nil && err == nil {
		return
	}
	logger.Info("configuration:")
	switch {
	case err != nil:
		cverrors.LogErrors(logger, "configuration errors found", err)
	case cfg != nil:
		logger.Info(cfg.DetailString())
	}
}

// validate returns an error if this config is invalid.
// it enforces Home/cosmovisor is a valid directory and exists,
// and that Name is set
func (cfg *Config) validate() []error {
	var errs []error

	// validate EnvName
	if cfg.Name == "" {
		errs = append(errs, fmt.Errorf("%s is not set", EnvName))
	}

	// validate EnvHome
	switch {
	case cfg.Home == "":
		errs = append(errs, fmt.Errorf("%s is not set", EnvHome))
	case !filepath.IsAbs(cfg.Home):
		errs = append(errs, fmt.Errorf("%s must be an absolute path", EnvHome))
	default:
		switch info, err := os.Stat(cfg.Root()); {
		case err != nil:
			errs = append(errs, fmt.Errorf("cannot stat home dir: %w", err))
		case !info.IsDir():
			errs = append(errs, fmt.Errorf("%s is not a directory", cfg.Root()))
		}
	}

	// check the DataBackupPath
	if cfg.UnsafeSkipBackup {
		return errs
	}

	// if UnsafeSkipBackup is false, validate DataBackupPath
	switch {
	case cfg.DataBackupPath == "":
		errs = append(errs, fmt.Errorf("%s must not be empty", EnvDataBackupPath))
	case !filepath.IsAbs(cfg.DataBackupPath):
		errs = append(errs, fmt.Errorf("%s must be an absolute path", cfg.DataBackupPath))
	default:
		switch info, err := os.Stat(cfg.DataBackupPath); {
		case err != nil:
			errs = append(errs, fmt.Errorf("%q must be a valid directory: %w", cfg.DataBackupPath, err))
		case !info.IsDir():
			errs = append(errs, fmt.Errorf("%q must be a valid directory", cfg.DataBackupPath))
		}
	}

	return errs
}

// SetCurrentUpgrade sets the named upgrade to be the current link, returns error if this binary doesn't exist
func (cfg *Config) SetCurrentUpgrade(u upgradetypes.Plan) (rerr error) {
	// ensure named upgrade exists
	bin := cfg.UpgradeBin(u.Name)

	if err := plan.EnsureBinary(bin); err != nil {
		return err
	}

	// set a symbolic link
	link := filepath.Join(cfg.Root(), currentLink)
	safeName := url.PathEscape(u.Name)
	upgrade := filepath.Join(cfg.Root(), upgradesDir, safeName)

	// remove link if it exists
	if _, err := os.Stat(link); err == nil {
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("failed to remove existing link: %w", err)
		}
	}

	// point to the new directory
	if err := os.Symlink(upgrade, link); err != nil {
		return fmt.Errorf("creating current symlink: %w", err)
	}

	cfg.currentUpgrade = u
	f, err := os.Create(filepath.Join(upgrade, upgradetypes.UpgradeInfoFilename))
	if err != nil {
		return err
	}
	defer func() {
		cerr := f.Close()
		if rerr == nil {
			rerr = cerr
		}
	}()

	bz, err := json.Marshal(u)
	if err != nil {
		return err
	}
	_, err = f.Write(bz)
	return err
}

func (cfg *Config) UpgradeInfo() (upgradetypes.Plan, error) {
	if cfg.currentUpgrade.Name != "" {
		return cfg.currentUpgrade, nil
	}

	filename := filepath.Join(cfg.Root(), currentLink, upgradetypes.UpgradeInfoFilename)
	_, err := os.Lstat(filename)
	var u upgradetypes.Plan
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
	return cfg.currentUpgrade, nil

returnError:
	cfg.currentUpgrade.Name = "_"
	return cfg.currentUpgrade, fmt.Errorf("failed to read %q: %w", filename, err)
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
		{EnvRestartDelay, cfg.RestartDelay.String()},
		{EnvInterval, cfg.PollInterval.String()},
		{EnvSkipBackup, fmt.Sprintf("%t", cfg.UnsafeSkipBackup)},
		{EnvDataBackupPath, cfg.DataBackupPath},
		{EnvPreupgradeMaxRetries, fmt.Sprintf("%d", cfg.PreupgradeMaxRetries)},
		{EnvDisableLogs, fmt.Sprintf("%t", cfg.DisableLogs)},
	}

	derivedEntries := []struct{ name, value string }{
		{"Root Dir", cfg.Root()},
		{"Upgrade Dir", cfg.BaseUpgradeDir()},
		{"Genesis Bin", cfg.GenesisBin()},
		{"Monitored File", cfg.UpgradeInfoFilePath()},
		{"Data Backup Dir", cfg.DataBackupPath},
	}

	var sb strings.Builder
	sb.WriteString("Configurable Values:\n")
	for _, kv := range configEntries {
		fmt.Fprintf(&sb, "  %s: %s\n", kv.name, kv.value)
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
		fmt.Fprintf(&sb, dFmt, kv.name, kv.value)
	}
	return sb.String()
}
