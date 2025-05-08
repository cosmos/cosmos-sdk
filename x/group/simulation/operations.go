// Deprecated: This package is deprecated and will be removed in the next major release. The `x/group` module will be moved to a separate repo `github.com/cosmos/cosmos-sdk-legacy`.
package simulation

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/group"        //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/group/keeper" //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const unsetGroupID = 100000000000000

// group message types
// will be removed in the future
var (
	TypeMsgCreateGroup                     = sdk.MsgTypeURL(&group.MsgCreateGroup{})
	TypeMsgUpdateGroupMembers              = sdk.MsgTypeURL(&group.MsgUpdateGroupMembers{})
	TypeMsgUpdateGroupAdmin                = sdk.MsgTypeURL(&group.MsgUpdateGroupAdmin{})
	TypeMsgUpdateGroupMetadata             = sdk.MsgTypeURL(&group.MsgUpdateGroupMetadata{})
	TypeMsgCreateGroupWithPolicy           = sdk.MsgTypeURL(&group.MsgCreateGroupWithPolicy{})
	TypeMsgCreateGroupPolicy               = sdk.MsgTypeURL(&group.MsgCreateGroupPolicy{})
	TypeMsgUpdateGroupPolicyAdmin          = sdk.MsgTypeURL(&group.MsgUpdateGroupPolicyAdmin{})
	TypeMsgUpdateGroupPolicyDecisionPolicy = sdk.MsgTypeURL(&group.MsgUpdateGroupPolicyDecisionPolicy{})
	TypeMsgUpdateGroupPolicyMetadata       = sdk.MsgTypeURL(&group.MsgUpdateGroupPolicyMetadata{})
	TypeMsgSubmitProposal                  = sdk.MsgTypeURL(&group.MsgSubmitProposal{})
	TypeMsgWithdrawProposal                = sdk.MsgTypeURL(&group.MsgWithdrawProposal{})
	TypeMsgVote                            = sdk.MsgTypeURL(&group.MsgVote{})
	TypeMsgExec                            = sdk.MsgTypeURL(&group.MsgExec{})
	TypeMsgLeaveGroup                      = sdk.MsgTypeURL(&group.MsgLeaveGroup{})
)

// Simulation operation weights constants
// will be removed in the future
const (
	OpMsgCreateGroup                     = "op_weight_msg_create_group"
	OpMsgUpdateGroupAdmin                = "op_weight_msg_update_group_admin"
	OpMsgUpdateGroupMetadata             = "op_wieght_msg_update_group_metadata"
	OpMsgUpdateGroupMembers              = "op_weight_msg_update_group_members"
	OpMsgCreateGroupPolicy               = "op_weight_msg_create_group_account"
	OpMsgCreateGroupWithPolicy           = "op_weight_msg_create_group_with_policy"
	OpMsgUpdateGroupPolicyAdmin          = "op_weight_msg_update_group_account_admin"
	OpMsgUpdateGroupPolicyDecisionPolicy = "op_weight_msg_update_group_account_decision_policy"
	OpMsgUpdateGroupPolicyMetaData       = "op_weight_msg_update_group_account_metadata"
	OpMsgSubmitProposal                  = "op_weight_msg_submit_proposal"
	OpMsgWithdrawProposal                = "op_weight_msg_withdraw_proposal"
	OpMsgVote                            = "op_weight_msg_vote"
	OpMsgExec                            = "ops_weight_msg_exec"
	OpMsgLeaveGroup                      = "ops_weight_msg_leave_group"
)

// If update group or group policy txn's executed, `SimulateMsgVote` & `SimulateMsgExec` txn's returns `noOp`.
// That's why we have less weight for update group & group-policy txn's.
// will be removed in the future
const (
	WeightMsgCreateGroup                     = 100
	WeightMsgCreateGroupPolicy               = 50
	WeightMsgSubmitProposal                  = 90
	WeightMsgVote                            = 90
	WeightMsgExec                            = 90
	WeightMsgLeaveGroup                      = 5
	WeightMsgUpdateGroupMetadata             = 5
	WeightMsgUpdateGroupAdmin                = 5
	WeightMsgUpdateGroupMembers              = 5
	WeightMsgUpdateGroupPolicyAdmin          = 5
	WeightMsgUpdateGroupPolicyDecisionPolicy = 5
	WeightMsgUpdateGroupPolicyMetadata       = 5
	WeightMsgWithdrawProposal                = 20
	WeightMsgCreateGroupWithPolicy           = 50
)

