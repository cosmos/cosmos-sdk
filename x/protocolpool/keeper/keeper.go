package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Keeper struct {
	environment   appmodule.Environment
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper

	cdc codec.BinaryCodec

	authority string

	// State
	Schema         collections.Schema
	BudgetProposal collections.Map[sdk.AccAddress, types.Budget]
	ContinuousFund collections.Map[sdk.AccAddress, types.ContinuousFund]
	// RecipientFundPercentage key: RecipientAddr | value: Percentage in math.Int
	RecipientFundPercentage collections.Map[sdk.AccAddress, math.Int]
	// RecipientFundDistribution key: RecipientAddr | value: Claimable amount
	RecipientFundDistribution collections.Map[sdk.AccAddress, math.Int]
	// ToDistribute is to keep track of funds distributed
	ToDistribute collections.Item[math.Int]
}

func NewKeeper(cdc codec.BinaryCodec, env appmodule.Environment, ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper, authority string,
) Keeper {
	// ensure pool module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	// ensure stream account is set
	if addr := ak.GetModuleAddress(types.StreamAccount); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.StreamAccount))
	}

	sb := collections.NewSchemaBuilder(env.KVStoreService)

	keeper := Keeper{
		environment:               env,
		authKeeper:                ak,
		bankKeeper:                bk,
		stakingKeeper:             sk,
		cdc:                       cdc,
		authority:                 authority,
		BudgetProposal:            collections.NewMap(sb, types.BudgetKey, "budget", sdk.AccAddressKey, codec.CollValue[types.Budget](cdc)),
		ContinuousFund:            collections.NewMap(sb, types.ContinuousFundKey, "continuous_fund", sdk.AccAddressKey, codec.CollValue[types.ContinuousFund](cdc)),
		RecipientFundPercentage:   collections.NewMap(sb, types.RecipientFundPercentageKey, "recipient_fund_percentage", sdk.AccAddressKey, sdk.IntValue),
		RecipientFundDistribution: collections.NewMap(sb, types.RecipientFundDistributionKey, "recipient_fund_distribution", sdk.AccAddressKey, sdk.IntValue),
		ToDistribute:              collections.NewItem(sb, types.ToDistributeKey, "to_distribute", sdk.IntValue),
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
	return k.environment.Logger.With(log.ModuleKey, "x/"+types.ModuleName)
}

// FundCommunityPool allows an account to directly fund the community fund pool.
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}

// DistributeFromCommunityPool distributes funds from the protocolpool module account to
// a receiver address.
func (k Keeper) DistributeFromCommunityPool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
}

// DistributeFromStreamFunds distributes funds from the protocolpool's stream module account to
// a receiver address.
func (k Keeper) DistributeFromStreamFunds(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.StreamAccount, receiveAddr, amount)
}

// GetCommunityPool gets the community pool balance.
func (k Keeper) GetCommunityPool(ctx context.Context) (sdk.Coins, error) {
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAccount == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccount)
	}
	return k.bankKeeper.GetAllBalances(ctx, moduleAccount.GetAddress()), nil
}

func (k Keeper) withdrawContinuousFund(ctx context.Context, recipientAddr string) (sdk.Coin, error) {
	recipient, err := k.authKeeper.AddressCodec().StringToBytes(recipientAddr)
	if err != nil {
		return sdk.Coin{}, sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	cf, err := k.ContinuousFund.Get(ctx, recipient)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return sdk.Coin{}, fmt.Errorf("no continuous fund found for recipient: %s", recipientAddr)
		}
		return sdk.Coin{}, fmt.Errorf("get continuous fund failed for recipient: %s", recipientAddr)
	}
	if cf.Expiry != nil && cf.Expiry.Before(k.environment.HeaderService.GetHeaderInfo(ctx).Time) {
		return sdk.Coin{}, fmt.Errorf("cannot withdraw continuous funds: continuous fund expired for recipient: %s", recipientAddr)
	}

	toDistributeAmount, err := k.ToDistribute.Get(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}

	if !toDistributeAmount.Equal(math.ZeroInt()) {
		err = k.iterateAndUpdateFundsDistribution(ctx, toDistributeAmount)
		if err != nil {
			return sdk.Coin{}, fmt.Errorf("error while iterating all the continuous funds: %w", err)
		}
	}

	// withdraw continuous fund
	withdrawnAmount, err := k.withdrawRecipientFunds(ctx, recipientAddr)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("error while withdrawing recipient funds for recipient: %s", recipientAddr)
	}

	return withdrawnAmount, nil
}

