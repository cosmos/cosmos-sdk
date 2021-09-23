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
		cosmovisor.EnvHome, cosmovisor.EnvName, cosmovisor.EnvDownloadBin,
		cosmovisor.EnvRestartUpgrade, cosmovisor.EnvSkipBackup, cosmovisor.EnvInterval,
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
func (s *HelpTestSuite) setEnv(t *testing.T, envVals []string) {
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

func (s *HelpTestSuite) TestShouldGiveHelp() {
	initialEnv := s.clearEnv()
	defer s.setEnv(nil, initialEnv)

	s.T().Run("name not set", func(t *testing.T) {
		actual := ShouldGiveHelp([]string{})
		assert.True(t, actual)
	})

	s.Require().NoError(os.Setenv(cosmovisor.EnvName, "somename"), "setting name environment variable")

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "name set nil args",
			args:     nil,
			expected: false,
		},
		{
			name:     "name set empty args",
			args:     []string{},
			expected: false,
		},
		{
			name:     "name set help first",
			args:     []string{"-h"},
			expected: true,
		},
		{
			name:     "name set -h first",
			args:     []string{"-h"},
			expected: true,
		},
		{
			name:     "name set --help first",
			args:     []string{"--help"},
			expected: true,
		},
		{
			name:     "name set help second",
			args:     []string{"arg1", "-h"},
			expected: false,
		},
		{
			name:     "name set -h second",
			args:     []string{"arg1", "-h"},
			expected: false,
		},
		{
			name:     "name set --help second",
			args:     []string{"arg1", "--help"},
			expected: false,
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
		cosmovisor.EnvHome, cosmovisor.EnvName, cosmovisor.EnvDownloadBin,
		cosmovisor.EnvRestartUpgrade, cosmovisor.EnvSkipBackup, cosmovisor.EnvInterval,
	}

	actual := GetHelpText()
	for _, piece := range expectedPieces {
		s.Assert().Contains(actual, piece)
	}
}