// WeightedOperations returns all the operations from the module with their respective weights
// migrate to the msg factories instead, this method will be removed in the future
func WeightedOperations(
	registry cdctypes.InterfaceRegistry,
	appParams simtypes.AppParams, _ codec.JSONCodec, txGen client.TxConfig,
	ak group.AccountKeeper, bk group.BankKeeper, k keeper.Keeper,
	_ cdctypes.AnyUnpacker,
) simulation.WeightedOperations {
	var (
		weightMsgCreateGroup                     int
		weightMsgUpdateGroupAdmin                int
		weightMsgUpdateGroupMetadata             int
		weightMsgUpdateGroupMembers              int
		weightMsgCreateGroupPolicy               int
		weightMsgUpdateGroupPolicyAdmin          int
		weightMsgUpdateGroupPolicyDecisionPolicy int
		weightMsgUpdateGroupPolicyMetadata       int
		weightMsgSubmitProposal                  int
		weightMsgVote                            int
		weightMsgExec                            int
		weightMsgLeaveGroup                      int
		weightMsgWithdrawProposal                int
		weightMsgCreateGroupWithPolicy           int
	)

	appParams.GetOrGenerate(OpMsgCreateGroup, &weightMsgCreateGroup, nil, func(_ *rand.Rand) {
		weightMsgCreateGroup = WeightMsgCreateGroup
	})
	appParams.GetOrGenerate(OpMsgCreateGroupPolicy, &weightMsgCreateGroupPolicy, nil, func(_ *rand.Rand) {
		weightMsgCreateGroupPolicy = WeightMsgCreateGroupPolicy
	})
	appParams.GetOrGenerate(OpMsgLeaveGroup, &weightMsgLeaveGroup, nil, func(_ *rand.Rand) {
		weightMsgLeaveGroup = WeightMsgLeaveGroup
	})
	appParams.GetOrGenerate(OpMsgCreateGroupWithPolicy, &weightMsgCreateGroupWithPolicy, nil, func(_ *rand.Rand) {
		weightMsgCreateGroupWithPolicy = WeightMsgCreateGroupWithPolicy
	})
	appParams.GetOrGenerate(OpMsgSubmitProposal, &weightMsgSubmitProposal, nil, func(_ *rand.Rand) {
		weightMsgSubmitProposal = WeightMsgSubmitProposal
	})
	appParams.GetOrGenerate(OpMsgVote, &weightMsgVote, nil, func(_ *rand.Rand) {
		weightMsgVote = WeightMsgVote
	})
	appParams.GetOrGenerate(OpMsgExec, &weightMsgExec, nil, func(_ *rand.Rand) {
		weightMsgExec = WeightMsgExec
	})
	appParams.GetOrGenerate(OpMsgUpdateGroupMetadata, &weightMsgUpdateGroupMetadata, nil, func(_ *rand.Rand) {
		weightMsgUpdateGroupMetadata = WeightMsgUpdateGroupMetadata
	})
	appParams.GetOrGenerate(OpMsgUpdateGroupAdmin, &weightMsgUpdateGroupAdmin, nil, func(_ *rand.Rand) {
		weightMsgUpdateGroupAdmin = WeightMsgUpdateGroupAdmin
	})
	appParams.GetOrGenerate(OpMsgUpdateGroupMembers, &weightMsgUpdateGroupMembers, nil, func(_ *rand.Rand) {
		weightMsgUpdateGroupMembers = WeightMsgUpdateGroupMembers
	})
	appParams.GetOrGenerate(OpMsgUpdateGroupPolicyAdmin, &weightMsgUpdateGroupPolicyAdmin, nil, func(_ *rand.Rand) {
		weightMsgUpdateGroupPolicyAdmin = WeightMsgUpdateGroupPolicyAdmin
	})
	appParams.GetOrGenerate(OpMsgUpdateGroupPolicyDecisionPolicy, &weightMsgUpdateGroupPolicyDecisionPolicy, nil, func(_ *rand.Rand) {
		weightMsgUpdateGroupPolicyDecisionPolicy = WeightMsgUpdateGroupPolicyDecisionPolicy
	})
	appParams.GetOrGenerate(OpMsgUpdateGroupPolicyMetaData, &weightMsgUpdateGroupPolicyMetadata, nil, func(_ *rand.Rand) {
		weightMsgUpdateGroupPolicyMetadata = WeightMsgUpdateGroupPolicyMetadata
	})
	appParams.GetOrGenerate(OpMsgWithdrawProposal, &weightMsgWithdrawProposal, nil, func(_ *rand.Rand) {
		weightMsgWithdrawProposal = WeightMsgWithdrawProposal
	})

	pCdc := codec.NewProtoCodec(registry)

	state := NewSharedState()

	// create two proposals for weightedOperations
	var createProposalOps simulation.WeightedOperations
	for range 2 {
		createProposalOps = append(createProposalOps, simulation.NewWeightedOperation(
			weightMsgSubmitProposal,
			simulateMsgSubmitProposal(pCdc, txGen, ak, bk, k, state),
		))
	}

	wPreCreateProposalOps := simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateGroup,
			SimulateMsgCreateGroup(pCdc, txGen, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateGroupPolicy,
			simulateMsgCreateGroupPolicy(pCdc, txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateGroupWithPolicy,
			SimulateMsgCreateGroupWithPolicy(pCdc, txGen, ak, bk),
		),
	}

	wPostCreateProposalOps := simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			WeightMsgWithdrawProposal,
			simulateMsgWithdrawProposal(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgVote,
			simulateMsgVote(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgExec,
			simulateMsgExec(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateGroupMetadata,
			simulateMsgUpdateGroupMetadata(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateGroupAdmin,
			simulateMsgUpdateGroupAdmin(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateGroupMembers,
			simulateMsgUpdateGroupMembers(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateGroupPolicyAdmin,
			simulateMsgUpdateGroupPolicyAdmin(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateGroupPolicyDecisionPolicy,
			simulateMsgUpdateGroupPolicyDecisionPolicy(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateGroupPolicyMetadata,
			simulateMsgUpdateGroupPolicyMetadata(txGen, ak, bk, k, state),
		),
		simulation.NewWeightedOperation(
			weightMsgLeaveGroup,
			simulateMsgLeaveGroup(txGen, k, ak, bk, state),
		),
	}

	return append(wPreCreateProposalOps, append(createProposalOps, wPostCreateProposalOps...)...)
}

// SimulateMsgCreateGroup generates a MsgCreateGroup with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgCreateGroup(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		acc, _ := simtypes.RandomAcc(r, accounts)
		account := ak.GetAccount(ctx, acc.Address)
		accAddr := acc.Address.String()

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroup, "fee error"), nil, err
		}

		members := genGroupMembers(r, accounts)
		msg := &group.MsgCreateGroup{Admin: accAddr, Members: members, Metadata: simtypes.RandStringOfLength(r, 10)}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroup, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgCreateGroupWithPolicy generates a MsgCreateGroupWithPolicy with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgCreateGroupWithPolicy(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		acc, _ := simtypes.RandomAcc(r, accounts)
		account := ak.GetAccount(ctx, acc.Address)
		accAddr := acc.Address.String()

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroup, "fee error"), nil, err
		}

		members := genGroupMembers(r, accounts)
		decisionPolicy := &group.ThresholdDecisionPolicy{
			Threshold: fmt.Sprintf("%d", simtypes.RandIntBetween(r, 1, 10)),
			Windows: &group.DecisionPolicyWindows{
				VotingPeriod: time.Second * time.Duration(30*24*60*60),
			},
		}

		msg := &group.MsgCreateGroupWithPolicy{
			Admin:               accAddr,
			Members:             members,
			GroupMetadata:       simtypes.RandStringOfLength(r, 10),
			GroupPolicyMetadata: simtypes.RandStringOfLength(r, 10),
			GroupPolicyAsAdmin:  r.Float32() < 0.5,
		}
		err = msg.SetDecisionPolicy(decisionPolicy)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to set decision policy"), nil, err
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupWithPolicy, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgCreateGroupPolicy generates a NewMsgCreateGroupPolicy with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgCreateGroupPolicy(
	cdc *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgCreateGroupPolicy(cdc, txGen, ak, bk, k, NewSharedState())
}

func simulateMsgCreateGroupPolicy(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupInfo, acc, account, err := randomGroup(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupPolicy, ""), nil, err
		}
		if groupInfo == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupPolicy, ""), nil, nil
		}
		groupID := groupInfo.Id

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupPolicy, "fee error"), nil, err
		}

		msg, err := group.NewMsgCreateGroupPolicy(
			acc.Address,
			groupID,
			simtypes.RandStringOfLength(r, 10),
			&group.ThresholdDecisionPolicy{
				Threshold: fmt.Sprintf("%d", simtypes.RandIntBetween(r, 1, 10)),
				Windows: &group.DecisionPolicyWindows{
					VotingPeriod: time.Second * time.Duration(30*24*60*60),
				},
			},
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupPolicy, err.Error()), nil, err
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupPolicy, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			fmt.Printf("ERR DELIVER %v\n", err)
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgSubmitProposal generates a NewMsgSubmitProposal with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgSubmitProposal(
	cdc *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgSubmitProposal(cdc, txGen, ak, bk, k, NewSharedState())
}

