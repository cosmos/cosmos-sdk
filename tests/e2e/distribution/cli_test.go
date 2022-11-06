//go:build e2e
// +build e2e

package distribution

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCQueryTestSuite))
}

func TestWithdrawAllSuite(t *testing.T) {
	suite.Run(t, new(WithdrawAllTestSuite))
}
