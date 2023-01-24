package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/simulation"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *simapp.SimApp
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

func (s *SimTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T(), false)
	s.freshCtx()
}

// freshCtx creates a new context and sets it to this SimTestSuite's ctx field.
func (s *SimTestSuite) freshCtx() {
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

// createTestingAccounts creates testing accounts with a default balance.
func (s *SimTestSuite) createTestingAccounts(r *rand.Rand, count int) []simtypes.Account {
	return s.createTestingAccountsWithPower(r, count, 200)
}

// createTestingAccountsWithPower creates new accounts with the specified power (coins amount).
func (s *SimTestSuite) createTestingAccountsWithPower(r *rand.Rand, count int, power int64) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, count)

	initAmt := sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(bankutil.FundAccount(s.app.BankKeeper, s.ctx, account.Address, initCoins))
	}

	return accounts
}

// setSanctionParamsAboveGovDeposit looks up the x/gov min deposit and sets the
// sanction params to be larger by 5 (for sanction) and 10 (for unsanction).
// If there's no gov min dep, sets params to 5stake and 10stake respectively.
func (s *SimTestSuite) setSanctionParamsAboveGovDeposit() {
	sancParams := &sanction.Params{
		ImmediateSanctionMinDeposit:   nil,
		ImmediateUnsanctionMinDeposit: nil,
	}

	for _, coin := range s.app.GovKeeper.GetDepositParams(s.ctx).MinDeposit {
		sanctCoin := sdk.NewCoin(coin.Denom, coin.Amount.AddRaw(5))
		unsanctCoin := sdk.NewCoin(coin.Denom, coin.Amount.AddRaw(10))
		sancParams.ImmediateSanctionMinDeposit = sancParams.ImmediateSanctionMinDeposit.Add(sanctCoin)
		sancParams.ImmediateUnsanctionMinDeposit = sancParams.ImmediateUnsanctionMinDeposit.Add(unsanctCoin)
	}

	if sancParams.ImmediateSanctionMinDeposit.IsZero() {
		sancParams.ImmediateSanctionMinDeposit = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}
	}
	if sancParams.ImmediateUnsanctionMinDeposit.IsZero() {
		sancParams.ImmediateUnsanctionMinDeposit = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)}
	}

	s.Require().NoError(s.app.SanctionKeeper.SetParams(s.ctx, sancParams), "SanctionKeeper.SetParams")
}

// getLastGovProp gets the last gov prop to be submitted.
func (s *SimTestSuite) getLastGovProp() *govv1.Proposal {
	props := s.app.GovKeeper.GetProposals(s.ctx)
	if len(props) == 0 {
		return nil
	}
	return props[len(props)-1]
}

// getWeightedOpsArgs creates a standard WeightedOpsArgs.
func (s *SimTestSuite) getWeightedOpsArgs() simulation.WeightedOpsArgs {
	return simulation.WeightedOpsArgs{
		AppParams:  make(simtypes.AppParams),
		JSONCodec:  s.app.AppCodec(),
		ProtoCodec: codec.NewProtoCodec(s.app.InterfaceRegistry()),
		AK:         s.app.AccountKeeper,
		BK:         s.app.BankKeeper,
		GK:         s.app.GovKeeper,
		SK:         &s.app.SanctionKeeper,
	}
}

// nextBlock ends the current block, commits it, and starts a new one.
// This is needed because some tests would run out of gas if all done in a single block.
func (s *SimTestSuite) nextBlock() {
	s.Require().NotPanics(func() { s.app.EndBlock(abci.RequestEndBlock{}) }, "app.EndBlock")
	s.Require().NotPanics(func() { s.app.Commit() }, "app.Commit")
	s.Require().NotPanics(func() {
		s.app.BeginBlock(abci.RequestBeginBlock{
			Header: tmproto.Header{
				Height: s.app.LastBlockHeight() + 1,
			},
			LastCommitInfo:      abci.LastCommitInfo{},
			ByzantineValidators: nil,
		})
	}, "app.BeginBlock")
	s.freshCtx()
}

func (s *SimTestSuite) TestWeightedOperations() {
	s.setSanctionParamsAboveGovDeposit()
	testutil.RequireNotPanicsNoError(s.T(), func() error {
		accs := simtypes.RandomAccounts(rand.New(rand.NewSource(500)), 10)
		addrs := make([]sdk.AccAddress, len(accs))
		for i, acc := range accs {
			addrs[i] = acc.Address
		}
		return s.app.SanctionKeeper.SanctionAddresses(s.ctx, addrs...)
	})

	govPropType := sdk.MsgTypeURL(&govv1.MsgSubmitProposal{})

	expected := []struct {
		comment string
		weight  int
	}{
		{comment: "sanction", weight: simulation.DefaultWeightSanction},
		{comment: "immediate sanction", weight: simulation.DefaultWeightSanctionImmediate},
		{comment: "unsanction", weight: simulation.DefaultWeightUnsanction},
		{comment: "immediate unsanction", weight: simulation.DefaultWeightUnsanctionImmediate},
		{comment: "update params", weight: simulation.DefaultWeightUpdateParams},
	}

	weightedOps := simulation.WeightedOperations(
		make(simtypes.AppParams), s.app.AppCodec(), codec.NewProtoCodec(s.app.InterfaceRegistry()),
		s.app.AccountKeeper, s.app.BankKeeper, s.app.GovKeeper, s.app.SanctionKeeper,
	)

	s.Require().Len(weightedOps, len(expected), "weighted ops")

	accountCount := 10
	r := rand.New(rand.NewSource(1))
	accs := s.createTestingAccounts(r, accountCount)

	for i, actual := range weightedOps {
		exp := expected[i]
		s.Run(exp.comment, func() {
			var operationMsg simtypes.OperationMsg
			var futureOps []simtypes.FutureOperation
			var err error
			testFunc := func() {
				operationMsg, futureOps, err = actual.Op()(r, s.app.BaseApp, s.ctx, accs, "")
			}
			s.Require().NotPanics(testFunc, "calling op")
			s.T().Logf("operationMsg.Msg: %s", operationMsg.Msg)
			s.Assert().NoError(err, "op error")
			s.Assert().Equal(exp.weight, actual.Weight(), "op weight")
			s.Assert().True(operationMsg.OK, "op msg ok")
			s.Assert().Equal(exp.comment, operationMsg.Comment, "op msg comment")
			s.Assert().Equal("gov", operationMsg.Route, "op msg route")
			s.Assert().Equal(govPropType, operationMsg.Name, "op msg name")
			s.Assert().Len(futureOps, accountCount, "future ops")
			// Note: As of writing this, the content of operationMsg.Msg comes from MsgSubmitProposal.GetSignBytes.
			// But for some reason, it's also wrapped in '{"type":"{msg.Type}","value":"{msg.GetSignBytes}"}'.
			// The sign bytes are json, but the MsgSubmitProposal.Messages field's json marshals as just the value
			// instead of the Any that it is (i.e. there's no type_url). That makes it impossible to know from
			// that operationMsg.Msg field what type of messages are in the proposal Messages.
			// For this specific case, both MsgSanction and MsgUnsanction look exactly the same,
			// it's just: '{"addresses":[...]}'
			// So, long story short (too late), there's nothing worthwhile to check in the operationMsg.Msg field.
		})
	}
}