func simulateMsgSubmitProposal(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		g, groupPolicy, _, _, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, ""), nil, err
		}
		if g == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, "no group found"), nil, nil
		}
		if groupPolicy == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, "no group policy found"), nil, nil
		}
		groupID := g.Id
		groupPolicyAddr := groupPolicy.Address

		// Return a no-op if we know the proposal cannot be created
		policy, err := groupPolicy.GetDecisionPolicy()
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, ""), nil, nil
		}
		err = policy.Validate(*g, group.DefaultConfig())
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, ""), nil, nil
		}

		// Pick a random member from the group
		acc, account, err := randomMember(sdkCtx, r, k, ak, accounts, groupID)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, ""), nil, err
		}
		if account == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, "no group member found"), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, "fee error"), nil, err
		}

		msg := &group.MsgSubmitProposal{
			GroupPolicyAddress: groupPolicyAddr,
			Proposers:          []string{acc.Address.String()},
			Metadata:           simtypes.RandStringOfLength(r, 10),
			Title:              "Test Proposal",
			Summary:            "Summary of the proposal",
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgSubmitProposal, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgUpdateGroupAdmin generates a MsgUpdateGroupAdmin with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgUpdateGroupAdmin(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgUpdateGroupAdmin(txGen, ak, bk, k, NewSharedState())
}

func simulateMsgUpdateGroupAdmin(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupInfo, acc, account, err := randomGroup(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAdmin, ""), nil, err
		}
		if groupInfo == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAdmin, ""), nil, nil
		}
		groupID := groupInfo.Id

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAdmin, "fee error"), nil, err
		}

		if len(accounts) == 1 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAdmin, "can't set a new admin with only one account"), nil, nil
		}
		newAdmin, _ := simtypes.RandomAcc(r, accounts)
		// disallow setting current admin as new admin
		for acc.PubKey.Equals(newAdmin.PubKey) {
			newAdmin, _ = simtypes.RandomAcc(r, accounts)
		}

		msg := &group.MsgUpdateGroupAdmin{
			GroupId:  groupID,
			Admin:    account.GetAddress().String(),
			NewAdmin: newAdmin.Address.String(),
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAdmin, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgUpdateGroupMetadata generates a MsgUpdateGroupMetadata with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgUpdateGroupMetadata(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgUpdateGroupMetadata(txGen, ak, bk, k, NewSharedState())
}

func simulateMsgUpdateGroupMetadata(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupInfo, acc, account, err := randomGroup(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMetadata, ""), nil, err
		}
		if groupInfo == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMetadata, ""), nil, nil
		}
		groupID := groupInfo.Id

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMetadata, "fee error"), nil, err
		}

		msg := &group.MsgUpdateGroupMetadata{
			GroupId:  groupID,
			Admin:    account.GetAddress().String(),
			Metadata: simtypes.RandStringOfLength(r, 10),
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMetadata, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgUpdateGroupMembers generates a MsgUpdateGroupMembers with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgUpdateGroupMembers(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgUpdateGroupMembers(txGen, ak, bk, k, NewSharedState())
}

func simulateMsgUpdateGroupMembers(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupInfo, acc, account, err := randomGroup(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMembers, ""), nil, err
		}
		if groupInfo == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMembers, ""), nil, nil
		}
		groupID := groupInfo.Id

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMembers, "fee error"), nil, err
		}

		members := genGroupMembers(r, accounts)
		ctx := sdk.UnwrapSDKContext(sdkCtx)
		res, err := k.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: groupID})
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMembers, "group members"), nil, err
		}

		// set existing random group member weight to zero to remove from the group
		existigMembers := res.Members
		if len(existigMembers) > 1 {
			memberToRemove := existigMembers[r.Intn(len(existigMembers))]
			var isDuplicateMember bool
			for idx, m := range members {
				if m.Address == memberToRemove.Member.Address {
					members[idx].Weight = "0"
					isDuplicateMember = true
					break
				}
			}

			if !isDuplicateMember {
				m := memberToRemove.Member
				m.Weight = "0"
				members = append(members, group.MemberToMemberRequest(m))
			}
		}

		msg := &group.MsgUpdateGroupMembers{
			GroupId:       groupID,
			Admin:         acc.Address.String(),
			MemberUpdates: members,
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMembers, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgUpdateGroupPolicyAdmin generates a MsgUpdateGroupPolicyAdmin with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgUpdateGroupPolicyAdmin(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgUpdateGroupPolicyAdmin(txGen, ak, bk, k, NewSharedState())
}

func simulateMsgUpdateGroupPolicyAdmin(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		_, groupPolicy, acc, account, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyAdmin, ""), nil, err
		}
		if groupPolicy == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyAdmin, "no group policy found"), nil, nil
		}
		groupPolicyAddr := groupPolicy.Address

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyAdmin, "fee error"), nil, err
		}

		if len(accounts) == 1 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyAdmin, "can't set a new admin with only one account"), nil, nil
		}
		newAdmin, _ := simtypes.RandomAcc(r, accounts)
		// disallow setting current admin as new admin
		for acc.PubKey.Equals(newAdmin.PubKey) {
			newAdmin, _ = simtypes.RandomAcc(r, accounts)
		}

		msg := &group.MsgUpdateGroupPolicyAdmin{
			Admin:              acc.Address.String(),
			GroupPolicyAddress: groupPolicyAddr,
			NewAdmin:           newAdmin.Address.String(),
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyAdmin, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgUpdateGroupPolicyDecisionPolicy generates a NewMsgUpdateGroupPolicyDecisionPolicy with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgUpdateGroupPolicyDecisionPolicy(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgUpdateGroupPolicyDecisionPolicy(txGen, ak, bk, k, NewSharedState())
}

func simulateMsgUpdateGroupPolicyDecisionPolicy(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		_, groupPolicy, acc, account, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyDecisionPolicy, ""), nil, err
		}
		if groupPolicy == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyDecisionPolicy, "no group policy found"), nil, nil
		}
		groupPolicyAddr := groupPolicy.Address

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyDecisionPolicy, "fee error"), nil, err
		}

		groupPolicyBech32, err := sdk.AccAddressFromBech32(groupPolicyAddr)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyDecisionPolicy, fmt.Sprintf("fail to decide bech32 address: %s", err.Error())), nil, nil
		}

		msg, err := group.NewMsgUpdateGroupPolicyDecisionPolicy(acc.Address, groupPolicyBech32, &group.ThresholdDecisionPolicy{
			Threshold: fmt.Sprintf("%d", simtypes.RandIntBetween(r, 1, 10)),
			Windows: &group.DecisionPolicyWindows{
				VotingPeriod: time.Second * time.Duration(simtypes.RandIntBetween(r, 100, 1000)),
			},
		})
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyDecisionPolicy, err.Error()), nil, err
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyDecisionPolicy, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgUpdateGroupPolicyMetadata generates a MsgUpdateGroupPolicyMetadata with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgUpdateGroupPolicyMetadata(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgUpdateGroupPolicyMetadata(txGen, ak, bk, k, NewSharedState())
}

