package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/nft"        //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/nft/keeper" //nolint:staticcheck // deprecated and to be removed
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// will be removed in the future
const (
	// OpWeightMsgSend Simulation operation weights constants
	OpWeightMsgSend = "op_weight_msg_send"

	// WeightSend nft operations weights
	WeightSend = 100
)

// TypeMsgSend will be removed in the future
var TypeMsgSend = sdk.MsgTypeURL(&nft.MsgSend{})

// WeightedOperations returns all the operations from the module with their respective weights
// migrate to the msg factories instead, this method will be removed in the future
func WeightedOperations(
	registry cdctypes.InterfaceRegistry,
	appParams simtypes.AppParams,
	_ codec.JSONCodec,
	txCfg client.TxConfig,
	ak nft.AccountKeeper,
	bk nft.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgSend int

	appParams.GetOrGenerate(OpWeightMsgSend, &weightMsgSend, nil,
		func(_ *rand.Rand) {
			weightMsgSend = WeightSend
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSend,
			SimulateMsgSend(codec.NewProtoCodec(registry), txCfg, ak, bk, k),
		),
	}
}

// SimulateMsgSend generates a MsgSend with random values.
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgSend(
	_ *codec.ProtoCodec,
	txCfg client.TxConfig,
	ak nft.AccountKeeper,
	bk nft.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		sender, _ := simtypes.RandomAcc(r, accs)
		receiver, _ := simtypes.RandomAcc(r, accs)

		if sender.Address.Equals(receiver.Address) {
			return simtypes.NoOpMsg(nft.ModuleName, TypeMsgSend, "sender and receiver are same"), nil, nil
		}

		senderAcc := ak.GetAccount(ctx, sender.Address)
		spendableCoins := bk.SpendableCoins(ctx, sender.Address)
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(nft.ModuleName, TypeMsgSend, err.Error()), nil, err
		}

		spendLimit := spendableCoins.Sub(fees...)
		if spendLimit == nil {
			return simtypes.NoOpMsg(nft.ModuleName, TypeMsgSend, "spend limit is nil"), nil, nil
		}

		n, err := randNFT(ctx, r, k, senderAcc.GetAddress())
		if err != nil {
			return simtypes.NoOpMsg(nft.ModuleName, TypeMsgSend, err.Error()), nil, err
		}

		senderStr, err := ak.AddressCodec().BytesToString(senderAcc.GetAddress().Bytes())
		if err != nil {
			return simtypes.NoOpMsg(nft.ModuleName, TypeMsgSend, err.Error()), nil, err
		}

		recieverStr, err := ak.AddressCodec().BytesToString(receiver.Address.Bytes())
		if err != nil {
			return simtypes.NoOpMsg(nft.ModuleName, TypeMsgSend, err.Error()), nil, err
		}

		msg := &nft.MsgSend{
			ClassId:  n.ClassId,
			Id:       n.Id,
			Sender:   senderStr,
			Receiver: recieverStr,
		}

		tx, err := simtestutil.GenSignedMockTx(
			r,
			txCfg,
			[]sdk.Msg{msg},
			fees,
			simtestutil.DefaultGenTxGas,
			chainID,
			[]uint64{senderAcc.GetAccountNumber()},
			[]uint64{senderAcc.GetSequence()},
			sender.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(nft.ModuleName, TypeMsgSend, "unable to generate mock tx"), nil, err
		}

		if _, _, err = app.SimDeliver(txCfg.TxEncoder(), tx); err != nil {
			return simtypes.NoOpMsg(nft.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// randNFT picks a random NFT from a class belonging to the specified owner(minter).
func randNFT(ctx sdk.Context, r *rand.Rand, k keeper.Keeper, minter sdk.AccAddress) (nft.NFT, error) {
	c, err := randClass(ctx, r, k)
	if err != nil {
		return nft.NFT{}, err
	}
	ns := k.GetNFTsOfClassByOwner(ctx, c.Id, minter)
	if len(ns) > 0 {
		return ns[r.Intn(len(ns))], nil
	}

	n := nft.NFT{
		ClassId: c.Id,
		Id:      simtypes.RandStringOfLength(r, 10),
		Uri:     simtypes.RandStringOfLength(r, 10),
	}
	err = k.Mint(ctx, n, minter)
	if err != nil {
		return nft.NFT{}, err
	}
	return n, nil
}

// randClass picks a random Class.
func randClass(ctx sdk.Context, r *rand.Rand, k keeper.Keeper) (nft.Class, error) {
	classes := k.GetClasses(ctx)
	if len(classes) == 0 {
		c := nft.Class{
			Id:          simtypes.RandStringOfLength(r, 10),
			Name:        simtypes.RandStringOfLength(r, 10),
			Symbol:      simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 10),
			Uri:         simtypes.RandStringOfLength(r, 10),
		}
		err := k.SaveClass(ctx, c)
		if err != nil {
			return nft.Class{}, err
		}
		return c, nil
	}
	return *classes[r.Intn(len(classes))], nil
}
