package keeper

import (
	"context"
	"cosmossdk.io/math"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"cosmossdk.io/errors"
)

// Keeper of the nft store
type Keeper struct {
	appmodule.Environment
	cdc codec.BinaryCodec
	bk  nft.BankKeeper
	ac  address.Codec
}

// NewKeeper creates a new nft Keeper instance
func NewKeeper(env appmodule.Environment,
	cdc codec.BinaryCodec, ak nft.AccountKeeper, bk nft.BankKeeper,
) Keeper {
	// ensure nft module account is set
	if addr := ak.GetModuleAddress(nft.ModuleName); addr == nil {
		panic("the nft module account has not been set")
	}
	return Keeper{
		Environment: env,
		cdc:         cdc,
		bk:          bk,
		ac:          ak.AddressCodec(),
	}
}

// Stake locks an NFT for a specified duration
func (k Keeper) Stake(ctx context.Context, classId string, nftId string, owner sdk.AccAddress, stakeDuration uint64) error {
	// Implementation of staking logic
	// This is a placeholder and needs to be implemented based on your requirements
	return nil
}

// New method to set the creator
func (k Keeper) setCreator(ctx context.Context, classID string, nftID string, creator string) {
	store := k.KVStoreService.OpenKVStore(ctx)
	key := creatorStoreKey(classID, nftID)
	store.Set(key, []byte(creator))
}

// Helper function to generate the creator store key
func creatorStoreKey(classID, nftID string) []byte {
	return []byte(fmt.Sprintf("%s/creator/%s/%s", nft.ModuleName, classID, nftID))
}

func (k Keeper) WithdrawRoyaltiesInternal(ctx context.Context, classID string, nftID string, role string, recipient sdk.AccAddress) (sdk.Coin, error) {
	store := k.KVStoreService.OpenKVStore(ctx)
	key := royaltyStoreKey(classID, nftID)
	var accumulatedRoyalties nft.AccumulatedRoyalties
	bz, err := store.Get(key)
	if err != nil {
		return sdk.Coin{}, err
	}
	if bz == nil {
		return sdk.Coin{}, fmt.Errorf("no royalties found for NFT %s in class %s", nftID, classID)
	}
	k.cdc.MustUnmarshal(bz, &accumulatedRoyalties)

	var amount sdk.Coin
	platformWalletAddress := sdk.MustAccAddressFromBech32("cosmos1d9ms9wf4yx3vky2kp6fc7t3qm9p8ps33g49c9s")

	switch role {
	case "creator":
		amount, err = sdk.ParseCoinNormalized(accumulatedRoyalties.CreatorRoyalties)
		if err != nil {
			return sdk.Coin{}, err
		}
		accumulatedRoyalties.CreatorRoyalties = sdk.NewCoin(amount.Denom, math.ZeroInt()).String()
	case "platform":
		if !recipient.Equals(platformWalletAddress) {
			return sdk.Coin{}, fmt.Errorf("unauthorized withdrawal attempt for platform royalties")
		}
		amount, err = sdk.ParseCoinNormalized(accumulatedRoyalties.PlatformRoyalties)
		if err != nil {
			return sdk.Coin{}, err
		}
		accumulatedRoyalties.PlatformRoyalties = sdk.NewCoin(amount.Denom, math.ZeroInt()).String()
	case "owner":
		amount, err = sdk.ParseCoinNormalized(accumulatedRoyalties.OwnerRoyalties)
		if err != nil {
			return sdk.Coin{}, err
		}
		accumulatedRoyalties.OwnerRoyalties = sdk.NewCoin(amount.Denom, math.ZeroInt()).String()
	default:
		return sdk.Coin{}, fmt.Errorf("invalid role: %s", role)
	}

	if amount.IsZero() {
		return sdk.Coin{}, fmt.Errorf("no royalties to withdraw for role %s", role)
	}

	err = k.bk.SendCoinsFromModuleToAccount(ctx, nft.ModuleName, recipient, sdk.NewCoins(amount))
	if err != nil {
		return sdk.Coin{}, err
	}

	// Update the accumulated royalties
	err = store.Set(key, k.cdc.MustMarshal(&accumulatedRoyalties))
	if err != nil {
		return sdk.Coin{}, err
	}

	// Emit event using the EventManager from the context
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err = sdkCtx.EventManager().EmitTypedEvent(&nft.EventRoyaltyWithdraw{
		ClassId: classID,
		Id:      nftID,
		Role:    role,
		Amount:  amount.String(),
	})
	if err != nil {
		return sdk.Coin{}, err
	}

	return amount, nil
}