func simulateMsgUpdateGroupPolicyMetadata(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		_, groupPolicy, acc, account, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyMetadata, ""), nil, err
		}
		if groupPolicy == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyMetadata, "no group policy found"), nil, nil
		}
		groupPolicyAddr := groupPolicy.Address

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyMetadata, "fee error"), nil, err
		}

		msg := &group.MsgUpdateGroupPolicyMetadata{
			Admin:              acc.Address.String(),
			GroupPolicyAddress: groupPolicyAddr,
			Metadata:           simtypes.RandStringOfLength(r, 10),
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyMetadata, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgWithdrawProposal generates a MsgWithdrawProposal with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgWithdrawProposal(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgWithdrawProposal(txGen, ak, bk, k, NewSharedState())
}

// simulateMsgWithdrawProposal generates a MsgWithdrawProposal with random values
func simulateMsgWithdrawProposal(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		g, groupPolicy, _, _, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, ""), nil, err
		}
		if g == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, "no group found"), nil, nil
		}
		if groupPolicy == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, "no group policy found"), nil, nil
		}

		groupPolicyAddr := groupPolicy.Address
		policy, err := groupPolicy.GetDecisionPolicy()
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, err.Error()), nil, nil
		}
		err = policy.Validate(*g, group.DefaultConfig())
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, err.Error()), nil, nil
		}

		proposalsResult, err := k.ProposalsByGroupPolicy(sdkCtx, &group.QueryProposalsByGroupPolicyRequest{Address: groupPolicyAddr})
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, "fail to query group info"), nil, err
		}

		proposals := proposalsResult.GetProposals()
		if len(proposals) == 0 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, "no proposals found"), nil, nil
		}

		var proposal *group.Proposal
		proposalID := -1

		for _, p := range proposals {
			if p.Status == group.PROPOSAL_STATUS_SUBMITTED {
				timeout := p.VotingPeriodEnd
				proposal = p
				proposalID = int(p.Id)
				if timeout.Before(sdkCtx.BlockTime()) || timeout.Equal(sdkCtx.BlockTime()) {
					return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, "voting period ended: skipping"), nil, nil
				}
				break
			}
		}

		// return no-op if no proposal found
		if proposalID == -1 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, "no proposals found"), nil, nil
		}

		// select a random proposer
		proposers := proposal.Proposers
		n := randIntInRange(r, len(proposers))
		proposerIdx := findAccount(accounts, proposers[n])
		proposer := accounts[proposerIdx]
		proposerAcc := ak.GetAccount(sdkCtx, proposer.Address)

		spendableCoins := bk.SpendableCoins(sdkCtx, proposer.Address)
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgWithdrawProposal, "fee error"), nil, err
		}

		msg := &group.MsgWithdrawProposal{
			ProposalId: uint64(proposalID),
			Address:    proposer.Address.String(),
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{proposerAcc.GetAccountNumber()},
			[]uint64{proposerAcc.GetSequence()},
			proposer.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyMetadata, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			if strings.Contains(err.Error(), "group was modified") || strings.Contains(err.Error(), "group policy was modified") {
				return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "no-op:group/group-policy was modified"), nil, nil
			}
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgVote generates a MsgVote with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgVote(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgVote(txGen, ak, bk, k, NewSharedState())
}

