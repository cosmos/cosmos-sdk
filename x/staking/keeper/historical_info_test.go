package keeper_test

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// IsValSetSorted reports whether valset is sorted.
func IsValSetSorted(data []types.Validator, powerReduction sdk.Int) bool {
	n := len(data)
	for i := n - 1; i > 0; i-- {
		if types.ValidatorsByVotingPower(data).Less(i, i-1, powerReduction) {
			return false
		}
	}
	return true
}

func (suite *KeeperTestSuite) TestHistoricalInfo() {
	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 50, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	validators := make([]types.Validator, len(addrVals))

	for i, valAddr := range addrVals {
		validators[i] = teststaking.NewValidator(suite.T(), valAddr, PKs[i])
	}

	hi := types.NewHistoricalInfo(suite.ctx.BlockHeader(), validators, suite.stakingKeeper.PowerReduction(suite.ctx))
	suite.stakingKeeper.SetHistoricalInfo(suite.ctx, 2, &hi)

	recv, found := suite.stakingKeeper.GetHistoricalInfo(suite.ctx, 2)
	suite.Require().True(found, "HistoricalInfo not found after set")
	suite.Require().Equal(hi, recv, "HistoricalInfo not equal")
	suite.Require().True(IsValSetSorted(recv.Valset, suite.stakingKeeper.PowerReduction(suite.ctx)), "HistoricalInfo validators is not sorted")

	suite.stakingKeeper.DeleteHistoricalInfo(suite.ctx, 2)

	recv, found = suite.stakingKeeper.GetHistoricalInfo(suite.ctx, 2)
	suite.Require().False(found, "HistoricalInfo found after delete")
	suite.Require().Equal(types.HistoricalInfo{}, recv, "HistoricalInfo is not empty")
}

func (suite *KeeperTestSuite) TestTrackHistoricalInfo() {
	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 50, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	// set historical entries in params to 5
	params := types.DefaultParams()
	params.HistoricalEntries = 5
	suite.stakingKeeper.SetParams(suite.ctx, params)

	// set historical info at 5, 4 which should be pruned
	// and check that it has been stored
	h4 := tmproto.Header{
		ChainID: "HelloChain",
		Height:  4,
	}
	h5 := tmproto.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	valSet := []types.Validator{
		teststaking.NewValidator(suite.T(), addrVals[0], PKs[0]),
		teststaking.NewValidator(suite.T(), addrVals[1], PKs[1]),
	}
	hi4 := types.NewHistoricalInfo(h4, valSet, suite.stakingKeeper.PowerReduction(suite.ctx))
	hi5 := types.NewHistoricalInfo(h5, valSet, suite.stakingKeeper.PowerReduction(suite.ctx))
	suite.stakingKeeper.SetHistoricalInfo(suite.ctx, 4, &hi4)
	suite.stakingKeeper.SetHistoricalInfo(suite.ctx, 5, &hi5)
	recv, found := suite.stakingKeeper.GetHistoricalInfo(suite.ctx, 4)
	suite.Require().True(found)
	suite.Require().Equal(hi4, recv)
	recv, found = suite.stakingKeeper.GetHistoricalInfo(suite.ctx, 5)
	suite.Require().True(found)
	suite.Require().Equal(hi5, recv)

	// genesis validator
	genesisVals := suite.stakingKeeper.GetAllValidators(suite.ctx)
	suite.Require().Len(genesisVals, 1)

	// Set bonded validators in keeper
	val1 := teststaking.NewValidator(suite.T(), addrVals[2], PKs[2])
	val1.Status = types.Bonded // when not bonded, consensus power is Zero
	val1.Tokens = suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	suite.stakingKeeper.SetValidator(suite.ctx, val1)
	suite.stakingKeeper.SetLastValidatorPower(suite.ctx, val1.GetOperator(), 10)
	val2 := teststaking.NewValidator(suite.T(), addrVals[3], PKs[3])
	val1.Status = types.Bonded
	val2.Tokens = suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 80)
	suite.stakingKeeper.SetValidator(suite.ctx, val2)
	suite.stakingKeeper.SetLastValidatorPower(suite.ctx, val2.GetOperator(), 80)

	vals := []types.Validator{val1, genesisVals[0], val2}
	suite.Require().True(IsValSetSorted(vals, suite.stakingKeeper.PowerReduction(suite.ctx)))

	// Set Header for BeginBlock context
	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  10,
	}

	ctx := suite.ctx.WithBlockHeader(header)

	suite.stakingKeeper.TrackHistoricalInfo(ctx)

	// Check HistoricalInfo at height 10 is persisted
	expected := types.HistoricalInfo{
		Header: header,
		Valset: vals,
	}
	recv, found = suite.stakingKeeper.GetHistoricalInfo(ctx, 10)
	suite.Require().True(found, "GetHistoricalInfo failed after BeginBlock")
	suite.Require().Equal(expected, recv, "GetHistoricalInfo returned unexpected result")

	// Check HistoricalInfo at height 5, 4 is pruned
	recv, found = suite.stakingKeeper.GetHistoricalInfo(ctx, 4)
	suite.Require().False(found, "GetHistoricalInfo did not prune earlier height")
	suite.Require().Equal(types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 4 is not empty after prune")
	recv, found = suite.stakingKeeper.GetHistoricalInfo(ctx, 5)
	suite.Require().False(found, "GetHistoricalInfo did not prune first prune height")
	suite.Require().Equal(types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 5 is not empty after prune")
}

func (suite *KeeperTestSuite) TestGetAllHistoricalInfo() {
	// clear historical info
	infos := suite.stakingKeeper.GetAllHistoricalInfo(suite.ctx)
	suite.Require().Len(infos, 1)
	suite.stakingKeeper.DeleteHistoricalInfo(suite.ctx, infos[0].Header.Height)

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 50, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	valSet := []types.Validator{
		teststaking.NewValidator(suite.T(), addrVals[0], PKs[0]),
		teststaking.NewValidator(suite.T(), addrVals[1], PKs[1]),
	}

	header1 := tmproto.Header{ChainID: "HelloChain", Height: 10}
	header2 := tmproto.Header{ChainID: "HelloChain", Height: 11}
	header3 := tmproto.Header{ChainID: "HelloChain", Height: 12}

	hist1 := types.HistoricalInfo{Header: header1, Valset: valSet}
	hist2 := types.HistoricalInfo{Header: header2, Valset: valSet}
	hist3 := types.HistoricalInfo{Header: header3, Valset: valSet}

	expHistInfos := []types.HistoricalInfo{hist1, hist2, hist3}

	for i, hi := range expHistInfos {
		suite.stakingKeeper.SetHistoricalInfo(suite.ctx, int64(10+i), &hi)
	}

	infos = suite.stakingKeeper.GetAllHistoricalInfo(suite.ctx)
	suite.Require().Equal(expHistInfos, infos)
}