func (k Keeper) withdrawRecipientFunds(ctx context.Context, recipientAddr string) (sdk.Coin, error) {
	recipient, err := k.authKeeper.AddressCodec().StringToBytes(recipientAddr)
	if err != nil {
		return sdk.Coin{}, sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	// get allocated continuous fund
	fundsAllocated, err := k.RecipientFundDistribution.Get(ctx, recipient)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return sdk.Coin{}, types.ErrNoRecipientFund
		}
		return sdk.Coin{}, err
	}

	denom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Distribute funds to the recipient from pool module account
	withdrawnAmount := sdk.NewCoin(denom, fundsAllocated)
	err = k.DistributeFromStreamFunds(ctx, sdk.NewCoins(withdrawnAmount), recipient)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("error while distributing funds to the recipient %s: %v", recipientAddr, err)
	}

	// reset fund distribution
	err = k.RecipientFundDistribution.Set(ctx, recipient, math.ZeroInt())
	if err != nil {
		return sdk.Coin{}, err
	}
	return withdrawnAmount, nil
}

// SetToDistribute sets the amount to be distributed among recipients.
// This could be only set by the authority address.
func (k Keeper) SetToDistribute(ctx context.Context, amount sdk.Coins, addr string) error {
	authAddr, err := k.authKeeper.AddressCodec().StringToBytes(addr)
	if err != nil {
		return err
	}
	hasPermission, err := k.hasPermission(ctx, authAddr)
	if err != nil {
		return err
	}
	if !hasPermission {
		return sdkerrors.ErrUnauthorized
	}

	denom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	totalStreamFundsPercentage := math.ZeroInt()
	err = k.RecipientFundPercentage.Walk(ctx, nil, func(key sdk.AccAddress, value math.Int) (stop bool, err error) {
		totalStreamFundsPercentage = totalStreamFundsPercentage.Add(value)
		return false, nil
	})
	if totalStreamFundsPercentage.GT(math.NewInt(100)) {
		return fmt.Errorf("total funds percentage cannot exceed 100")
	}
	if err != nil {
		return err
	}

	// send streaming funds to the stream module account
	if err := k.sendFundsToStreamModule(ctx, denom, totalStreamFundsPercentage); err != nil {
		return err
	}

	err = k.ToDistribute.Set(ctx, amount.AmountOf(denom))
	if err != nil {
		return fmt.Errorf("error while setting ToDistribute: %v", err)
	}
	return nil
}

func (k Keeper) sendFundsToStreamModule(ctx context.Context, denom string, percentage math.Int) error {
	totalPoolAmt, err := k.GetCommunityPool(ctx)
	if err != nil {
		return err
	}

	poolAmt := totalPoolAmt.AmountOf(denom)
	poolAmtDec := sdk.NewDecCoins(sdk.NewDecCoin(denom, poolAmt))
	amt := poolAmtDec.MulDec(math.LegacyNewDecFromIntWithPrec(percentage, 2))
	streamAmt := sdk.NewCoins(sdk.NewCoin(denom, amt.AmountOf(denom).TruncateInt()))

	// Send streaming funds to the StreamModuleAccount
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.StreamAccount, streamAmt); err != nil {
		return err
	}

	return nil
}

func (k Keeper) hasPermission(ctx context.Context, addr []byte) (bool, error) {
	authority := k.GetAuthority()
	authAcc, err := k.authKeeper.AddressCodec().StringToBytes(authority)
	if err != nil {
		return false, err
	}

	return bytes.Equal(authAcc, addr), nil
}