func (s *SimTestSuite) TestSendGovMsg() {
	r := rand.New(rand.NewSource(1))
	accounts := s.createTestingAccounts(r, 10)
	accounts = append(accounts, s.createTestingAccountsWithPower(r, 1, 0)...)
	accounts = append(accounts, s.createTestingAccountsWithPower(r, 1, 1)...)
	acctZero := accounts[len(accounts)-2]
	acctOne := accounts[len(accounts)-1]
	acctOneBalance := s.app.BankKeeper.SpendableCoins(s.ctx, acctOne.Address)
	var acctOneBalancePlusOne sdk.Coins
	for _, c := range acctOneBalance {
		acctOneBalancePlusOne = acctOneBalancePlusOne.Add(sdk.NewCoin(c.Denom, c.Amount.AddRaw(1)))
	}

	tests := []struct {
		name            string
		sender          simtypes.Account
		msg             sdk.Msg
		deposit         sdk.Coins
		comment         string
		expSkip         bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
		expInErr        []string
	}{
		{
			name:   "no spendable coins",
			sender: acctZero,
			msg: &sanction.MsgSanction{
				Addresses: []string{accounts[4].Address.String(), accounts[5].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "should not matter",
			expSkip:         true,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "sender has no spendable coins",
			expInErr:        nil,
		},
		{
			name:   "not enough coins for deposit",
			sender: acctOne,
			msg: &sanction.MsgSanction{
				Addresses: []string{accounts[5].Address.String(), accounts[6].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			deposit:         acctOneBalancePlusOne,
			comment:         "should not be this",
			expSkip:         true,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "sender has insufficient balance to cover deposit",
			expInErr:        nil,
		},
		{
			name:            "nil msg",
			sender:          accounts[0],
			msg:             nil,
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "will not get returned",
			expSkip:         true,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    "/",
			expOpMsgComment: "wrapping MsgSanction as Any",
			expInErr:        []string{"Expecting non nil value to create a new Any", "failed packing protobuf message to Any"},
		},
		{
			name: "gen and deliver returns error",
			sender: simtypes.Account{
				PrivKey: accounts[0].PrivKey,
				PubKey:  acctOne.PubKey,
				Address: acctOne.Address,
				ConsKey: accounts[0].ConsKey,
			},
			msg: &sanction.MsgSanction{
				Addresses: []string{accounts[6].Address.String(), accounts[7].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			deposit:         acctOneBalance,
			comment:         "this should be ignored",
			expSkip:         true,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "unable to deliver tx",
			expInErr:        []string{"pubKey does not match signer address", "invalid pubkey"},
		},
		{
			name:   "all good",
			sender: accounts[1],
			msg: &sanction.MsgSanction{
				Addresses: []string{accounts[2].Address.String(), accounts[3].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "this is a test comment",
			expSkip:         false,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "this is a test comment",
			expInErr:        nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			args := &simulation.SendGovMsgArgs{
				WeightedOpsArgs: s.getWeightedOpsArgs(),
				R:               rand.New(rand.NewSource(1)),
				App:             s.app.BaseApp,
				Ctx:             s.ctx,
				Accs:            accounts,
				ChainID:         "send-gov-test",
				Sender:          tc.sender,
				Msg:             tc.msg,
				Deposit:         tc.deposit,
				Comment:         tc.comment,
			}

			var skip bool
			var opMsg simtypes.OperationMsg
			var err error
			testFunc := func() {
				skip, opMsg, err = simulation.SendGovMsg(args)
			}
			s.Require().NotPanics(testFunc, "SendGovMsg")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "SendGovMsg error")
			s.Assert().Equal(tc.expSkip, skip, "SendGovMsg result skip bool")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "SendGovMsg result op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "SendGovMsg result op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "SendGovMsg result op msg comment")
			if !tc.expSkip && !skip {
				// If we don't expect a skip, and we didn't get one,
				// get the last gov prop and make sure it's the one we just sent.
				expMsgs := []sdk.Msg{tc.msg}
				prop := s.getLastGovProp()
				if s.Assert().NotNil(prop, "last gov prop") {
					msgs, err := prop.GetMsgs()
					if s.Assert().NoError(err, "error from prop.GetMsgs() on the last gov prop") {
						s.Assert().Equal(expMsgs, msgs, "messages in the last gov prop")
					}
				}
			}
		})
	}
}

func (s *SimTestSuite) TestOperationMsgVote() {
	s.setSanctionParamsAboveGovDeposit()

	// I don't think the ChainID matters in here, but just for consistency...
	chainID := "test-op-msg-vote"

	sancParams := s.app.SanctionKeeper.GetParams(s.ctx)
	sanctMinDep := sancParams.ImmediateSanctionMinDeposit
	unsanctMinDep := sancParams.ImmediateUnsanctionMinDeposit

	r := rand.New(rand.NewSource(1))
	accounts := s.createTestingAccounts(r, 10)

	// Create a couple gov props that we can vote on.
	// Note that I'm sending enough deposit for them to be immediate.
	// That shouldn't come into play, but if weird things are happening in here....
	var skip bool
	var opMsg simtypes.OperationMsg
	var err error
	testSendGovSanct := func() {
		skip, opMsg, err = simulation.SendGovMsg(&simulation.SendGovMsgArgs{
			WeightedOpsArgs: s.getWeightedOpsArgs(),
			R:               r,
			App:             s.app.BaseApp,
			Ctx:             s.ctx,
			Accs:            accounts,
			ChainID:         chainID,
			Sender:          accounts[0],
			Msg: &sanction.MsgSanction{
				Addresses: []string{accounts[8].Address.String(), accounts[9].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			Deposit: sanctMinDep,
			Comment: "sanction",
		})
	}
	testSendGovUnsanct := func() {
		skip, opMsg, err = simulation.SendGovMsg(&simulation.SendGovMsgArgs{
			WeightedOpsArgs: s.getWeightedOpsArgs(),
			R:               r,
			App:             s.app.BaseApp,
			Ctx:             s.ctx,
			Accs:            accounts,
			ChainID:         chainID,
			Sender:          accounts[0],
			Msg: &sanction.MsgUnsanction{
				Addresses: []string{accounts[6].Address.String(), accounts[7].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			Deposit: unsanctMinDep,
			Comment: "unsanction",
		})
	}

	s.Require().NotPanics(testSendGovSanct, "SendGovMsg with MsgSanction")
	s.Require().NoError(err, "SendGovMsg with MsgSanction result error")
	s.Require().False(skip, "SendGovMsg with MsgSanction result skip")
	s.Require().Equal("sanction", opMsg.Comment, "SendGovMsg with MsgSanction result op msg comment")
	govPropSanct := s.getLastGovProp()

	s.Require().NotPanics(testSendGovUnsanct, "SendGovMsg with MsgUnsanction")
	s.Require().NoError(err, "SendGovMsg with MsgUnsanction result error")
	s.Require().False(skip, "SendGovMsg with MsgUnsanction result skip")
	s.Require().Equal("unsanction", opMsg.Comment, "SendGovMsg with MsgUnsanction result op msg comment")
	govPropUnsanct := s.getLastGovProp()

	tests := []struct {
		name            string
		voter           simtypes.Account
		govPropID       uint64
		vote            govv1.VoteOption
		comment         string
		expInErr        []string
		expOpMsgOK      bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
	}{
		{
			name: "gen and deliver returns error",
			voter: simtypes.Account{
				PrivKey: accounts[1].PrivKey,
				PubKey:  accounts[0].PubKey,
				Address: accounts[0].Address,
				ConsKey: accounts[1].ConsKey,
			},
			govPropID:       govPropSanct.Id,
			vote:            govv1.OptionYes,
			comment:         "this should be ignored",
			expInErr:        []string{"pubKey does not match signer address", "invalid pubkey"},
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgVote{}),
			expOpMsgComment: "unable to deliver tx",
		},
		{
			name:            "sanction yes",
			voter:           accounts[0],
			govPropID:       govPropSanct.Id,
			vote:            govv1.OptionYes,
			comment:         "sanction-yes",
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgVote{}),
			expOpMsgComment: "sanction-yes",
		},
		{
			name:            "sanction no",
			voter:           accounts[1],
			govPropID:       govPropSanct.Id,
			vote:            govv1.OptionNo,
			comment:         "sanction-no",
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgVote{}),
			expOpMsgComment: "sanction-no",
		},
		{
			name:            "unsanction yes",
			voter:           accounts[0],
			govPropID:       govPropUnsanct.Id,
			vote:            govv1.OptionYes,
			comment:         "unsanction-yes",
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgVote{}),
			expOpMsgComment: "unsanction-yes",
		},
		{
			name:            "unsanction no",
			voter:           accounts[1],
			govPropID:       govPropUnsanct.Id,
			vote:            govv1.OptionNo,
			comment:         "unsanction-no",
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgVote{}),
			expOpMsgComment: "unsanction-no",
		},
		{
			// Since we sent enough deposit to make it immediate,
			// accounts[9] should have at least a temp sanction.
			// So it shouldn't be able to pay the fees on any message.
			name:            "attempt to vote from a sanctioned account",
			voter:           accounts[9],
			govPropID:       govPropSanct.Id,
			vote:            govv1.OptionNo,
			comment:         "don't sanction me bro",
			expInErr:        []string{"account is sanctioned", "insufficient funds"},
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgVote{}),
			expOpMsgComment: "unable to deliver tx",
		},
	}

	wopArgs := s.getWeightedOpsArgs()

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var op simtypes.Operation
			testFunc := func() {
				op = simulation.OperationMsgVote(&wopArgs, tc.voter, tc.govPropID, tc.vote, tc.comment)
			}
			s.Require().NotPanics(testFunc, "OperationMsgVote")
			var fops []simtypes.FutureOperation
			testOp := func() {
				opMsg, fops, err = op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, accounts, chainID)
			}
			s.Require().NotPanics(testOp, "calling Operation returned by OperationMsgVote")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "op error")
			s.Assert().Equal(tc.expOpMsgOK, opMsg.OK, "op msg ok")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "op msg comment")
			if tc.expOpMsgOK && opMsg.OK {
				// If we were expecting a success and there was a success,
				// get the prop again and check that the vote went through.
				vote, found := s.app.GovKeeper.GetVote(s.ctx, tc.govPropID, tc.voter.Address)
				if s.Assert().True(found, "GetVote(%d) found bool", tc.govPropID) {
					if s.Assert().Len(vote.Options, 1, "vote options") {
						s.Assert().Equal(tc.vote, vote.Options[0].Option, "vote option")
						s.Assert().Equal("1.000000000000000000", vote.Options[0].Weight, "vote option weight")
					}
				}
			}
			s.Assert().Empty(fops, "future ops")
		})
	}
}

