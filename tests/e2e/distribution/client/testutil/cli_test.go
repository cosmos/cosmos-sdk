//go:build e2e
// +build e2e

package testutil

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/distribution/client/testutil"
)

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(testutil.IntegrationTestSuite))
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(testutil.GRPCQueryTestSuite))
}

func TestWithdrawAllSuite(t *testing.T) {
	suite.Run(t, new(testutil.WithdrawAllTestSuite))
}
