package main

import (
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/tools/cosmovisor"
)

type HelpTestSuite struct {
	suite.Suite
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