// GetAccumulatedRoyalties retrieves the accumulated royalties for an NFT
func (k Keeper) GetAccumulatedRoyalties(ctx context.Context, classID, nftID string) (nft.AccumulatedRoyalties, bool) {
	store := k.KVStoreService.OpenKVStore(ctx)
	key := royaltyStoreKey(classID, nftID)
	bz, err := store.Get(key)
	if err != nil {
		return nft.AccumulatedRoyalties{}, false
	}
	if bz == nil {
		return nft.AccumulatedRoyalties{}, false
	}

	var royalties nft.AccumulatedRoyalties
	k.cdc.MustUnmarshal(bz, &royalties)
	return royalties, true
}

func (k Keeper) IncrementTotalPlays(ctx context.Context, classID, nftID string, playCount uint64) error {
	nft, found := k.GetNFT(ctx, classID, nftID)
	if !found {
		return fmt.Errorf("NFT not found: %s/%s", classID, nftID)
	}

	nft.TotalPlays += playCount
	return k.updateNFT(ctx, nft)
}

// AddToTotalRoyalties adds to the total royalties generated for an NFT
func (k Keeper) AddToTotalRoyalties(ctx context.Context, classID, nftID string, amount sdk.Coin) error {
	nft, found := k.GetNFT(ctx, classID, nftID)
	if !found {
		return fmt.Errorf("NFT not found: %s/%s", classID, nftID)
	}

	currentTotal, err := sdk.ParseCoinsNormalized(nft.TotalRoyaltiesGenerated)
	if err != nil {
		return err
	}

	newTotal := currentTotal.Add(amount)
	nft.TotalRoyaltiesGenerated = newTotal.String()

	return k.updateNFT(ctx, nft)
}

func (k Keeper) updateNFT(ctx context.Context, nft nft.NFT) error {
	k.setNFT(ctx, nft)
	return nil
}
func (k Keeper) TotalPlays(ctx context.Context, r *nft.QueryTotalPlaysRequest) (*nft.QueryTotalPlaysResponse, error) {
	nftData, found := k.GetNFT(ctx, r.ClassId, r.Id)
	if !found {
		return nil, errors.Wrapf(nft.ErrNFTNotExists, "NFT %s not found in class %s", r.Id, r.ClassId)
	}

	return &nft.QueryTotalPlaysResponse{
		TotalPlays: nftData.TotalPlays,
	}, nil
}

func (k Keeper) TotalRoyalties(ctx context.Context, request *nft.QueryTotalRoyaltiesRequest) (*nft.QueryTotalRoyaltiesResponse, error) {
	nftData, found := k.GetNFT(ctx, request.ClassId, request.Id)
	if !found {
		return nil, errors.Wrapf(nft.ErrNFTNotExists, "NFT %s not found in class %s", request.Id, request.ClassId)
	}

	// Assuming TotalRoyaltiesGenerated is stored as a string in the NFT struct
	// and represents the total amount of royalties generated
	return &nft.QueryTotalRoyaltiesResponse{
		TotalRoyalties: nftData.TotalRoyaltiesGenerated,
	}, nil
}

