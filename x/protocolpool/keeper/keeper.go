package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Keeper struct {
	storeService storetypes.KVStoreService
	authKeeper   types.AccountKeeper
	bankKeeper   types.BankKeeper

	cdc codec.BinaryCodec

	authority string

	// State
	Schema         collections.Schema
	BudgetProposal collections.Map[sdk.AccAddress, types.BudgetProposal]
	DistrInfo      collections.Map[sdk.AccAddress, sdk.Coin]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService,
	ak types.AccountKeeper, bk types.BankKeeper, authority string,
) Keeper {
	// ensure pool module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	sb := collections.NewSchemaBuilder(storeService)

	keeper := Keeper{
		storeService:   storeService,
		authKeeper:     ak,
		bankKeeper:     bk,
		cdc:            cdc,
		authority:      authority,
		BudgetProposal: collections.NewMap(sb, types.BudgetKey, "budget", sdk.AccAddressKey, codec.CollValue[types.BudgetProposal](cdc)),
		DistrInfo:      collections.NewMap(sb, types.DistrInfoKey, "distribution_info", sdk.AccAddressKey, codec.CollValue[sdk.Coin](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	keeper.Schema = schema

	return keeper
}

// GetAuthority returns the x/protocolpool module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With(log.ModuleKey, "x/"+types.ModuleName)
}

// FundCommunityPool allows an account to directly fund the community fund pool.
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}

// DistributeFromFeePool distributes funds from the protocolpool module account to
// a receiver address.
func (k Keeper) DistributeFromFeePool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
}

// GetCommunityPool get the community pool balance.
func (k Keeper) GetCommunityPool(ctx context.Context) (sdk.Coins, error) {
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAccount == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccount)
	}
	return k.bankKeeper.GetAllBalances(ctx, moduleAccount.GetAddress()), nil
}

func (k Keeper) AppendDistributionInfo(ctx sdk.Context, distributionInfo types.DistributionInfo) error {
	amount, err := k.DistrInfo.Get(ctx, distributionInfo.Address)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			err := k.DistrInfo.Set(ctx, distributionInfo.Address, distributionInfo.Amount)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	updatedAmount := distributionInfo.Amount.Add(amount)
	return k.DistrInfo.Set(ctx, distributionInfo.Address, updatedAmount)
}

func (k Keeper) ClaimFunds(ctx context.Context, recipient sdk.AccAddress) (amount sdk.Coin, err error) {
	// get claimable funds from distribution info
	amount, err = k.getClaimableFunds(ctx, recipient)
	if err != nil {
		return sdk.Coin{}, err
	}

	// distribute amount from feepool
	err = k.DistributeFromFeePool(ctx, sdk.NewCoins(amount), recipient)
	if err != nil {
		return sdk.Coin{}, err
	}

	// remove the recipient from the DistributionInfo
	err = k.DistrInfo.Remove(ctx, recipient)
	if err != nil {
		return sdk.Coin{}, err
	}

	return amount, nil
}

func (k Keeper) getClaimableFunds(ctx context.Context, recipient sdk.AccAddress) (amount sdk.Coin, err error) {
	amount, err = k.DistrInfo.Get(ctx, recipient)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return sdk.Coin{}, fmt.Errorf("no claimable funds are present for recipient: %s", recipient.String())
		}
		return sdk.Coin{}, err
	}
	return amount, nil
}

func (k Keeper) validateBudgetProposal(bp types.BudgetProposal) error {
	if err := validateAmount(sdk.NewCoins(*bp.TotalBudget)); err != nil {
		return err
	}

	if bp.StartTime <= 0 {
		return fmt.Errorf("start time should be positive")
	}

	if bp.RemainingTranches <= 0 {
		return fmt.Errorf("cannot set tranches <= 0")
	}

	if bp.Period <= 0 {
		return fmt.Errorf("period should be a positive integer")
	}

	return nil
}
