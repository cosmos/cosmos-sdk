package simulation

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgTransferNFT     = "op_weight_msg_transfer_nft"
	OpWeightMsgEditNFTMetadata = "op_weight_msg_edit_nft_metadata"
	OpWeightMsgMintNFT         = "op_weight_msg_mint_nft"
	OpWeightMsgBurnNFT         = "op_weight_msg_burn_nft"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper,
	k keeper.Keeper) simulation.WeightedOperations {

	var (
		weightMsgTransferNFT     int
		weightMsgEditNFTMetadata int
		weightMsgMintNFT         int
		weightMsgBurnNFT         int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgTransferNFT, &weightMsgTransferNFT, nil,
		func(_ *rand.Rand) { weightMsgTransferNFT = 33 })

	appParams.GetOrGenerate(cdc, OpWeightMsgEditNFTMetadata, &weightMsgEditNFTMetadata, nil,
		func(_ *rand.Rand) { weightMsgTransferNFT = 5 })

	appParams.GetOrGenerate(cdc, OpWeightMsgMintNFT, &weightMsgMintNFT, nil,
		func(_ *rand.Rand) { weightMsgMintNFT = 10 })

	appParams.GetOrGenerate(cdc, OpWeightMsgBurnNFT, &weightMsgBurnNFT, nil,
		func(_ *rand.Rand) { weightMsgBurnNFT = 5 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightMsgTransferNFT,
			SimulateMsgTransferNFT(ak, k),
		),
		simulation.NewWeigthedOperation(
			weightMsgEditNFTMetadata,
			SimulateMsgEditNFTMetadata(ak, k),
		),
		simulation.NewWeigthedOperation(
			weightMsgMintNFT,
			SimulateMsgMintNFT(ak, k),
		),
		simulation.NewWeigthedOperation(
			weightMsgBurnNFT,
			SimulateMsgBurnNFT(ak, k),
		),
	}
}

// SimulateMsgTransferNFT simulates the transfer of an NFT
func SimulateMsgTransferNFT(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		ownerAddr, denom, nftID := getRandomNFTFromOwner(ctx, k, r)
		if ownerAddr.Empty() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		recipientAccount, _ := simulation.RandomAcc(r, accs)

		msg := types.NewMsgTransferNFT(
			ownerAddr, // sender
			recipientAccount.Address,
			denom,
			nftID,
		)

		simAccount, found := simulation.FindAccount(accs, ownerAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("account %s not found", ownerAddr)
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgEditNFTMetadata simulates an edit metadata transaction
func SimulateMsgEditNFTMetadata(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		ownerAddr, denom, nftID := getRandomNFTFromOwner(ctx, k, r)
		if ownerAddr.Empty() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgEditNFTMetadata(
			ownerAddr,
			nftID,
			denom,
			simulation.RandStringOfLength(r, 45), // tokenURI
		)

		simAccount, found := simulation.FindAccount(accs, ownerAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("account %s not found", ownerAddr)
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgMintNFT simulates a mint of an NFT
func SimulateMsgMintNFT(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		simAccount, _ := simulation.RandomAcc(r, accs)
		recipientAccount, _ := simulation.RandomAcc(r, accs)

		msg := types.NewMsgMintNFT(
			simAccount.Address,                   // sender
			recipientAccount.Address,             // recipient
			simulation.RandStringOfLength(r, 10), // nft ID
			simulation.RandStringOfLength(r, 10), // denom
			simulation.RandStringOfLength(r, 45), // tokenURI
		)

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgBurnNFT simulates a burn of an existing NFT
func SimulateMsgBurnNFT(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		ownerAddr, denom, nftID := getRandomNFTFromOwner(ctx, k, r)
		if ownerAddr.Empty() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgBurnNFT(ownerAddr, nftID, denom)

		simAccount, found := simulation.FindAccount(accs, ownerAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("account %s not found", ownerAddr)
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}
