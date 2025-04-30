package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"                    //nolint:staticcheck // deprecated and to be removed
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper" //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/group/simulation"         //nolint:staticcheck // deprecated and to be removed
	grouptestutil "github.com/cosmos/cosmos-sdk/x/group/testutil"
)

type SimTestSuite struct {
	suite.Suite

	ctx               sdk.Context
	app               *runtime.App
	codec             codec.Codec
	interfaceRegistry codectypes.InterfaceRegistry
	txConfig          client.TxConfig
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	groupKeeper       groupkeeper.Keeper
}

func (suite *SimTestSuite) SetupTest() {
	app, err := simtestutil.Setup(
		depinject.Configs(
			grouptestutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&suite.codec,
		&suite.interfaceRegistry,
		&suite.txConfig,
		&suite.accountKeeper,
		&suite.bankKeeper,
		&suite.groupKeeper,
	)
	suite.Require().NoError(err)

	suite.app = app
	suite.ctx = app.NewContext(false)
}

func (suite *SimTestSuite) TestWeightedOperations() {
	cdc := suite.codec
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(suite.interfaceRegistry, appParams, cdc, suite.txConfig, suite.accountKeeper,
		suite.bankKeeper, suite.groupKeeper, cdc,
	)

	s := rand.NewSource(2)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.WeightMsgCreateGroup, group.ModuleName, simulation.TypeMsgCreateGroup},
		{simulation.WeightMsgCreateGroupPolicy, group.ModuleName, simulation.TypeMsgCreateGroupPolicy},
		{simulation.WeightMsgCreateGroupWithPolicy, group.ModuleName, simulation.TypeMsgCreateGroupWithPolicy},
		{simulation.WeightMsgSubmitProposal, group.ModuleName, simulation.TypeMsgSubmitProposal},
		{simulation.WeightMsgSubmitProposal, group.ModuleName, simulation.TypeMsgSubmitProposal},
		{simulation.WeightMsgWithdrawProposal, group.ModuleName, simulation.TypeMsgWithdrawProposal},
		{simulation.WeightMsgVote, group.ModuleName, simulation.TypeMsgVote},
		{simulation.WeightMsgExec, group.ModuleName, simulation.TypeMsgExec},
		{simulation.WeightMsgUpdateGroupMetadata, group.ModuleName, simulation.TypeMsgUpdateGroupMetadata},
		{simulation.WeightMsgUpdateGroupAdmin, group.ModuleName, simulation.TypeMsgUpdateGroupAdmin},
		{simulation.WeightMsgUpdateGroupMembers, group.ModuleName, simulation.TypeMsgUpdateGroupMembers},
		{simulation.WeightMsgUpdateGroupPolicyAdmin, group.ModuleName, simulation.TypeMsgUpdateGroupPolicyAdmin},
		{simulation.WeightMsgUpdateGroupPolicyDecisionPolicy, group.ModuleName, simulation.TypeMsgUpdateGroupPolicyDecisionPolicy},
		{simulation.WeightMsgUpdateGroupPolicyMetadata, group.ModuleName, simulation.TypeMsgUpdateGroupPolicyMetadata},
		{simulation.WeightMsgLeaveGroup, group.ModuleName, simulation.TypeMsgLeaveGroup},
	}

	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, "")
		suite.Require().NoError(err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(testutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins))
	}

	return accounts
}