func (k Keeper) iterateAndUpdateFundsDistribution(ctx context.Context, toDistributeAmount math.Int) error {
	totalPercentageToBeDistributed := math.ZeroInt()

	// Create a map to store keys & values from RecipientFundPercentage during the first iteration
	recipientFundMap := make(map[string]math.Int)

	// Calculate totalPercentageToBeDistributed and store values
	err := k.RecipientFundPercentage.Walk(ctx, nil, func(key sdk.AccAddress, value math.Int) (stop bool, err error) {
		addr, err := k.authKeeper.AddressCodec().BytesToString(key)
		if err != nil {
			return true, err
		}
		totalPercentageToBeDistributed = totalPercentageToBeDistributed.Add(value)
		recipientFundMap[addr] = value
		return false, nil
	})
	if err != nil {
		return err
	}
	if totalPercentageToBeDistributed.GT(math.NewInt(100)) {
		return fmt.Errorf("total funds percentage cannot exceed 100")
	}

	denom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}
	toDistributeDec := sdk.NewDecCoins(sdk.NewDecCoin(denom, toDistributeAmount))

	// Calculate the funds to be distributed based on the total percentage to be distributed
	totalAmountToBeDistributed := toDistributeDec.MulDec(math.LegacyNewDecFromIntWithPrec(totalPercentageToBeDistributed, 2))
	totalDistrAmount := totalAmountToBeDistributed.AmountOf(denom)

	for keyStr, value := range recipientFundMap {
		// Calculate the funds to be distributed based on the percentage
		decValue := math.LegacyNewDecFromIntWithPrec(value, 2)
		percentage := math.LegacyNewDecFromIntWithPrec(totalPercentageToBeDistributed, 2)
		recipientAmount := totalDistrAmount.Mul(decValue).Quo(percentage)
		recipientCoins := recipientAmount.TruncateInt()

		key, err := k.authKeeper.AddressCodec().StringToBytes(keyStr)
		if err != nil {
			return err
		}

		// Set funds to be claimed
		toClaim, err := k.RecipientFundDistribution.Get(ctx, key)
		if err != nil {
			return err
		}
		amount := toClaim.Add(recipientCoins)
		err = k.RecipientFundDistribution.Set(ctx, key, amount)
		if err != nil {
			return err
		}
	}

	// Set the coins to be distributed from toDistribute to 0
	return k.ToDistribute.Set(ctx, math.ZeroInt())
}

func (k Keeper) claimFunds(ctx context.Context, recipientAddr string) (amount sdk.Coin, err error) {
	recipient, err := k.authKeeper.AddressCodec().StringToBytes(recipientAddr)
	if err != nil {
		return sdk.Coin{}, sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	// get claimable funds from distribution info
	amount, err = k.getClaimableFunds(ctx, recipientAddr)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("error getting claimable funds: %w", err)
	}

	// distribute amount from community pool
	err = k.DistributeFromCommunityPool(ctx, sdk.NewCoins(amount), recipient)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("error distributing from community pool: %w", err)
	}

	return amount, nil
}

func (k Keeper) getClaimableFunds(ctx context.Context, recipientAddr string) (amount sdk.Coin, err error) {
	recipient, err := k.authKeeper.AddressCodec().StringToBytes(recipientAddr)
	if err != nil {
		return sdk.Coin{}, sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	budget, err := k.BudgetProposal.Get(ctx, recipient)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return sdk.Coin{}, fmt.Errorf("no budget found for recipient: %s", recipientAddr)
		}
		return sdk.Coin{}, err
	}

	// check if the distribution is completed
	if budget.TranchesLeft == 0 && budget.ClaimedAmount != nil {
		// check that claimed amount is equal to total budget
		if budget.ClaimedAmount.Equal(budget.TotalBudget) {
			// remove the entry of budget ended recipient
			if err := k.BudgetProposal.Remove(ctx, recipient); err != nil {
				return sdk.Coin{}, err
			}
			// Return the end of the budget
			return sdk.Coin{}, fmt.Errorf("budget ended for recipient: %s", recipientAddr)
		}
	}

	currentTime := k.environment.HeaderService.GetHeaderInfo(ctx).Time
	startTime := budget.StartTime

	// Check if the start time is reached
	if currentTime.Before(*startTime) {
		return sdk.Coin{}, fmt.Errorf("distribution has not started yet")
	}

	if budget.NextClaimFrom == nil || budget.NextClaimFrom.IsZero() {
		budget.NextClaimFrom = budget.StartTime
	}

	if budget.TranchesLeft == 0 && budget.ClaimedAmount == nil {
		budget.TranchesLeft = budget.Tranches
		zeroCoin := sdk.NewCoin(budget.TotalBudget.Denom, math.ZeroInt())
		budget.ClaimedAmount = &zeroCoin
	}

	return k.calculateClaimableFunds(ctx, recipient, budget, currentTime)
}