func simulateMsgVote(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		g, groupPolicy, _, _, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, ""), nil, err
		}
		if g == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "no group found"), nil, nil
		}
		if groupPolicy == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "no group policy found"), nil, nil
		}
		groupPolicyAddr := groupPolicy.Address

		// Pick a random member from the group
		acc, account, err := randomMember(sdkCtx, r, k, ak, accounts, g.Id)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, ""), nil, err
		}
		if account == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "no group member found"), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "fee error"), nil, err
		}

		proposalsResult, err := k.ProposalsByGroupPolicy(sdkCtx, &group.QueryProposalsByGroupPolicyRequest{Address: groupPolicyAddr})
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "fail to query group info"), nil, err
		}
		proposals := proposalsResult.GetProposals()
		if len(proposals) == 0 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "no proposals found"), nil, nil
		}

		proposalID := -1

		for _, p := range proposals {
			if p.Status == group.PROPOSAL_STATUS_SUBMITTED {
				timeout := p.VotingPeriodEnd
				proposalID = int(p.Id)
				if timeout.Before(sdkCtx.BlockTime()) || timeout.Equal(sdkCtx.BlockTime()) {
					return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "voting period ended: skipping"), nil, nil
				}
				break
			}
		}

		// return no-op if no proposal found
		if proposalID == -1 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "no proposals found"), nil, nil
		}

		// Ensure member hasn't already voted
		res, _ := k.VoteByProposalVoter(sdkCtx, &group.QueryVoteByProposalVoterRequest{
			Voter:      acc.Address.String(),
			ProposalId: uint64(proposalID),
		})
		if res != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "member has already voted"), nil, nil
		}

		msg := &group.MsgVote{
			ProposalId: uint64(proposalID),
			Voter:      acc.Address.String(),
			Option:     group.VOTE_OPTION_YES,
			Metadata:   simtypes.RandStringOfLength(r, 10),
		}
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyMetadata, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			if strings.Contains(err.Error(), "group was modified") || strings.Contains(err.Error(), "group policy was modified") {
				return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "no-op:group/group-policy was modified"), nil, nil
			}
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgExec generates a MsgExec with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgExec(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return simulateMsgExec(txGen, ak, bk, k, NewSharedState())
}

