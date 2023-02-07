package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/tools/cosmovisor"
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
//
//	initialEnv := clearEnv()
//	defer setEnv(nil, initialEnv)
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

func (s *HelpTestSuite) TestGetHelpText() {
	expectedPieces := []string{
		"Cosmovisor",
		cosmovisor.EnvName, cosmovisor.EnvHome,
		"https://docs.cosmos.network/main/tooling/cosmovisor",
	}

	actual := GetHelpText()
	for _, piece := range expectedPieces {
		s.Assert().Contains(actual, piece)
	}
}
