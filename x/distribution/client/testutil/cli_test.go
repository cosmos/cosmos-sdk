//go:build norace
// +build norace

package testutil

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.NumValidators = 1
	suite.Run(t, NewIntegrationTestSuite(cfg))
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCQueryTestSuite))
}

func TestWithdrawAllSuite(t *testing.T) {
	cfg1 := network.DefaultConfig()
	cfg1.NumValidators = 2
	suite.Run(t, NewWithdrawAllTestSuite(cfg1))
}
