package cosmovisor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/x/upgrade/plan"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// environment variable names
const (
	EnvHome                     = "DAEMON_HOME"
	EnvName                     = "DAEMON_NAME"
	EnvDownloadBin              = "DAEMON_ALLOW_DOWNLOAD_BINARIES"
	EnvDownloadMustHaveChecksum = "DAEMON_DOWNLOAD_MUST_HAVE_CHECKSUM"
	EnvRestartUpgrade           = "DAEMON_RESTART_AFTER_UPGRADE"
	EnvRestartDelay             = "DAEMON_RESTART_DELAY"
	EnvShutdownGrace            = "DAEMON_SHUTDOWN_GRACE"
	EnvSkipBackup               = "UNSAFE_SKIP_BACKUP"
	EnvDataBackupPath           = "DAEMON_DATA_BACKUP_DIR"
	EnvInterval                 = "DAEMON_POLL_INTERVAL"
	EnvPreupgradeMaxRetries     = "DAEMON_PREUPGRADE_MAX_RETRIES"
	EnvDisableLogs              = "COSMOVISOR_DISABLE_LOGS"
	EnvColorLogs                = "COSMOVISOR_COLOR_LOGS"
	EnvTimeFormatLogs           = "COSMOVISOR_TIMEFORMAT_LOGS"
	EnvCustomPreupgrade         = "COSMOVISOR_CUSTOM_PREUPGRADE"
	EnvDisableRecase            = "COSMOVISOR_DISABLE_RECASE"
)

const (
	rootName    = "cosmovisor"
	genesisDir  = "genesis"
	upgradesDir = "upgrades"
	currentLink = "current"
)

// Config is the information passed in to control the daemon
type Config struct {
	Home                     string
	Name                     string
	AllowDownloadBinaries    bool
	DownloadMustHaveChecksum bool
	RestartAfterUpgrade      bool
	RestartDelay             time.Duration
	ShutdownGrace            time.Duration
	PollInterval             time.Duration
	UnsafeSkipBackup         bool
	DataBackupPath           string
	PreupgradeMaxRetries     int
	DisableLogs              bool
	ColorLogs                bool
	TimeFormatLogs           string
	CustomPreupgrade         string
	DisableRecase            bool

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
	return filepath.Join(cfg.Home, "data", upgradetypes.UpgradeInfoFilename)
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
		Home:             os.Getenv(EnvHome),
		Name:             os.Getenv(EnvName),
		DataBackupPath:   os.Getenv(EnvDataBackupPath),
		CustomPreupgrade: os.Getenv(EnvCustomPreupgrade),
	}

	if cfg.DataBackupPath == "" {
		cfg.DataBackupPath = cfg.Home
	}

	var err error
	if cfg.AllowDownloadBinaries, err = BooleanOption(EnvDownloadBin, false); err != nil {
		errs = append(errs, err)
	}
	if cfg.DownloadMustHaveChecksum, err = BooleanOption(EnvDownloadMustHaveChecksum, false); err != nil {
		errs = append(errs, err)
	}
	if cfg.RestartAfterUpgrade, err = BooleanOption(EnvRestartUpgrade, true); err != nil {
		errs = append(errs, err)
	}
	if cfg.UnsafeSkipBackup, err = BooleanOption(EnvSkipBackup, false); err != nil {
		errs = append(errs, err)
	}
	if cfg.DisableLogs, err = BooleanOption(EnvDisableLogs, false); err != nil {
		errs = append(errs, err)
	}
	if cfg.ColorLogs, err = BooleanOption(EnvColorLogs, true); err != nil {
		errs = append(errs, err)
	}
	if cfg.TimeFormatLogs, err = TimeFormatOptionFromEnv(EnvTimeFormatLogs, time.Kitchen); err != nil {
		errs = append(errs, err)
	}
	if cfg.DisableRecase, err = BooleanOption(EnvDisableRecase, false); err != nil {
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

	cfg.ShutdownGrace = 0 // default value but makes it explicit
	shutdownGrace := os.Getenv(EnvShutdownGrace)
	if shutdownGrace != "" {
		val, err := parseEnvDuration(shutdownGrace)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid: %s: %w", EnvShutdownGrace, err))
		} else {
			cfg.ShutdownGrace = val
		}
	}

	envPreupgradeMaxRetriesVal := os.Getenv(EnvPreupgradeMaxRetries)
	if cfg.PreupgradeMaxRetries, err = strconv.Atoi(envPreupgradeMaxRetriesVal); err != nil && envPreupgradeMaxRetriesVal != "" {
		errs = append(errs, fmt.Errorf("%s could not be parsed to int: %w", EnvPreupgradeMaxRetries, err))
	}

	errs = append(errs, cfg.validate()...)
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return cfg, nil
}