func (k Keeper) Burn(ctx context.Context, classID, nftID string) error {
	if !k.HasClass(ctx, classID) {
		return errors.Wrap(nft.ErrClassNotExists, classID)
	}

	if !k.HasNFT(ctx, classID, nftID) {
		return errors.Wrap(nft.ErrNFTNotExists, nftID)
	}

	// Get the NFT
	nftData, found := k.GetNFT(ctx, classID, nftID)
	if !found {
		return errors.Wrap(nft.ErrNFTNotExists, nftID)
	}

	// Get accumulated royalties
	royalties, found := k.GetAccumulatedRoyalties(ctx, classID, nftID)
	if !found {
		return errors.Wrap(sdkerrors.ErrNotFound, "royalties not found")
	}

	// Redistribute royalties
	err := k.redistributeRoyalties(ctx, nftData, royalties)
	if err != nil {
		return err
	}

	// Burn the NFT
	err = k.burnWithNoCheck(ctx, classID, nftID)
	if err != nil {
		return err
	}

	// Clear accumulated royalties
	err = k.ClearAccumulatedRoyalties(ctx, classID, nftID)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) redistributeRoyalties(ctx context.Context, nftData nft.NFT, royalties nft.AccumulatedRoyalties) error {
	creatorAddr, err := sdk.AccAddressFromBech32(nftData.Creator)
	if err != nil {
		return err
	}

	ownerAddr, err := sdk.AccAddressFromBech32(nftData.Owner)
	if err != nil {
		return err
	}

	platformAddr := sdk.MustAccAddressFromBech32("cosmos1d9ms9wf4yx3vky2kp6fc7t3qm9p8ps33g49c9s")

	// Redistribute creator royalties
	creatorAmount, err := sdk.ParseCoinsNormalized(royalties.CreatorRoyalties)
	if err != nil {
		return err
	}
	if !creatorAmount.IsZero() {
		err = k.bk.SendCoinsFromModuleToAccount(ctx, nft.ModuleName, creatorAddr, creatorAmount)
		if err != nil {
			return err
		}
	}

	// Redistribute platform royalties
	platformAmount, err := sdk.ParseCoinsNormalized(royalties.PlatformRoyalties)
	if err != nil {
		return err
	}
	if !platformAmount.IsZero() {
		err = k.bk.SendCoinsFromModuleToAccount(ctx, nft.ModuleName, platformAddr, platformAmount)
		if err != nil {
			return err
		}
	}

	// Redistribute owner royalties
	ownerAmount, err := sdk.ParseCoinsNormalized(royalties.OwnerRoyalties)
	if err != nil {
		return err
	}
	if !ownerAmount.IsZero() {
		err = k.bk.SendCoinsFromModuleToAccount(ctx, nft.ModuleName, ownerAddr, ownerAmount)
		if err != nil {
			return err
		}
	}

	return nil
}

// Add this function if it doesn't exist in your keeper.go
func (k Keeper) ClearAccumulatedRoyalties(ctx context.Context, classID string, nftID string) error {
	store := k.KVStoreService.OpenKVStore(ctx)
	key := royaltyStoreKey(classID, nftID)

	err := store.Delete(key)
	if err != nil {
		return fmt.Errorf("failed to clear accumulated royalties: %w", err)
	}

	return nil
}

// burnWithNoCheck defines a method for burning a nft from a specific account.
// Note: this method does not check whether the class already exists in nft.
// The upper-layer application needs to check it when it needs to use it
func (k Keeper) burnWithNoCheck(ctx context.Context, classID, nftID string) error {
	owner := k.GetOwner(ctx, classID, nftID)
	nftStore := k.getNFTStore(ctx, classID)
	nftStore.Delete([]byte(nftID))

	k.deleteOwner(ctx, classID, nftID, owner)
	k.decrTotalSupply(ctx, classID)

	ownerStr, err := k.ac.BytesToString(owner.Bytes())
	if err != nil {
		return err
	}

	return k.EventService.EventManager(ctx).Emit(&nft.EventBurn{
		ClassId: classID,
		Id:      nftID,
		Owner:   ownerStr,
	})
}
