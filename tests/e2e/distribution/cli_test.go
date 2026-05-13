//go:build e2e

package distribution

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, NewE2ETestSuite(false))
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, NewGRPCQueryTestSuite(false))
}

func TestWithdrawAllSuite(t *testing.T) {
	suite.Run(t, NewWithdrawAllTestSuite(false))
}
