package keeper

import (
	"time"

	"context"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"fmt"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"

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

func (k Keeper) StakeNFT(goCtx context.Context, msg *nft.MsgStakeNFT) (*nft.MsgStakeNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}

	err = k.StakeNFTInternal(ctx, msg.ClassId, msg.Id, sender, msg.StakeDuration)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"nft_staked",
			sdk.NewAttribute("class_id", msg.ClassId),
			sdk.NewAttribute("id", msg.Id),
			sdk.NewAttribute("owner", msg.Sender),
			sdk.NewAttribute("stake_duration", fmt.Sprintf("%d", msg.StakeDuration)),
		),
	)
	return &nft.MsgStakeNFTResponse{}, nil
}

// StakeNFT stakes an NFT for a specified duration
func (k Keeper) StakeNFTInternal(ctx context.Context, classID string, nftID string, owner sdk.AccAddress, stakeDuration uint64) error {
	nftData, found := k.GetNFT(ctx, classID, nftID)
	if !found {
		return errors.Wrap(nft.ErrNFTNotExists, nftID)
	}
	if nftData.Owner != owner.String() {
		return errors.Wrap(sdkerrors.ErrUnauthorized, "only the owner can stake the NFT")
	}
	if nftData.Staked {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "NFT is already staked")
	}

	// Get current time
	currentTime := uint64(time.Now().Unix())

	// Set staking status
	nftData.Staked = true
	nftData.StakeEndTime = currentTime + stakeDuration

	// Save updated NFT
	k.setNFT(ctx, nftData)

	return nil
}

// HandleStakeNFTMsg handles the MsgStakeNFT message
func (k Keeper) HandleStakeNFTMsg(goCtx context.Context, msg *nft.MsgStakeNFT) (*nft.MsgStakeNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}

	err = k.StakeNFTInternal(ctx, msg.ClassId, msg.Id, sender, msg.StakeDuration)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"nft_staked",
			sdk.NewAttribute("class_id", msg.ClassId),
			sdk.NewAttribute("id", msg.Id),
			sdk.NewAttribute("owner", msg.Sender),
			sdk.NewAttribute("stake_duration", fmt.Sprintf("%d", msg.StakeDuration)),
		),
	)
	return &nft.MsgStakeNFTResponse{}, nil
}

func (k Keeper) UnstakeNFT(goCtx context.Context, msg *nft.MsgUnstakeNFT) (*nft.MsgUnstakeNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}

	err = k.UnstakeNFTInternal(ctx, msg.ClassId, msg.Id, sender)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"nft_unstaked",
			sdk.NewAttribute("class_id", msg.ClassId),
			sdk.NewAttribute("id", msg.Id),
			sdk.NewAttribute("owner", msg.Sender),
		),
	)

	return &nft.MsgUnstakeNFTResponse{}, nil
}

// UnstakeNFT unstakes a previously staked NFT
func (k Keeper) UnstakeNFTInternal(ctx context.Context, classID string, nftID string, owner sdk.AccAddress) error {
	nftData, found := k.GetNFT(ctx, classID, nftID)
	if !found {
		return errors.Wrap(nft.ErrNFTNotExists, nftID)
	}
	if nftData.Owner != owner.String() {
		return errors.Wrap(sdkerrors.ErrUnauthorized, "only the owner can unstake the NFT")
	}
	if !nftData.Staked {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "NFT is not staked")
	}

	// Unstake the NFT
	nftData.Staked = false
	nftData.StakeEndTime = 0

	// Save updated NFT
	k.setNFT(ctx, nftData)

	return nil
}

// HandleUnstakeNFTMsg handles the MsgUnstakeNFT message
func (k Keeper) HandleUnstakeNFTMsg(goCtx context.Context, msg *nft.MsgUnstakeNFT) (*nft.MsgUnstakeNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}

	err = k.UnstakeNFTInternal(ctx, msg.ClassId, msg.Id, sender)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"nft_unstaked",
			sdk.NewAttribute("class_id", msg.ClassId),
			sdk.NewAttribute("id", msg.Id),
			sdk.NewAttribute("owner", msg.Sender),
		),
	)

	return &nft.MsgUnstakeNFTResponse{}, nil
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

func listedNFTKey(classID, nftID string) []byte {
	return []byte(fmt.Sprintf("listed_nft/%s/%s", classID, nftID))
}

