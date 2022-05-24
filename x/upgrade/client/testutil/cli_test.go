package testutil

import (
	"github.com/Stride-Labs/cosmos-sdk/testutil/network"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.NumValidators = 1
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
