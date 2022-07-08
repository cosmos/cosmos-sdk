package keeper_test

import (
	"testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	"github.com/cosmos/cosmos-sdk/x/crisis/testutil"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	// app    *simapp.SimApp
	ctx    sdk.Context
	keeper keeper.Keeper
}

func (s *KeeperTestSuite) SetupSuite(t *testing.T) {
	app, err := simtestutil.Setup(testutil.AppConfig,
		&s.keeper,
	)
	s.Require().NoError(err)

	// app := simapp.Setup(t, false)
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1}})
	ctx := app.NewContext(true, tmproto.Header{})

	// s.app = app
	s.ctx = ctx
}

func (s *KeeperTestSuite) TestMsgUpdateParams() {
	// default params
	constantFee := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)) // 4%

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority:   "invalid",
				ConstantFee: constantFee,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid constant fee",
			input: &types.MsgUpdateParams{
				Authority:   s.keeper.GetAuthority(),
				ConstantFee: sdk.Coin{},
			},
			expErr: true,
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority:   s.keeper.GetAuthority(),
				ConstantFee: constantFee,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.keeper.UpdateParams(s.ctx, tc.input)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