// GetListedNFT returns a single listed NFT
func (k Keeper) GetListedNFT(ctx context.Context, classID, nftID string) (nft.ListedNFT, bool) {
	store := k.KVStoreService.OpenKVStore(ctx)
	key := listedNFTKey(classID, nftID)
	bz, err := store.Get(key)
	if err != nil {
		return nft.ListedNFT{}, false
	}
	if bz == nil {
		return nft.ListedNFT{}, false
	}

	var listedNFT nft.ListedNFT
	k.cdc.MustUnmarshal(bz, &listedNFT)
	return listedNFT, true
}

// GetListedNFTs returns all NFTs currently listed on the marketplace
func (k Keeper) GetListedNFTs(ctx context.Context, pagination *query.PageRequest) ([]*nft.ListedNFT, *query.PageResponse, error) {
	var listedNFTs []*nft.ListedNFT
	store := k.KVStoreService.OpenKVStore(ctx)

	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), []byte("listed_nft/")), pagination, func(key []byte, value []byte) error {
		var listedNFT nft.ListedNFT
		k.cdc.MustUnmarshal(value, &listedNFT)
		listedNFTs = append(listedNFTs, &listedNFT)
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return listedNFTs, pageRes, nil
}

// ListNFTOnMarketplace handles the MsgListNFT message
func (k Keeper) ListNFTOnMarketplace(ctx context.Context, classID, nftID string, owner sdk.AccAddress, price sdk.Coin) error {
	if !k.HasNFT(ctx, classID, nftID) {
		return errors.Wrap(nft.ErrNFTNotExists, nftID)
	}

	nftData, _ := k.GetNFT(ctx, classID, nftID)
	nftData.Listed = true
	k.setNFT(ctx, nftData)
	if nftData.Owner != owner.String() {
		return errors.Wrap(sdkerrors.ErrUnauthorized, "not the owner of the NFT")

	}

	listedNFT := nft.ListedNFT{
		ClassId: classID,
		Id:      nftID,
		Owner:   owner.String(),
		Price:   price.String(),
	}

	store := k.KVStoreService.OpenKVStore(ctx)
	key := listedNFTKey(classID, nftID)
	bz := k.cdc.MustMarshal(&listedNFT)
	err := store.Set(key, bz)
	if err != nil {
		return err
	}

	return k.EventService.EventManager(ctx).Emit(&nft.EventMarketplace{
		ClassId: classID,
		Id:      nftID,
		Seller:  owner.String(),
		Price:   price.String(),
		Action:  "list",
	})
}

// BuyNFTFromMarketplace allows a user to buy an NFT from the marketplace
func (k Keeper) BuyNFTFromMarketplace(ctx context.Context, classID, nftID string, buyer sdk.AccAddress) error {
	store := k.KVStoreService.OpenKVStore(ctx)
	key := listedNFTKey(classID, nftID)
	bz, err := store.Get(key)
	if err != nil {
		return err
	}
	if bz == nil {
		return errors.Wrap(sdkerrors.ErrNotFound, "NFT not listed")
	}

	var listedNFT nft.ListedNFT
	k.cdc.MustUnmarshal(bz, &listedNFT)

	price, err := sdk.ParseCoinsNormalized(listedNFT.Price)
	if err != nil {
		return err
	}

	seller, err := sdk.AccAddressFromBech32(listedNFT.Owner)
	if err != nil {
		return err
	}

	// Transfer funds from buyer to seller
	err = k.bk.SendCoins(ctx, buyer, seller, price)
	if err != nil {
		return err
	}

	// Transfer NFT ownership
	err = k.Transfer(ctx, classID, nftID, seller, buyer)
	if err != nil {
		return err
	}

	// Remove NFT from listing
	err = store.Delete(key)
	if err != nil {
		return err
	}

	nftData, _ := k.GetNFT(ctx, classID, nftID)
	nftData.Listed = false
	k.setNFT(ctx, nftData)

	return k.EventService.EventManager(ctx).Emit(&nft.EventMarketplace{
		ClassId: classID,
		Id:      nftID,
		Seller:  listedNFT.Owner,
		Buyer:   buyer.String(),
		Price:   listedNFT.Price,
		Action:  "buy",
	})
}