func (suite *SimTestSuite) TestSimulateCreateGroup() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 1)

	_, err := suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	acc := accounts[0]

	// execute operation
	op := simulation.SimulateMsgCreateGroup(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgCreateGroup
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(acc.Address.String(), msg.Admin)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateCreateGroupWithPolicy() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 1)

	_, err := suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	acc := accounts[0]

	// execute operation
	op := simulation.SimulateMsgCreateGroupWithPolicy(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgCreateGroupWithPolicy
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(acc.Address.String(), msg.Admin)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateCreateGroupPolicy() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 1)
	acc := accounts[0]

	// setup a group
	_, err := suite.groupKeeper.CreateGroup(suite.ctx,
		&group.MsgCreateGroup{
			Admin: acc.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: acc.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgCreateGroupPolicy(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgCreateGroupPolicy
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(acc.Address.String(), msg.Admin)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateSubmitProposal() {
	// setup 1 account
	s := rand.NewSource(2)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 1)
	acc := accounts[0]

	// setup a group
	ctx := suite.ctx
	groupRes, err := suite.groupKeeper.CreateGroup(ctx,
		&group.MsgCreateGroup{
			Admin: acc.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: acc.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	// setup a group account
	accountReq := &group.MsgCreateGroupPolicy{
		Admin:   acc.Address.String(),
		GroupId: groupRes.GroupId,
	}
	err = accountReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)
	groupPolicyRes, err := suite.groupKeeper.CreateGroupPolicy(ctx, accountReq)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgSubmitProposal(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgSubmitProposal
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(groupPolicyRes.Address, msg.GroupPolicyAddress)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestWithdrawProposal() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)
	acc := accounts[0]

	// setup a group
	ctx := suite.ctx
	addr := acc.Address.String()
	groupRes, err := suite.groupKeeper.CreateGroup(ctx,
		&group.MsgCreateGroup{
			Admin: addr,
			Members: []group.MemberRequest{
				{
					Address: addr,
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	// setup a group account
	accountReq := &group.MsgCreateGroupPolicy{
		Admin:   addr,
		GroupId: groupRes.GroupId,
	}
	err = accountReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)
	groupPolicyRes, err := suite.groupKeeper.CreateGroupPolicy(ctx, accountReq)
	suite.Require().NoError(err)

	// setup a proposal
	proposalReq, err := group.NewMsgSubmitProposal(groupPolicyRes.Address, []string{addr}, []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: groupPolicyRes.Address,
			ToAddress:   addr,
			Amount:      sdk.Coins{sdk.NewInt64Coin("token", 100)},
		},
	}, "", 0, "MsgSend", "this is a test proposal")
	suite.Require().NoError(err)
	_, err = suite.groupKeeper.SubmitProposal(ctx, proposalReq)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgWithdrawProposal(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgWithdrawProposal
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(addr, msg.Address)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateVote() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 1)
	acc := accounts[0]

	// setup a group
	ctx := suite.ctx
	addr := acc.Address.String()
	groupRes, err := suite.groupKeeper.CreateGroup(ctx,
		&group.MsgCreateGroup{
			Admin: addr,
			Members: []group.MemberRequest{
				{
					Address: addr,
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	// setup a group account
	accountReq := &group.MsgCreateGroupPolicy{
		Admin:    addr,
		GroupId:  groupRes.GroupId,
		Metadata: "",
	}
	err = accountReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)
	groupPolicyRes, err := suite.groupKeeper.CreateGroupPolicy(ctx, accountReq)
	suite.Require().NoError(err)

	// setup a proposal
	proposalReq, err := group.NewMsgSubmitProposal(groupPolicyRes.Address, []string{addr}, []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: groupPolicyRes.Address,
			ToAddress:   addr,
			Amount:      sdk.Coins{sdk.NewInt64Coin("token", 100)},
		},
	}, "", 0, "MsgSend", "this is a test proposal")
	suite.Require().NoError(err)
	_, err = suite.groupKeeper.SubmitProposal(ctx, proposalReq)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgVote(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgVote
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(addr, msg.Voter)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateExec() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 1)
	acc := accounts[0]

	// setup a group
	ctx := suite.ctx
	addr := acc.Address.String()
	groupRes, err := suite.groupKeeper.CreateGroup(ctx,
		&group.MsgCreateGroup{
			Admin: addr,
			Members: []group.MemberRequest{
				{
					Address: addr,
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	// setup a group account
	accountReq := &group.MsgCreateGroupPolicy{
		Admin:   addr,
		GroupId: groupRes.GroupId,
	}
	err = accountReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)
	groupPolicyRes, err := suite.groupKeeper.CreateGroupPolicy(ctx, accountReq)
	suite.Require().NoError(err)

	// setup a proposal
	proposalReq, err := group.NewMsgSubmitProposal(groupPolicyRes.Address, []string{addr}, []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: groupPolicyRes.Address,
			ToAddress:   addr,
			Amount:      sdk.Coins{sdk.NewInt64Coin("token", 100)},
		},
	}, "", 0, "MsgSend", "this is a test proposal")
	suite.Require().NoError(err)
	proposalRes, err := suite.groupKeeper.SubmitProposal(ctx, proposalReq)
	suite.Require().NoError(err)

	// vote
	_, err = suite.groupKeeper.Vote(ctx, &group.MsgVote{
		ProposalId: proposalRes.ProposalId,
		Voter:      addr,
		Option:     group.VOTE_OPTION_YES,
		Exec:       1,
	})
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgExec(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgExec
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(addr, msg.Executor)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateUpdateGroupAdmin() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	acc := accounts[0]

	// setup a group
	_, err := suite.groupKeeper.CreateGroup(suite.ctx,
		&group.MsgCreateGroup{
			Admin: acc.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: acc.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgUpdateGroupAdmin(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgUpdateGroupAdmin
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(acc.Address.String(), msg.Admin)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateUpdateGroupMetadata() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	acc := accounts[0]

	// setup a group
	_, err := suite.groupKeeper.CreateGroup(suite.ctx,
		&group.MsgCreateGroup{
			Admin: acc.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: acc.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgUpdateGroupMetadata(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgUpdateGroupMetadata
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(acc.Address.String(), msg.Admin)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateUpdateGroupMembers() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	acc := accounts[0]

	// setup a group
	_, err := suite.groupKeeper.CreateGroup(suite.ctx,
		&group.MsgCreateGroup{
			Admin: acc.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: acc.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgUpdateGroupMembers(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgUpdateGroupMembers
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(acc.Address.String(), msg.Admin)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateUpdateGroupPolicyAdmin() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	acc := accounts[0]

	// setup a group
	ctx := suite.ctx
	groupRes, err := suite.groupKeeper.CreateGroup(ctx,
		&group.MsgCreateGroup{
			Admin: acc.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: acc.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	// setup a group account
	accountReq := &group.MsgCreateGroupPolicy{
		Admin:   acc.Address.String(),
		GroupId: groupRes.GroupId,
	}
	err = accountReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)
	groupPolicyRes, err := suite.groupKeeper.CreateGroupPolicy(ctx, accountReq)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgUpdateGroupPolicyAdmin(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgUpdateGroupPolicyAdmin
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(groupPolicyRes.Address, msg.GroupPolicyAddress)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateUpdateGroupPolicyDecisionPolicy() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	acc := accounts[0]

	// setup a group
	ctx := suite.ctx
	groupRes, err := suite.groupKeeper.CreateGroup(ctx,
		&group.MsgCreateGroup{
			Admin: acc.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: acc.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	// setup a group account
	accountReq := &group.MsgCreateGroupPolicy{
		Admin:   acc.Address.String(),
		GroupId: groupRes.GroupId,
	}
	err = accountReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)
	groupPolicyRes, err := suite.groupKeeper.CreateGroupPolicy(ctx, accountReq)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgUpdateGroupPolicyDecisionPolicy(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgUpdateGroupPolicyDecisionPolicy
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(groupPolicyRes.Address, msg.GroupPolicyAddress)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateUpdateGroupPolicyMetadata() {
	// setup 1 account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	acc := accounts[0]

	// setup a group
	ctx := suite.ctx
	groupRes, err := suite.groupKeeper.CreateGroup(ctx,
		&group.MsgCreateGroup{
			Admin: acc.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: acc.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	suite.Require().NoError(err)

	// setup a group account
	accountReq := &group.MsgCreateGroupPolicy{
		Admin:   acc.Address.String(),
		GroupId: groupRes.GroupId,
	}
	err = accountReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)
	groupPolicyRes, err := suite.groupKeeper.CreateGroupPolicy(ctx, accountReq)
	suite.Require().NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgUpdateGroupPolicyMetadata(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.groupKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgUpdateGroupPolicyMetadata
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(groupPolicyRes.Address, msg.GroupPolicyAddress)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateLeaveGroup() {
	s := rand.NewSource(1)
	r := rand.New(s)
	require := suite.Require()

	// setup 4 account
	accounts := suite.getTestingAccounts(r, 4)
	admin := accounts[0]
	member1 := accounts[1]
	member2 := accounts[2]
	member3 := accounts[3]

	// setup a group
	ctx := suite.ctx
	groupRes, err := suite.groupKeeper.CreateGroup(ctx,
		&group.MsgCreateGroup{
			Admin: admin.Address.String(),
			Members: []group.MemberRequest{
				{
					Address: member1.Address.String(),
					Weight:  "1",
				},
				{
					Address: member2.Address.String(),
					Weight:  "2",
				},
				{
					Address: member3.Address.String(),
					Weight:  "1",
				},
			},
		},
	)
	require.NoError(err)

	// setup a group account
	accountReq := &group.MsgCreateGroupPolicy{
		Admin:    admin.Address.String(),
		GroupId:  groupRes.GroupId,
		Metadata: "",
	}
	require.NoError(accountReq.SetDecisionPolicy(group.NewThresholdDecisionPolicy("3", time.Hour, time.Hour)))
	_, err = suite.groupKeeper.CreateGroupPolicy(ctx, accountReq)
	require.NoError(err)

	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgLeaveGroup(nil, suite.txConfig, suite.groupKeeper, suite.accountKeeper, suite.bankKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg group.MsgLeaveGroup
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(groupRes.GroupId, msg.GroupId)
	suite.Require().Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