func TestMaxCoins(t *testing.T) {
	// Not using SimTestSuite for this one since it doesn't need the infrastructure.

	// cz is a short way to convert a string to Coins.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		a    sdk.Coins
		b    sdk.Coins
		exp  sdk.Coins
	}{
		{
			name: "nil nil",
			a:    nil,
			b:    nil,
			exp:  sdk.Coins{},
		},
		{
			name: "nil empty",
			a:    nil,
			b:    sdk.Coins{},
			exp:  sdk.Coins{},
		},
		{
			name: "empty nil",
			a:    sdk.Coins{},
			b:    nil,
			exp:  sdk.Coins{},
		},
		{
			name: "empty empty",
			a:    sdk.Coins{},
			b:    sdk.Coins{},
			exp:  sdk.Coins{},
		},
		{
			name: "one denom nil",
			a:    cz("5acoin"),
			b:    nil,
			exp:  cz("5acoin"),
		},
		{
			name: "one denom empty",
			a:    cz("5acoin"),
			b:    sdk.Coins{},
			exp:  cz("5acoin"),
		},
		{
			name: "nil one denom",
			a:    nil,
			b:    cz("3bcoin"),
			exp:  cz("3bcoin"),
		},
		{
			name: "empty one denom",
			a:    sdk.Coins{},
			b:    cz("3bcoin"),
			exp:  cz("3bcoin"),
		},
		{
			name: "two denoms nil",
			a:    cz("1aone,2atwo"),
			b:    nil,
			exp:  cz("1aone,2atwo"),
		},
		{
			name: "two denoms empty",
			a:    cz("1aone,2atwo"),
			b:    sdk.Coins{},
			exp:  cz("1aone,2atwo"),
		},
		{
			name: "nil two denoms",
			a:    nil,
			b:    cz("4bone,5btwo"),
			exp:  cz("4bone,5btwo"),
		},
		{
			name: "empty two denoms",
			a:    sdk.Coins{},
			b:    cz("4bone,5btwo"),
			exp:  cz("4bone,5btwo"),
		},
		{
			name: "different denoms",
			a:    cz("99acoin"),
			b:    cz("101bcoin"),
			exp:  cz("99acoin,101bcoin"),
		},
		{
			name: "both have same denom a bigger",
			a:    cz("2sharecoin"),
			b:    cz("1sharecoin"),
			exp:  cz("2sharecoin"),
		},
		{
			name: "both have same denom b bigger",
			a:    cz("4sharecoin"),
			b:    cz("5sharecoin"),
			exp:  cz("5sharecoin"),
		},
		{
			name: "each with unique denoms",
			a:    cz("3aonecoin,8atwocoin"),
			b:    cz("4bonecoin,9btwocoin"),
			exp:  cz("3aonecoin,8atwocoin,4bonecoin,9btwocoin"),
		},
		{
			name: "one denom smaller vs two denoms",
			a:    cz("1share"),
			b:    cz("2bcoin,2share"),
			exp:  cz("2bcoin,2share"),
		},
		{
			name: "one denom larger vs two denoms",
			a:    cz("3share"),
			b:    cz("2bcoin,2share"),
			exp:  cz("2bcoin,3share"),
		},
		{
			name: "two denoms vs one denom smaller",
			a:    cz("2acoin,2share"),
			b:    cz("1share"),
			exp:  cz("2acoin,2share"),
		},
		{
			name: "two denoms vs one denom larger",
			a:    cz("2acoin,2share"),
			b:    cz("3share"),
			exp:  cz("2acoin,3share"),
		},
		{
			name: "multiple denoms one shared a bigger",
			a:    cz("9aonlycoin,22sharecoin"),
			b:    cz("6bonlycoin,7bonlytwo,21sharecoin"),
			exp:  cz("9aonlycoin,6bonlycoin,7bonlytwo,22sharecoin"),
		},
		{
			name: "multiple denoms one shared b bigger",
			a:    cz("9aonlycoin,22sharecoin"),
			b:    cz("6bonlycoin,7bonlytwo,23sharecoin"),
			exp:  cz("9aonlycoin,6bonlycoin,7bonlytwo,23sharecoin"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = simulation.MaxCoins(tc.a, tc.b)
			}
			require.NotPanics(t, testFunc, "MaxCoins")
			assert.Equal(t, tc.exp.String(), actual.String(), "MaxCoins result")
		})
	}
}

