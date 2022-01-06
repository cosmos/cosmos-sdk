package simulation

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/cosmos/cosmos-sdk/x/group"
)

var initialGroupID = uint64(100000000000000)

// group message types
var (
	TypeMsgCreateGroup                      = sdk.MsgTypeURL(&group.MsgCreateGroup{})
	TypeMsgUpdateGroupMembers               = sdk.MsgTypeURL(&group.MsgUpdateGroupMembers{})
	TypeMsgUpdateGroupAdmin                 = sdk.MsgTypeURL(&group.MsgUpdateGroupAdmin{})
	TypeMsgUpdateGroupMetadata              = sdk.MsgTypeURL(&group.MsgUpdateGroupMetadata{})
	TypeMsgCreateGroupAccount               = sdk.MsgTypeURL(&group.MsgCreateGroupAccount{})
	TypeMsgUpdateGroupAccountAdmin          = sdk.MsgTypeURL(&group.MsgUpdateGroupAccountAdmin{})
	TypeMsgUpdateGroupAccountDecisionPolicy = sdk.MsgTypeURL(&group.MsgUpdateGroupAccountDecisionPolicy{})
	TypeMsgUpdateGroupAccountMetadata       = sdk.MsgTypeURL(&group.MsgUpdateGroupAccountMetadata{})
	TypeMsgCreateProposal                   = sdk.MsgTypeURL(&group.MsgCreateProposal{})
	TypeMsgVote                             = sdk.MsgTypeURL(&group.MsgVote{})
	TypeMsgExec                             = sdk.MsgTypeURL(&group.MsgExec{})
)

// Simulation operation weights constants
const (
	OpMsgCreateGroup                      = "op_weight_msg_create_group"
	OpMsgUpdateGroupAdmin                 = "op_weight_msg_update_group_admin"
	OpMsgUpdateGroupMetadata              = "op_wieght_msg_update_group_metadata"
	OpMsgUpdateGroupMembers               = "op_weight_msg_update_group_members"
	OpMsgCreateGroupAccount               = "op_weight_msg_create_group_account"
	OpMsgUpdateGroupAccountAdmin          = "op_weight_msg_update_group_account_admin"
	OpMsgUpdateGroupAccountDecisionPolicy = "op_weight_msg_update_group_account_decision_policy"
	OpMsgUpdateGroupAccountMetaData       = "op_weight_msg_update_group_account_metadata"
	OpMsgCreateProposal                   = "op_weight_msg_create_proposal"
	OpMsgVote                             = "op_weight_msg_vote"
	OpMsgExec                             = "ops_weight_msg_exec"
)

// If update group or group account txn's executed, `SimulateMsgVote` & `SimulateMsgExec` txn's returns `noOp`.
// That's why we have less weight for update group & group-account txn's.
const (
	WeightMsgCreateGroup                      = 100
	WeightMsgCreateGroupAccount               = 100
	WeightMsgCreateProposal                   = 90
	WeightMsgVote                             = 90
	WeightMsgExec                             = 90
	WeightMsgUpdateGroupMetadata              = 5
	WeightMsgUpdateGroupAdmin                 = 5
	WeightMsgUpdateGroupMembers               = 5
	WeightMsgUpdateGroupAccountAdmin          = 5
	WeightMsgUpdateGroupAccountDecisionPolicy = 5
	WeightMsgUpdateGroupAccountMetadata       = 5
)

