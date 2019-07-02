package nft

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// DONTCOVER

// SimulateMsgTransferNFT simulates the transfer of an NFT
func SimulateMsgTransferNFT(k Keeper) simulation.Operation {
	handler := GenericHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		ownerAddr, denom, nftID := getRandomNFTFromOwner(ctx, k, r)
		if ownerAddr.Empty() {
			return simulation.NoOpMsg(), nil, nil
		}

		msg := NewMsgTransferNFT(
			ownerAddr,                             // sender
			simulation.RandomAcc(r, accs).Address, // recipient
			denom,
			nftID,
		)

		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		ctx, write := ctx.CacheContext()
		ok := handler(ctx, msg).IsOK()
		if ok {
			write()
		}

		opMsg = simulation.NewOperationMsg(msg, ok, "")
		return opMsg, nil, nil
	}
}

// SimulateMsgEditNFTMetadata simulates an edit metadata transaction
func SimulateMsgEditNFTMetadata(k Keeper) simulation.Operation {
	handler := GenericHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		ownerAddr, denom, nftID := getRandomNFTFromOwner(ctx, k, r)
		if ownerAddr.Empty() {
			return simulation.NoOpMsg(), nil, nil
		}

		msg := NewMsgEditNFTMetadata(
			ownerAddr,
			denom,
			nftID,
			simulation.RandStringOfLength(r, 15), // name
			simulation.RandStringOfLength(r, 50), // description
			simulation.RandStringOfLength(r, 30), // image
			simulation.RandStringOfLength(r, 45), // tokenURI
		)

		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		ctx, write := ctx.CacheContext()
		ok := handler(ctx, msg).IsOK()
		if ok {
			write()
		}

		opMsg = simulation.NewOperationMsg(msg, ok, "")
		return opMsg, nil, nil
	}
}

func getRandomNFTFromOwner(ctx sdk.Context, k Keeper, r *rand.Rand) (address sdk.AccAddress, denom, nftID string) {
	owners := k.GetOwners(ctx)
	if len(owners) == 0 {
		return nil, "", ""
	}

	// get random validator
	i := r.Intn(len(owners))
	owner := owners[i]

	// get random collection from owner's balance
	i = r.Intn(len(owner.IDCollections))
	idsCollection := owner.IDCollections[i] // nfts IDs
	denom = idsCollection.Denom

	// get random nft from collection
	i = r.Intn(len(idsCollection.IDs))
	nftID = idsCollection.IDs[i]

	return owner.Address, denom, nftID
}
