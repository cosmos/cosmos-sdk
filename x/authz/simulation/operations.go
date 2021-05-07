package simulation

import (
	"context"
	"errors"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"

	banktype "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// authz message types
var (
	TypeMsgGrant  = sdk.MsgTypeURL(&authz.MsgGrant{})
	TypeMsgRevoke = sdk.MsgTypeURL(&authz.MsgRevoke{})
	TypeMsgExec   = sdk.MsgTypeURL(&authz.MsgExec{})
	sendLimit     = sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10)))
)

// Simulation operation weights constants
const (
	OpWeightMsgGrant = "op_weight_msg_grant"
	OpWeightRevoke   = "op_weight_msg_revoke"
	OpWeightExec     = "op_weight_msg_execute"
)

// authz operations weights
const (
	WeightGrant  = 100
	WeightRevoke = 100
	WeightExec   = 100
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak authz.AccountKeeper, bk authz.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker, protoCdc *codec.ProtoCodec) simulation.WeightedOperations {

	var (
		weightMsgGrant int
		weightRevoke   int
		weightExec     int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgGrant, &weightMsgGrant, nil,
		func(_ *rand.Rand) {
			weightMsgGrant = WeightGrant
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightRevoke, &weightRevoke, nil,
		func(_ *rand.Rand) {
			weightRevoke = WeightRevoke
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightExec, &weightExec, nil,
		func(_ *rand.Rand) {
			weightExec = WeightExec
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgGrant,
			SimulateMsgGrantAuthorization(ak, bk, k, protoCdc),
		),
		simulation.NewWeightedOperation(
			weightRevoke,
			SimulateMsgRevokeAuthorization(ak, bk, k, protoCdc),
		),
		simulation.NewWeightedOperation(
			weightExec,
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
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrant, "granter and grantee are same"), nil, nil
		}

		granterAcc := ak.GetAccount(ctx, granter.Address)
		spendableCoins := bk.SpendableCoins(ctx, granter.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrant, err.Error()), nil, err
		}

		spendLimit := spendableCoins.Sub(fees)
		if spendLimit == nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrant, "spend limit is nil"), nil, nil
		}

		expiration := ctx.BlockTime().AddDate(1, 0, 0)
		msg, err := authz.NewMsgGrant(granter.Address, grantee.Address, generateRandomAuthorization(r, spendLimit), expiration)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrant, err.Error()), nil, err
		}

		txCfg := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := authz.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.Grant(context.Background(), msg)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrant, err.Error()), nil, err
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
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgGrant, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txCfg.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, sdk.MsgTypeURL(svcMsgClientConn.GetMsgs()[0]), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, nil
	}
}

func generateRandomAuthorization(r *rand.Rand, spendLimit sdk.Coins) authz.Authorization {
	authorizations := make([]authz.Authorization, 2)
	authorizations[0] = banktype.NewSendAuthorization(spendLimit)
	authorizations[1] = authz.NewGenericAuthorization("/cosmos.bank.v1beta1.MsgSend")

	return authorizations[r.Intn(len(authorizations))]
}

// SimulateMsgRevokeAuthorization generates a MsgRevokeAuthorization with random values.
// nolint: funlen
func SimulateMsgRevokeAuthorization(ak authz.AccountKeeper, bk authz.BankKeeper, k keeper.Keeper, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var granterAddr, granteeAddr sdk.AccAddress
		var grant authz.Grant
		hasGrant := false

		k.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, g authz.Grant) bool {
			grant = g
			granterAddr = granter
			granteeAddr = grantee
			hasGrant = true
			return true
		})

		if !hasGrant {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevoke, "no grants"), nil, nil
		}

		if _, ok := simtypes.FindAccount(accs, granterAddr); !ok {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevoke, "Account not found"), nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "account not found")
		}

		spendableCoins := bk.SpendableCoins(ctx, granterAddr)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevoke, "fee error"), nil, err
		}

		authorizationGrant := grant.GetAuthorization()
		msg := authz.NewMsgRevoke(granterAddr, granteeAddr, authorizationGrant.MsgTypeURL())
		txCfg := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := authz.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.Revoke(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevoke, err.Error()), nil, err
		}

		granterKeys, _ := simtypes.FindAccount(accs, granterAddr)
		granterAcc := ak.GetAccount(ctx, granterAddr)

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
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevoke, err.Error()), nil, err
		}

		_, _, err = app.Deliver(txCfg.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, svcMsgClientConn.GetMsgs()[0].String(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, nil
	}
}

// SimulateMsgExecAuthorization generates a MsgExecAuthorized with random values.
func SimulateMsgExecAuthorization(ak authz.AccountKeeper, bk authz.BankKeeper, k keeper.Keeper, cdc cdctypes.AnyUnpacker, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
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
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, "Not found"), nil, nil
		}

		if _, ok := simtypes.FindAccount(accs, granteeAddr); !ok {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgRevoke, "Account not found"), nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "account not found")
		}

		// granterAccount := ak.GetAccount(ctx, granterAddr)

		granterspendableCoins := bk.SpendableCoins(ctx, granterAddr)
		if granterspendableCoins.Empty() {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, "no coins"), nil, nil
		}

		if targetGrant.Expiration.Before(ctx.BlockHeader().Time) {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, "grant expired"), nil, nil
		}

		granteeSpendableCoins := bk.SpendableCoins(ctx, granteeAddr)
		fees, err := simtypes.RandomFees(r, ctx, granteeSpendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, "fee error"), nil, err
		}

		execMsg := banktype.NewMsgSend(
			granterAddr,
			granteeAddr,
			sendLimit,
		)

		msg := authz.NewMsgExec(granteeAddr, []sdk.Msg{execMsg})

		var sendGrant authz.Authorization
		switch targetGrant.Authorization.TypeUrl {
		case "/cosmos.authz.v1beta1.GenericAuthorization":
			sendGrant = authz.NewGenericAuthorization("/cosmos.bank.v1beta1.MsgSend")
		case "/cosmos.bank.v1beta1.SendAuthorization":
			sendGrant = banktype.NewSendAuthorization(sendLimit)
		default:
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, "authorization type error"), nil, errors.New("invalid authorization")

		}

		_, err = sendGrant.Accept(ctx, execMsg)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, err.Error()), nil, nil
		}

		txCfg := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := authz.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.Exec(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, err.Error()), nil, err
		}

		granteeAcc := ak.GetAccount(ctx, granteeAddr)
		grantee1, _ := simtypes.FindAccount(accs, granteeAddr)
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
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, err.Error()), nil, err
		}

		_, _, err = app.Deliver(txCfg.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, err.Error()), nil, err
		}

		err = msg.UnpackInterfaces(cdc)
		if err != nil {
			return simtypes.NoOpMsg(authz.ModuleName, TypeMsgExec, "unmarshal error"), nil, err
		}

		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, nil
	}
}