func (cfg *Config) Logger(dst io.Writer) log.Logger {
	var logger log.Logger

	if cfg.DisableLogs {
		logger = log.NewNopLogger()
	} else {
		logger = log.NewLogger(dst,
			log.ColorOption(cfg.ColorLogs),
			log.TimeFormatOption(cfg.TimeFormatLogs)).With(log.ModuleKey, "cosmovisor")
	}

	return logger
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
func BooleanOption(name string, defaultVal bool) (bool, error) {
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

// checks and validates env option
func TimeFormatOptionFromEnv(env, defaultVal string) (string, error) {
	val, set := os.LookupEnv(env)
	if !set {
		return defaultVal, nil
	}
	switch val {
	case "layout":
		return time.Layout, nil
	case "ansic":
		return time.ANSIC, nil
	case "unixdate":
		return time.UnixDate, nil
	case "rubydate":
		return time.RubyDate, nil
	case "rfc822":
		return time.RFC822, nil
	case "rfc822z":
		return time.RFC822Z, nil
	case "rfc850":
		return time.RFC850, nil
	case "rfc1123":
		return time.RFC1123, nil
	case "rfc1123z":
		return time.RFC1123Z, nil
	case "rfc3339":
		return time.RFC3339, nil
	case "rfc3339nano":
		return time.RFC3339Nano, nil
	case "kitchen":
		return time.Kitchen, nil
	case "":
		return "", nil
	}
	return "", fmt.Errorf("env variable %q must have a timeformat value (\"layout|ansic|unixdate|rubydate|rfc822|rfc822z|rfc850|rfc1123|rfc1123z|rfc3339|rfc3339nano|kitchen\"), got %q", EnvTimeFormatLogs, val)
}

// DetailString returns a multi-line string with details about this config.
func (cfg Config) DetailString() string {
	configEntries := []struct{ name, value string }{
		{EnvHome, cfg.Home},
		{EnvName, cfg.Name},
		{EnvDownloadBin, fmt.Sprintf("%t", cfg.AllowDownloadBinaries)},
		{EnvDownloadMustHaveChecksum, fmt.Sprintf("%t", cfg.DownloadMustHaveChecksum)},
		{EnvRestartUpgrade, fmt.Sprintf("%t", cfg.RestartAfterUpgrade)},
		{EnvRestartDelay, cfg.RestartDelay.String()},
		{EnvShutdownGrace, cfg.ShutdownGrace.String()},
		{EnvInterval, cfg.PollInterval.String()},
		{EnvSkipBackup, fmt.Sprintf("%t", cfg.UnsafeSkipBackup)},
		{EnvDataBackupPath, cfg.DataBackupPath},
		{EnvPreupgradeMaxRetries, fmt.Sprintf("%d", cfg.PreupgradeMaxRetries)},
		{EnvDisableLogs, fmt.Sprintf("%t", cfg.DisableLogs)},
		{EnvColorLogs, fmt.Sprintf("%t", cfg.ColorLogs)},
		{EnvTimeFormatLogs, cfg.TimeFormatLogs},
		{EnvCustomPreupgrade, cfg.CustomPreupgrade},
		{EnvDisableRecase, fmt.Sprintf("%t", cfg.DisableRecase)},
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