func simulateMsgExec(
	txGen client.TxConfig,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	k keeper.Keeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		_, groupPolicy, acc, account, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgExec, ""), nil, err
		}
		if groupPolicy == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgExec, "no group policy found"), nil, nil
		}
		groupPolicyAddr := groupPolicy.Address

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgExec, "fee error"), nil, err
		}

		proposalsResult, err := k.ProposalsByGroupPolicy(sdkCtx, &group.QueryProposalsByGroupPolicyRequest{Address: groupPolicyAddr})
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgExec, "fail to query group info"), nil, err
		}
		proposals := proposalsResult.GetProposals()
		if len(proposals) == 0 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgExec, "no proposals found"), nil, nil
		}

		proposalID := -1

		for _, proposal := range proposals {
			if proposal.Status == group.PROPOSAL_STATUS_ACCEPTED {
				proposalID = int(proposal.Id)
				break
			}
		}

		// return no-op if no proposal found
		if proposalID == -1 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgExec, "no proposals found"), nil, nil
		}

		msg := &group.MsgExec{
			ProposalId: uint64(proposalID),
			Executor:   acc.Address.String(),
		}
		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupPolicyMetadata, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			if strings.Contains(err.Error(), "group was modified") || strings.Contains(err.Error(), "group policy was modified") {
				return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "no-op:group/group-policy was modified"), nil, nil
			}
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

