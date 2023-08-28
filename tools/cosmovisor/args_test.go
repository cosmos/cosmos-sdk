package cosmovisor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/x/upgrade/plan"
)

type argsTestSuite struct {
	suite.Suite
}

func TestArgsTestSuite(t *testing.T) {
	suite.Run(t, new(argsTestSuite))
}

// cosmovisorEnv are the string values of environment variables used to configure Cosmovisor.
type cosmovisorEnv struct {
	Home                     string
	Name                     string
	DownloadBin              string
	DownloadMustHaveChecksum string
	RestartUpgrade           string
	RestartDelay             string
	SkipBackup               string
	DataBackupPath           string
	Interval                 string
	PreupgradeMaxRetries     string
	DisableLogs              string
	ColorLogs                string
	TimeFormatLogs           string
	CustomPreupgrade         string
	DisableRecase            string
	ShutdownGrace            string
}

type envMap struct {
	val        string
	allowEmpty bool
}

// ToMap creates a map of the cosmovisorEnv where the keys are the env var names.
func (c cosmovisorEnv) ToMap() map[string]envMap {
	return map[string]envMap{
		EnvHome:                     {val: c.Home, allowEmpty: false},
		EnvName:                     {val: c.Name, allowEmpty: false},
		EnvDownloadBin:              {val: c.DownloadBin, allowEmpty: false},
		EnvDownloadMustHaveChecksum: {val: c.DownloadMustHaveChecksum, allowEmpty: false},
		EnvRestartUpgrade:           {val: c.RestartUpgrade, allowEmpty: false},
		EnvRestartDelay:             {val: c.RestartDelay, allowEmpty: false},
		EnvShutdownGrace:            {val: c.ShutdownGrace, allowEmpty: false},
		EnvSkipBackup:               {val: c.SkipBackup, allowEmpty: false},
		EnvDataBackupPath:           {val: c.DataBackupPath, allowEmpty: false},
		EnvInterval:                 {val: c.Interval, allowEmpty: false},
		EnvPreupgradeMaxRetries:     {val: c.PreupgradeMaxRetries, allowEmpty: false},
		EnvDisableLogs:              {val: c.DisableLogs, allowEmpty: false},
		EnvColorLogs:                {val: c.ColorLogs, allowEmpty: false},
		EnvTimeFormatLogs:           {val: c.TimeFormatLogs, allowEmpty: true},
		EnvCustomPreupgrade:         {val: c.CustomPreupgrade, allowEmpty: true},
		EnvDisableRecase:            {val: c.DisableRecase, allowEmpty: true},
	}
}

// Set sets the field in this cosmovisorEnv corresponding to the provided envVar to the given envVal.
func (c *cosmovisorEnv) Set(envVar, envVal string) {
	switch envVar {
	case EnvHome:
		c.Home = envVal
	case EnvName:
		c.Name = envVal
	case EnvDownloadBin:
		c.DownloadBin = envVal
	case EnvDownloadMustHaveChecksum:
		c.DownloadMustHaveChecksum = envVal
	case EnvRestartUpgrade:
		c.RestartUpgrade = envVal
	case EnvRestartDelay:
		c.RestartDelay = envVal
	case EnvShutdownGrace:
		c.ShutdownGrace = envVal
	case EnvSkipBackup:
		c.SkipBackup = envVal
	case EnvDataBackupPath:
		c.DataBackupPath = envVal
	case EnvInterval:
		c.Interval = envVal
	case EnvPreupgradeMaxRetries:
		c.PreupgradeMaxRetries = envVal
	case EnvDisableLogs:
		c.DisableLogs = envVal
	case EnvColorLogs:
		c.ColorLogs = envVal
	case EnvTimeFormatLogs:
		c.TimeFormatLogs = envVal
	case EnvCustomPreupgrade:
		c.CustomPreupgrade = envVal
	case EnvDisableRecase:
		c.DisableRecase = envVal
	default:
		panic(fmt.Errorf("Unknown environment variable [%s]. Cannot set field to [%s]. ", envVar, envVal))
	}
}