func (k Keeper) calculateClaimableFunds(ctx context.Context, recipient sdk.AccAddress, budget types.Budget, currentTime time.Time) (amount sdk.Coin, err error) {
	// Calculate the time elapsed since the last claim time
	timeElapsed := currentTime.Sub(*budget.NextClaimFrom)

	// Check the time elapsed has passed period length
	if timeElapsed < *budget.Period {
		return sdk.Coin{}, fmt.Errorf("budget period has not passed yet")
	}

	// Calculate how many periods have passed
	periodsPassed := int64(timeElapsed) / int64(*budget.Period)

	// Calculate the amount to distribute for all passed periods
	coinsToDistribute := math.NewInt(periodsPassed).Mul(budget.TotalBudget.Amount.QuoRaw(int64(budget.Tranches)))
	amount = sdk.NewCoin(budget.TotalBudget.Denom, coinsToDistribute)

	// update the budget's remaining tranches
	budget.TranchesLeft -= uint64(periodsPassed)

	// update the ClaimedAmount
	claimedAmount := budget.ClaimedAmount.Add(amount)
	budget.ClaimedAmount = &claimedAmount

	// Update the last claim time for the budget
	nextClaimFrom := budget.NextClaimFrom.Add(*budget.Period)
	budget.NextClaimFrom = &nextClaimFrom

	k.Logger(ctx).Debug(fmt.Sprintf("Processing budget for recipient: %s. Amount: %s", budget.RecipientAddress, coinsToDistribute.String()))

	// Save the updated budget in the state
	if err := k.BudgetProposal.Set(ctx, recipient, budget); err != nil {
		return sdk.Coin{}, fmt.Errorf("error while updating the budget for recipient %s", budget.RecipientAddress)
	}

	return amount, nil
}

func (k Keeper) validateAndUpdateBudgetProposal(ctx context.Context, bp types.MsgSubmitBudgetProposal) (*types.Budget, error) {
	if bp.TotalBudget.IsZero() {
		return nil, fmt.Errorf("invalid budget proposal: total budget cannot be zero")
	}

	if err := validateAmount(sdk.NewCoins(*bp.TotalBudget)); err != nil {
		return nil, fmt.Errorf("invalid budget proposal: %w", err)
	}

	currentTime := k.environment.HeaderService.GetHeaderInfo(ctx).Time
	if bp.StartTime.IsZero() || bp.StartTime == nil {
		bp.StartTime = &currentTime
	}

	if currentTime.After(*bp.StartTime) {
		return nil, fmt.Errorf("invalid budget proposal: start time cannot be less than the current block time")
	}

	if bp.Tranches == 0 {
		return nil, fmt.Errorf("invalid budget proposal: tranches must be greater than zero")
	}

	if bp.Period == nil || *bp.Period == 0 {
		return nil, fmt.Errorf("invalid budget proposal: period length should be greater than zero")
	}

	// Create and return an updated budget proposal
	updatedBudget := types.Budget{
		RecipientAddress: bp.RecipientAddress,
		TotalBudget:      bp.TotalBudget,
		StartTime:        bp.StartTime,
		Tranches:         bp.Tranches,
		Period:           bp.Period,
	}

	return &updatedBudget, nil
}

// validateContinuousFund validates the fields of the CreateContinuousFund message.
func (k Keeper) validateContinuousFund(ctx context.Context, msg types.MsgCreateContinuousFund) error {
	// Validate percentage
	if msg.Percentage.IsZero() || msg.Percentage.IsNil() {
		return fmt.Errorf("percentage cannot be zero or empty")
	}
	if msg.Percentage.IsNegative() {
		return fmt.Errorf("percentage cannot be negative")
	}
	if msg.Percentage.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("percentage cannot be greater than or equal to one")
	}

	// Validate expiry
	currentTime := k.environment.HeaderService.GetHeaderInfo(ctx).Time
	if msg.Expiry != nil && msg.Expiry.Compare(currentTime) == -1 {
		return fmt.Errorf("expiry time cannot be less than the current block time")
	}

	return nil
}