const GroupMemberWeight = 40

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak group.AccountKeeper,
	bk group.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker) simulation.WeightedOperations {
	var (
		weightMsgCreateGroup                      int
		weightMsgUpdateGroupAdmin                 int
		weightMsgUpdateGroupMetadata              int
		weightMsgUpdateGroupMembers               int
		weightMsgCreateGroupAccount               int
		weightMsgUpdateGroupAccountAdmin          int
		weightMsgUpdateGroupAccountDecisionPolicy int
		weightMsgUpdateGroupAccountMetadata       int
		weightMsgCreateProposal                   int
		weightMsgVote                             int
		weightMsgExec                             int
	)

	appParams.GetOrGenerate(cdc, OpMsgCreateGroup, &weightMsgCreateGroup, nil,
		func(_ *rand.Rand) {
			weightMsgCreateGroup = WeightMsgCreateGroup
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgCreateGroupAccount, &weightMsgCreateGroupAccount, nil,
		func(_ *rand.Rand) {
			weightMsgCreateGroupAccount = WeightMsgCreateGroupAccount
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgCreateProposal, &weightMsgCreateProposal, nil,
		func(_ *rand.Rand) {
			weightMsgCreateProposal = WeightMsgCreateProposal
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgVote, &weightMsgVote, nil,
		func(_ *rand.Rand) {
			weightMsgVote = WeightMsgVote
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgExec, &weightMsgExec, nil,
		func(_ *rand.Rand) {
			weightMsgExec = WeightMsgExec
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgUpdateGroupMetadata, &weightMsgUpdateGroupMetadata, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateGroupMetadata = WeightMsgUpdateGroupMetadata
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgUpdateGroupAdmin, &weightMsgUpdateGroupAdmin, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateGroupAdmin = WeightMsgUpdateGroupAdmin
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgUpdateGroupMembers, &weightMsgUpdateGroupMembers, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateGroupMembers = WeightMsgUpdateGroupMembers
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgUpdateGroupAccountAdmin, &weightMsgUpdateGroupAccountAdmin, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateGroupAccountAdmin = WeightMsgUpdateGroupAccountAdmin
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgUpdateGroupAccountDecisionPolicy, &weightMsgUpdateGroupAccountDecisionPolicy, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateGroupAccountDecisionPolicy = WeightMsgUpdateGroupAccountDecisionPolicy
		},
	)
	appParams.GetOrGenerate(cdc, OpMsgUpdateGroupAccountMetaData, &weightMsgUpdateGroupAccountMetadata, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateGroupAccountMetadata = WeightMsgUpdateGroupAccountMetadata
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateGroup,
			SimulateMsgCreateGroup(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateGroupAccount,
			SimulateMsgCreateGroupAccount(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateProposal,
			SimulateMsgCreateProposal(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgVote,
			SimulateMsgVote(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgExec,
			SimulateMsgExec(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateGroupMetadata,
			SimulateMsgUpdateGroupMetadata(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateGroupAdmin,
			SimulateMsgUpdateGroupAdmin(ak, bk, k),
		),
		// simulation.NewWeightedOperation(
		// 	weightMsgUpdateGroupMembers,
		// 	SimulateMsgUpdateGroupMembers(ak, bk, k),
		// ),
		// simulation.NewWeightedOperation(
		// 	weightMsgUpdateGroupAccountAdmin,
		// 	SimulateMsgUpdateGroupAccountAdmin(ak, bk, k),
		// ),
		// simulation.NewWeightedOperation(
		// 	weightMsgUpdateGroupAccountDecisionPolicy,
		// 	SimulateMsgUpdateGroupAccountDecisionPolicy(ak, bk, k, appCdc),
		// ),
		// simulation.NewWeightedOperation(
		// 	weightMsgUpdateGroupAccountMetadata,
		// 	SimulateMsgUpdateGroupAccountMetadata(ak, bk, k),
		// ),
	}
}

// SimulateMsgCreateGroup generates a MsgCreateGroup with random values
func SimulateMsgCreateGroup(ak group.AccountKeeper, bk group.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		acc, _ := simtypes.RandomAcc(r, accounts)
		account := ak.GetAccount(ctx, acc.Address)
		accAddr := acc.Address.String()

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroup, "fee error"), nil, err
		}

		members := []group.Member{
			{
				Address:  accAddr,
				Weight:   fmt.Sprintf("%d", GroupMemberWeight),
				Metadata: []byte(simtypes.RandStringOfLength(r, 10)),
			},
		}

		msg := &group.MsgCreateGroup{Admin: accAddr, Members: members, Metadata: []byte(simtypes.RandStringOfLength(r, 10))}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
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
			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, err
	}
}

// SimulateMsgCreateGroupAccount generates a NewMsgCreateGroupAccount with random values
func SimulateMsgCreateGroupAccount(ak group.AccountKeeper, bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupID, acc, account, err := randomGroup(r, k, ak, sdkCtx, accounts)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, ""), nil, err
		}
		if groupID == 0 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, ""), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, "fee error"), nil, err
		}

		msg, err := group.NewMsgCreateGroupAccount(
			acc.Address,
			groupID,
			[]byte(simtypes.RandStringOfLength(r, 10)),
			&group.ThresholdDecisionPolicy{
				Threshold: "20",
				Timeout:   time.Second * time.Duration(30*24*60*60),
			},
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, err.Error()), nil, err
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			fmt.Printf("ERR DELIVER %v\n", err)
			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, err
	}
}

