package simulation

import (
	"context"
	"errors"
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
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
	banktype "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtype "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// authz message types
const (
	TypeMsgGrantAuthorization  = "/cosmos.authz.v1beta1.Msg/GrantAuthorization"
	TypeMsgRevokeAuthorization = "/cosmos.authz.v1beta1.Msg/RevokeAuthorization"
	TypeMsgExecAuthorization   = "/cosmos.authz.v1beta1.Msg/ExecAuthorized"
)

// Simulation operation weights constants
const (
	OpWeightMsgGrantAuthorization = "op_weight_msg_grant_authorization"
	OpWeightRevokeAuthorization   = "op_weight_msg_revoke_authorization"
	OpWeightExecAuthorization     = "op_weight_msg_execute_authorization"
)

// authz operations weights
const (
	WeightGrantAuthorization  = 100
	WeightRevokeAuthorization = 80
	WeightExecAuthorization   = 80
)

var sendLimit = sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10)))

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker, protoCdc *codec.ProtoCodec) simulation.WeightedOperations {

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
func SimulateMsgGrantAuthorization(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
	protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		granter := accs[0]
		grantee := accs[1]

		granterAcc := ak.GetAccount(ctx, granter.Address)
		spendableCoins := bk.SpendableCoins(ctx, granter.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
		}

		spendLimit := spendableCoins.Sub(fees)
		if spendLimit == nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, "spend limit is nil"), nil, nil
		}

		expiration := ctx.BlockTime().AddDate(1, 0, 0)
		msg, err := types.NewMsgGrantAuthorization(granter.Address, grantee.Address,
			generateRandomAuthorization(r, spendLimit), expiration)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
		}

		txCfg := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := types.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.GrantAuthorization(context.Background(), msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
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
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txCfg.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, svcMsgClientConn.GetMsgs()[0].Type(), "unable to deliver tx"), nil, err
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

func generateRandomAuthorizationType(r *rand.Rand) string {
	authorizationTypes := make([]string, 2)
	authorizationTypes[0] = banktype.SendAuthorization{}.MethodName()
	authorizationTypes[1] = "/cosmos.gov.v1beta1.Msg/SubmitProposal"

	return authorizationTypes[r.Intn(len(authorizationTypes))]
}

// SimulateMsgRevokeAuthorization generates a MsgRevokeAuthorization with random values.
// nolint: funlen
func SimulateMsgRevokeAuthorization(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, protoCdc *codec.ProtoCodec) simtypes.Operation {
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
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, "no authorizations exists"), nil, nil
		}

		spendableCoins := bk.SpendableCoins(ctx, granter)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, "fee error"), nil, err
		}

		txCfg := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := types.NewMsgClient(svcMsgClientConn)

		authorization := grant.Authorization.GetCachedValue().(exported.Authorization)

		msg := types.NewMsgRevokeAuthorization(granter, grantee, authorization.MethodName())
		_, err = authzMsgClient.RevokeAuthorization(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, err.Error()), nil, err
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
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, err.Error()), nil, err
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
func SimulateMsgExecAuthorization(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, cdc cdctypes.AnyUnpacker, protoCdc *codec.ProtoCodec) simtypes.Operation {
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
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, "no authorizations exists"), nil, nil
		}

		if grant.Expiration.Before(ctx.BlockHeader().Time) {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, "grant expired"), nil, nil
		}

		authorization := grant.Authorization.GetCachedValue().(exported.Authorization)

		execMsg := sdk.ServiceMsg{
			MethodName: authorization.MethodName(),
		}
		switch authorization.MethodName() {
		case banktype.SendAuthorization{}.MethodName():
			execMsg.Request = banktype.NewMsgSend(
				granter,
				grantee,
				sendLimit,
			)
		case "/cosmos.gov.v1beta1.Msg/SubmitProposal":
			proposal, err := govtype.NewMsgSubmitProposal(govtype.NewTextProposal(simtypes.RandStringOfLength(r, 10), simtypes.RandStringOfLength(r, 50)), sendLimit, granter)
			if err != nil {
				panic(err)
			}
			execMsg.Request = proposal
		default:
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, "unknown authorization"), nil, errors.New("unknown authorization")
		}

		granteespendableCoins := bk.SpendableCoins(ctx, grantee)
		fees, err := simtypes.RandomFees(r, ctx, granteespendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, "fee error"), nil, err
		}

		msg := types.NewMsgExecAuthorized(grantee, []sdk.ServiceMsg{execMsg})
		_, _, err = authorization.Accept(ctx, execMsg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, err.Error()), nil, err
		}

		txCfg := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := types.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.ExecAuthorized(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, err.Error()), nil, err
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
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, err.Error()), nil, err
		}

		_, _, err = app.Deliver(txCfg.TxEncoder(), tx)
		if err != nil {
			if strings.Contains(err.Error(), "insufficient fee") {
				return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, "insufficient fee"), nil, nil
			}
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, err.Error()), nil, err
		}

		err = msg.UnpackInterfaces(cdc)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecAuthorization, "unmarshal error"), nil, err
		}

		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, nil
	}
}
