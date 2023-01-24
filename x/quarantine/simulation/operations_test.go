package simulation_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *simapp.SimApp
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, account.Address, initCoins))
	}

	return accounts
}

func (s *SimTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T(), false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (s *SimTestSuite) TestWeightedOperations() {
	cdc := s.app.AppCodec()
	appParams := make(simtypes.AppParams)

	expected := []struct {
		weight      int
		opsMsgRoute string
		opsMsgName  string
	}{
		{simulation.WeightMsgOptIn, simulation.TypeMsgOptIn, simulation.TypeMsgOptIn},
		{simulation.WeightMsgOptOut, simulation.TypeMsgOptOut, simulation.TypeMsgOptOut},
		{simulation.WeightMsgAccept, simulation.TypeMsgAccept, simulation.TypeMsgAccept},
		{simulation.WeightMsgDecline, simulation.TypeMsgDecline, simulation.TypeMsgDecline},
		{simulation.WeightMsgUpdateAutoResponses, simulation.TypeMsgUpdateAutoResponses, simulation.TypeMsgUpdateAutoResponses},
	}

	weightedOps := simulation.WeightedOperations(
		appParams, cdc,
		s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper, cdc,
	)

	s.Require().Len(weightedOps, len(expected), "weighted ops")

	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 10)

	for i, actual := range weightedOps {
		exp := expected[i]
		parts := strings.Split(exp.opsMsgName, ".")
		s.Run(parts[len(parts)-1], func() {
			operationMsg, futureOps, err := actual.Op()(r, s.app.BaseApp, s.ctx, accs, "")
			s.Assert().NoError(err, "op error")
			s.Assert().Equal(exp.weight, actual.Weight(), "op weight")
			s.Assert().Equal(exp.opsMsgRoute, operationMsg.Route, "op route")
			s.Assert().Equal(exp.opsMsgName, operationMsg.Name, "op name")
			s.Assert().Nil(futureOps, "future ops")
		})
	}
}

func (s *SimTestSuite) TestSimulateMsgOptIn() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	s.app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  s.app.LastBlockHeight() + 1,
			AppHash: s.app.LastCommitID().Hash,
		},
	})

	op := simulation.SimulateMsgOptIn(s.app.AccountKeeper, s.app.BankKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgOptIn
	err = s.app.AppCodec().UnmarshalJSON(opMsg.Msg, &msg)
	s.Assert().NoError(err, "UnmarshalJSON on opMsg.Msg for MsgOptIn")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Len(futureOps, 0)
}

func (s *SimTestSuite) TestSimulateMsgOptOut() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, accounts[0].Address)
	s.Require().NoError(err, "SetOptIn on accounts[0]")

	s.app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  s.app.LastBlockHeight() + 1,
			AppHash: s.app.LastCommitID().Hash,
		},
	})

	op := simulation.SimulateMsgOptOut(s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgOptOut
	err = s.app.AppCodec().UnmarshalJSON(opMsg.Msg, &msg)
	s.Assert().NoError(err, "UnmarshalJSON on opMsg.Msg for MsgOptIn")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Len(futureOps, 0)
}

func (s *SimTestSuite) TestSimulateMsgAccept() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, accounts[0].Address)
	s.Require().NoError(err, "SetOptIn on accounts[0]")
	spendableCoins := s.app.BankKeeper.SpendableCoins(s.ctx, accounts[1].Address)
	toSend, err := simtypes.RandomFees(r, s.ctx, spendableCoins)
	s.Require().NoError(err, "RandomFees(%q)", spendableCoins.String())
	err = s.app.BankKeeper.SendCoins(s.ctx, accounts[1].Address, accounts[0].Address, toSend)
	s.Require().NoError(err, "SendCoins")

	s.app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  s.app.LastBlockHeight() + 1,
			AppHash: s.app.LastCommitID().Hash,
		},
	})

	op := simulation.SimulateMsgAccept(s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgAccept
	err = s.app.AppCodec().UnmarshalJSON(opMsg.Msg, &msg)
	s.Assert().NoError(err, "UnmarshalJSON on opMsg.Msg for MsgOptIn")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Len(futureOps, 0)
}

func (s *SimTestSuite) TestSimulateMsgDecline() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, accounts[0].Address)
	s.Require().NoError(err, "SetOptIn on accounts[0]")
	spendableCoins := s.app.BankKeeper.SpendableCoins(s.ctx, accounts[1].Address)
	toSend, err := simtypes.RandomFees(r, s.ctx, spendableCoins)
	s.Require().NoError(err, "RandomFees(%q)", spendableCoins.String())
	err = s.app.BankKeeper.SendCoins(s.ctx, accounts[1].Address, accounts[0].Address, toSend)
	s.Require().NoError(err, "SendCoins")

	s.app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  s.app.LastBlockHeight() + 1,
			AppHash: s.app.LastCommitID().Hash,
		},
	})

	op := simulation.SimulateMsgDecline(s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgDecline
	err = s.app.AppCodec().UnmarshalJSON(opMsg.Msg, &msg)
	s.Assert().NoError(err, "UnmarshalJSON on opMsg.Msg for MsgOptIn")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Len(futureOps, 0)
}

func (s *SimTestSuite) TestSimulateMsgUpdateAutoResponses() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, accounts[0].Address)
	s.Require().NoError(err, "SetOptIn on accounts[0]")

	s.app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  s.app.LastBlockHeight() + 1,
			AppHash: s.app.LastCommitID().Hash,
		},
	})

	op := simulation.SimulateMsgUpdateAutoResponses(s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgUpdateAutoResponses
	err = s.app.AppCodec().UnmarshalJSON(opMsg.Msg, &msg)
	s.Assert().NoError(err, "UnmarshalJSON on opMsg.Msg for MsgOptIn")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Len(futureOps, 0)
	s.Assert().GreaterOrEqual(len(msg.Updates), 1, "number of updates")
}
