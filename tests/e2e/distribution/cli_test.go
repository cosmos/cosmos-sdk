//go:build e2e
// +build e2e

package distribution

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, NewE2ETestSuite(false))
	suite.Run(t, NewE2ETestSuite(true))
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, NewGRPCQueryTestSuite(false))
	suite.Run(t, NewGRPCQueryTestSuite(true))
}

func TestWithdrawAllSuite(t *testing.T) {
	suite.Run(t, NewWithdrawAllTestSuite(false))
	suite.Run(t, NewWithdrawAllTestSuite(true))
}
