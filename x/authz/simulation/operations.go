package simulation

import (
	"context"
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
	banktype "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// authz message types
var (
	TypeMsgGrantAuthorization  = sdk.MsgTypeURL(&authz.MsgGrant{})
	TypeMsgRevokeAuthorization = sdk.MsgTypeURL(&authz.MsgRevoke{})
	TypeMsgExecDelegated       = sdk.MsgTypeURL(&authz.MsgExec{})
)

// Simulation operation weights constants
const (
	OpWeightMsgGrantAuthorization = "op_weight_msg_grant"
	OpWeightRevokeAuthorization   = "op_weight_msg_revoke"
	OpWeightExecAuthorized        = "op_weight_msg_execute"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak authz.AccountKeeper, bk authz.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker, protoCdc *codec.ProtoCodec) simulation.WeightedOperations {

	var (
		weightMsgGrantAuthorization int
		weightRevokeAuthorization   int
		weightExecAuthorization     int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgGrantAuthorization, &weightMsgGrantAuthorization, nil,
		func(_ *rand.Rand) {
			weightMsgGrantAuthorization = WeightGrantAuthorization
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightRevokeAuthorization, &weightRevokeAuthorization, nil,
		func(_ *rand.Rand) {
			weightRevokeAuthorization = WeightRevokeAuthorization
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightExecAuthorization, &weightExecAuthorization, nil,
		func(_ *rand.Rand) {
			weightExecAuthorization = WeightExecAuthorization
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgGrantAuthorization,
			SimulateMsgGrantAuthorization(ak, bk, k, protoCdc),
		),
		simulation.NewWeightedOperation(
			weightRevokeAuthorization,
			SimulateMsgRevokeAuthorization(ak, bk, k, protoCdc),
		),
		simulation.NewWeightedOperation(
			weightExecAuthorization,
			SimulateMsgExecAuthorization(ak, bk, k, appCdc, protoCdc),
		),
	}
}

// SimulateMsgGrantAuthorization generates a MsgGrantAuthorization with random values.
// nolint: funlen
func SimulateMsgGrantAuthorization(ak authz.AccountKeeper, bk authz.BankKeeper, _ keeper.Keeper,
	protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		granter, _ := simtypes.RandomAcc(r, accs)
		grantee, _ := simtypes.RandomAcc(r, accs)

		if granter.Address.Equals(grantee.Address) {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, "granter and grantee are same"), nil, nil
		}

		granterAcc := ak.GetAccount(ctx, granter.Address)
		spendableCoins := bk.SpendableCoins(ctx, granter.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
		}

		spendLimit := spendableCoins.Sub(fees)
		if spendLimit == nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrantAuthorization, "spend limit is nil"), nil, nil
		}

		expiration := ctx.BlockTime().AddDate(1, 0, 0)
		msg, err := types.NewMsgGrantAuthorization(granter.Address, grantee.Address,
			generateRandomAuthorization(r, spendLimit), expiration)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
		}

		txCfg := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := authz.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.Grant(context.Background(), msg)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
		}

		tx, err := helpers.GenTx(
			txCfg,
			svcMsgClientConn.GetMsgs(),
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{granterAcc.GetAccountNumber()},
			[]uint64{granterAcc.GetSequence()},
			granter.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrantAuthorization, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txCfg.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, sdk.MsgTypeURL(svcMsgClientConn.GetMsgs()[0]), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, nil
	}
}

func generateRandomAuthorization(r *rand.Rand, spendLimit sdk.Coins) exported.Authorization {
	authorizations := make([]exported.Authorization, 2)
	authorizations[0] = banktype.NewSendAuthorization(spendLimit)
	authorizations[1] = types.NewGenericAuthorization("/cosmos.gov.v1beta1.Msg/SubmitProposal")

	return authorizations[r.Intn(len(authorizations))]
}

