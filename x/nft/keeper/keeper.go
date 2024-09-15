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

// WithdrawRoyaltiesInternal allows the withdrawal of accumulated royalties for a specific role
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
	platformWalletAddress, err := sdk.AccAddressFromBech32("cosmos1d9ms9wf4yx3vky2kp6fc7t3qm9p8ps33g49c9s")
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("invalid platform wallet address format")
	}

	switch role {
	case "creator":
		amount, err = sdk.ParseCoinNormalized(accumulatedRoyalties.CreatorRoyalties)
		accumulatedRoyalties.CreatorRoyalties = sdk.NewCoin(amount.Denom, math.ZeroInt()).String()
	case "platform":
		// Check if the recipient matches the platform wallet address
		if !recipient.Equals(platformWalletAddress) {
			return sdk.Coin{}, fmt.Errorf("unauthorized withdrawal attempt for platform royalties")
		}
		amount, err = sdk.ParseCoinNormalized(accumulatedRoyalties.PlatformRoyalties)
		accumulatedRoyalties.PlatformRoyalties = sdk.NewCoin(amount.Denom, math.ZeroInt()).String()
	case "owner":
		amount, err = sdk.ParseCoinNormalized(accumulatedRoyalties.OwnerRoyalties)
		accumulatedRoyalties.OwnerRoyalties = sdk.NewCoin(amount.Denom, math.ZeroInt()).String()
	default:
		return sdk.Coin{}, fmt.Errorf("invalid role: %s", role)
	}

	if err != nil {
		return sdk.Coin{}, err
	}

	if amount.IsZero() {
		return sdk.Coin{}, fmt.Errorf("no royalties to withdraw for role %s", role)
	}

	err = k.bk.SendCoinsFromModuleToAccount(ctx, nft.ModuleName, recipient, sdk.NewCoins(amount))
	if err != nil {
		return sdk.Coin{}, err
	}

	store.Set(key, k.cdc.MustMarshal(&accumulatedRoyalties))

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
	//TODO implement me
	panic("implement me")
}

func (k Keeper) TotalRoyalties(ctx context.Context, request *nft.QueryTotalRoyaltiesRequest) (*nft.QueryTotalRoyaltiesResponse, error) {
	//TODO implement me
	panic("implement me")
}
