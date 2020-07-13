package cmd

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	crisistestutil "github.com/cosmos/cosmos-sdk/x/crisis/client/testutil"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/client/testutil"
)

// TestIntegrationTestSuites runs integration tests for all the Cosmos SDK modules.
// Apps can re-use the integration test suites for the modules they import.
func TestIntegrationTestSuites(t *testing.T) {
	// because of the way testify test suites are setup, this should run test
	// suites in parallel
	t.Parallel()

	// apps should setup their own config based on the default config, but at
	// a minimum changing the AppConstructor to use their own app
	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	// run the integration test suite for each module used in the app
	suite.Run(t, banktestutil.NewIntegrationTestSuite(cfg))
	suite.Run(t, crisistestutil.NewIntegrationTestSuite(cfg))
	suite.Run(t, distrtestutil.NewIntegrationTestSuite(cfg))
}
