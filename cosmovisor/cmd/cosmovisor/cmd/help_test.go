package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

type HelpTestSuite struct {
	suite.Suite
}

func TestHelpTestSuite(t *testing.T) {
	suite.Run(t, new(HelpTestSuite))
}

// cosmovisorHelpEnv are some string values of environment variables used to configure Cosmovisor.
type cosmovisorHelpEnv struct {
	Home string
	Name string
}

// ToMap creates a map of the cosmovisorHelpEnv where the keys are the env var names.
func (c cosmovisorHelpEnv) ToMap() map[string]string {
	return map[string]string{
		cosmovisor.EnvHome: c.Home,
		cosmovisor.EnvName: c.Name,
	}
}

// Set sets the field in this cosmovisorHelpEnv corresponding to the provided envVar to the given envVal.
func (c *cosmovisorHelpEnv) Set(envVar, envVal string) {
	switch envVar {
	case cosmovisor.EnvHome:
		c.Home = envVal
	case cosmovisor.EnvName:
		c.Name = envVal
	default:
		panic(fmt.Errorf("Unknown environment variable [%s]. Ccannot set field to [%s]. ", envVar, envVal))
	}
}

// clearEnv clears environment variables and returns what they were.
// Designed to be used like this:
//    initialEnv := clearEnv()
//    defer setEnv(nil, initialEnv)
func (s *HelpTestSuite) clearEnv() *cosmovisorHelpEnv {
	s.T().Logf("Clearing environment variables.")
	rv := cosmovisorHelpEnv{}
	for envVar := range rv.ToMap() {
		rv.Set(envVar, os.Getenv(envVar))
		s.Require().NoError(os.Unsetenv(envVar))
	}
	return &rv
}

// setEnv sets environment variables to the values provided.
// If t is not nil, and there's a problem, the test will fail immediately.
// If t is nil, problems will just be logged using s.T().
func (s *HelpTestSuite) setEnv(t *testing.T, env *cosmovisorHelpEnv) {
	if t == nil {
		s.T().Logf("Restoring environment variables.")
	}
	for envVar, envVal := range env.ToMap() {
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
	defer s.setEnv(nil, initialEnv)

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
	defer s.setEnv(nil, initialEnv)

	s.setEnv(s.T(), &cosmovisorHelpEnv{"/testhome", "testname"})

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
			expected: true,
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
