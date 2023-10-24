//go:build e2e
// +build e2e

package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	v1 "cosmossdk.io/x/gov/types/v1"

	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func TestE2ETestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	suite.Run(t, NewE2ETestSuite(cfg))
}

func TestDepositTestSuite(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	genesisState := v1.DefaultGenesisState()
	maxDepPeriod := time.Duration(20) * time.Second
	votingPeriod := time.Duration(8) * time.Second
	genesisState.Params.MaxDepositPeriod = &maxDepPeriod
	genesisState.Params.VotingPeriod = &votingPeriod
	bz, err := cfg.Codec.MarshalJSON(genesisState)
	require.NoError(t, err)
	cfg.GenesisState["gov"] = bz
	suite.Run(t, NewDepositTestSuite(cfg))
}
