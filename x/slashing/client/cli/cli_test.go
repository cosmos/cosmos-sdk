package cli_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

// SetupSuite executes bootstrapping logic before all the tests, i.e. once before
// the entire suite, start executing.
func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownSuite performs cleanup logic after all the tests, i.e. once after the
// entire suite, has finished executing.
func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// func TestCLISlashingGetParams(t *testing.T) {
// 	t.SkipNow() // TODO: Bring back once viper is refactored.
// 	t.Parallel()
// 	f := cli.InitFixtures(t)

// 	// start simd server
// 	proc := f.SDStart()
// 	t.Cleanup(func() { proc.Stop(false) })

// 	params := testutil.QuerySlashingParams(f)
// 	require.Equal(t, int64(100), params.SignedBlocksWindow)
// 	require.Equal(t, sdk.NewDecWithPrec(5, 1), params.MinSignedPerWindow)

// 	sinfo := testutil.QuerySigningInfo(f, f.SDTendermint("show-validator"))
// 	require.Equal(t, int64(0), sinfo.StartHeight)
// 	require.False(t, sinfo.Tombstoned)

// 	// Cleanup testing directories
// 	f.Cleanup()
// }