// clearEnv clears environment variables and what they were.
// Designed to be used like this:
//
//	initialEnv := clearEnv()
//	defer setEnv(nil, initialEnv)
func (s *argsTestSuite) clearEnv() *cosmovisorEnv {
	s.T().Logf("Clearing environment variables.")
	rv := cosmovisorEnv{}
	for envVar := range rv.ToMap() {
		rv.Set(envVar, os.Getenv(envVar))
		s.Require().NoError(os.Unsetenv(envVar))
	}
	return &rv
}

// setEnv sets environment variables to the values provided.
// If t is not nil, and there's a problem, the test will fail immediately.
// If t is nil, problems will just be logged using s.T().
func (s *argsTestSuite) setEnv(t *testing.T, env *cosmovisorEnv) { //nolint:thelper // false positive
	if t == nil {
		s.T().Logf("Restoring environment variables.")
	}
	for envVar, envVal := range env.ToMap() {
		var err error
		var msg string
		if len(envVal.val) != 0 || envVal.allowEmpty {
			err = os.Setenv(envVar, envVal.val)
			msg = fmt.Sprintf("setting %s to %s", envVar, envVal.val)
		} else {
			err = os.Unsetenv(envVar)
			msg = fmt.Sprintf("unsetting %s", envVar)
		}
		switch {
		case t != nil:
			require.NoError(t, err, msg)
		case err != nil:
			s.T().Logf("error %s: %v", msg, err)
		default:
			s.T().Logf("done %s", msg)
		}
	}
}

func (s *argsTestSuite) TestConfigPaths() {
	cases := map[string]struct {
		cfg           Config
		upgradeName   string
		expectRoot    string
		expectGenesis string
		expectUpgrade string
	}{
		"simple": {
			cfg:           Config{Home: "/foo", Name: "myd"},
			upgradeName:   "bar",
			expectRoot:    fmt.Sprintf("/foo/%s", rootName),
			expectGenesis: fmt.Sprintf("/foo/%s/genesis/bin/myd", rootName),
			expectUpgrade: fmt.Sprintf("/foo/%s/upgrades/bar/bin/myd", rootName),
		},
		"handle space": {
			cfg:           Config{Home: "/longer/prefix/", Name: "yourd"},
			upgradeName:   "some spaces",
			expectRoot:    fmt.Sprintf("/longer/prefix/%s", rootName),
			expectGenesis: fmt.Sprintf("/longer/prefix/%s/genesis/bin/yourd", rootName),
			expectUpgrade: "/longer/prefix/cosmovisor/upgrades/some%20spaces/bin/yourd",
		},
		"handle casing": {
			cfg:           Config{Home: "/longer/prefix/", Name: "appd"},
			upgradeName:   "myUpgrade",
			expectRoot:    fmt.Sprintf("/longer/prefix/%s", rootName),
			expectGenesis: fmt.Sprintf("/longer/prefix/%s/genesis/bin/appd", rootName),
			expectUpgrade: "/longer/prefix/cosmovisor/upgrades/myUpgrade/bin/appd",
		},
	}

	for _, tc := range cases {
		s.Require().Equal(tc.cfg.Root(), filepath.FromSlash(tc.expectRoot))
		s.Require().Equal(tc.cfg.GenesisBin(), filepath.FromSlash(tc.expectGenesis))
		s.Require().Equal(tc.cfg.UpgradeBin(tc.upgradeName), filepath.FromSlash(tc.expectUpgrade))
	}
}

