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

	"github.com/cosmos/cosmos-sdk/cosmovisor/errors"
)

type argsTestSuite struct {
	suite.Suite

	envVars []string
}

func TestArgsTestSuite(t *testing.T) {
	suite.Run(t, new(argsTestSuite))
}

func (s *argsTestSuite) SetupSuite() {
	s.envVars = []string{EnvHome, EnvName, EnvDownloadBin, EnvRestartUpgrade, EnvSkipBackup, EnvInterval, EnvPreupgradeMaxRetries}
}

// clearEnv clears environment variables and returns the values
// in the same order as the entries in s.envVars.
// Designed to be used like this:
//    initialEnv := clearEnv()
//    defer setEnv(nil, initialEnv)
func (s *argsTestSuite) clearEnv() []string {
	s.T().Logf("Clearing environment variables.")
	rv := make([]string, len(s.envVars))
	for i, envVar := range s.envVars {
		rv[i] = os.Getenv(envVar)
		s.Require().NoError(os.Unsetenv(envVar))
	}
	return rv
}

// setEnv sets environment variables to the values provided.
// Ordering of envVals is the same as the entries in s.envVars.
// If t is not nil, and there's a problem, the test will fail immediately.
// If t is nil, problems will just be logged using s.T().
func (s *argsTestSuite) setEnv(t *testing.T, envVals []string) {
	if t == nil {
		s.T().Logf("Restoring environment variables.")
	}
	for i := 0; i < len(envVals) && i < len(s.envVars); i++ {
		envVar := s.envVars[i]
		envVal := envVals[i]
		var err error
		var msg string
		if len(envVal) != 0 {
			err = os.Setenv(envVar, envVal)
			msg = fmt.Sprintf("setting %s to %s", envVar, envVal)
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
	}

	for _, tc := range cases {
		s.Require().Equal(tc.cfg.Root(), filepath.FromSlash(tc.expectRoot))
		s.Require().Equal(tc.cfg.GenesisBin(), filepath.FromSlash(tc.expectGenesis))
		s.Require().Equal(tc.cfg.UpgradeBin(tc.upgradeName), filepath.FromSlash(tc.expectUpgrade))
	}
}

// Test validate
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
			cfg:   Config{Home: absPath, Name: "bind"},
			valid: true,
		},
		"happy with download": {
			cfg:   Config{Home: absPath, Name: "bind", AllowDownloadBinaries: true},
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
		"relative path": {
			cfg:   Config{Home: relPath, Name: "bind"},
			valid: false,
		},
		"no upgrade manager subdir": {
			cfg:   Config{Home: testdata, Name: "bind"},
			valid: false,
		},
		"no such dir": {
			cfg:   Config{Home: filepath.FromSlash("/no/such/dir"), Name: "bind"},
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

	cfg := Config{Home: absPath, Name: "dummyd"}
	s.Require().Len(cfg.validate(), 0, "validation errors")

	s.Require().NoError(EnsureBinary(cfg.GenesisBin()))

	cases := map[string]struct {
		upgrade string
		hasBin  bool
	}{
		"proper":         {"chain2", true},
		"no binary":      {"nobin", false},
		"not executable": {"noexec", false},
		"no directory":   {"foobarbaz", false},
	}

	for _, tc := range cases {
		err := EnsureBinary(cfg.UpgradeBin(tc.upgrade))
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
		v, err := booleanOption(name, def)
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

func (s *argsTestSuite) TestDetailString() {
	home := "/home"
	name := "test-name"
	allowDownloadBinaries := true
	restartAfterUpgrade := true
	pollInterval := 406 * time.Millisecond
	unsafeSkipBackup := false
	preupgradeMaxRetries := 8
	cfg := &Config{
		Home:                  home,
		Name:                  name,
		AllowDownloadBinaries: allowDownloadBinaries,
		RestartAfterUpgrade:   restartAfterUpgrade,
		PollInterval:          pollInterval,
		UnsafeSkipBackup:      unsafeSkipBackup,
		PreupgradeMaxRetries:  preupgradeMaxRetries,
	}

	expectedPieces := []string{
		"Configurable Values:",
		fmt.Sprintf("%s: %s", EnvHome, home),
		fmt.Sprintf("%s: %s", EnvName, name),
		fmt.Sprintf("%s: %t", EnvDownloadBin, allowDownloadBinaries),
		fmt.Sprintf("%s: %t", EnvRestartUpgrade, restartAfterUpgrade),
		fmt.Sprintf("%s: %s", EnvInterval, pollInterval),
		fmt.Sprintf("%s: %t", EnvSkipBackup, unsafeSkipBackup),
		fmt.Sprintf("%s: %d", EnvPreupgradeMaxRetries, preupgradeMaxRetries),
		"Derived Values:",
		fmt.Sprintf("Root Dir: %s", home),
		fmt.Sprintf("Upgrade Dir: %s", home),
		fmt.Sprintf("Genesis Bin: %s", home),
		fmt.Sprintf("Monitored Upgrade Info File: %s", home),
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

	newConfig := func(home, name string, downloadBin, restartUpgrade, skipBackup bool, interval, preupgradeMaxRetries int) *Config {
		return &Config{
			Home:                  home,
			Name:                  name,
			AllowDownloadBinaries: downloadBin,
			RestartAfterUpgrade:   restartUpgrade,
			PollInterval:          time.Millisecond * time.Duration(interval),
			UnsafeSkipBackup:      skipBackup,
			PreupgradeMaxRetries:  preupgradeMaxRetries,
		}
	}

	tests := []struct {
		name             string
		envVals          []string
		expectedCfg      *Config
		expectedErrCount int
	}{
		// EnvHome, EnvName, EnvDownloadBin, EnvRestartUpgrade, EnvSkipBackup, EnvInterval, EnvPreupgradeMaxRetries
		{
			name:             "all bad",
			envVals:          []string{"", "", "bad", "bad", "bad", "bad", "bad"},
			expectedCfg:      nil,
			expectedErrCount: 7,
		},
		{
			name:             "all good",
			envVals:          []string{absPath, "testname", "true", "false", "true", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", true, false, true, 303, 1),
			expectedErrCount: 0,
		},
		{
			name:             "nothing set",
			envVals:          []string{"", "", "", "", "", "", ""},
			expectedCfg:      nil,
			expectedErrCount: 2,
		},
		// Note: Home and Name tests are done in TestValidate
		{
			name:             "download bin bad",
			envVals:          []string{absPath, "testname", "bad", "false", "true", "303", "1"},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "download bin not set",
			envVals:          []string{absPath, "testname", "", "false", "true", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", false, false, true, 303, 1),
			expectedErrCount: 0,
		},
		{
			name:             "download bin true",
			envVals:          []string{absPath, "testname", "true", "false", "true", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", true, false, true, 303, 1),
			expectedErrCount: 0,
		},
		{
			name:             "download bin false",
			envVals:          []string{absPath, "testname", "false", "false", "true", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", false, false, true, 303, 1),
			expectedErrCount: 0,
		},
		// EnvHome, EnvName, EnvDownloadBin, EnvRestartUpgrade, EnvSkipBackup, EnvInterval, EnvPreupgradeMaxRetries
		{
			name:             "restart upgrade bad",
			envVals:          []string{absPath, "testname", "true", "bad", "true", "303", "1"},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "restart upgrade not set",
			envVals:          []string{absPath, "testname", "true", "", "true", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", true, true, true, 303, 1),
			expectedErrCount: 0,
		},
		{
			name:             "restart upgrade true",
			envVals:          []string{absPath, "testname", "true", "true", "true", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", true, true, true, 303, 1),
			expectedErrCount: 0,
		},
		{
			name:             "restart upgrade true",
			envVals:          []string{absPath, "testname", "true", "false", "true", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", true, false, true, 303, 1),
			expectedErrCount: 0,
		},
		// EnvHome, EnvName, EnvDownloadBin, EnvRestartUpgrade, EnvSkipBackup, EnvInterval, EnvPreupgradeMaxRetries
		{
			name:             "skip unsafe backups bad",
			envVals:          []string{absPath, "testname", "true", "false", "bad", "303", "1"},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "skip unsafe backups not set",
			envVals:          []string{absPath, "testname", "true", "false", "", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", true, false, false, 303, 1),
			expectedErrCount: 0,
		},
		{
			name:             "skip unsafe backups true",
			envVals:          []string{absPath, "testname", "true", "false", "true", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", true, false, true, 303, 1),
			expectedErrCount: 0,
		},
		{
			name:             "skip unsafe backups false",
			envVals:          []string{absPath, "testname", "true", "false", "false", "303", "1"},
			expectedCfg:      newConfig(absPath, "testname", true, false, false, 303, 1),
			expectedErrCount: 0,
		},
		// EnvHome, EnvName, EnvDownloadBin, EnvRestartUpgrade, EnvSkipBackup, EnvInterval, EnvPreupgradeMaxRetries
		{
			name:             "poll interval bad",
			envVals:          []string{absPath, "testname", "false", "false", "false", "bad", "1"},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "poll interval 0",
			envVals:          []string{absPath, "testname", "false", "false", "false", "0", "1"},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "poll interval not set",
			envVals:          []string{absPath, "testname", "false", "false", "false", "", "1"},
			expectedCfg:      newConfig(absPath, "testname", false, false, false, 300, 1),
			expectedErrCount: 0,
		},
		{
			name:             "poll interval 987",
			envVals:          []string{absPath, "testname", "false", "false", "false", "987", "1"},
			expectedCfg:      newConfig(absPath, "testname", false, false, false, 987, 1),
			expectedErrCount: 0,
		},
		// EnvHome, EnvName, EnvDownloadBin, EnvRestartUpgrade, EnvSkipBackup, EnvInterval, EnvPreupgradeMaxRetries
		{
			name:             "prepupgrade max retries bad",
			envVals:          []string{absPath, "testname", "false", "false", "false", "406", "bad"},
			expectedCfg:      nil,
			expectedErrCount: 1,
		},
		{
			name:             "prepupgrade max retries 0",
			envVals:          []string{absPath, "testname", "false", "false", "false", "406", "0"},
			expectedCfg:      newConfig(absPath, "testname", false, false, false, 406, 0),
			expectedErrCount: 0,
		},
		{
			name:             "prepupgrade max retries not set",
			envVals:          []string{absPath, "testname", "false", "false", "false", "406", ""},
			expectedCfg:      newConfig(absPath, "testname", false, false, false, 406, 0),
			expectedErrCount: 0,
		},
		{
			name:             "prepupgrade max retries 5",
			envVals:          []string{absPath, "testname", "false", "false", "false", "406", "5"},
			expectedCfg:      newConfig(absPath, "testname", false, false, false, 406, 5),
			expectedErrCount: 0,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			s.setEnv(t, tc.envVals)
			cfg, err := GetConfigFromEnv()
			if tc.expectedErrCount == 0 {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					errCount := 1
					if multi, isMulti := err.(*errors.MultiError); isMulti {
						errCount = multi.Len()
					}
					assert.Equal(t, tc.expectedErrCount, errCount, "error count")
				}
			}
			assert.Equal(t, tc.expectedCfg, cfg, "config")
		})
	}
}