func (s *SimTestSuite) TestSimulateGovMsgSanction() {
	chainID := "test-simulate-gov-msg-sanction"
	votingPeriod := 2 * time.Minute
	depositPeriod := 1 * time.Second
	govMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 3))
	sanctMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 7))
	unsanctMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))

	// resetParams resets the params to the values defined above.
	resetParams := func(t *testing.T, ctx sdk.Context) {
		require.NotPanics(t, func() {
			s.app.GovKeeper.SetVotingParams(ctx, govv1.VotingParams{VotingPeriod: &votingPeriod})
		}, "gov SetVotingParams")
		require.NotPanics(t, func() {
			s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{
				MinDeposit:       govMinDep,
				MaxDepositPeriod: &depositPeriod,
			})
		}, "gov SetDepositParams")
		testutil.RequireNotPanicsNoError(t, func() error {
			return s.app.SanctionKeeper.SetParams(ctx, &sanction.Params{
				ImmediateSanctionMinDeposit:   sanctMinDep,
				ImmediateUnsanctionMinDeposit: unsanctMinDep,
			})
		}, "sanction SetParams")
	}

	// Create a random number generator for use only in generating accounts.
	accsRand := rand.New(rand.NewSource(1))

	tests := []struct {
		name            string
		setup           func(t *testing.T, ctx sdk.Context)
		accs            []simtypes.Account
		expInErr        []string
		expOpMsgOK      bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
	}{
		{
			name: "gov min deposit equals immediate sanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NotPanics(t, func() {
					s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{MinDeposit: sanctMinDep})
				}, "gov SetDepositParams")
			},
			accs:            simtypes.RandomAccounts(accsRand, 10),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "cannot sanction without it being immediate",
		},
		{
			name: "gov min deposit greater than immediate sanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NotPanics(t, func() {
					s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{MinDeposit: sanctMinDep.Add(govMinDep...)})
				}, "gov SetDepositParams")
			},
			accs:            simtypes.RandomAccounts(accsRand, 10),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "cannot sanction without it being immediate",
		},
		{
			name:            "problem sending gov msg",
			accs:            s.createTestingAccountsWithPower(accsRand, 10, 0),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "sender has no spendable coins",
		},
		{
			name:            "all good",
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "sanction",
		},
	}

	wopArgs := s.getWeightedOpsArgs()
	voteType := sdk.MsgTypeURL(&govv1.MsgVote{})

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resetParams(s.T(), s.ctx)
			if tc.setup != nil {
				tc.setup(s.T(), s.ctx)
			}
			var op simtypes.Operation
			testFunc := func() {
				op = simulation.SimulateGovMsgSanction(&wopArgs)
			}
			s.Require().NotPanics(testFunc, "SimulateGovMsgSanction")
			var opMsg simtypes.OperationMsg
			var fops []simtypes.FutureOperation
			var err error
			testOp := func() {
				opMsg, fops, err = op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, tc.accs, chainID)
			}
			s.Require().NotPanics(testOp, "SimulateGovMsgSanction op execution")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "op error")
			s.Assert().Equal(tc.expOpMsgOK, opMsg.OK, "op msg ok")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "op msg comment")
			if !tc.expOpMsgOK && !opMsg.OK {
				s.Assert().Empty(fops, "future ops")
			}
			if tc.expOpMsgOK && opMsg.OK {
				s.Assert().Equal(len(tc.accs), len(fops), "number of future ops")
				// If we were expecting it to be okay, and it was, run all the future ops too.
				// Some of them might fail (due to being sanctioned),
				// but all the ones that went through should be YES votes.
				maxBlockTime := s.ctx.BlockHeader().Time.Add(votingPeriod)
				prop := s.getLastGovProp()
				s.Assert().Equal(govMinDep.String(), sdk.NewCoins(prop.TotalDeposit...).String(), "prop deposit")
				preVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
				// There shouldn't be any votes yet.
				if !s.Assert().Empty(preVotes, "votes before running future ops") {
					for i, fop := range fops {
						s.Assert().LessOrEqual(fop.BlockTime, maxBlockTime, "future op %d block time", i+1)
						s.Assert().Equal(0, fop.BlockHeight, "future op %d block height", i+1)
						var fopMsg simtypes.OperationMsg
						var ffops []simtypes.FutureOperation
						testFop := func() {
							fopMsg, ffops, err = fop.Op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, tc.accs, chainID)
						}
						if !s.Assert().NotPanics(testFop, "future op %d execution", i+1) {
							continue
						}
						if err != nil {
							s.T().Logf("future op %d returned an error, but that's kind of expected: %v", i+1, err)
							continue
						}
						if !fopMsg.OK {
							s.T().Logf("future op %d returned not okay, but that's kind of expected: %q", i+1, fopMsg.Comment)
							continue
						}
						s.Assert().Empty(ffops, "future ops returned by future op %d", i+1)
						s.Assert().Equal(voteType, fopMsg.Name, "future op %d msg name", i+1)
						s.Assert().Equal(tc.expOpMsgComment, fopMsg.Comment, "future op %d msg comment", i+1)
					}
					// Now there should be some votes.
					postVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
					for i, vote := range postVotes {
						if s.Assert().Len(vote.Options, 1, "vote %d options count", i+1) {
							s.Assert().Equal(govv1.OptionYes, vote.Options[0].Option, "vote %d option", i+1)
							s.Assert().Equal("1.000000000000000000", vote.Options[1].Weight, "vote %d weight", i+1)
						}
					}
				}
				// Now, get the message and count the number of addresses listed that were provided.
				providedAddrs := make(map[string]bool)
				for _, acc := range tc.accs {
					providedAddrs[acc.Address.String()] = true
				}
				msgs, err := prop.GetMsgs()
				if s.Assert().NoError(err, "getting messages from the proposal") {
					if s.Assert().Len(msgs, 1, "number of messages in the proposal") {
						msg, ok := msgs[0].(*sanction.MsgSanction)
						if s.Assert().True(ok, "could not cast prop msg to MsgSanction") {
							s.Assert().NotEmpty(msg.Addresses, "msg Addresses")
							var inMsg []string
							for _, addr := range msg.Addresses {
								if providedAddrs[addr] {
									inMsg = append(inMsg, addr)
								}
							}
							s.Assert().Empty(inMsg, "provided accs that ended up in the gov prop msg")
						}
					}
				}
			}
			s.nextBlock()
		})
	}
}

