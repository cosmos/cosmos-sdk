package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

type keeper struct {
	cdc        codec.Codec
	storeKey   sdk.StoreKey
	bankKeeper bankkeeper.Keeper
	moduleName string
}

func (k keeper) nftDenom(collectionId, tokenId string) string {
	return fmt.Sprintf("%s/%s/%s", k.moduleName, collectionId, tokenId)
}

func (k keeper) issueCollection(ctx sdk.Context, id string, name string, schema string, sender sdk.AccAddress) error {
	return k.setCollectionInfo(ctx, nft.NewCollectionInfo(id, name, schema, sender))
}

func (k keeper) setCollectionInfo(ctx sdk.Context, info nft.CollectionInfo) error {
	if k.hasCollectionId(ctx, info.Id) {
		return sdkerrors.Wrapf(nft.ErrInvalidCollection, "collection ID %s already exists", info.Id)
	}

	if k.hasCollectionName(ctx, info.Name) {
		return sdkerrors.Wrapf(nft.ErrInvalidCollectionName, "collection name %s already exists", info.Name)
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&info)
	store.Set(KeyCollectionInfo(info.Id), bz)
	store.Set(KeyCollectionName(info.Name), []byte(info.Id))
	return nil
}

func (k keeper) hasCollectionId(ctx sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(KeyCollectionInfo(id))
}

func (k keeper) hasCollectionName(ctx sdk.Context, name string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(KeyCollectionName(name))
}

func (k keeper) mintNFT(
	ctx sdk.Context, collectionId, tokenID, tokenNm,
	tokenURI, tokenData string, owner sdk.AccAddress,
) error {

	if !k.hasCollectionId(ctx, collectionId) {
		return sdkerrors.Wrapf(nft.ErrInvalidCollection, "collection ID %s does not exist", collectionId)
	}

	denom := k.nftDenom(collectionId, tokenID)

	if !k.bankKeeper.GetSupply(ctx, denom).IsZero() {
		return sdkerrors.Wrapf(nft.ErrNFTAlreadyExists, "NFT %s already exists in collection %s", tokenID, collectionId)
	}

	coins := []sdk.Coin{sdk.NewInt64Coin(denom, 1)}

	if err := k.bankKeeper.MintCoins(ctx, k.moduleName, coins); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.moduleName, owner, coins); err != nil {
		return err
	}

	k.bankKeeper.SetDenomMetaData(ctx, types.Metadata{
		Description: tokenData,
		Base:        denom,
		Name:        tokenNm,
		Uri:         tokenURI,
	})

	k.increaseSupply(ctx, collectionId)

	return nil
}

func (k keeper) burnNFT(ctx sdk.Context, collectionId, tokenID string, owner sdk.AccAddress) error {
	denom := k.nftDenom(collectionId, tokenID)

	// note that this will burn fractional balances when bank supports it
	bal := k.bankKeeper.GetBalance(ctx, owner, denom)
	if bal.IsZero() {
		return sdkerrors.Wrapf(nft.ErrUnauthorized, "%s does not own NFT %s", owner, denom)
	}

	coins := sdk.Coins{bal}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, k.moduleName, coins); err != nil {
		return err
	}

	return k.bankKeeper.BurnCoins(ctx, k.moduleName, coins)
	// TODO delete metadata
}
