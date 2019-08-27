package operations

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// DONTCOVER

// SimulateMsgTransferNFT simulates the transfer of an NFT
func SimulateMsgTransferNFT(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		ownerAddr, denom, nftID := getRandomNFTFromOwner(ctx, k, r)
		if ownerAddr.Empty() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgTransferNFT(
			ownerAddr,                             // sender
			simulation.RandomAcc(r, accs).Address, // recipient
			denom,
			nftID,
		)

		acc, found := simulation.FindAccount(accs, ownerAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("account %s not found", ownerAddr)
		}

		fromAcc := ak.GetAccount(ctx, acc.Address)
		tx := simapp.GenTx([]sdk.Msg{msg},
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{acc.PrivKey}...)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgEditNFTMetadata simulates an edit metadata transaction
func SimulateMsgEditNFTMetadata(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

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

		acc, found := simulation.FindAccount(accs, ownerAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("account %s not found", ownerAddr)
		}

		fromAcc := ak.GetAccount(ctx, acc.Address)
		tx := simapp.GenTx([]sdk.Msg{msg},
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{acc.PrivKey}...)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgMintNFT simulates a mint of an NFT
func SimulateMsgMintNFT(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		acc := simulation.RandomAcc(r, accs)

		msg := types.NewMsgMintNFT(
			acc.Address,                           // sender
			simulation.RandomAcc(r, accs).Address, // recipient
			simulation.RandStringOfLength(r, 10),  // nft ID
			simulation.RandStringOfLength(r, 10),  // denom
			simulation.RandStringOfLength(r, 45),  // tokenURI
		)

		fromAcc := ak.GetAccount(ctx, acc.Address)
		tx := simapp.GenTx([]sdk.Msg{msg},
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{acc.PrivKey}...)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgBurnNFT simulates a burn of an existing NFT
func SimulateMsgBurnNFT(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		ownerAddr, denom, nftID := getRandomNFTFromOwner(ctx, k, r)
		if ownerAddr.Empty() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgBurnNFT(ownerAddr, nftID, denom)

		acc, found := simulation.FindAccount(accs, ownerAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("account %s not found", ownerAddr)
		}

		fromAcc := ak.GetAccount(ctx, acc.Address)
		tx := simapp.GenTx([]sdk.Msg{msg},
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{acc.PrivKey}...)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}