func (s *SimTestSuite) TestSimulateGovMsgSanctionImmediate() {
	chainID := "test-simulate-gov-msg-immediate-sanction"
	votingPeriod := 2 * time.Minute
	depositPeriod := 1 * time.Second
	govMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 3))
	sanctMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 7))
	unsanctMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))

	// resetParams resets the params to the values defined above.
	resetParams := func(t *testing.T, ctx sdk.Context) {
		require.NotPanics(t, func() {
			s.app.GovKeeper.SetVotingParams(ctx, govv1.VotingParams{VotingPeriod: &votingPeriod})
		}, "gov SetVotingParams")
		require.NotPanics(t, func() {
			s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{
				MinDeposit:       govMinDep,
				MaxDepositPeriod: &depositPeriod,
			})
		}, "gov SetDepositParams")
		testutil.RequireNotPanicsNoError(t, func() error {
			return s.app.SanctionKeeper.SetParams(ctx, &sanction.Params{
				ImmediateSanctionMinDeposit:   sanctMinDep,
				ImmediateUnsanctionMinDeposit: unsanctMinDep,
			})
		}, "sanction SetParams")
	}

	// Create a random number generator for use only in generating accounts.
	accsRand := rand.New(rand.NewSource(1))

	tests := []struct {
		name            string
		setup           func(t *testing.T, ctx sdk.Context)
		r               *rand.Rand
		accs            []simtypes.Account
		expInErr        []string
		expOpMsgOK      bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
		expVote         govv1.VoteOption
		expDeposit      sdk.Coins
	}{
		{
			name: "immediate min deposit is zero",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SetParams(ctx, &sanction.Params{
						ImmediateSanctionMinDeposit:   sdk.Coins{},
						ImmediateUnsanctionMinDeposit: sdk.Coins{},
					})
				}, "sanction SetParams")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            simtypes.RandomAccounts(accsRand, 10),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "immediate sanction min deposit is zero",
		},
		{
			name:            "gov min deposit less than immediate sanction",
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 10),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate sanction",
			expVote:         govv1.OptionYes,
			expDeposit:      sanctMinDep,
		},
		{
			name: "gov min deposit equals immediate sanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NotPanics(t, func() {
					s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{
						MinDeposit:       sanctMinDep,
						MaxDepositPeriod: &depositPeriod,
					})
				}, "gov SetDepositParams")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 10),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate sanction",
			expVote:         govv1.OptionYes,
			expDeposit:      sanctMinDep,
		},
		{
			name: "gov min deposit greater than immediate sanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NotPanics(t, func() {
					s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{
						MinDeposit:       sanctMinDep.Add(govMinDep...),
						MaxDepositPeriod: &depositPeriod,
					})
				}, "gov SetDepositParams")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 10),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate sanction",
			expVote:         govv1.OptionYes,
			expDeposit:      sanctMinDep.Add(govMinDep...),
		},
		{
			name:            "problem sending gov msg",
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccountsWithPower(accsRand, 10, 0),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "sender has no spendable coins",
		},
		{
			name:            "all good yes vote",
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate sanction",
			expVote:         govv1.OptionYes,
			expDeposit:      sanctMinDep,
		},
		{
			name:            "all good no vote",
			r:               rand.New(rand.NewSource(0)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate sanction",
			expVote:         govv1.OptionNo,
			expDeposit:      sanctMinDep,
		},
	}

	wopArgs := s.getWeightedOpsArgs()
	voteType := sdk.MsgTypeURL(&govv1.MsgVote{})

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resetParams(s.T(), s.ctx)
			if tc.setup != nil {
				tc.setup(s.T(), s.ctx)
			}
			var op simtypes.Operation
			testFunc := func() {
				op = simulation.SimulateGovMsgSanctionImmediate(&wopArgs)
			}
			s.Require().NotPanics(testFunc, "SimulateGovMsgSanctionImmediate")
			var opMsg simtypes.OperationMsg
			var fops []simtypes.FutureOperation
			var err error
			testOp := func() {
				opMsg, fops, err = op(tc.r, s.app.BaseApp, s.ctx, tc.accs, chainID)
			}
			s.Require().NotPanics(testOp, "SimulateGovMsgSanctionImmediate op execution")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "op error")
			s.Assert().Equal(tc.expOpMsgOK, opMsg.OK, "op msg ok")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "op msg comment")
			if !tc.expOpMsgOK && !opMsg.OK {
				s.Assert().Empty(fops, "future ops")
			}
			if tc.expOpMsgOK && opMsg.OK {
				s.Assert().Equal(len(tc.accs), len(fops), "number of future ops")
				// If we were expecting it to be okay, and it was, run all the future ops too.
				// Some of them might fail (due to being sanctioned),
				// but all the ones that went through should be YES votes.
				maxBlockTime := s.ctx.BlockHeader().Time.Add(votingPeriod)
				prop := s.getLastGovProp()
				s.Assert().Equal(tc.expDeposit.String(), sdk.NewCoins(prop.TotalDeposit...).String(), "prop deposit")
				preVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
				// There shouldn't be any votes yet.
				if !s.Assert().Empty(preVotes, "votes before running future ops") {
					for i, fop := range fops {
						s.Assert().LessOrEqual(fop.BlockTime, maxBlockTime, "future op %d block time", i+1)
						s.Assert().Equal(0, fop.BlockHeight, "future op %d block height", i+1)
						var fopMsg simtypes.OperationMsg
						var ffops []simtypes.FutureOperation
						testFop := func() {
							fopMsg, ffops, err = fop.Op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, tc.accs, chainID)
						}
						if !s.Assert().NotPanics(testFop, "future op %d execution", i+1) {
							continue
						}
						if err != nil {
							s.T().Logf("future op %d returned an error, but that's kind of expected: %v", i+1, err)
							continue
						}
						if !fopMsg.OK {
							s.T().Logf("future op %d returned not okay, but that's kind of expected: %q", i+1, fopMsg.Comment)
							continue
						}
						s.Assert().Empty(ffops, "future ops returned by future op %d", i+1)
						s.Assert().Equal(voteType, fopMsg.Name, "future op %d msg name", i+1)
						s.Assert().Equal(tc.expOpMsgComment, fopMsg.Comment, "future op %d msg comment", i+1)
					}
					// Now there should be some votes.
					postVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
					for i, vote := range postVotes {
						if s.Assert().Len(vote.Options, 1, "vote %d options count", i+1) {
							s.Assert().Equal(tc.expVote, vote.Options[0].Option, "vote %d option", i+1)
							s.Assert().Equal("1.000000000000000000", vote.Options[1].Weight, "vote %d weight", i+1)
						}
					}
				}
				// Now, get the message and count the number of addresses listed that were provided.
				providedAddrs := make(map[string]bool)
				for _, acc := range tc.accs {
					providedAddrs[acc.Address.String()] = true
				}
				msgs, err := prop.GetMsgs()
				if s.Assert().NoError(err, "getting messages from the proposal") {
					if s.Assert().Len(msgs, 1, "number of messages in the proposal") {
						msg, ok := msgs[0].(*sanction.MsgSanction)
						if s.Assert().True(ok, "could not cast prop msg to MsgSanction") {
							s.Assert().NotEmpty(msg.Addresses, "msg Addresses")
							var inMsg []string
							for _, addr := range msg.Addresses {
								if providedAddrs[addr] {
									inMsg = append(inMsg, addr)
								}
							}
							s.Assert().Empty(inMsg, "provided accs that ended up in the gov prop msg")
						}
					}
				}
			}
			s.nextBlock()
		})
	}
}

