package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/epochs/types"
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
		"start in future": {
			addedEpochInfo: types.EpochInfo{
				Identifier:              defaultIdentifier,
				StartTime:               startBlockTime.Add(time.Hour),
				Duration:                defaultDuration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
			expEpochInfo: types.EpochInfo{
				Identifier:              defaultIdentifier,
				StartTime:               startBlockTime.Add(time.Hour),
				Duration:                defaultDuration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
			expErr: false,
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockHeight(startBlockHeight).WithBlockTime(startBlockTime)
			err := s.EpochsKeeper.AddEpochInfo(s.Ctx, test.addedEpochInfo)
			if !test.expErr {
				s.Require().NoError(err)
				actualEpochInfo, err := s.EpochsKeeper.EpochInfo.Get(s.Ctx, test.addedEpochInfo.Identifier)
				s.Require().NoError(err)
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
	err := s.EpochsKeeper.AddEpochInfo(s.Ctx, epochInfo)
	s.Require().NoError(err)
	epochInfoSaved, err := s.EpochsKeeper.EpochInfo.Get(s.Ctx, "monthly")
	s.Require().NoError(err)
	// setup expected epoch info
	expectedEpochInfo := epochInfo
	expectedEpochInfo.StartTime = s.Ctx.BlockTime()
	expectedEpochInfo.CurrentEpochStartHeight = s.Ctx.BlockHeight()
	s.Require().Equal(expectedEpochInfo, epochInfoSaved)

	allEpochs, err := s.EpochsKeeper.AllEpochInfos(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(allEpochs, 5)
	s.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	s.Require().Equal(allEpochs[1].Identifier, "hour")
	s.Require().Equal(allEpochs[2].Identifier, "minute")
	s.Require().Equal(allEpochs[3].Identifier, "monthly")
	s.Require().Equal(allEpochs[4].Identifier, "week")
}

func (s *KeeperTestSuite) TestNumBlocksSinceEpochStart() {
	s.SetupTest()

	startBlockHeight := int64(100)
	startBlockTime := time.Unix(1656907200, 0).UTC()
	duration := time.Hour

	s.Ctx = s.Ctx.WithBlockHeight(startBlockHeight).WithBlockTime(startBlockTime)

	tests := map[string]struct {
		setupEpoch        types.EpochInfo
		advanceBlockDelta int64
		advanceTimeDelta  time.Duration
		expErr            bool
		expBlocksSince    int64
	}{
		"same block as start": {
			setupEpoch: types.EpochInfo{
				Identifier:              "epoch_same_block",
				StartTime:               startBlockTime,
				Duration:                duration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: startBlockHeight,
				CurrentEpochStartTime:   startBlockTime,
				EpochCountingStarted:    true,
			},
			advanceBlockDelta: 0,
			advanceTimeDelta:  0,
			expErr:            false,
			expBlocksSince:    0,
		},
		"after 5 blocks": {
			setupEpoch: types.EpochInfo{
				Identifier:              "epoch_after_five",
				StartTime:               startBlockTime,
				Duration:                duration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: startBlockHeight,
				CurrentEpochStartTime:   startBlockTime,
				EpochCountingStarted:    true,
			},
			advanceBlockDelta: 5,
			advanceTimeDelta:  time.Minute * 5, // just to simulate realistic advancement
			expErr:            false,
			expBlocksSince:    5,
		},
		"epoch not started yet": {
			setupEpoch: types.EpochInfo{
				Identifier:              "epoch_future",
				StartTime:               startBlockTime.Add(time.Hour),
				Duration:                duration,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
			advanceBlockDelta: 0,
			advanceTimeDelta:  0,
			expErr:            true,
			expBlocksSince:    0,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockHeight(startBlockHeight).WithBlockTime(startBlockTime)

			err := s.EpochsKeeper.AddEpochInfo(s.Ctx, tc.setupEpoch)
			s.Require().NoError(err)

			// Advance block height and time if needed
			s.Ctx = s.Ctx.WithBlockHeight(startBlockHeight + tc.advanceBlockDelta).
				WithBlockTime(startBlockTime.Add(tc.advanceTimeDelta))

			blocksSince, err := s.EpochsKeeper.NumBlocksSinceEpochStart(s.Ctx, tc.setupEpoch.Identifier)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expBlocksSince, blocksSince)
			}
		})
	}
}
