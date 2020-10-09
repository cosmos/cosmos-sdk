package cli_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctmcli "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/client/cli"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestCreateClientCmd() {
	val := s.network.Validators[0]

	// TODO: figure out how to generate correct json bytes for consensus state
	// add tests for:
	// - empty client id
	// - invalid header
	// - invalid contents in file
	// - invalid/emtpy fraction
	// - successful parsed fraction
	// - invalid trusting period
	// - invalid unbonding period
	// - invalid max clock drift
	// - invalid consensus params
	// - invalid proof specs
	// - invalid contents for proof specs file

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"success",
			[]string{
				"clientid",
				"header todo",
				ibctesting.TrustingPeriod.String(),
				ibctesting.UnbondingPeriod.String(),
				ibctesting.MaxClockDrift.String(),
				ibctesting.DefaultConsensusParams.String(),
				fmt.Sprintf("--%s=%s", ibctmcli.FlagTrustLevel, "default"),
				fmt.Sprintf("--%s=%s", ibctmcli.FlagProofSpecs, "default"),
				fmt.Sprintf("--%s=%s", ibctmcli.FlagUpgradePath, "upgrade/upgradedClient"),
				fmt.Sprintf("--%s=true", ibctmcli.FlagAllowUpdateAfterExpiry),
				fmt.Sprintf("--%s=true", ibctmcli.FlagAllowUpdateAfterMisbehaviour),
			},
			false,
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			bz, err := clitestutil.ExecTestCLICmd(clientCtx, ibctmcli.NewCreateClientCmd(), tc.args)

			s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(bz.Bytes(), tc.respType), bz.String())

			txResp := tc.respType.(*sdk.TxResponse)
			s.Require().Equal(tc.expectedCode, txResp.Code)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