// Test validate
// add more test in test validate
func (s *argsTestSuite) TestValidate() {
	relPath := filepath.Join("testdata", "validate")
	absPath, err := filepath.Abs(relPath)
	s.Require().NoError(err)

	testdata, err := filepath.Abs("testdata")
	s.Require().NoError(err)

	cases := map[string]struct {
		cfg   Config
		valid bool
	}{
		"happy": {
			cfg:   Config{Home: absPath, Name: "bind", DataBackupPath: absPath},
			valid: true,
		},
		"happy with download": {
			cfg:   Config{Home: absPath, Name: "bind", AllowDownloadBinaries: true, DataBackupPath: absPath},
			valid: true,
		},
		"happy with skip data backup": {
			cfg:   Config{Home: absPath, Name: "bind", UnsafeSkipBackup: true, DataBackupPath: absPath},
			valid: true,
		},
		"happy with skip data backup and empty data backup path": {
			cfg:   Config{Home: absPath, Name: "bind", UnsafeSkipBackup: true, DataBackupPath: ""},
			valid: true,
		},
		"happy with skip data backup and no such data backup path dir": {
			cfg:   Config{Home: absPath, Name: "bind", UnsafeSkipBackup: true, DataBackupPath: filepath.FromSlash("/no/such/dir")},
			valid: true,
		},
		"happy with skip data backup and relative data backup path": {
			cfg:   Config{Home: absPath, Name: "bind", UnsafeSkipBackup: true, DataBackupPath: relPath},
			valid: true,
		},
		"missing home": {
			cfg:   Config{Name: "bind"},
			valid: false,
		},
		"missing name": {
			cfg:   Config{Home: absPath},
			valid: false,
		},
		"relative home path": {
			cfg:   Config{Home: relPath, Name: "bind"},
			valid: false,
		},
		"no upgrade manager subdir": {
			cfg:   Config{Home: testdata, Name: "bind"},
			valid: false,
		},
		"no such home dir": {
			cfg:   Config{Home: filepath.FromSlash("/no/such/dir"), Name: "bind"},
			valid: false,
		},
		"empty data backup path": {
			cfg:   Config{Home: absPath, Name: "bind", DataBackupPath: ""},
			valid: false,
		},
		"no such data backup path dir": {
			cfg:   Config{Home: absPath, Name: "bind", DataBackupPath: filepath.FromSlash("/no/such/dir")},
			valid: false,
		},
		"relative data backup path": {
			cfg:   Config{Home: absPath, Name: "bind", DataBackupPath: relPath},
			valid: false,
		},
	}

	for _, tc := range cases {
		errs := tc.cfg.validate()
		if tc.valid {
			s.Require().Len(errs, 0)
		} else {
			s.Require().Greater(len(errs), 0, "number of errors returned")
		}
	}
}