func (s *SimTestSuite) TestSimulateGovMsgUnsanction() {
	chainID := "test-simulate-gov-msg-unsanction"
	votingPeriod := 2 * time.Minute
	depositPeriod := 1 * time.Second
	govMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 3))
	sanctMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 7))
	unsanctMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))

	// resetState resets the params to the values defined above and deletes all sanctions and temp sanctions.
	resetState := func(t *testing.T, ctx sdk.Context) {
		require.NotPanics(t, func() {
			s.app.GovKeeper.SetVotingParams(ctx, govv1.VotingParams{VotingPeriod: &votingPeriod})
		}, "gov SetVotingParams")
		require.NotPanics(t, func() {
			s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{
				MinDeposit:       govMinDep,
				MaxDepositPeriod: &depositPeriod,
			})
		}, "gov SetDepositParams")
		testutil.RequireNotPanicsNoError(t, func() error {
			return s.app.SanctionKeeper.SetParams(ctx, &sanction.Params{
				ImmediateSanctionMinDeposit:   sanctMinDep,
				ImmediateUnsanctionMinDeposit: unsanctMinDep,
			})
		}, "sanction SetParams")
		var sanctionedAddrs []sdk.AccAddress
		require.NotPanics(t, func() {
			s.app.SanctionKeeper.IterateSanctionedAddresses(ctx, func(addr sdk.AccAddress) bool {
				sanctionedAddrs = append(sanctionedAddrs, addr)
				return false
			})
		}, "IterateSanctionedAddresses")
		require.NotPanics(t, func() {
			s.app.SanctionKeeper.IterateTemporaryEntries(ctx, nil, func(addr sdk.AccAddress, _ uint64, _ bool) bool {
				sanctionedAddrs = append(sanctionedAddrs, addr)
				return false
			})
		}, "IterateTemporaryEntries")
		testutil.RequireNotPanicsNoError(t, func() error {
			return s.app.SanctionKeeper.UnsanctionAddresses(ctx, sanctionedAddrs...)
		}, "UnsanctionAddresses")
	}

	// Create a random number generator for use only in generating accounts.
	accsRand := rand.New(rand.NewSource(1))

	randAddrs := func(count int) []sdk.AccAddress {
		accs := simtypes.RandomAccounts(accsRand, count)
		addrs := make([]sdk.AccAddress, len(accs))
		for i, acc := range accs {
			addrs[i] = acc.Address
		}
		return addrs
	}

	tests := []struct {
		name            string
		setup           func(t *testing.T, ctx sdk.Context)
		accs            []simtypes.Account
		expInErr        []string
		expOpMsgOK      bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
		expAddrCount    int
	}{
		{
			name:            "no addresses sanctioned",
			accs:            simtypes.RandomAccounts(accsRand, 10),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
			expOpMsgComment: "no addresses are sanctioned",
		},
		{
			name: "gov min deposit equals immediate unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(5)...)
				})
				require.NotPanics(t, func() {
					s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{MinDeposit: unsanctMinDep})
				}, "gov SetDepositParams")
			},
			accs:            simtypes.RandomAccounts(accsRand, 10),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
			expOpMsgComment: "cannot unsanction without it being immediate",
		},
		{
			name: "gov min deposit greater than immediate unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(5)...)
				})
				require.NotPanics(t, func() {
					s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{MinDeposit: unsanctMinDep.Add(govMinDep...)})
				}, "gov SetDepositParams")
			},
			accs:            simtypes.RandomAccounts(accsRand, 10),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
			expOpMsgComment: "cannot unsanction without it being immediate",
		},
		{
			name: "problem sending gov msg",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(5)...)
				})
			},
			accs:            s.createTestingAccountsWithPower(accsRand, 10, 0),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
			expOpMsgComment: "sender has no spendable coins",
		},
		{
			name: "3 addrs to unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(3)...)
				}, "SanctionAddresses")
			},
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "unsanction",
			expAddrCount:    3,
		},
		{
			name: "10 addrs to unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(10)...)
				}, "SanctionAddresses")
			},
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "unsanction",
			expAddrCount:    4,
		},
		{
			name: "39 addrs to unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(39)...)
				}, "SanctionAddresses")
			},
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "unsanction",
			expAddrCount:    9,
		},
		{
			name: "40 addrs to unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(40)...)
				}, "SanctionAddresses")
			},
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "unsanction",
			expAddrCount:    10,
		},
	}

	wopArgs := s.getWeightedOpsArgs()
	voteType := sdk.MsgTypeURL(&govv1.MsgVote{})

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resetState(s.T(), s.ctx)
			if tc.setup != nil {
				tc.setup(s.T(), s.ctx)
			}
			var op simtypes.Operation
			testFunc := func() {
				op = simulation.SimulateGovMsgUnsanction(&wopArgs)
			}
			s.Require().NotPanics(testFunc, "SimulateGovMsgUnsanction")
			var opMsg simtypes.OperationMsg
			var fops []simtypes.FutureOperation
			var err error
			testOp := func() {
				opMsg, fops, err = op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, tc.accs, chainID)
			}
			s.Require().NotPanics(testOp, "SimulateGovMsgUnsanction op execution")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "op error")
			s.Assert().Equal(tc.expOpMsgOK, opMsg.OK, "op msg ok")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "op msg comment")
			if !tc.expOpMsgOK && !opMsg.OK {
				s.Assert().Empty(fops, "future ops")
			}
			if tc.expOpMsgOK && opMsg.OK {
				s.Assert().Equal(len(tc.accs), len(fops), "number of future ops")
				// If we were expecting it to be okay, and it was, run all the future ops too.
				// Some of them might fail (due to being sanctioned),
				// but all the ones that went through should be YES votes.
				maxBlockTime := s.ctx.BlockHeader().Time.Add(votingPeriod)
				prop := s.getLastGovProp()
				s.Assert().Equal(govMinDep.String(), sdk.NewCoins(prop.TotalDeposit...).String(), "prop deposit")
				preVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
				// There shouldn't be any votes yet.
				if !s.Assert().Empty(preVotes, "votes before running future ops") {
					for i, fop := range fops {
						s.Assert().LessOrEqual(fop.BlockTime, maxBlockTime, "future op %d block time", i+1)
						s.Assert().Equal(0, fop.BlockHeight, "future op %d block height", i+1)
						var fopMsg simtypes.OperationMsg
						var ffops []simtypes.FutureOperation
						testFop := func() {
							fopMsg, ffops, err = fop.Op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, tc.accs, chainID)
						}
						if !s.Assert().NotPanics(testFop, "future op %d execution", i+1) {
							continue
						}
						if err != nil {
							s.T().Logf("future op %d returned an error, but that's kind of expected: %v", i+1, err)
							continue
						}
						if !fopMsg.OK {
							s.T().Logf("future op %d returned not okay, but that's kind of expected: %q", i+1, fopMsg.Comment)
							continue
						}
						s.Assert().Empty(ffops, "future ops returned by future op %d", i+1)
						s.Assert().Equal(voteType, fopMsg.Name, "future op %d msg name", i+1)
						s.Assert().Equal(tc.expOpMsgComment, fopMsg.Comment, "future op %d msg comment", i+1)
					}
					// Now there should be some votes.
					postVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
					for i, vote := range postVotes {
						if s.Assert().Len(vote.Options, 1, "vote %d options count", i+1) {
							s.Assert().Equal(govv1.OptionYes, vote.Options[0].Option, "vote %d option", i+1)
							s.Assert().Equal("1.000000000000000000", vote.Options[1].Weight, "vote %d weight", i+1)
						}
					}
				}
				// Now, get the message and count the number of addresses listed that were provided.
				providedAddrs := make(map[string]bool)
				for _, acc := range tc.accs {
					providedAddrs[acc.Address.String()] = true
				}
				msgs, err := prop.GetMsgs()
				if s.Assert().NoError(err, "getting messages from the proposal") {
					if s.Assert().Len(msgs, 1, "number of messages in the proposal") {
						msg, ok := msgs[0].(*sanction.MsgUnsanction)
						if s.Assert().True(ok, "could not cast prop msg to MsgUnsanction") {
							s.Assert().Len(msg.Addresses, tc.expAddrCount, "msg Addresses")
							var inMsg []string
							for _, addr := range msg.Addresses {
								if providedAddrs[addr] {
									inMsg = append(inMsg, addr)
								}
							}
							s.Assert().Empty(inMsg, "provided accs that ended up in the gov prop msg")
						}
					}
				}
			}
			s.nextBlock()
		})
	}
}