// SimulateMsgCreateProposal generates a NewMsgCreateProposal with random values
func SimulateMsgCreateProposal(ak group.AccountKeeper, bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupPolicyAddr, acc, account, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateProposal, ""), nil, err
		}
		if groupPolicyAddr == "" {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateProposal, ""), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateProposal, "fee error"), nil, err
		}

		msg := group.MsgCreateProposal{
			Address:   groupPolicyAddr,
			Proposers: []string{acc.Address.String()},
			Metadata:  []byte(simtypes.RandStringOfLength(r, 10)),
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{&msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateProposal, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		// err = msg.UnpackInterfaces(appCdc)
		// if err != nil {
		// 	return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateProposal, "unmarshal error"), nil, err
		// }
		return simtypes.NewOperationMsg(&msg, true, "", nil), nil, err
	}
}

// SimulateMsgUpdateGroupAdmin generates a MsgUpdateGroupAdmin with random values
func SimulateMsgUpdateGroupAdmin(ak group.AccountKeeper, bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupID, acc, account, err := randomGroup(r, k, ak, sdkCtx, accounts)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, ""), nil, err
		}
		if groupID == 0 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, ""), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAdmin, "fee error"), nil, err
		}

		newAdmin, _ := simtypes.RandomAcc(r, accounts)
		// disallow setting current admin as new admin
		for acc.PubKey.Equals(newAdmin.PubKey) {
			newAdmin, _ = simtypes.RandomAcc(r, accounts)
		}

		msg := group.MsgUpdateGroupAdmin{
			GroupId:  groupID,
			Admin:    account.GetAddress().String(),
			NewAdmin: newAdmin.Address.String(),
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{&msg},
			fees,
			helpers.DefaultGenTxGas,
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
			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(&msg, true, "", nil), nil, err
	}
}

// SimulateMsgUpdateGroupMetadata generates a MsgUpdateGroupMetadata with random values
func SimulateMsgUpdateGroupMetadata(ak group.AccountKeeper, bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupID, acc, account, err := randomGroup(r, k, ak, sdkCtx, accounts)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, ""), nil, err
		}
		if groupID == 0 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, ""), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMetadata, "fee error"), nil, err
		}

		msg := group.MsgUpdateGroupMetadata{
			GroupId:  groupID,
			Admin:    account.GetAddress().String(),
			Metadata: []byte(simtypes.RandStringOfLength(r, 10)),
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{&msg},
			fees,
			helpers.DefaultGenTxGas,
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
			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(&msg, true, "", nil), nil, err
	}
}

// // SimulateMsgUpdateGroupMembers generates a MsgUpdateGroupMembers with random values
// func SimulateMsgUpdateGroupMembers(ak group.AccountKeeper,
// 	bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
// 	return func(
// 		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
// 		acc1 := accounts[0]
// 		acc2 := accounts[1]
// 		acc3 := accounts[2]
// 		account := ak.GetAccount(sdkCtx, acc1.Address)

