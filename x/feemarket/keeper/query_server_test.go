package keeper_test

import (
	"cosmossdk.io/math"

	"cosmossdk.io/x/feemarket/types"
)

func (s *KeeperTestSuite) TestParamsRequest() {
	s.Run("can get default params", func() {
		req := &types.ParamsRequest{}
		resp, err := s.queryServer.Params(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(types.DefaultParams(), resp.Params)

		params, err := s.feeMarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.Params, params)
	})

	s.Run("can get updated params", func() {
		params := types.Params{
			Alpha:               math.LegacyMustNewDecFromStr("0.1"),
			Beta:                math.LegacyMustNewDecFromStr("0.1"),
			Gamma:               math.LegacyMustNewDecFromStr("0.1"),
			Delta:               math.LegacyMustNewDecFromStr("0.1"),
			MinBaseGasPrice:     math.LegacyNewDec(10),
			MinLearningRate:     math.LegacyMustNewDecFromStr("0.1"),
			MaxLearningRate:     math.LegacyMustNewDecFromStr("0.1"),
			MaxBlockUtilization: 10,
			Window:              1,
			Enabled:             true,
		}
		err := s.feeMarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		req := &types.ParamsRequest{}
		resp, err := s.queryServer.Params(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(params, resp.Params)

		params, err = s.feeMarketKeeper.GetParams(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.Params, params)
	})
}

func (s *KeeperTestSuite) TestStateRequest() {
	s.Run("can get default state", func() {
		req := &types.StateRequest{}
		resp, err := s.queryServer.State(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(types.DefaultState(), resp.State)

		state, err := s.feeMarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.State, state)
	})

	s.Run("can get updated state", func() {
		state := types.State{
			BaseGasPrice: math.LegacyOneDec(),
			LearningRate: math.LegacyOneDec(),
			Window:       []uint64{1},
			Index:        0,
		}
		err := s.feeMarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		req := &types.StateRequest{}
		resp, err := s.queryServer.State(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Equal(state, resp.State)

		state, err = s.feeMarketKeeper.GetState(s.ctx)
		s.Require().NoError(err)

		s.Require().Equal(resp.State, state)
	})
}

func (s *KeeperTestSuite) TestBaseFeeRequest() {
	s.Run("can get gas price", func() {
		req := &types.GasPriceRequest{
			Denom: "stake",
		}
		resp, err := s.queryServer.GasPrice(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		gasPrice, err := s.feeMarketKeeper.GetMinGasPrice(s.ctx, req.GetDenom())
		s.Require().NoError(err)

		s.Require().Equal(resp.GetPrice(), gasPrice)
	})

	s.Run("can get updated gas price", func() {
		state := types.State{
			BaseGasPrice: math.LegacyOneDec(),
		}
		err := s.feeMarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		params := types.Params{
			FeeDenom: "test",
		}
		err = s.feeMarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		req := &types.GasPriceRequest{
			Denom: "stake",
		}
		resp, err := s.queryServer.GasPrice(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		gasPrice, err := s.feeMarketKeeper.GetMinGasPrice(s.ctx, req.GetDenom())
		s.Require().NoError(err)

		s.Require().Equal(resp.GetPrice(), gasPrice)
	})

	s.Run("can get updated gas price < 1", func() {
		state := types.State{
			BaseGasPrice: math.LegacyMustNewDecFromStr("0.005"),
		}
		err := s.feeMarketKeeper.SetState(s.ctx, state)
		s.Require().NoError(err)

		params := types.Params{
			FeeDenom: "test",
		}
		err = s.feeMarketKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		req := &types.GasPriceRequest{
			Denom: "stake",
		}
		resp, err := s.queryServer.GasPrice(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		fee, err := s.feeMarketKeeper.GetMinGasPrice(s.ctx, req.GetDenom())
		s.Require().NoError(err)

		s.Require().Equal(resp.GetPrice(), fee)
	})
}