func (s *SimTestSuite) TestSimulateGovMsgUnsanctionImmediate() {
	chainID := "test-simulate-gov-msg-immediate-unsanction"
	votingPeriod := 2 * time.Minute
	depositPeriod := 1 * time.Second
	govMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 3))
	sanctMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 7))
	unsanctMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))

	// resetState resets the params to the values defined above and deletes all sanctions and temp sanctions.
	resetState := func(t *testing.T, ctx sdk.Context) {
		require.NotPanics(t, func() {
			s.app.GovKeeper.SetVotingParams(ctx, govv1.VotingParams{VotingPeriod: &votingPeriod})
		}, "gov SetVotingParams")
		require.NotPanics(t, func() {
			s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{
				MinDeposit:       govMinDep,
				MaxDepositPeriod: &depositPeriod,
			})
		}, "gov SetDepositParams")
		testutil.RequireNotPanicsNoError(t, func() error {
			return s.app.SanctionKeeper.SetParams(ctx, &sanction.Params{
				ImmediateSanctionMinDeposit:   sanctMinDep,
				ImmediateUnsanctionMinDeposit: unsanctMinDep,
			})
		}, "sanction SetParams")
		var sanctionedAddrs []sdk.AccAddress
		require.NotPanics(t, func() {
			s.app.SanctionKeeper.IterateSanctionedAddresses(ctx, func(addr sdk.AccAddress) bool {
				sanctionedAddrs = append(sanctionedAddrs, addr)
				return false
			})
		}, "IterateSanctionedAddresses")
		require.NotPanics(t, func() {
			s.app.SanctionKeeper.IterateTemporaryEntries(ctx, nil, func(addr sdk.AccAddress, _ uint64, _ bool) bool {
				sanctionedAddrs = append(sanctionedAddrs, addr)
				return false
			})
		}, "IterateTemporaryEntries")
		testutil.RequireNotPanicsNoError(t, func() error {
			return s.app.SanctionKeeper.UnsanctionAddresses(ctx, sanctionedAddrs...)
		}, "UnsanctionAddresses")
	}

	// Create a random number generator for use only in generating accounts.
	accsRand := rand.New(rand.NewSource(1))

	randAddrs := func(count int) []sdk.AccAddress {
		accs := simtypes.RandomAccounts(accsRand, count)
		addrs := make([]sdk.AccAddress, len(accs))
		for i, acc := range accs {
			addrs[i] = acc.Address
		}
		return addrs
	}

	tests := []struct {
		name            string
		setup           func(t *testing.T, ctx sdk.Context)
		r               *rand.Rand
		accs            []simtypes.Account
		expInErr        []string
		expOpMsgOK      bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
		expVote         govv1.VoteOption
		expDeposit      sdk.Coins
		expAddrCount    int
	}{
		{
			name:            "no addrs to sanction",
			r:               rand.New(rand.NewSource(1)),
			accs:            simtypes.RandomAccounts(accsRand, 10),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
			expOpMsgComment: "no addresses are sanctioned",
		},
		{
			name: "immediate min deposit is zero",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(5)...)
				}, "SanctionAddresses")
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SetParams(ctx, &sanction.Params{
						ImmediateSanctionMinDeposit:   sdk.Coins{},
						ImmediateUnsanctionMinDeposit: sdk.Coins{},
					})
				}, "sanction SetParams")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            simtypes.RandomAccounts(accsRand, 10),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
			expOpMsgComment: "immediate unsanction min deposit is zero",
			expAddrCount:    4,
		},
		{
			name: "gov min deposit less than immediate unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(5)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 10),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionYes,
			expDeposit:      unsanctMinDep,
			expAddrCount:    4,
		},
		{
			name: "gov min deposit equals immediate unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(5)...)
				}, "SanctionAddresses")
				require.NotPanics(t, func() {
					s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{
						MinDeposit:       unsanctMinDep,
						MaxDepositPeriod: &depositPeriod,
					})
				}, "gov SetDepositParams")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 10),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionYes,
			expDeposit:      unsanctMinDep,
			expAddrCount:    4,
		},
		{
			name: "gov min deposit greater than immediate unsanction",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(5)...)
				}, "SanctionAddresses")
				require.NotPanics(t, func() {
					s.app.GovKeeper.SetDepositParams(ctx, govv1.DepositParams{
						MinDeposit:       unsanctMinDep.Add(govMinDep...),
						MaxDepositPeriod: &depositPeriod,
					})
				}, "gov SetDepositParams")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 10),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionYes,
			expDeposit:      unsanctMinDep.Add(govMinDep...),
			expAddrCount:    4,
		},
		{
			name: "problem sending gov msg",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(5)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccountsWithPower(accsRand, 10, 0),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
			expOpMsgComment: "sender has no spendable coins",
		},
		{
			name: "3 addrs to unsanction yes",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(3)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionYes,
			expDeposit:      unsanctMinDep,
			expAddrCount:    3,
		},
		{
			name: "3 addrs to unsanction no",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(3)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(0)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionNo,
			expDeposit:      unsanctMinDep,
			expAddrCount:    3,
		},
		{
			name: "10 addrs to unsanction yes",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(10)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionYes,
			expDeposit:      unsanctMinDep,
			expAddrCount:    4,
		},
		{
			name: "10 addrs to unsanction no",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(10)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(0)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionNo,
			expDeposit:      unsanctMinDep,
			expAddrCount:    4,
		},
		{
			name: "39 addrs to unsanction yes",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(39)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionYes,
			expDeposit:      unsanctMinDep,
			expAddrCount:    9,
		},
		{
			name: "39 addrs to unsanction no",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(39)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(0)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionNo,
			expDeposit:      unsanctMinDep,
			expAddrCount:    9,
		},
		{
			name: "40 addrs to unsanction yes",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(40)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionYes,
			expDeposit:      unsanctMinDep,
			expAddrCount:    10,
		},
		{
			name: "40 addrs to unsanction no",
			setup: func(t *testing.T, ctx sdk.Context) {
				testutil.RequireNotPanicsNoError(t, func() error {
					return s.app.SanctionKeeper.SanctionAddresses(ctx, randAddrs(40)...)
				}, "SanctionAddresses")
			},
			r:               rand.New(rand.NewSource(0)),
			accs:            s.createTestingAccounts(accsRand, 20),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "immediate unsanction",
			expVote:         govv1.OptionNo,
			expDeposit:      unsanctMinDep,
			expAddrCount:    10,
		},
	}

	wopArgs := s.getWeightedOpsArgs()
	voteType := sdk.MsgTypeURL(&govv1.MsgVote{})

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resetState(s.T(), s.ctx)
			if tc.setup != nil {
				tc.setup(s.T(), s.ctx)
			}
			var op simtypes.Operation
			testFunc := func() {
				op = simulation.SimulateGovMsgUnsanctionImmediate(&wopArgs)
			}
			s.Require().NotPanics(testFunc, "SimulateGovMsgUnsanctionImmediate")
			var opMsg simtypes.OperationMsg
			var fops []simtypes.FutureOperation
			var err error
			testOp := func() {
				opMsg, fops, err = op(tc.r, s.app.BaseApp, s.ctx, tc.accs, chainID)
			}
			s.Require().NotPanics(testOp, "SimulateGovMsgUnsanctionImmediate op execution")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "op error")
			s.Assert().Equal(tc.expOpMsgOK, opMsg.OK, "op msg ok")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "op msg comment")
			if !tc.expOpMsgOK && !opMsg.OK {
				s.Assert().Empty(fops, "future ops")
			}
			if tc.expOpMsgOK && opMsg.OK {
				s.Assert().Equal(len(tc.accs), len(fops), "number of future ops")
				// If we were expecting it to be okay, and it was, run all the future ops too.
				// Some of them might fail (due to being sanctioned),
				// but all the ones that went through should be YES votes.
				maxBlockTime := s.ctx.BlockHeader().Time.Add(votingPeriod)
				prop := s.getLastGovProp()
				s.Assert().Equal(tc.expDeposit.String(), sdk.NewCoins(prop.TotalDeposit...).String(), "prop deposit")
				preVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
				// There shouldn't be any votes yet.
				if !s.Assert().Empty(preVotes, "votes before running future ops") {
					for i, fop := range fops {
						s.Assert().LessOrEqual(fop.BlockTime, maxBlockTime, "future op %d block time", i+1)
						s.Assert().Equal(0, fop.BlockHeight, "future op %d block height", i+1)
						var fopMsg simtypes.OperationMsg
						var ffops []simtypes.FutureOperation
						testFop := func() {
							fopMsg, ffops, err = fop.Op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, tc.accs, chainID)
						}
						if !s.Assert().NotPanics(testFop, "future op %d execution", i+1) {
							continue
						}
						if err != nil {
							s.T().Logf("future op %d returned an error, but that's kind of expected: %v", i+1, err)
							continue
						}
						if !fopMsg.OK {
							s.T().Logf("future op %d returned not okay, but that's kind of expected: %q", i+1, fopMsg.Comment)
							continue
						}
						s.Assert().Empty(ffops, "future ops returned by future op %d", i+1)
						s.Assert().Equal(voteType, fopMsg.Name, "future op %d msg name", i+1)
						s.Assert().Equal(tc.expOpMsgComment, fopMsg.Comment, "future op %d msg comment", i+1)
					}
					// Now there should be some votes.
					postVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
					for i, vote := range postVotes {
						if s.Assert().Len(vote.Options, 1, "vote %d options count", i+1) {
							s.Assert().Equal(tc.expVote, vote.Options[0].Option, "vote %d option", i+1)
							s.Assert().Equal("1.000000000000000000", vote.Options[1].Weight, "vote %d weight", i+1)
						}
					}
				}
				// Now, get the message and count the number of addresses listed that were provided.
				providedAddrs := make(map[string]bool)
				for _, acc := range tc.accs {
					providedAddrs[acc.Address.String()] = true
				}
				msgs, err := prop.GetMsgs()
				if s.Assert().NoError(err, "getting messages from the proposal") {
					if s.Assert().Len(msgs, 1, "number of messages in the proposal") {
						msg, ok := msgs[0].(*sanction.MsgUnsanction)
						if s.Assert().True(ok, "could not cast prop msg to MsgUnsanction") {
							s.Assert().Len(msg.Addresses, tc.expAddrCount, "msg Addresses")
							var inMsg []string
							for _, addr := range msg.Addresses {
								if providedAddrs[addr] {
									inMsg = append(inMsg, addr)
								}
							}
							s.Assert().Empty(inMsg, "provided accs that ended up in the gov prop msg")
						}
					}
				}
			}
			s.nextBlock()
		})
	}
}