// 		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
// 		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMembers, "fee error"), nil, err
// 		}

// 		ctx := sdk.WrapSDKContext(sdkCtx)

// 		groupAdmin, groupID, op, err := getGroupDetails(ctx, k, acc1)
// 		if err != nil {
// 			return op, nil, err
// 		}
// 		if groupAdmin == "" {
// 			return op, nil, nil
// 		}

// 		groupAccounts, opMsg, err := groupAccountsByAdmin(ctx, k, groupAdmin)
// 		if groupAccounts == nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMembers, opMsg), nil, err
// 		}

// 		members := []group.Member{
// 			{
// 				Address:  acc2.Address.String(),
// 				Weight:   fmt.Sprintf("%d", GroupMemberWeight),
// 				Metadata: []byte(simtypes.RandStringOfLength(r, 10)),
// 			},
// 			{
// 				Address:  acc3.Address.String(),
// 				Weight:   fmt.Sprintf("%d", GroupMemberWeight),
// 				Metadata: []byte(simtypes.RandStringOfLength(r, 10)),
// 			},
// 		}

// 		msg := group.MsgUpdateGroupMembers{
// 			GroupId:       groupID,
// 			Admin:         groupAdmin,
// 			MemberUpdates: members,
// 		}

// 		txGen := simappparams.MakeTestEncodingConfig().TxConfig
// 		tx, err := helpers.GenTx(
// 			txGen,
// 			[]sdk.Msg{&msg},
// 			fees,
// 			helpers.DefaultGenTxGas,
// 			chainID,
// 			[]uint64{account.GetAccountNumber()},
// 			[]uint64{account.GetSequence()},
// 			acc1.PrivKey,
// 		)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupMembers, "unable to generate mock tx"), nil, err
// 		}

// 		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
// 		}

// 		return simtypes.NewOperationMsg(&msg, true, "", nil), nil, err
// 	}
// }

// // SimulateMsgUpdateGroupAccountAdmin generates a MsgUpdateGroupAccountAdmin with random values
// func SimulateMsgUpdateGroupAccountAdmin(ak group.AccountKeeper, bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
// 	return func(
// 		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
// 		acc1 := accounts[0]
// 		acc2 := accounts[1]

// 		account := ak.GetAccount(sdkCtx, acc1.Address)

// 		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
// 		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountAdmin, "fee error"), nil, err
// 		}

// 		ctx := sdk.WrapSDKContext(sdkCtx)

// 		groupAdmin, _, op, err := getGroupDetails(ctx, k, acc1)
// 		if err != nil {
// 			return op, nil, err
// 		}
// 		if groupAdmin == "" {
// 			return op, nil, nil
// 		}

// 		groupAccounts, opMsg, err := groupAccountsByAdmin(ctx, k, groupAdmin)
// 		if groupAccounts == nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountAdmin, opMsg), nil, err
// 		}

// 		msg := group.MsgUpdateGroupAccountAdmin{
// 			Admin:    groupAccounts[0].Admin,
// 			Address:  groupAccounts[0].Address,
// 			NewAdmin: acc2.Address.String(),
// 		}

// 		txGen := simappparams.MakeTestEncodingConfig().TxConfig
// 		tx, err := helpers.GenTx(
// 			txGen,
// 			[]sdk.Msg{&msg},
// 			fees,
// 			helpers.DefaultGenTxGas,
// 			chainID,
// 			[]uint64{account.GetAccountNumber()},
// 			[]uint64{account.GetSequence()},
// 			acc1.PrivKey,
// 		)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountAdmin, "unable to generate mock tx"), nil, err
// 		}

// 		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
// 		}

// 		return simtypes.NewOperationMsg(&msg, true, "", nil), nil, err
// 	}
// }

