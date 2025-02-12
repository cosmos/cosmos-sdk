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

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"

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
	EnvGRPCAddress              = "DAEMON_GRPC_ADDRESS"
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

	cfgFileName  = "config"
	cfgExtension = "toml"
)

// Config is the information passed in to control the daemon
type Config struct {
	Home                     string        `toml:"daemon_home" mapstructure:"daemon_home"`
	Name                     string        `toml:"daemon_name" mapstructure:"daemon_name"`
	AllowDownloadBinaries    bool          `toml:"daemon_allow_download_binaries" mapstructure:"daemon_allow_download_binaries" default:"false"`
	DownloadMustHaveChecksum bool          `toml:"daemon_download_must_have_checksum" mapstructure:"daemon_download_must_have_checksum" default:"false"`
	RestartAfterUpgrade      bool          `toml:"daemon_restart_after_upgrade" mapstructure:"daemon_restart_after_upgrade" default:"true"`
	RestartDelay             time.Duration `toml:"daemon_restart_delay" mapstructure:"daemon_restart_delay"`
	ShutdownGrace            time.Duration `toml:"daemon_shutdown_grace" mapstructure:"daemon_shutdown_grace"`
	PollInterval             time.Duration `toml:"daemon_poll_interval" mapstructure:"daemon_poll_interval" default:"300ms"`
	UnsafeSkipBackup         bool          `toml:"unsafe_skip_backup" mapstructure:"unsafe_skip_backup" default:"false"`
	DataBackupPath           string        `toml:"daemon_data_backup_dir" mapstructure:"daemon_data_backup_dir"`
	PreUpgradeMaxRetries     int           `toml:"daemon_preupgrade_max_retries" mapstructure:"daemon_preupgrade_max_retries" default:"0"`
	GRPCAddress              string        `toml:"daemon_grpc_address" mapstructure:"daemon_grpc_address"`
	DisableLogs              bool          `toml:"cosmovisor_disable_logs" mapstructure:"cosmovisor_disable_logs" default:"false"`
	ColorLogs                bool          `toml:"cosmovisor_color_logs" mapstructure:"cosmovisor_color_logs" default:"true"`
	TimeFormatLogs           string        `toml:"cosmovisor_timeformat_logs" mapstructure:"cosmovisor_timeformat_logs" default:"kitchen"`
	CustomPreUpgrade         string        `toml:"cosmovisor_custom_preupgrade" mapstructure:"cosmovisor_custom_preupgrade" default:""`
	DisableRecase            bool          `toml:"cosmovisor_disable_recase" mapstructure:"cosmovisor_disable_recase" default:"false"`

	// currently running upgrade
	currentUpgrade upgradetypes.Plan
}

// Root returns the root directory where all info lives
func (cfg *Config) Root() string {
	return filepath.Join(cfg.Home, rootName)
}

// DefaultCfgPath returns the default path to the configuration file.
func (cfg *Config) DefaultCfgPath() string {
	return filepath.Join(cfg.Root(), cfgFileName+"."+cfgExtension)
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

// UpgradeInfoBatchFilePath is the same as UpgradeInfoFilePath but with a batch suffix.
func (cfg *Config) UpgradeInfoBatchFilePath() string {
	return cfg.UpgradeInfoFilePath() + ".batch"
}

// SymLinkToGenesis creates a symbolic link from "./current" to the genesis directory.
func (cfg *Config) SymLinkToGenesis() (string, error) {
	// workdir is set to cosmovisor directory so relative
	// symlinks are getting resolved correctly
	if err := os.Symlink(genesisDir, currentLink); err != nil {
		return "", err
	}

	res, err := filepath.EvalSymlinks(cfg.GenesisBin())
	if err != nil {
		return "", err
	}

	// and return the genesis binary
	return res, nil
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
	// if it is there, ensure it is a symlink
	info, err := os.Lstat(cur)
	if err != nil || (info.Mode()&os.ModeSymlink == 0) {
		// Create symlink to the genesis
		return cfg.SymLinkToGenesis()
	}

	res, err := filepath.EvalSymlinks(cur)
	if err != nil {
		// Create symlink to the genesis
		return cfg.SymLinkToGenesis()
	}

	// and return the binary
	binpath := filepath.Join(res, "bin", cfg.Name)

	return binpath, nil
}

// GetConfigFromFile will read the configuration from the config file at the given path.
// If the file path is not provided, it will read the configuration from the ENV variables.
// If a file path is provided and ENV variables are set, they will override the values in the file.
func GetConfigFromFile(filePath string) (*Config, error) {
	if filePath == "" {
		return GetConfigFromEnv(false)
	}

	// ensure the file exist
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("config not found: at %s : %w", filePath, err)
	}

	v := viper.New()
	// read the configuration from the file
	v.SetConfigFile(filePath)
	// load the env variables
	// if the env variable is set, it will override the value provided by the config
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	var (
		err  error
		errs []error
	)

	if cfg.TimeFormatLogs, err = getTimeFormatOption(cfg.TimeFormatLogs); err != nil {
		errs = append(errs, err)
	}

	errs = append(errs, cfg.validate()...)
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return cfg, nil
}