// SimulateMsgLeaveGroup generates a MsgLeaveGroup with random values
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgLeaveGroup(
	_ *codec.ProtoCodec,
	txGen client.TxConfig,
	k keeper.Keeper,
	ak group.AccountKeeper,
	bk group.BankKeeper,
) simtypes.Operation {
	return simulateMsgLeaveGroup(txGen, k, ak, bk, NewSharedState())
}

func simulateMsgLeaveGroup(
	txGen client.TxConfig,
	k keeper.Keeper,
	ak group.AccountKeeper,
	bk group.BankKeeper,
	s *SharedState,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupInfo, policyInfo, _, _, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts, s)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgLeaveGroup, ""), nil, err
		}

		if policyInfo == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgLeaveGroup, "no policy found"), nil, nil
		}

		// Pick a random member from the group
		acc, account, err := randomMember(sdkCtx, r, k, ak, accounts, groupInfo.Id)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgLeaveGroup, ""), nil, err
		}
		if account == nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgLeaveGroup, "no group member found"), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, acc.Address)
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgLeaveGroup, "fee error"), nil, err
		}

		msg := &group.MsgLeaveGroup{
			Address: acc.Address.String(),
			GroupId: groupInfo.Id,
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgLeaveGroup, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, sdk.MsgTypeURL(msg), err.Error()), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, err
	}
}

func randomGroup(r *rand.Rand, k keeper.Keeper, ak group.AccountKeeper,
	ctx sdk.Context, accounts []simtypes.Account, s *SharedState,
) (groupInfo *group.GroupInfo, acc simtypes.Account, account sdk.AccountI, err error) {
	groupID := k.GetGroupSequence(ctx)

	if initialGroupID := s.getMinGroupID(); initialGroupID == unsetGroupID {
		s.setMinGroupID(groupID)
	} else if initialGroupID < groupID {
		groupID = uint64(simtypes.RandIntBetween(r, int(initialGroupID+1), int(groupID+1)))
	}

	// when groupID is 0, it proves that SimulateMsgCreateGroup has never been called. that is, no group exists in the chain
	if groupID == 0 {
		return nil, simtypes.Account{}, nil, nil
	}

	res, err := k.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: groupID})
	if err != nil {
		return nil, simtypes.Account{}, nil, err
	}

	groupInfo = res.Info
	groupAdmin := groupInfo.Admin
	found := -1
	for i := range accounts {
		if accounts[i].Address.String() == groupAdmin {
			found = i
			break
		}
	}
	if found < 0 {
		return nil, simtypes.Account{}, nil, nil
	}
	acc = accounts[found]
	account = ak.GetAccount(ctx, acc.Address)
	return groupInfo, acc, account, nil
}