// // SimulateMsgUpdateGroupAccountDecisionPolicy generates a NewMsgUpdateGroupAccountDecisionPolicyRequest with random values
// func SimulateMsgUpdateGroupAccountDecisionPolicy(ak group.AccountKeeper,
// 	bk group.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker) simtypes.Operation {
// 	return func(
// 		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
// 		acc1 := accounts[0]
// 		account := ak.GetAccount(sdkCtx, acc1.Address)

// 		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
// 		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountDecisionPolicy, "fee error"), nil, err
// 		}

// 		ctx := sdk.WrapSDKContext(sdkCtx)

// 		groupAdmin, _, op, err := getGroupDetails(ctx, k, acc1)
// 		if err != nil {
// 			return op, nil, err
// 		}
// 		if groupAdmin == "" {
// 			return op, nil, nil
// 		}

// 		groupAccounts, opMsg, err := groupAccountsByAdmin(ctx, k, groupAdmin)
// 		if groupAccounts == nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountDecisionPolicy, opMsg), nil, err
// 		}

// 		adminBech32, err := sdk.AccAddressFromBech32(groupAccounts[0].Admin)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountDecisionPolicy, fmt.Sprintf("fail to decide bech32 address: %s", err.Error())), nil, nil
// 		}

// 		groupAccountBech32, err := sdk.AccAddressFromBech32(groupAccounts[0].Address)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountDecisionPolicy, fmt.Sprintf("fail to decide bech32 address: %s", err.Error())), nil, nil
// 		}

// 		msg, err := group.NewMsgUpdateGroupAccountDecisionPolicyRequest(adminBech32, groupAccountBech32, &group.ThresholdDecisionPolicy{
// 			Threshold: fmt.Sprintf("%d", simtypes.RandIntBetween(r, 1, 20)),
// 			Timeout:   time.Second * time.Duration(simtypes.RandIntBetween(r, 100, 1000)),
// 		})
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountDecisionPolicy, err.Error()), nil, err
// 		}

// 		txGen := simappparams.MakeTestEncodingConfig().TxConfig
// 		tx, err := helpers.GenTx(
// 			txGen,
// 			[]sdk.Msg{msg},
// 			fees,
// 			helpers.DefaultGenTxGas,
// 			chainID,
// 			[]uint64{account.GetAccountNumber()},
// 			[]uint64{account.GetSequence()},
// 			acc1.PrivKey,
// 		)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountDecisionPolicy, "unable to generate mock tx"), nil, err
// 		}

// 		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
// 		}
// 		err = msg.UnpackInterfaces(appCdc)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountDecisionPolicy, "unmarshal error"), nil, err
// 		}
// 		return simtypes.NewOperationMsg(msg, true, "", nil), nil, err
// 	}
// }

// // SimulateMsgUpdateGroupAccountMetadata generates a MsgUpdateGroupAccountMetadata with random values
// func SimulateMsgUpdateGroupAccountMetadata(ak group.AccountKeeper,
// 	bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
// 	return func(
// 		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
// 		acc1 := accounts[0]

// 		account := ak.GetAccount(sdkCtx, acc1.Address)

// 		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
// 		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountMetadata, "fee error"), nil, err
// 		}

// 		ctx := sdk.WrapSDKContext(sdkCtx)

// 		groupAdmin, _, op, err := getGroupDetails(ctx, k, acc1)
// 		if err != nil {
// 			return op, nil, err
// 		}
// 		if groupAdmin == "" {
// 			return op, nil, nil
// 		}

// 		groupAccounts, opMsg, err := groupAccountsByAdmin(ctx, k, groupAdmin)
// 		if groupAccounts == nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountMetadata, opMsg), nil, err
// 		}

// 		msg := group.MsgUpdateGroupAccountMetadata{
// 			Admin:    groupAccounts[0].Admin,
// 			Address:  groupAccounts[0].Address,
// 			Metadata: []byte(simtypes.RandStringOfLength(r, 10)),
// 		}

