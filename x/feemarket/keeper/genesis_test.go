package keeper_test

import (
	"cosmossdk.io/x/feemarket/types"
)

func (s *KeeperTestSuite) TestInitGenesis() {
	s.Run("default genesis should not panic", func() {
		s.Require().NotPanics(func() {
			s.feeMarketKeeper.InitGenesis(s.ctx, *types.DefaultGenesisState())
		})
	})

	s.Run("default AIMD genesis should not panic", func() {
		s.Require().NotPanics(func() {
			s.feeMarketKeeper.InitGenesis(s.ctx, *types.DefaultAIMDGenesisState())
		})
	})

	s.Run("bad genesis state should panic", func() {
		gs := types.DefaultGenesisState()
		gs.Params.Window = 0
		s.Require().Panics(func() {
			s.feeMarketKeeper.InitGenesis(s.ctx, *gs)
		})
	})

	s.Run("mismatch in params and state for window should panic", func() {
		gs := types.DefaultAIMDGenesisState()
		gs.Params.Window = 1

		s.Require().Panics(func() {
			s.feeMarketKeeper.InitGenesis(s.ctx, *gs)
		})
	})
}

func (s *KeeperTestSuite) TestExportGenesis() {
	s.Run("export genesis should not panic for default eip-1559", func() {
		gs := types.DefaultGenesisState()
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)

		var exportedGenesis *types.GenesisState
		s.Require().NotPanics(func() {
			exportedGenesis = s.feeMarketKeeper.ExportGenesis(s.ctx)
		})

		s.Require().Equal(gs, exportedGenesis)
	})

	s.Run("export genesis should not panic for default AIMD eip-1559", func() {
		gs := types.DefaultAIMDGenesisState()
		s.feeMarketKeeper.InitGenesis(s.ctx, *gs)

		var exportedGenesis *types.GenesisState
		s.Require().NotPanics(func() {
			exportedGenesis = s.feeMarketKeeper.ExportGenesis(s.ctx)
		})

		s.Require().Equal(gs, exportedGenesis)
	})
}