// SimulateMsgRevokeAuthorization generates a MsgRevokeAuthorization with random values.
// nolint: funlen
func SimulateMsgRevokeAuthorization(ak authz.AccountKeeper, bk authz.BankKeeper, k keeper.Keeper, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var granter, grantee sdk.AccAddress
		var grant types.AuthorizationGrant
		hasGrant := false

		k.IterateGrants(ctx, func(grntr, grntee sdk.AccAddress, g types.AuthorizationGrant) bool {
			grant = g
			granter = grntr
			grantee = grntee
			hasGrant = true
			return true
		})

		if !hasGrant {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevokeAuthorization, "no grants"), nil, nil
		}

		granter, ok := simtypes.FindAccount(accs, granterAddr)
		if !ok {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevokeAuthorization, "Account not found"), nil, nil
		}
		account := ak.GetAccount(ctx, granter.Address)

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevokeAuthorization, "fee error"), nil, err
		}

		auth := targetGrant.GetAuthorization()
		msg := authz.NewMsgRevoke(granterAddr, granteeAddr, auth.MsgTypeURL())
		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := authz.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.Revoke(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevokeAuthorization, err.Error()), nil, err
		}

		granterKeys, _ := simtypes.FindAccount(accs, granter)
		granterAcc := ak.GetAccount(ctx, granter)

		tx, err := helpers.GenTx(
			txCfg,
			svcMsgClientConn.GetMsgs(),
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{granterAcc.GetAccountNumber()},
			[]uint64{granterAcc.GetSequence()},
			granterKeys.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevokeAuthorization, err.Error()), nil, err
		}

		_, _, err = app.Deliver(txCfg.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, svcMsgClientConn.GetMsgs()[0].Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, nil
	}
}

// SimulateMsgExecAuthorization generates a MsgExecAuthorized with random values.
// nolint: funlen
func SimulateMsgExecuteAuthorized(ak authz.AccountKeeper, bk authz.BankKeeper, k keeper.Keeper, cdc cdctypes.AnyUnpacker, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var grant types.AuthorizationGrant
		hasGrant := false
		var targetGrant authz.Grant
		var granterAddr sdk.AccAddress
		var granteeAddr sdk.AccAddress
		k.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant authz.Grant) bool {
			targetGrant = grant
			granterAddr = granter
			granteeAddr = grantee
			hasGrant = true
			return true
		})

		if !hasGrant {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, "Not found"), nil, nil
		}

		grantee, _ := simtypes.FindAccount(accs, granteeAddr)
		granterAccount := ak.GetAccount(ctx, granterAddr)
		
		granterspendableCoins := bk.SpendableCoins(ctx, granterAccount.GetAddress())
		if granterspendableCoins.Empty() {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, "no coins"), nil, nil
		}

		if targetGrant.Expiration.Before(ctx.BlockHeader().Time) {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, "grant expired"), nil, nil
		}

		granteeSpendableCoins := bk.SpendableCoins(ctx, grantee)
		fees, err := simtypes.RandomFees(r, ctx, granteeSpendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, "fee error"), nil, err
		}

		authorization := grant.Authorization.GetCachedValue().(exported.Authorization)

		execMsg := banktype.NewMsgSend(
			granterAddr,
			granteeAddr,
			sendCoins,
		)

		msg := authz.NewMsgExec(grantee.Address, []sdk.Msg{execMsg})
		sendGrant := targetGrant.Authorization.GetCachedValue().(*banktype.SendAuthorization)
		_, err = sendGrant.Accept(ctx, execMsg)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, err.Error()), nil, nil
		}

		txCfg := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := authz.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.Exec(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, err.Error()), nil, err
		}

		granteeAcc := ak.GetAccount(ctx, grantee)
		grantee1, _ := simtypes.FindAccount(accs, grantee)
		tx, err := helpers.GenTx(
			txCfg,
			svcMsgClientConn.GetMsgs(),
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{granteeAcc.GetAccountNumber()},
			[]uint64{granteeAcc.GetSequence()},
			grantee1.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, err.Error()), nil, err
		}

		_, _, err = app.Deliver(txCfg.TxEncoder(), tx)
		if err != nil {
			if strings.Contains(err.Error(), "insufficient fee") {
				return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, "insufficient fee"), nil, nil
			}
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, err.Error()), nil, err
		}

		err = msg.UnpackInterfaces(cdc)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExecDelegated, "unmarshal error"), nil, err
		}

		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, nil
	}
}