// 		txGen := simappparams.MakeTestEncodingConfig().TxConfig
// 		tx, err := helpers.GenTx(
// 			txGen,
// 			[]sdk.Msg{&msg},
// 			fees,
// 			helpers.DefaultGenTxGas,
// 			chainID,
// 			[]uint64{account.GetAccountNumber()},
// 			[]uint64{account.GetSequence()},
// 			acc1.PrivKey,
// 		)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountMetadata, "unable to generate mock tx"), nil, err
// 		}

// 		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
// 		if err != nil {
// 			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
// 		}

// 		return simtypes.NewOperationMsg(&msg, true, "", nil), nil, err
// 	}
// }

// SimulateMsgVote generates a MsgVote with random values
func SimulateMsgVote(ak group.AccountKeeper,
	bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupPolicyAddr, acc1, account, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, ""), nil, err
		}
		if groupPolicyAddr == "" {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, ""), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "fee error"), nil, err
		}

		ctx := sdk.WrapSDKContext(sdkCtx)
		proposalsResult, err := k.ProposalsByGroupAccount(ctx, &group.QueryProposalsByGroupAccountRequest{Address: groupPolicyAddr})
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "fail to query group info"), nil, err
		}
		proposals := proposalsResult.GetProposals()
		if len(proposals) == 0 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "no proposals found"), nil, nil
		}

		proposalID := -1

		for _, proposal := range proposals {
			if proposal.Status == group.ProposalStatusSubmitted {
				timeout := proposal.Timeout
				if err != nil {
					return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "error: while converting to timestamp"), nil, err
				}
				proposalID = int(proposal.ProposalId)
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

		msg := group.MsgVote{
			ProposalId: uint64(proposalID),
			Voter:      acc1.Address.String(),
			Choice:     group.Choice_CHOICE_YES,
			Metadata:   []byte(simtypes.RandStringOfLength(r, 10)),
		}
		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{&msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc1.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountMetadata, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)

		if err != nil {
			if strings.Contains(err.Error(), "group was modified") || strings.Contains(err.Error(), "group account was modified") {
				return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "no-op:group/group-account was modified"), nil, nil
			}
			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(&msg, true, "", nil), nil, err
	}
}

// // SimulateMsgExec generates a MsgExec with random values
func SimulateMsgExec(ak group.AccountKeeper,
	bk group.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, sdkCtx sdk.Context, accounts []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		groupPolicyAddr, acc1, account, err := randomGroupPolicy(r, k, ak, sdkCtx, accounts)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, ""), nil, err
		}
		if groupPolicyAddr == "" {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, ""), nil, nil
		}

		spendableCoins := bk.SpendableCoins(sdkCtx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, sdkCtx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "fee error"), nil, err
		}

		ctx := sdk.WrapSDKContext(sdkCtx)
		proposalsResult, err := k.ProposalsByGroupAccount(ctx, &group.QueryProposalsByGroupAccountRequest{Address: groupPolicyAddr})
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "fail to query group info"), nil, err
		}
		proposals := proposalsResult.GetProposals()
		if len(proposals) == 0 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "no proposals found"), nil, nil
		}

		proposalID := -1

		for _, proposal := range proposals {
			if proposal.Status == group.ProposalStatusClosed {
				proposalID = int(proposal.ProposalId)
				break
			}
		}

		// return no-op if no proposal found
		if proposalID == -1 {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgVote, "no proposals found"), nil, nil
		}

		msg := group.MsgExec{
			ProposalId: uint64(proposalID),
			Signer:     acc1.Address.String(),
		}
		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{&msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			acc1.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(group.ModuleName, TypeMsgUpdateGroupAccountMetadata, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			if strings.Contains(err.Error(), "group was modified") || strings.Contains(err.Error(), "group account was modified") {
				return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "no-op:group/group-account was modified"), nil, nil
			}
			return simtypes.NoOpMsg(group.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(&msg, true, "", nil), nil, err
	}
}