func (s *SimTestSuite) TestSimulateGovMsgUpdateParams() {
	chainID := "test-simulate-gov-msg-update-params"
	votingPeriod := 2 * time.Minute
	depositPeriod := 1 * time.Second
	govMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2))
	s.Require().NotPanics(func() {
		s.app.GovKeeper.SetVotingParams(s.ctx, govv1.VotingParams{VotingPeriod: &votingPeriod})
	}, "gov SetVotingParams")
	s.Require().NotPanics(func() {
		s.app.GovKeeper.SetDepositParams(s.ctx, govv1.DepositParams{
			MinDeposit:       govMinDep,
			MaxDepositPeriod: &depositPeriod,
		})
	}, "gov SetDepositParams")

	// Create a random number generator for use only in generating accounts.
	accsRand := rand.New(rand.NewSource(1))

	tests := []struct {
		name            string
		r               *rand.Rand
		accs            []simtypes.Account
		expInErr        []string
		expOpMsgOK      bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
		expParams       *sanction.Params
	}{
		{
			name:            "problem sending gov msg",
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccountsWithPower(accsRand, 10, 0),
			expOpMsgOK:      false,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgUpdateParams{}),
			expOpMsgComment: "sender has no spendable coins",
		},
		{
			name:            "all good seed 1",
			r:               rand.New(rand.NewSource(1)),
			accs:            s.createTestingAccounts(accsRand, 10),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "update params",
			expParams: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 821+1)},
			},
		},
		{
			name:            "all good seed 100",
			r:               rand.New(rand.NewSource(100)),
			accs:            s.createTestingAccounts(accsRand, 10),
			expOpMsgOK:      true,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "update params",
			expParams: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 24+1)},
				ImmediateUnsanctionMinDeposit: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 39+1)},
			},
		},
	}

	wopArgs := s.getWeightedOpsArgs()
	voteType := sdk.MsgTypeURL(&govv1.MsgVote{})

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var op simtypes.Operation
			testFunc := func() {
				op = simulation.SimulateGovMsgUpdateParams(&wopArgs)
			}
			s.Require().NotPanics(testFunc, "SimulateGovMsgUpdateParams")
			var opMsg simtypes.OperationMsg
			var fops []simtypes.FutureOperation
			var err error
			testOp := func() {
				opMsg, fops, err = op(tc.r, s.app.BaseApp, s.ctx, tc.accs, chainID)
			}
			s.Require().NotPanics(testOp, "SimulateGovMsgUpdateParams op execution")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "op error")
			s.Assert().Equal(tc.expOpMsgOK, opMsg.OK, "op msg ok")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "op msg comment")
			if !tc.expOpMsgOK && !opMsg.OK {
				s.Assert().Empty(fops, "future ops")
			}
			if tc.expOpMsgOK && opMsg.OK {
				s.Assert().Equal(len(tc.accs), len(fops), "number of future ops")
				// If we were expecting it to be okay, and it was, run all the future ops too.
				// Some of them might fail (due to being sanctioned),
				// but all the ones that went through should be YES votes.
				maxBlockTime := s.ctx.BlockHeader().Time.Add(votingPeriod)
				prop := s.getLastGovProp()
				s.Assert().Equal(govMinDep.String(), sdk.NewCoins(prop.TotalDeposit...).String(), "prop deposit")
				preVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
				// There shouldn't be any votes yet.
				if !s.Assert().Empty(preVotes, "votes before running future ops") {
					for i, fop := range fops {
						s.Assert().LessOrEqual(fop.BlockTime, maxBlockTime, "future op %d block time", i+1)
						s.Assert().Equal(0, fop.BlockHeight, "future op %d block height", i+1)
						var fopMsg simtypes.OperationMsg
						var ffops []simtypes.FutureOperation
						testFop := func() {
							fopMsg, ffops, err = fop.Op(rand.New(rand.NewSource(1)), s.app.BaseApp, s.ctx, tc.accs, chainID)
						}
						if !s.Assert().NotPanics(testFop, "future op %d execution", i+1) {
							continue
						}
						if err != nil {
							s.T().Logf("future op %d returned an error, but that's kind of expected: %v", i+1, err)
							continue
						}
						if !fopMsg.OK {
							s.T().Logf("future op %d returned not okay, but that's kind of expected: %q", i+1, fopMsg.Comment)
							continue
						}
						s.Assert().Empty(ffops, "future ops returned by future op %d", i+1)
						s.Assert().Equal(voteType, fopMsg.Name, "future op %d msg name", i+1)
						s.Assert().Equal(tc.expOpMsgComment, fopMsg.Comment, "future op %d msg comment", i+1)
					}
					// Now there should be some votes.
					postVotes := s.app.GovKeeper.GetVotes(s.ctx, prop.Id)
					for i, vote := range postVotes {
						if s.Assert().Len(vote.Options, 1, "vote %d options count", i+1) {
							s.Assert().Equal(govv1.OptionYes, vote.Options[0].Option, "vote %d option", i+1)
							s.Assert().Equal("1.000000000000000000", vote.Options[1].Weight, "vote %d weight", i+1)
						}
					}
				}
				// Now, get the message and check its content.
				msgs, err := prop.GetMsgs()
				if s.Assert().NoError(err, "getting messages from the proposal") {
					if s.Assert().Len(msgs, 1, "number of messages in the proposal") {
						msg, ok := msgs[0].(*sanction.MsgUpdateParams)
						if s.Assert().True(ok, "could not cast prop msg to MsgUpdateParams") {
							if !s.Assert().Equal(tc.expParams, msg.Params, "params in gov prop") && tc.expParams != nil && msg.Params != nil {
								s.Assert().Equal(tc.expParams.ImmediateSanctionMinDeposit.String(),
									msg.Params.ImmediateSanctionMinDeposit.String(),
									"ImmediateSanctionMinDeposit")
								s.Assert().Equal(tc.expParams.ImmediateUnsanctionMinDeposit.String(),
									msg.Params.ImmediateUnsanctionMinDeposit.String(),
									"ImmediateUnsanctionMinDeposit")
							}
						}
					}
				}
			}
			s.nextBlock()
		})
	}
}
