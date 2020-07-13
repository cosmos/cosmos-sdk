package cmd

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	crisistestutil "github.com/cosmos/cosmos-sdk/x/crisis/client/testutil"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/client/testutil"
)

func TestIntegrationTestSuites(t *testing.T) {
	t.Parallel()

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	suite.Run(t, banktestutil.NewIntegrationTestSuite(cfg))
	suite.Run(t, crisistestutil.NewIntegrationTestSuite(cfg))
	suite.Run(t, distrtestutil.NewIntegrationTestSuite(cfg))
}
