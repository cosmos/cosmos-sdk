package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
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
	BudgetProposal collections.Map[sdk.AccAddress, types.Budget]
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
		BudgetProposal: collections.NewMap(sb, types.BudgetKey, "budget", sdk.AccAddressKey, codec.CollValue[types.Budget](cdc)),
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

func (k Keeper) claimFunds(ctx context.Context, recipient sdk.AccAddress) (amount sdk.Coin, err error) {
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

	return amount, nil
}

func (k Keeper) getClaimableFunds(ctx context.Context, recipient sdk.AccAddress) (amount sdk.Coin, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	budget, err := k.BudgetProposal.Get(ctx, recipient)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return sdk.Coin{}, fmt.Errorf("no claimable funds are present for recipient: %s", recipient.String())
		}
		return sdk.Coin{}, err
	}

	// check if the distribution is completed
	if budget.Tranches <= 0 {
		// remove the entry of budget ended recipient
		err := k.BudgetProposal.Remove(ctx, recipient)
		if err != nil {
			return sdk.Coin{}, err
		}
		// Log the end of the budget
		k.Logger(ctx).Debug(fmt.Sprintf("Budget ended for recipient: %s", recipient.String()))
		return sdk.Coin{}, nil
	}

	currentTime := sdkCtx.BlockTime().Unix()

	// Check if the start time is reached
	if currentTime < budget.StartTime {
		return sdk.Coin{}, fmt.Errorf("distribution has not started yet")
	}

	// Calculate the number of blocks elapsed since the start time
	blocksElapsed := sdkCtx.BlockHeight() - budget.Period

	// Check if its time to distribute funds based on period intervals
	if blocksElapsed > 0 {
		// Calculate how many periods have passed
		periodsPassed := blocksElapsed / budget.Period

		if periodsPassed > 0 {
			// Calculate the amount to distribute for all passed periods
			coinsToDistribute := math.NewInt(periodsPassed).Mul(budget.TotalBudget.Amount.QuoRaw(budget.Tranches))
			amount := sdk.NewCoin(budget.TotalBudget.Denom, coinsToDistribute)

			// update the budget's remaining tranches
			budget.Tranches -= periodsPassed

			// update the TotalBudget amount
			budget.TotalBudget.Amount.Sub(coinsToDistribute)

			k.Logger(ctx).Info(fmt.Sprintf("Processing budget for recipient: %s. Amount: %s", budget.RecipientAddress, coinsToDistribute.String()))

			// Save the updated budget in the state
			err = k.BudgetProposal.Set(ctx, recipient, budget)
			if err != nil {
				return sdk.Coin{}, fmt.Errorf("error while updating the budget for recipient %s", budget.RecipientAddress)
			}

			return amount, nil
		} else {
			return sdk.Coin{}, fmt.Errorf("budget period has not passed yet")
		}

	}
	return sdk.Coin{}, nil
}

func (k Keeper) validateBudgetProposal(ctx context.Context, bp types.MsgSubmitBudgetProposal) error {
	if bp.TotalBudget.IsZero() {
		return fmt.Errorf("total budget cannot be zero")
	}

	if err := validateAmount(sdk.NewCoins(*bp.TotalBudget)); err != nil {
		return err
	}

	if bp.StartTime <= 0 {
		return fmt.Errorf("start time must be a positive value")
	}

	if bp.Tranches <= 0 {
		return fmt.Errorf("remaining tranches must be a positive value")
	}

	if bp.Period <= 0 {
		return fmt.Errorf("period should be a positive value")
	}

	return nil
}
