package main

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/tools/cosmovisor"
)

type HelpTestSuite struct {
	suite.Suite
}

func TestHelpTestSuite(t *testing.T) {
	suite.Run(t, new(HelpTestSuite))
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