// GetConfigFromEnv will read the environmental variables into a config
// and then validate it is reasonable
func GetConfigFromEnv(skipValidate bool) (*Config, error) {
	var errs []error
	cfg := &Config{
		Home:             os.Getenv(EnvHome),
		Name:             os.Getenv(EnvName),
		DataBackupPath:   os.Getenv(EnvDataBackupPath),
		CustomPreUpgrade: os.Getenv(EnvCustomPreupgrade),
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

	envPreUpgradeMaxRetriesVal := os.Getenv(EnvPreupgradeMaxRetries)
	if cfg.PreUpgradeMaxRetries, err = strconv.Atoi(envPreUpgradeMaxRetriesVal); err != nil && envPreUpgradeMaxRetriesVal != "" {
		errs = append(errs, fmt.Errorf("%s could not be parsed to int: %w", EnvPreupgradeMaxRetries, err))
	}

	cfg.GRPCAddress = os.Getenv(EnvGRPCAddress)
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = "localhost:9090"
	}

	if !skipValidate {
		errs = append(errs, cfg.validate()...)
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
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
		return 0, errors.New("must be greater than 0")
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
	safeName := url.PathEscape(u.Name)
	upgrade := filepath.Join(upgradesDir, safeName)

	// remove link if it exists
	if _, err := os.Stat(currentLink); err == nil {
		if err := os.Remove(currentLink); err != nil {
			return fmt.Errorf("failed to remove existing link: %w", err)
		}
	}

	// point to the new directory
	if err := os.Symlink(upgrade, currentLink); err != nil {
		return fmt.Errorf("creating current symlink: %w", err)
	}

	cfg.currentUpgrade = u
	f, err := os.Create(filepath.Join(cfg.Root(), upgrade, upgradetypes.UpgradeInfoFilename))
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

// UpgradeInfo returns the current upgrade info
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

// BooleanOption checks and validate env option
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

// TimeFormatOptionFromEnv checks and validates the time format option
func TimeFormatOptionFromEnv(env, defaultVal string) (string, error) {
	val, set := os.LookupEnv(env)
	if !set {
		return defaultVal, nil
	}

	return getTimeFormatOption(val)
}

func getTimeFormatOption(val string) (string, error) {
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

// ValueToTimeFormatOption converts the time format option to the env value
func ValueToTimeFormatOption(format string) string {
	switch format {
	case time.Layout:
		return "layout"
	case time.ANSIC:
		return "ansic"
	case time.UnixDate:
		return "unixdate"
	case time.RubyDate:
		return "rubydate"
	case time.RFC822:
		return "rfc822"
	case time.RFC822Z:
		return "rfc822z"
	case time.RFC850:
		return "rfc850"
	case time.RFC1123:
		return "rfc1123"
	case time.RFC1123Z:
		return "rfc1123z"
	case time.RFC3339:
		return "rfc3339"
	case time.RFC3339Nano:
		return "rfc3339nano"
	case time.Kitchen:
		return "kitchen"
	default:
		return ""
	}
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
		{EnvPreupgradeMaxRetries, fmt.Sprintf("%d", cfg.PreUpgradeMaxRetries)},
		{EnvDisableLogs, fmt.Sprintf("%t", cfg.DisableLogs)},
		{EnvColorLogs, fmt.Sprintf("%t", cfg.ColorLogs)},
		{EnvTimeFormatLogs, cfg.TimeFormatLogs},
		{EnvCustomPreupgrade, cfg.CustomPreUpgrade},
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

// Export exports the configuration to a file at the cosmovisor root directory.
func (cfg Config) Export() (string, error) {
	// always use the default path
	path := filepath.Clean(cfg.DefaultCfgPath())

	// check if config file already exists ask user if they want to overwrite it
	if _, err := os.Stat(path); err == nil {
		// ask user if they want to overwrite the file
		if !askForConfirmation(fmt.Sprintf("file %s already exists, do you want to overwrite it?", path)) {
			cfg.Logger(os.Stdout).Info("file already exists, not overriding")
			return path, nil
		}
	}

	// create the file
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("failed to create configuration file: %w", err)
	}

	// convert the time value to its format option
	cfg.TimeFormatLogs = ValueToTimeFormatOption(cfg.TimeFormatLogs)

	defer file.Close()

	// write the configuration to the file
	err = toml.NewEncoder(file).Encode(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to encode configuration: %w", err)
	}

	return path, nil
}

func askForConfirmation(str string) bool {
	var response string
	fmt.Printf("%s [y/n]: ", str)
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}

	return strings.ToLower(response) == "y"
}
