package cmd

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type HelpTestSuite struct {
	suite.Suite

	envVars []string
}

func TestHelpTestSuite(t *testing.T) {
	suite.Run(t, new(HelpTestSuite))
}

func (s *HelpTestSuite) SetupSuite() {
	s.envVars = []string{
		cosmovisor.EnvHome, cosmovisor.EnvName, cosmovisor.EnvDownloadBin, cosmovisor.EnvRestartUpgrade,
		cosmovisor.EnvSkipBackup, cosmovisor.EnvInterval, cosmovisor.EnvPreupgradeMaxRetries,
	}
}

// clearEnv clears environment variables and returns the values
// in the same order as the entries in s.envVars.
// Designed to be used like this:
//    initialEnv := clearEnv()
//    defer setEnv(nil, initialEnv)
func (s *HelpTestSuite) clearEnv() []string {
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
func (s *HelpTestSuite) setEnv(t *testing.T, envVals ...string) {
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

func (s *HelpTestSuite) TestShouldGiveHelpEnvVars() {
	initialEnv := s.clearEnv()
	defer s.setEnv(nil, initialEnv...)

	emptyVal := ""
	homeVal := "/somehome"
	nameVal := "somename"

	tests := []struct {
		name     string
		envHome  *string
		envName  *string
		expected bool
	}{
		{
			name:     "home set name set",
			envHome:  &homeVal,
			envName:  &nameVal,
			expected: false,
		},
		{
			name:     "home not set name not set",
			envHome:  nil,
			envName:  nil,
			expected: true,
		},
		{
			name:     "home empty name not set",
			envHome:  &emptyVal,
			envName:  nil,
			expected: true,
		},
		{
			name:     "home set name not set",
			envHome:  &homeVal,
			envName:  nil,
			expected: true,
		},
		{
			name:     "home not set name empty",
			envHome:  nil,
			envName:  &emptyVal,
			expected: true,
		},
		{
			name:     "home empty name empty",
			envHome:  &emptyVal,
			envName:  &emptyVal,
			expected: true,
		},
		{
			name:     "home set name empty",
			envHome:  &homeVal,
			envName:  &emptyVal,
			expected: true,
		},
		{
			name:     "home not set name set",
			envHome:  nil,
			envName:  &nameVal,
			expected: true,
		},
		{
			name:     "home empty name set",
			envHome:  &emptyVal,
			envName:  &nameVal,
			expected: true,
		},
	}

	prepEnv := func(t *testing.T, envVar string, envVal *string) {
		if envVal == nil {
			require.NoError(t, os.Unsetenv(cosmovisor.EnvHome))
		} else {
			require.NoError(t, os.Setenv(envVar, *envVal))
		}
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			prepEnv(t, cosmovisor.EnvHome, tc.envHome)
			prepEnv(t, cosmovisor.EnvName, tc.envName)
			actual := ShouldGiveHelp(nil)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func (s HelpTestSuite) TestShouldGiveHelpArgs() {
	initialEnv := s.clearEnv()
	defer s.setEnv(nil, initialEnv...)

	s.setEnv(s.T(), "/testhome", "testname")

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "nil args",
			args:     nil,
			expected: false,
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: false,
		},
		{
			name:     "one arg random",
			args:     []string{"random"},
			expected: false,
		},
		{
			name:     "five args random",
			args:     []string{"random1", "--random2", "-r", "random4", "-random5"},
			expected: false,
		},
		{
			name:     "one arg help",
			args:     []string{"help"},
			expected: true,
		},
		{
			name:     " two args help first",
			args:     []string{"help", "arg2"},
			expected: true,
		},
		{
			name:     "two args help second",
			args:     []string{"arg1", "help"},
			expected: false,
		},
		{
			name:     "one arg -h",
			args:     []string{"-h"},
			expected: true,
		},
		{
			name:     "two args -h first",
			args:     []string{"-h", "arg2"},
			expected: true,
		},
		{
			name:     "two args -h second",
			args:     []string{"arg1", "-h"},
			expected: true,
		},
		{
			name:     "one arg --help",
			args:     []string{"--help"},
			expected: true,
		},
		{
			name:     "two args --help first",
			args:     []string{"--help", "arg2"},
			expected: true,
		},
		{
			name:     "two args --help second",
			args:     []string{"arg1", "--help"},
			expected: true,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := ShouldGiveHelp(tc.args)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func (s *HelpTestSuite) TestGetHelpText() {
	expectedPieces := []string{
		"Cosmosvisor",
		cosmovisor.EnvName, cosmovisor.EnvHome,
		"https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor/README.md",
	}

	actual := GetHelpText()
	for _, piece := range expectedPieces {
		s.Assert().Contains(actual, piece)
	}
}
