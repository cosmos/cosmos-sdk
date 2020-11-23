package keeper_test

import (
	"sort"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestHistoricalInfo() {
	addrDels := simapp.AddTestAddrsIncremental(app, suite.chainA.GetContext(), 50, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	validators := make([]stakingtypes.Validator, len(addrVals))

	for i, valAddr := range addrVals {
		validators[i] = teststaking.NewValidator(valAddr, PKs[i])
	}

	hi := types.NewHistoricalInfo(suite.chainA.GetContext().BlockHeader(), validators)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetHistoricalInfo(suite.chainA.GetContext(), 2, &hi)

	recv, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetHistoricalInfo(suite.chainA.GetContext(), 2)
	suite.Require().True(found, "HistoricalInfo not found after set")
	suite.Require().Equal(hi, recv, "HistoricalInfo not equal")
	suite.Require().True(sort.IsSorted(stakingtypes.ValidatorsByVotingPower(recv.Valset)), "HistoricalInfo validators is not sorted")

	suite.chainA.App.IBCKeeper.ClientKeeper.DeleteHistoricalInfo(suite.chainA.GetContext(), 2)

	recv, found = suite.chainA.App.IBCKeeper.ClientKeeper.GetHistoricalInfo(suite.chainA.GetContext(), 2)
	suite.Require().False(found, "HistoricalInfo found after delete")
	suite.Require().Equal(types.HistoricalInfo{}, recv, "HistoricalInfo is not empty")
}

func (suite *KeeperTestSuite) TestTrackHistoricalInfo() {
	addrDels := simapp.AddTestAddrsIncremental(app, suite.chainA.GetContext(), 50, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	// set historical entries in params to 5
	params := types.DefaultParams()
	params.HistoricalEntries = 5
	suite.chainA.App.IBCKeeper.ClientKeeper.SetParams(suite.chainA.GetContext(), params)

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
	valSet := []stakingtypes.Validator{
		teststaking.NewValidator(addrVals[0], PKs[0]),
		teststaking.NewValidator(addrVals[1], PKs[1]),
	}
	hi4 := types.NewHistoricalInfo(h4, valSet)
	hi5 := types.NewHistoricalInfo(h5, valSet)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetHistoricalInfo(suite.chainA.GetContext(), 4, &hi4)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetHistoricalInfo(suite.chainA.GetContext(), 5, &hi5)
	recv, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetHistoricalInfo(suite.chainA.GetContext(), 4)
	suite.Require().True(found)
	suite.Require().Equal(hi4, recv)
	recv, found = suite.chainA.App.IBCKeeper.ClientKeeper.GetHistoricalInfo(suite.chainA.GetContext(), 5)
	suite.Require().True(found)
	suite.Require().Equal(hi5, recv)

	// Set bonded validators in keeper
	val1 := teststaking.NewValidator(addrVals[2], PKs[2])
	suite.chainA.App.StakingKeeper.SetValidator(suite.chainA.GetContext(), val1)
	suite.chainA.App.StakingKeeper.SetLastValidatorPower(suite.chainA.GetContext(), val1.GetOperator(), 10)
	val2 := teststaking.NewValidator(addrVals[3], PKs[3])
	suite.chainA.App.StakingKeeper.SetValidator(suite.chainA.GetContext(), val2)
	suite.chainA.App.StakingKeeper.SetLastValidatorPower(suite.chainA.GetContext(), val2.GetOperator(), 80)

	vals := []stakingtypes.Validator{val1, val2}
	sort.Sort(stakingtypes.ValidatorsByVotingPower(vals))

	// Set Header for BeginBlock context
	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  10,
	}
	suite.chainA.GetContext() = suite.chainA.GetContext().WithBlockHeader(header)

	suite.chainA.App.IBCKeeper.ClientKeeper.TrackHistoricalInfo(suite.chainA.GetContext())

	// Check HistoricalInfo at height 10 is persisted
	expected := types.HistoricalInfo{
		Header: header,
		Valset: vals,
	}
	recv, found = suite.chainA.App.IBCKeeper.ClientKeeper.GetHistoricalInfo(suite.chainA.GetContext(), 10)
	suite.Require().True(found, "GetHistoricalInfo failed after BeginBlock")
	suite.Require().Equal(expected, recv, "GetHistoricalInfo returned eunexpected result")

	// Check HistoricalInfo at height 5, 4 is pruned
	recv, found = suite.chainA.App.IBCKeeper.ClientKeeper.GetHistoricalInfo(suite.chainA.GetContext(), 4)
	suite.Require().False(found, "GetHistoricalInfo did not prune earlier height")
	suite.Require().Equal(types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 4 is not empty after prune")
	recv, found = suite.chainA.App.IBCKeeper.ClientKeeper.GetHistoricalInfo(suite.chainA.GetContext(), 5)
	suite.Require().False(found, "GetHistoricalInfo did not prune first prune height")
	suite.Require().Equal(types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 5 is not empty after prune")
}

func (suite *KeeperTestSuite) TestGetAllHistoricalInfo() {
	addrDels := simapp.AddTestAddrsIncremental(app, suite.chainA.GetContext(), 50, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	valSet := []stakingtypes.Validator{
		teststaking.NewValidator(addrVals[0], PKs[0]),
		teststaking.NewValidator(addrVals[1], PKs[1]),
	}

	header1 := tmproto.Header{ChainID: "HelloChain", Height: 10}
	header2 := tmproto.Header{ChainID: "HelloChain", Height: 11}
	header3 := tmproto.Header{ChainID: "HelloChain", Height: 12}

	hist1 := types.HistoricalInfo{Header: header1, Valset: valSet}
	hist2 := types.HistoricalInfo{Header: header2, Valset: valSet}
	hist3 := types.HistoricalInfo{Header: header3, Valset: valSet}

	expHistInfos := []types.HistoricalInfo{hist1, hist2, hist3}

	for i, hi := range expHistInfos {
		suite.chainA.App.IBCKeeper.ClientKeeper.SetHistoricalInfo(suite.chainA.GetContext(), int64(10+i), &hi)
	}

	infos := suite.chainA.App.IBCKeeper.ClientKeeper.GetAllHistoricalInfo(suite.chainA.GetContext())
	suite.Require().Equal(expHistInfos, infos)
}