// ListedNFTs implements the ListedNFTs gRPC method
func (k Keeper) ListedNFTs(ctx context.Context, req *nft.QueryListedNFTsRequest) (*nft.QueryListedNFTsResponse, error) {
	listedNFTs, pageRes, err := k.GetListedNFTs(ctx, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &nft.QueryListedNFTsResponse{
		ListedNfts: listedNFTs,
		Pagination: pageRes,
	}, nil
}

// ListedNFT implements the ListedNFT gRPC method
func (k Keeper) ListedNFT(ctx context.Context, req *nft.QueryListedNFTRequest) (*nft.QueryListedNFTResponse, error) {
	listedNFT, found := k.GetListedNFT(ctx, req.ClassId, req.Id)
	if !found {
		return nil, errors.Wrap(sdkerrors.ErrNotFound, "listed NFT not found")
	}

	return &nft.QueryListedNFTResponse{
		ListedNft: &listedNFT,
	}, nil
}

// ListNFT handles the MsgListNFT message
func (k Keeper) ListNFT(goCtx context.Context, msg *nft.MsgListNFT) (*nft.MsgListNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	// Additional check to ensure the sender is the owner
	nftData, found := k.GetNFT(ctx, msg.ClassId, msg.Id)
	if !found {
		return nil, errors.Wrap(nft.ErrNFTNotExists, msg.Id)
	}
	if nftData.Owner != msg.Sender {
		return nil, errors.Wrap(sdkerrors.ErrUnauthorized, "not the owner of the NFT")
	}

	price, err := sdk.ParseCoinNormalized(msg.Price)
	if err != nil {
		return nil, err
	}

	err = k.ListNFTOnMarketplace(ctx, msg.ClassId, msg.Id, sender, price)
	if err != nil {
		return nil, err
	}

	return &nft.MsgListNFTResponse{}, nil
}

// BuyNFT handles the MsgBuyNFT message
func (k Keeper) BuyNFT(goCtx context.Context, msg *nft.MsgBuyNFT) (*nft.MsgBuyNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	buyer, err := sdk.AccAddressFromBech32(msg.Buyer)
	if err != nil {
		return nil, err
	}

	err = k.BuyNFTFromMarketplace(ctx, msg.ClassId, msg.Id, buyer)
	if err != nil {
		return nil, err
	}

	return &nft.MsgBuyNFTResponse{}, nil
}

// DelistNFTFromMarketplace removes an NFT from the marketplace listing
func (k Keeper) DelistNFTFromMarketplace(ctx context.Context, classID, nftID string, owner sdk.AccAddress) error {
	store := k.KVStoreService.OpenKVStore(ctx)
	key := listedNFTKey(classID, nftID)

	// Check if the NFT is listed
	bz, err := store.Get(key)
	if err != nil {
		return err
	}
	if bz == nil {
		return errors.Wrap(sdkerrors.ErrNotFound, "NFT not listed")
	}

	var listedNFT nft.ListedNFT
	k.cdc.MustUnmarshal(bz, &listedNFT)

	// Check if the sender is the owner of the NFT
	if listedNFT.Owner != owner.String() {
		return errors.Wrap(sdkerrors.ErrUnauthorized, "not the owner of the NFT")
	}

	// Remove NFT from listing
	err = store.Delete(key)
	if err != nil {
		return err
	}

	// Update the NFT's Listed status
	nftData, found := k.GetNFT(ctx, classID, nftID)
	if found {
		nftData.Listed = false
		k.setNFT(ctx, nftData)
	}

	return k.EventService.EventManager(ctx).Emit(&nft.EventMarketplace{
		ClassId: classID,
		Id:      nftID,
		Seller:  owner.String(),
		Action:  "delist",
	})
}

// DelistNFT handles the MsgDelistNFT message
func (k Keeper) DelistNFT(goCtx context.Context, msg *nft.MsgDelistNFT) (*nft.MsgDelistNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	// Additional check to ensure the sender is the owner
	nftData, found := k.GetNFT(ctx, msg.ClassId, msg.Id)
	if !found {
		return nil, errors.Wrap(nft.ErrNFTNotExists, msg.Id)
	}
	if nftData.Owner != msg.Sender {
		return nil, errors.Wrap(sdkerrors.ErrUnauthorized, "not the owner of the NFT")
	}

	err = k.DelistNFTFromMarketplace(ctx, msg.ClassId, msg.Id, sender)
	if err != nil {
		return nil, err
	}

	return &nft.MsgDelistNFTResponse{}, nil
}
