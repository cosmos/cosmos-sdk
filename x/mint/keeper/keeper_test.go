package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app       *simapp.SimApp
	ctx       sdk.Context
	msgServer types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(s.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	s.app = app
	s.ctx = ctx
	s.msgServer = keeper.NewMsgServerImpl(s.app.MintKeeper)
}

func (s *IntegrationTestSuite) TestParams() {
	testCases := []struct {
		name      string
		input     types.Params
		expectErr bool
	}{
		{
			name: "set invalid params",
			input: types.Params{
				MintDenom:           sdk.DefaultBondDenom,
				InflationRateChange: sdk.NewDecWithPrec(-13, 2),
				InflationMax:        sdk.NewDecWithPrec(20, 2),
				InflationMin:        sdk.NewDecWithPrec(7, 2),
				GoalBonded:          sdk.NewDecWithPrec(67, 2),
				BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			input: types.Params{
				MintDenom:           sdk.DefaultBondDenom,
				InflationRateChange: sdk.NewDecWithPrec(8, 2),
				InflationMax:        sdk.NewDecWithPrec(20, 2),
				InflationMin:        sdk.NewDecWithPrec(2, 2),
				GoalBonded:          sdk.NewDecWithPrec(37, 2),
				BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			expected := s.app.MintKeeper.GetParams(s.ctx)
			err := s.app.MintKeeper.SetParams(s.ctx, tc.input)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				expected = tc.input
				s.Require().NoError(err)
			}

			p := s.app.MintKeeper.GetParams(s.ctx)
			s.Require().Equal(expected, p)
		})
	}
}