func randomGroupPolicy(r *rand.Rand, k keeper.Keeper, ak group.AccountKeeper,
	ctx sdk.Context, accounts []simtypes.Account, s *SharedState,
) (groupInfo *group.GroupInfo, groupPolicyInfo *group.GroupPolicyInfo, acc simtypes.Account, account sdk.AccountI, err error) {
	groupInfo, _, _, err = randomGroup(r, k, ak, ctx, accounts, s)
	if err != nil {
		return nil, nil, simtypes.Account{}, nil, err
	}
	if groupInfo == nil {
		return nil, nil, simtypes.Account{}, nil, nil
	}
	groupID := groupInfo.Id

	result, err := k.GroupPoliciesByGroup(ctx, &group.QueryGroupPoliciesByGroupRequest{GroupId: groupID})
	if err != nil {
		return groupInfo, nil, simtypes.Account{}, nil, err
	}

	n := randIntInRange(r, len(result.GroupPolicies))
	if n < 0 {
		return groupInfo, nil, simtypes.Account{}, nil, nil
	}
	groupPolicyInfo = result.GroupPolicies[n]

	idx := findAccount(accounts, groupPolicyInfo.Admin)
	if idx < 0 {
		return groupInfo, nil, simtypes.Account{}, nil, nil
	}
	acc = accounts[idx]
	account = ak.GetAccount(ctx, acc.Address)
	return groupInfo, groupPolicyInfo, acc, account, nil
}

func randomMember(ctx context.Context, r *rand.Rand, k keeper.Keeper, ak group.AccountKeeper,
	accounts []simtypes.Account, groupID uint64,
) (acc simtypes.Account, account sdk.AccountI, err error) {
	res, err := k.GroupMembers(ctx, &group.QueryGroupMembersRequest{
		GroupId: groupID,
	})
	if err != nil {
		return simtypes.Account{}, nil, err
	}
	n := randIntInRange(r, len(res.Members))
	if n < 0 {
		return simtypes.Account{}, nil, err
	}
	idx := findAccount(accounts, res.Members[n].Member.Address)
	if idx < 0 {
		return simtypes.Account{}, nil, err
	}
	acc = accounts[idx]
	account = ak.GetAccount(sdk.UnwrapSDKContext(ctx), acc.Address)
	return acc, account, nil
}

func randIntInRange(r *rand.Rand, l int) int {
	if l == 0 {
		return -1
	}
	if l == 1 {
		return 0
	}
	return simtypes.RandIntBetween(r, 0, l-1)
}

func findAccount(accounts []simtypes.Account, addr string) (idx int) {
	idx = -1
	for i := range accounts {
		if accounts[i].Address.String() == addr {
			idx = i
			break
		}
	}
	return idx
}

func genGroupMembers(r *rand.Rand, accounts []simtypes.Account) []group.MemberRequest {
	if len(accounts) == 1 {
		return []group.MemberRequest{
			{
				Address:  accounts[0].Address.String(),
				Weight:   fmt.Sprintf("%d", simtypes.RandIntBetween(r, 1, 10)),
				Metadata: simtypes.RandStringOfLength(r, 10),
			},
		}
	}

	maxMembers := min(len(accounts), 5)
	membersLen := simtypes.RandIntBetween(r, 1, maxMembers)
	members := make([]group.MemberRequest, membersLen)

	for i := range membersLen {
		members[i] = group.MemberRequest{
			Address:  accounts[i].Address.String(),
			Weight:   fmt.Sprintf("%d", simtypes.RandIntBetween(r, 1, 10)),
			Metadata: simtypes.RandStringOfLength(r, 10),
		}
	}

	return members
}
