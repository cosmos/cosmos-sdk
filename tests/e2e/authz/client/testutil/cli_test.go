//go:build e2e
// +build e2e

package evidence

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
<<<<<<< HEAD:tests/e2e/authz/client/testutil/cli_test.go
	clienttestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
=======
>>>>>>> refactor(evidence): CLI tests using Tendermint Mock (#13056):tests/e2e/evidence/cli_test.go
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
