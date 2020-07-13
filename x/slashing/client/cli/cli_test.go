package cli_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
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

func (s *IntegrationTestSuite) TestGetCmdQuerySigningInfo() {
	val := s.network.Validators[0]

	valConsPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, val.PubKey)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  fmt.Stringer
		expected  fmt.Stringer
	}{
		{"invalid address", []string{"foo"}, true, nil, nil},
		{
			"valid address",
			[]string{
				valConsPubKey,
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&types.ValidatorSigningInfo{},
			types.ValidatorSigningInfo{
				Address:             sdk.ConsAddress(val.PubKey.Address()),
				StartHeight:         0,
				IndexOffset:         0,
				JailedUntil:         time.Unix(0, 0).UTC(),
				Tombstoned:          false,
				MissedBlocksCounter: 0,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQuerySigningInfo()
			_, out := testutil.ApplyMockIO(cmd)

			clientCtx := val.ClientCtx.WithOutput(out)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

			out.Reset()
			cmd.SetArgs(tc.args)

			err := cmd.ExecuteContext(ctx)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
