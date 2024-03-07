package keeper_test

import (
	"time"

	"cosmossdk.io/x/epochs/types"
	"cosmossdk.io/core/header"
)

func (s *KeeperTestSuite) TestAddEpochInfo() {
	defaultIdentifier := "default_add_epoch_info_id"
	defaultDuration := time.Hour
	startBlockHeight := int64(100)
	startBlockTime := time.Unix(1656907200, 0).UTC()
	tests := map[string]struct {
		addedEpochInfo types.EpochInfo
		expErr         bool
		expEpochInfo   types.EpochInfo
	}{
		"simple_add": {
			addedEpochInfo: types.EpochInfo{
				Identifier:              defaultIdentifier,
				StartTime:               time.Time{},
				Duration:                defaultDuration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
			expErr: false,
			expEpochInfo: types.EpochInfo{
				Identifier:              defaultIdentifier,
				StartTime:               startBlockTime,
				Duration:                defaultDuration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: startBlockHeight,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
		},
		"zero_duration": {
			addedEpochInfo: types.EpochInfo{
				Identifier:              defaultIdentifier,
				StartTime:               time.Time{},
				Duration:                time.Duration(0),
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
			expErr: true,
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithHeaderInfo(header.Info{Height: startBlockHeight, Time: startBlockTime})
			err := s.EpochsKeeper.AddEpochInfo(s.Ctx, test.addedEpochInfo)
			if !test.expErr {
				s.Require().NoError(err)
				actualEpochInfo := s.EpochsKeeper.GetEpochInfo(s.Ctx, test.addedEpochInfo.Identifier)
				s.Require().Equal(test.expEpochInfo, actualEpochInfo)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestDuplicateAddEpochInfo() {
	identifier := "duplicate_add_epoch_info"
	epochInfo := types.NewGenesisEpochInfo(identifier, time.Hour*24*30)
	err := s.EpochsKeeper.AddEpochInfo(s.Ctx, epochInfo)
	s.Require().NoError(err)
	err = s.EpochsKeeper.AddEpochInfo(s.Ctx, epochInfo)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestEpochLifeCycle() {
	s.SetupTest()

	epochInfo := types.NewGenesisEpochInfo("monthly", time.Hour*24*30)
	s.EpochsKeeper.AddEpochInfo(s.Ctx, epochInfo)
	epochInfoSaved := s.EpochsKeeper.GetEpochInfo(s.Ctx, "monthly")
	// setup expected epoch info
	expectedEpochInfo := epochInfo
	expectedEpochInfo.StartTime = s.Ctx.BlockTime()
	expectedEpochInfo.CurrentEpochStartHeight = s.Ctx.BlockHeight()
	s.Require().Equal(expectedEpochInfo, epochInfoSaved)

	allEpochs := s.EpochsKeeper.AllEpochInfos(s.Ctx)
	s.Require().Len(allEpochs, 4)
	s.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	s.Require().Equal(allEpochs[1].Identifier, "hour")
	s.Require().Equal(allEpochs[2].Identifier, "monthly")
	s.Require().Equal(allEpochs[3].Identifier, "week")
}