func (s *argsTestSuite) TestEnsureBin() {
	relPath := filepath.Join("testdata", "validate")
	absPath, err := filepath.Abs(relPath)
	s.Require().NoError(err)

	cfg := Config{Home: absPath, Name: "dummyd", DataBackupPath: absPath}
	s.Require().Len(cfg.validate(), 0, "validation errors")

	s.Require().NoError(plan.EnsureBinary(cfg.GenesisBin()))

	cases := map[string]struct {
		upgrade string
		hasBin  bool
	}{
		"proper":       {"chain2", true},
		"no binary":    {"nobin", false},
		"no directory": {"foobarbaz", false},
	}

	for _, tc := range cases {
		err := plan.EnsureBinary(cfg.UpgradeBin(tc.upgrade))
		if tc.hasBin {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

func (s *argsTestSuite) TestBooleanOption() {
	initialEnv := s.clearEnv()
	defer s.setEnv(nil, initialEnv)

	name := "COSMOVISOR_TEST_VAL"

	check := func(def, expected, isErr bool, msg string) {
		v, err := BooleanOption(name, def)
		if isErr {
			s.Require().Error(err)
			return
		}
		s.Require().NoError(err)
		s.Require().Equal(expected, v, msg)
	}

	os.Setenv(name, "")
	check(true, true, false, "should correctly set default value")
	check(false, false, false, "should correctly set default value")

	os.Setenv(name, "wrong")
	check(true, true, true, "should error on wrong value")
	os.Setenv(name, "truee")
	check(true, true, true, "should error on wrong value")

	os.Setenv(name, "false")
	check(true, false, false, "should handle false value")
	check(false, false, false, "should handle false value")
	os.Setenv(name, "faLSe")
	check(true, false, false, "should handle false value case not sensitive")
	check(false, false, false, "should handle false value case not sensitive")

	os.Setenv(name, "true")
	check(true, true, false, "should handle true value")
	check(false, true, false, "should handle true value")

	os.Setenv(name, "TRUE")
	check(true, true, false, "should handle true value case not sensitive")
	check(false, true, false, "should handle true value case not sensitive")
}

func (s *argsTestSuite) TestTimeFormat() {
	initialEnv := s.clearEnv()
	defer s.setEnv(nil, initialEnv)

	name := "COSMOVISOR_TEST_VAL"

	check := func(def, expected string, isErr bool, msg string) {
		v, err := TimeFormatOptionFromEnv(name, def)
		if isErr {
			s.Require().Error(err)
			return
		}
		s.Require().NoError(err)
		s.Require().Equal(expected, v, msg)
	}

	os.Unsetenv(name)
	check(time.Kitchen, time.Kitchen, false, "should correctly set default value")

	os.Setenv(name, "")
	check(time.Kitchen, "", false, "should correctly set to a none")

	os.Setenv(name, "wrong")
	check(time.Kitchen, "", true, "should error on wrong value")

	os.Setenv(name, "layout")
	check(time.Kitchen, time.Layout, false, "should handle layout value")
	os.Setenv(name, "ansic")
	check(time.Kitchen, time.ANSIC, false, "should handle ansic value")
	os.Setenv(name, "unixdate")
	check(time.Kitchen, time.UnixDate, false, "should handle unixdate value")
	os.Setenv(name, "rubydate")
	check(time.Kitchen, time.RubyDate, false, "should handle rubydate value")
	os.Setenv(name, "rfc822")
	check(time.Kitchen, time.RFC822, false, "should handle rfc822 value")
	os.Setenv(name, "rfc822z")
	check(time.Kitchen, time.RFC822Z, false, "should handle rfc822z value")
	os.Setenv(name, "rfc850")
	check(time.Kitchen, time.RFC850, false, "should handle rfc850 value")
	os.Setenv(name, "rfc1123")
	check(time.Kitchen, time.RFC1123, false, "should handle rfc1123 value")
	os.Setenv(name, "rfc1123z")
	check(time.Kitchen, time.RFC1123Z, false, "should handle rfc1123z value")
	os.Setenv(name, "rfc3339")
	check(time.Kitchen, time.RFC3339, false, "should handle rfc3339 value")
	os.Setenv(name, "rfc3339nano")
	check(time.Kitchen, time.RFC3339Nano, false, "should handle rfc3339nano value")
	os.Setenv(name, "kitchen")
	check(time.Kitchen, time.Kitchen, false, "should handle kitchen value")
}

func (s *argsTestSuite) TestDetailString() {
	home := "/home"
	name := "test-name"
	allowDownloadBinaries := true
	downloadMustHaveChecksum := true
	restartAfterUpgrade := true
	pollInterval := 406 * time.Millisecond
	unsafeSkipBackup := false
	dataBackupPath := "/home"
	preupgradeMaxRetries := 8
	cfg := &Config{
		Home:                     home,
		Name:                     name,
		AllowDownloadBinaries:    allowDownloadBinaries,
		DownloadMustHaveChecksum: downloadMustHaveChecksum,
		RestartAfterUpgrade:      restartAfterUpgrade,
		PollInterval:             pollInterval,
		UnsafeSkipBackup:         unsafeSkipBackup,
		DataBackupPath:           dataBackupPath,
		PreupgradeMaxRetries:     preupgradeMaxRetries,
	}

	expectedPieces := []string{
		"Configurable Values:",
		fmt.Sprintf("%s: %s", EnvHome, home),
		fmt.Sprintf("%s: %s", EnvName, name),
		fmt.Sprintf("%s: %t", EnvDownloadBin, allowDownloadBinaries),
		fmt.Sprintf("%s: %t", EnvDownloadMustHaveChecksum, downloadMustHaveChecksum),
		fmt.Sprintf("%s: %t", EnvRestartUpgrade, restartAfterUpgrade),
		fmt.Sprintf("%s: %s", EnvInterval, pollInterval),
		fmt.Sprintf("%s: %t", EnvSkipBackup, unsafeSkipBackup),
		fmt.Sprintf("%s: %s", EnvDataBackupPath, home),
		fmt.Sprintf("%s: %d", EnvPreupgradeMaxRetries, preupgradeMaxRetries),
		fmt.Sprintf("%s: %t", EnvDisableLogs, cfg.DisableLogs),
		fmt.Sprintf("%s: %t", EnvColorLogs, cfg.ColorLogs),
		fmt.Sprintf("%s: %s", EnvTimeFormatLogs, cfg.TimeFormatLogs),
		"Derived Values:",
		fmt.Sprintf("Root Dir: %s", home),
		fmt.Sprintf("Upgrade Dir: %s", home),
		fmt.Sprintf("Genesis Bin: %s", home),
		fmt.Sprintf("Monitored File: %s", home),
		fmt.Sprintf("Data Backup Dir: %s", home),
	}

	actual := cfg.DetailString()

	for _, piece := range expectedPieces {
		s.Assert().Contains(actual, piece)
	}
}

func (s *argsTestSuite) TestGetConfigFromEnv() {
	initialEnv := s.clearEnv()
	defer s.setEnv(nil, initialEnv)

	relPath := filepath.Join("testdata", "validate")
	absPath, perr := filepath.Abs(relPath)
	s.Require().NoError(perr)

	newConfig := func(
		home, name string,
		downloadBin bool,
		downloadMustHaveChecksum bool,
		restartUpgrade bool,
		restartDelay int,
		skipBackup bool,
		dataBackupPath string,
		interval, preupgradeMaxRetries int,
		disableLogs, colorLogs bool,
		timeFormatLogs string,
		customPreUpgrade string,
		disableRecase bool,
		shutdownGrace int,
	) *Config {
		return &Config{
			Home:                     home,
			Name:                     name,
			AllowDownloadBinaries:    downloadBin,
			DownloadMustHaveChecksum: downloadMustHaveChecksum,
			RestartAfterUpgrade:      restartUpgrade,
			RestartDelay:             time.Millisecond * time.Duration(restartDelay),
			PollInterval:             time.Millisecond * time.Duration(interval),
			UnsafeSkipBackup:         skipBackup,
			DataBackupPath:           dataBackupPath,
			PreupgradeMaxRetries:     preupgradeMaxRetries,
			DisableLogs:              disableLogs,
			ColorLogs:                colorLogs,
			TimeFormatLogs:           timeFormatLogs,
			CustomPreupgrade:         customPreUpgrade,
			DisableRecase:            disableRecase,
			ShutdownGrace:            time.Duration(shutdownGrace),
		}
	}

	tests := []struct {
		name             string
		envVals          cosmovisorEnv
		expectedCfg      *Config
		expectedErrCount int
	}{
		{
			name: "all bad",
			envVals: cosmovisorEnv{
				Home:                     "",
				Name:                     "",
				DownloadBin:              "bad",
				DownloadMustHaveChecksum: "bad",
				RestartUpgrade:           "bad",
				RestartDelay:             "bad",
				SkipBackup:               "bad",
				DataBackupPath:           "bad",
				Interval:                 "bad",
				PreupgradeMaxRetries:     "bad",
				TimeFormatLogs:           "bad",
				CustomPreupgrade:         "",
				DisableRecase:            "bad",
				ShutdownGrace:            "bad",
			},
			expectedCfg:      nil,
			expectedErrCount: 13,
		},
		{
			name:             "all good",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "false", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "true", "10s"},
			expectedCfg:      newConfig(absPath, "testname", true, true, false, 600, true, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", true, 10000000000),
			expectedErrCount: 0,
		},
		{
			name:             "nothing set",
			envVals:          cosmovisorEnv{"", "", "", "", "", "", "", "", "", "", "false", "false", "", "", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 3,
		},
		// Note: Home and Name tests are done in TestValidate
		// timeformat tests are done in the TestTimeFormat
		{
			name:             "download bin bad",
			envVals:          cosmovisorEnv{absPath, "testname", "bad", "true", "false", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "download bin not set",
			envVals:          cosmovisorEnv{absPath, "testname", "", "true", "false", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, true, absPath, 303, 1, false, true, time.Kitchen, "", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "download bin true",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "false", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", true, true, false, 600, true, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "download bin false",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, true, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "download ensure checksum true",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "false", "false", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", true, false, false, 600, true, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "restart upgrade bad",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "bad", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "restart upgrade not set",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", true, true, true, 600, true, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "restart upgrade true",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "true", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", true, true, true, 600, true, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "restart upgrade true",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "false", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", true, true, false, 600, true, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "skip unsafe backups bad",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "false", "600ms", "bad", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "skip unsafe backups not set",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "false", "600ms", "", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", true, true, false, 600, false, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "skip unsafe backups true",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "false", "600ms", "true", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", true, true, false, 600, true, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "skip unsafe backups false",
			envVals:          cosmovisorEnv{absPath, "testname", "true", "true", "false", "600ms", "false", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", true, true, false, 600, false, absPath, 303, 1, false, true, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "poll interval bad",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "bad", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "poll interval 0",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "0", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "poll interval not set",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "", "1", "false", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 300, 1, false, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "poll interval 600",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "600", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "poll interval 1s",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "1s", "1", "false", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 1000, 1, false, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "poll interval -3m",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "-3m", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "restart delay bad",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "bad", "false", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "restart delay 0",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "0", "false", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "restart delay not set",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "", "false", "", "303ms", "1", "false", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 0, false, absPath, 303, 1, false, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "restart delay 600",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600", "false", "", "300ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "restart delay 1s",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "1s", "false", "", "303ms", "1", "false", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 1000, false, absPath, 303, 1, false, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "restart delay -3m",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "-3m", "false", "", "303ms", "1", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "prepupgrade max retries bad",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "bad", "false", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "prepupgrade max retries 0",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "0", "false", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 0, false, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "prepupgrade max retries not set",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "false", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 0, false, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "prepupgrade max retries 5",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "5", "false", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 5, false, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "disable logs bad",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "5", "bad", "true", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "disable logs good",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "true", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 0, true, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "disable logs color bad",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "5", "true", "bad", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "disable logs color good",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "true", "false", "kitchen", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 0, true, false, time.Kitchen, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "disable logs timestamp",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "true", "false", "", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 0, true, false, "", "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "enable rf3339 logs timestamp",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "true", "true", "rfc3339", "preupgrade.sh", "", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 0, true, true, time.RFC3339, "preupgrade.sh", false, 0),
			expectedErrCount: 0,
		},
		{
			name:             "invalid logs timestamp format",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "true", "true", "invalid", "preupgrade.sh", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "disable recase good",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "true", "true", "rfc3339", "preupgrade.sh", "true", ""},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 0, true, true, time.RFC3339, "preupgrade.sh", true, 0),
			expectedErrCount: 0,
		},
		{
			name:             "disable recase bad",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "true", "true", "rfc3339", "preupgrade.sh", "bad", ""},
			expectedErrCount: 1,
		},
		{
			name:             "shutdown grace good",
			envVals:          cosmovisorEnv{absPath, "testname", "false", "true", "false", "600ms", "false", "", "406ms", "", "true", "true", "rfc3339", "preupgrade.sh", "true", "15s"},
			expectedCfg:      newConfig(absPath, "testname", false, true, false, 600, false, absPath, 406, 0, true, true, time.RFC3339, "preupgrade.sh", true, 15000000000),
			expectedErrCount: 0,
		},
	}

	for _, tc := range tests {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			s.setEnv(t, &tc.envVals)
			cfg, err := GetConfigFromEnv()
			if tc.expectedErrCount == 0 {
				assert.NoError(t, err)
			} else if assert.Error(t, err) {
				errCount := 1
				if errMulti, ok := err.(interface{ Unwrap() []error }); ok {
					errCount = len(errMulti.Unwrap())
				}
				assert.Equal(t, tc.expectedErrCount, errCount, "error count")
			}
			assert.Equal(t, tc.expectedCfg, cfg, "config")
		})
	}
}

var sink interface{}

func BenchmarkDetailString(b *testing.B) {
	cfg := &Config{
		Home: "/foo", Name: "myd",
		AllowDownloadBinaries: true,
		UnsafeSkipBackup:      true,
		PollInterval:          450 * time.Second,
		PreupgradeMaxRetries:  1e7,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sink = cfg.DetailString()
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}

	// Otherwise reset the sink.
	sink = (interface{})(nil)
}