func randomGroup(r *rand.Rand, k keeper.Keeper, ak group.AccountKeeper,
	ctx sdk.Context, accounts []simtypes.Account) (groupID uint64, acc simtypes.Account, account authtypes.AccountI, err error) {
	groupID = k.GetGroupSequence(ctx)

	switch {
	case groupID > initialGroupID:
		// select a random ID between [initialGroupID, groupID]
		groupID = uint64(simtypes.RandIntBetween(r, int(initialGroupID), int(groupID)))

	default:
		// This is called on the first call to this function
		// in order to update the global variable
		initialGroupID = groupID
	}

	res, err := k.GroupInfo(sdk.WrapSDKContext(ctx), &group.QueryGroupInfoRequest{GroupId: groupID})
	if err != nil {
		return 0, simtypes.Account{}, nil, err
	}

	groupAdmin := res.Info.Admin
	found := -1
	for i := range accounts {
		if accounts[i].Address.String() == groupAdmin {
			found = i
			break
		}
	}
	if found < 0 {
		return 0, simtypes.Account{}, nil, nil
	}
	acc = accounts[found]
	account = ak.GetAccount(ctx, acc.Address)
	return groupID, acc, account, nil
}

func randomGroupPolicy(r *rand.Rand, k keeper.Keeper, ak group.AccountKeeper,
	ctx sdk.Context, accounts []simtypes.Account) (groupPolicyAddr string, acc simtypes.Account, account authtypes.AccountI, err error) {
	groupID, acc, account, err := randomGroup(r, k, ak, ctx, accounts)
	if err != nil {
		return "", simtypes.Account{}, nil, err
	}
	if groupID == 0 {
		return "", simtypes.Account{}, nil, nil
	}

	result, err := k.GroupAccountsByGroup(sdk.WrapSDKContext(ctx), &group.QueryGroupAccountsByGroupRequest{GroupId: groupID})
	if err != nil {
		return "", simtypes.Account{}, nil, err
	}
	var n uint64
	l := len(result.GroupAccounts)
	if l == 1 {
		n = 0
	} else {
		n = uint64(simtypes.RandIntBetween(r, 0, l-1))
	}
	groupPolicyAddr = result.GroupAccounts[n].Address
	return groupPolicyAddr, acc, account, nil
}

// func randomActiveProposal(r *rand.Rand, k keeper.Keeper, ak group.AccountKeeper,
// 	ctx sdk.Context, accounts []simtypes.Account) (proposalID uint64, acc simtypes.Account, account authtypes.AccountI, err error) {
// 	groupPolicyAddr, acc, account, err := randomGroupPolicy(r, k, ak, ctx, accounts)
// }

func getGroupDetails(ctx context.Context, k keeper.Keeper, acc simtypes.Account) (groupAdmin string, groupID uint64, op simtypes.OperationMsg, err error) {
	groups, err := k.GroupsByAdmin(ctx, &group.QueryGroupsByAdminRequest{Admin: acc.Address.String()})
	if err != nil {
		return "", 0, simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, "fail to query groups"), err
	}

	if len(groups.Groups) == 0 {
		return "", 0, simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, ""), nil
	}

	groupAdmin = groups.Groups[0].Admin
	groupID = groups.Groups[0].GroupId

	return groupAdmin, groupID, simtypes.NoOpMsg(group.ModuleName, TypeMsgCreateGroupAccount, ""), nil
}

func groupAccountsByAdmin(ctx context.Context, k keeper.Keeper, admin string) ([]*group.GroupAccountInfo, string, error) {
	result, err := k.GroupAccountsByAdmin(ctx, &group.QueryGroupAccountsByAdminRequest{Admin: admin})
	if err != nil {
		return nil, "fail to query group info", err
	}

	groupAccounts := result.GetGroupAccounts()
	if len(groupAccounts) == 0 {
		return nil, "no group account found", nil
	}
	return groupAccounts, "", nil
}
