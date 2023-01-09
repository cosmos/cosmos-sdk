//go:build e2e
// +build e2e

package evidence

import (
	"testing"

	"cosmossdk.io/simapp"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/x/evidence/client/cli"
	"gotest.tools/v3/assert"
)

type fixture struct {
	cfg     network.Config
	network *network.Network
}

func initFixture(t *testing.T) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1

	network, err := network.New(t, t.TempDir(), cfg)
	assert.NilError(t, err)
	assert.NilError(t, network.WaitForNextBlock())

	return &fixture{
		cfg:     cfg,
		network: network,
	}
}

func TestGetQueryCmd(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	defer f.network.Cleanup()

	val := s.network.Validators[0]

	testCases := map[string]struct {
		args           []string
		expectedOutput string
		expectErr      bool
	}{
		"non-existent evidence": {
			[]string{"DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660"},
			"evidence DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660 not found",
			true,
		},
		"all evidence (default pagination)": {
			[]string{},
			"evidence: []\npagination:\n  next_key: null\n  total: \"0\"",
			false,
		},
	}

	for name, tc := range testCases {
		tc := tc

		s.Run(name, func(t. *testing.T) {
			cmd := cli.GetQueryCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}

			s.Require().Contains(strings.TrimSpace(out.String()), tc.expectedOutput)
		})
	}
}
