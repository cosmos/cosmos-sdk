package keeper_test

import (
	"time"

	"cosmossdk.io/collections"
	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/testutil"
	stakingtypes "cosmossdk.io/x/staking/types"
)

// IsValSetSorted reports whether valset is sorted.
func IsValSetSorted(data []stakingtypes.Validator, powerReduction math.Int) bool {
	n := len(data)
	for i := n - 1; i > 0; i-- {
		if stakingtypes.ValidatorsByVotingPower(data).Less(i, i-1, powerReduction) {
			return false
		}
	}
	return true
}

func (s *KeeperTestSuite) TestHistoricalInfo() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	_, addrVals := createValAddrs(50)

	validators := make([]stakingtypes.Validator, len(addrVals))

	for i, valAddr := range addrVals {
		validators[i] = testutil.NewValidator(s.T(), valAddr, PKs[i])
	}

	time := ctx.BlockTime()
	hi := stakingtypes.HistoricalRecord{
		Time:           &time,
		ValidatorsHash: ctx.CometInfo().ValidatorsHash,
		Apphash:        ctx.HeaderInfo().AppHash,
	}
	require.NoError(keeper.HistoricalInfo.Set(ctx, uint64(2), hi))

	recv, err := keeper.HistoricalInfo.Get(ctx, uint64(2))
	require.NoError(err, "HistoricalInfo found after set")
	require.Equal(hi, recv, "HistoricalInfo not equal")

	require.NoError(keeper.HistoricalInfo.Remove(ctx, uint64(2)))

	recv, err = keeper.HistoricalInfo.Get(ctx, uint64(2))
	require.ErrorIs(err, collections.ErrNotFound, "HistoricalInfo not found after delete")
	require.Equal(stakingtypes.HistoricalRecord{}, recv, "HistoricalInfo is not empty")
}

func (s *KeeperTestSuite) TestTrackHistoricalInfo() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	_, addrVals := createValAddrs(50)

	// set historical entries in params to 5
	params := stakingtypes.DefaultParams()
	params.HistoricalEntries = 5
	require.NoError(keeper.Params.Set(ctx, params))

	// set historical info at 5, 4 which should be pruned
	// and check that it has been stored
	t := time.Now().Round(0).UTC()
	hi4 := stakingtypes.HistoricalRecord{
		Time:           &t,
		ValidatorsHash: []byte("validatorHash"),
		Apphash:        []byte("AppHash"),
	}

	hi5 := stakingtypes.HistoricalRecord{
		Time:           &t,
		ValidatorsHash: []byte("validatorHash"),
		Apphash:        []byte("AppHash"),
	}

	require.NoError(keeper.HistoricalInfo.Set(ctx, uint64(4), hi4))
	require.NoError(keeper.HistoricalInfo.Set(ctx, uint64(5), hi5))
	recv, err := keeper.HistoricalInfo.Get(ctx, uint64(4))
	require.NoError(err)
	require.Equal(hi4, recv)
	recv, err = keeper.HistoricalInfo.Get(ctx, uint64(5))
	require.NoError(err)
	require.Equal(hi5, recv)

	// Set bonded validators in keeper
	val1 := testutil.NewValidator(s.T(), addrVals[2], PKs[2])
	val1.Status = stakingtypes.Bonded // when not bonded, consensus power is Zero
	val1.Tokens = keeper.TokensFromConsensusPower(ctx, 10)
	require.NoError(keeper.SetValidator(ctx, val1))
	valbz, err := keeper.ValidatorAddressCodec().StringToBytes(val1.GetOperator())
	require.NoError(err)
	require.NoError(keeper.SetLastValidatorPower(ctx, valbz, 10))
	val2 := testutil.NewValidator(s.T(), addrVals[3], PKs[3])
	val1.Status = stakingtypes.Bonded
	val2.Tokens = keeper.TokensFromConsensusPower(ctx, 80)
	require.NoError(keeper.SetValidator(ctx, val2))
	valbz, err = keeper.ValidatorAddressCodec().StringToBytes(val2.GetOperator())
	require.NoError(err)
	require.NoError(keeper.SetLastValidatorPower(ctx, valbz, 80))

	vals := []stakingtypes.Validator{val1, val2}
	require.True(IsValSetSorted(vals, keeper.PowerReduction(ctx)))

	// Set Header for BeginBlock context
	ctx = ctx.WithHeaderInfo(coreheader.Info{
		ChainID: "HelloChain",
		Height:  10,
		Time:    t,
	})

	require.NoError(keeper.TrackHistoricalInfo(ctx))

	// Check HistoricalInfo at height 10 is persisted
	expected := stakingtypes.HistoricalRecord{
		Time:           &t,
		ValidatorsHash: ctx.CometInfo().ValidatorsHash,
		Apphash:        ctx.HeaderInfo().AppHash,
	}
	recv, err = keeper.HistoricalInfo.Get(ctx, uint64(10))
	require.NoError(err, "GetHistoricalInfo failed after BeginBlock")
	require.Equal(expected, recv, "GetHistoricalInfo returned unexpected result")

	// Check HistoricalInfo at height 5, 4 is pruned
	recv, err = keeper.HistoricalInfo.Get(ctx, uint64(4))
	require.ErrorIs(err, collections.ErrNotFound, "GetHistoricalInfo did not prune earlier height")
	require.Equal(stakingtypes.HistoricalRecord{}, recv, "GetHistoricalInfo at height 4 is not empty after prune")
	recv, err = keeper.HistoricalInfo.Get(ctx, uint64(5))
	require.ErrorIs(err, collections.ErrNotFound, "GetHistoricalInfo did not prune first prune height")
	require.Equal(stakingtypes.HistoricalRecord{}, recv, "GetHistoricalInfo at height 5 is not empty after prune")
}

func (s *KeeperTestSuite) TestGetAllHistoricalInfo() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	t := time.Now().Round(0).UTC()

	hist1 := stakingtypes.HistoricalRecord{
		Time:           &t,
		ValidatorsHash: nil,
		Apphash:        nil,
	}
	hist2 := stakingtypes.HistoricalRecord{
		Time:           &t,
		ValidatorsHash: nil,
		Apphash:        nil,
	}
	hist3 := stakingtypes.HistoricalRecord{
		Time:           &t,
		ValidatorsHash: nil,
		Apphash:        nil,
	}

	expHistInfos := []stakingtypes.HistoricalRecord{hist1, hist2, hist3}

	for i, hi := range expHistInfos {
		require.NoError(keeper.HistoricalInfo.Set(ctx, uint64(int64(9+i)), hi))
	}

	var infos []stakingtypes.HistoricalRecord
	err := keeper.HistoricalInfo.Walk(ctx, nil, func(key uint64, info stakingtypes.HistoricalRecord) (stop bool, err error) {
		infos = append(infos, info)
		return false, nil
	})

	require.NoError(err)
	require.Equal(expHistInfos, infos)
}
