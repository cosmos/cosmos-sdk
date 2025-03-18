package keeper

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

// assert that this keeper can be used by x/distribution
var _ types.ExternalCommunityPoolKeeper = &Keeper{}

type Keeper struct {
	storeService store.KVStoreService

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper

	cdc codec.BinaryCodec

	authority string

	// State
	Schema collections.Schema
	// Budgets key: RecipientAddr | value: Budget
	Budgets        collections.Map[sdk.AccAddress, types.Budget]
	ContinuousFund collections.Map[sdk.AccAddress, types.ContinuousFund]
	// RecipientFundDistribution key: RecipientAddr | value: Claimable amount
	RecipientFundDistribution collections.Map[sdk.AccAddress, types.DistributionAmount]
	Distributions             collections.Map[time.Time, types.DistributionAmount] // key: time.Time, denom | value: amounts
	LastBalance               collections.Item[types.DistributionAmount]
	Params                    collections.Item[types.Params]
}

const (
	errModuleAccountNotSet = "%s module account has not been set"
)

func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, ak types.AccountKeeper, bk types.BankKeeper, authority string,
) Keeper {
	// ensure pool module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf(errModuleAccountNotSet, types.ModuleName))
	}
	// ensure stream account is set
	if addr := ak.GetModuleAddress(types.StreamAccount); addr == nil {
		panic(fmt.Sprintf(errModuleAccountNotSet, types.StreamAccount))
	}
	// ensure protocol pool distribution account is set
	if addr := ak.GetModuleAddress(types.ProtocolPoolDistrAccount); addr == nil {
		panic(fmt.Sprintf(errModuleAccountNotSet, types.ProtocolPoolDistrAccount))
	}

	sb := collections.NewSchemaBuilder(storeService)

	keeper := Keeper{
		storeService:              storeService,
		authKeeper:                ak,
		bankKeeper:                bk,
		cdc:                       cdc,
		authority:                 authority,
		Budgets:                   collections.NewMap(sb, types.BudgetKey, "budget", sdk.AccAddressKey, codec.CollValue[types.Budget](cdc)),
		ContinuousFund:            collections.NewMap(sb, types.ContinuousFundKey, "continuous_fund", sdk.AccAddressKey, codec.CollValue[types.ContinuousFund](cdc)),
		RecipientFundDistribution: collections.NewMap(sb, types.RecipientFundDistributionKey, "recipient_fund_distribution", sdk.AccAddressKey, codec.CollValue[types.DistributionAmount](cdc)),
		Distributions:             collections.NewMap(sb, types.DistributionsKey, "distributions", sdk.TimeKey, codec.CollValue[types.DistributionAmount](cdc)),
		LastBalance:               collections.NewItem(sb, types.LastBalanceKey, "last_balance", codec.CollValue[types.DistributionAmount](cdc)),
		Params:                    collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
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

// GetCommunityPoolModule gets the module name that funds should be sent to for the community pool.
// This is the address that x/distribution will send funds to for external management.
func (k Keeper) GetCommunityPoolModule() string {
	return types.ProtocolPoolDistrAccount
}

// FundCommunityPool allows an account to directly fund the community fund pool.
func (k Keeper) FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}

// DistributeFromCommunityPool distributes funds from the protocolpool module account to
// a receiver address.
func (k Keeper) DistributeFromCommunityPool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
}

// DistributeFromStreamFunds distributes funds from the protocolpool's stream module account to
// a receiver address.
func (k Keeper) DistributeFromStreamFunds(ctx sdk.Context, amount sdk.Coins, receiveAddr []byte) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.StreamAccount, receiveAddr, amount)
}

// GetCommunityPool gets the community pool balance.
func (k Keeper) GetCommunityPool(ctx sdk.Context) (sdk.Coins, error) {
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAccount == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", types.ModuleName)
	}
	return k.bankKeeper.GetAllBalances(ctx, moduleAccount.GetAddress()), nil
}

func (k Keeper) withdrawRecipientFunds(ctx sdk.Context, recipient []byte) (sdk.Coins, error) {
	// get allocated continuous fund
	fundsAllocated, err := k.RecipientFundDistribution.Get(ctx, recipient)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, types.ErrNoRecipientFound
		}
		return nil, err
	}
	// Distribute funds to the recipient from pool module account
	err = k.DistributeFromStreamFunds(ctx, fundsAllocated.Amount, recipient)
	if err != nil {
		return nil, fmt.Errorf("error while distributing funds: %w", err)
	}

	// reset fund distribution
	err = k.RecipientFundDistribution.Set(ctx, recipient, types.DistributionAmount{Amount: sdk.NewCoins()})
	if err != nil {
		return nil, err
	}
	return fundsAllocated.Amount, nil
}

// SetToDistribute sets the amount to be distributed among recipients.
func (k Keeper) SetToDistribute(ctx sdk.Context) error {
	// Get current balance of the intermediary module account
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ProtocolPoolDistrAccount)
	if moduleAccount == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", types.ProtocolPoolDistrAccount)
	}
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// only take into account the balances of denoms whitelisted in EnabledDistributionDenoms
	currentBalance := sdk.NewCoins()
	for _, denom := range params.EnabledDistributionDenoms {
		bal := k.bankKeeper.GetBalance(ctx, moduleAccount.GetAddress(), denom)
		currentBalance = currentBalance.Add(bal)
	}

	// if the balance is zero, return early
	if currentBalance.IsZero() {
		return nil
	}

	lastBalance, err := k.LastBalance.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			lastBalance = types.DistributionAmount{Amount: sdk.NewCoins()}
		} else {
			return err
		}
	}

	// Calculate the amount to be distributed
	amountToDistribute, anyNegative := currentBalance.SafeSub(lastBalance.Amount...)
	if anyNegative {
		return errors.New("error while calculating the amount to distribute, result can't be negative")
	}

	// Check if there are any recipients to distribute to, if not, send straight to the community pool and avoid
	// setting the distributions
	hasContinuousFunds := false
	err = k.ContinuousFund.Walk(ctx, nil, func(_ sdk.AccAddress, _ types.ContinuousFund) (bool, error) {
		hasContinuousFunds = true
		return true, nil
	})
	if err != nil {
		return err
	}

	// if there are no continuous funds, send all the funds to the community pool and reset the last balance
	if !hasContinuousFunds {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ProtocolPoolDistrAccount, types.ModuleName, amountToDistribute); err != nil {
			return err
		}

		if !lastBalance.Amount.IsZero() { // only reset if the last balance is not zero (so we leave it at zero)
			return k.LastBalance.Set(ctx, types.DistributionAmount{Amount: sdk.NewCoins()})
		}

		return nil
	}

	if err = k.Distributions.Set(ctx, ctx.BlockTime(), types.DistributionAmount{Amount: amountToDistribute}); err != nil {
		return fmt.Errorf("error while setting Distributions: %w", err)
	}

	// Update the last balance
	return k.LastBalance.Set(ctx, types.DistributionAmount{Amount: currentBalance})
}

func (k Keeper) IterateAndUpdateFundsDistribution(ctx sdk.Context) error {
	// first we get all the continuous funds, and keep a list of the ones that expired so we can delete later
	funds := []types.ContinuousFund{}
	toDelete := [][]byte{}
	err := k.ContinuousFund.Walk(ctx, nil, func(key sdk.AccAddress, cf types.ContinuousFund) (stop bool, err error) {
		funds = append(funds, cf)

		// check if the continuous fund has expired, and add it to the list of funds to delete
		if cf.Expiry != nil && cf.Expiry.Before(ctx.BlockTime()) {
			toDelete = append(toDelete, key)
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	// next we iterate over the distributions, calculate each recipient's share and the remaining pool funds
	distributeToRecipient := map[string]sdk.Coins{}
	effectiveDistributionAmounts := sdk.NewCoins() // amount assigned to distributions
	totalDistributionAmounts := sdk.NewCoins()     // total amount distributed to the pool, to then calculate the remaining pool funds

	if err = k.Distributions.Walk(ctx, nil, func(key time.Time, amount types.DistributionAmount) (stop bool, err error) {
		totalPercentageApplied := math.LegacyZeroDec()
		totalDistributionAmounts = totalDistributionAmounts.Add(amount.Amount...)

		for _, f := range funds {
			if f.Expiry != nil && f.Expiry.Before(key) {
				continue
			}

			totalPercentageApplied = totalPercentageApplied.Add(f.Percentage)

			_, ok := distributeToRecipient[f.Recipient]
			if !ok {
				distributeToRecipient[f.Recipient] = sdk.NewCoins()
			}

			for _, denom := range amount.Amount.Denoms() {
				am := sdk.NewCoin(denom, f.Percentage.MulInt(amount.Amount.AmountOf(denom)).TruncateInt())
				distributeToRecipient[f.Recipient] = distributeToRecipient[f.Recipient].Add(am)
				effectiveDistributionAmounts = effectiveDistributionAmounts.Add(am)
			}
		}

		// sanity check for max percentage
		if totalPercentageApplied.GT(math.LegacyOneDec()) {
			return true, errors.New("total funds percentage cannot exceed 100")
		}

		return false, nil
	}); err != nil {
		return err
	}

	// clear the distributions and reset the last balance
	if err = k.Distributions.Clear(ctx, nil); err != nil {
		return err
	}

	if err = k.LastBalance.Set(ctx, types.DistributionAmount{Amount: sdk.NewCoins()}); err != nil {
		return err
	}

	// send the funds to the stream account to be distributed later, and the remaining to the community pool
	if !effectiveDistributionAmounts.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ProtocolPoolDistrAccount, types.StreamAccount, effectiveDistributionAmounts); err != nil {
			return err
		}
	}

	poolFunds := totalDistributionAmounts.Sub(effectiveDistributionAmounts...)
	if !poolFunds.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ProtocolPoolDistrAccount, types.ModuleName, poolFunds); err != nil {
			return err
		}
	}

	// update the recipient fund distribution, first get the keys and sort them
	recipients := make([]string, 0, len(distributeToRecipient))
	for k2 := range distributeToRecipient {
		recipients = append(recipients, k2)
	}
	sort.Strings(recipients)

	for _, recipient := range recipients {
		// Set funds to be claimed
		bzAddr, err := k.authKeeper.AddressCodec().StringToBytes(recipient)
		if err != nil {
			return err
		}

		toClaim, err := k.RecipientFundDistribution.Get(ctx, bzAddr)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				toClaim = types.DistributionAmount{Amount: sdk.NewCoins()}
			} else {
				return err
			}
		}

		toClaim.Amount = toClaim.Amount.Add(distributeToRecipient[recipient]...)
		if err = k.RecipientFundDistribution.Set(ctx, bzAddr, toClaim); err != nil {
			return err
		}
	}

	// delete expired continuous funds
	for _, recipient := range toDelete {
		if err = k.ContinuousFund.Remove(ctx, recipient); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) claimFunds(ctx sdk.Context, recipientAddr string) (amount sdk.Coin, err error) {
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

func (k Keeper) getClaimableFunds(ctx sdk.Context, recipientAddr string) (amount sdk.Coin, err error) {
	recipient, err := k.authKeeper.AddressCodec().StringToBytes(recipientAddr)
	if err != nil {
		return sdk.Coin{}, sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	budget, err := k.Budgets.Get(ctx, recipient)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return sdk.Coin{}, fmt.Errorf("no budget found for recipient: %s", recipientAddr)
		}
		return sdk.Coin{}, err
	}

	totalBudgetAmountLeftToDistribute := budget.BudgetPerTranche.Amount.Mul(math.NewIntFromUint64(budget.TranchesLeft))
	totalBudgetAmountLeft := sdk.NewCoin(budget.BudgetPerTranche.Denom, totalBudgetAmountLeftToDistribute)
	zeroAmount := sdk.NewCoin(totalBudgetAmountLeft.Denom, math.ZeroInt())

	// check if the distribution is completed
	if budget.TranchesLeft == 0 && budget.ClaimedAmount != nil {
		// check that total budget amount left to distribute equals zero
		if totalBudgetAmountLeft.Equal(zeroAmount) {
			// remove the entry of budget ended recipient
			if err := k.Budgets.Remove(ctx, recipient); err != nil {
				return sdk.Coin{}, err
			}
			// Return the end of the budget
			return sdk.Coin{}, fmt.Errorf("budget ended for recipient: %s", recipientAddr)
		}
	}

	currentTime := ctx.BlockTime()

	// Check if the distribution time has not reached
	if !budget.LastClaimedAt.IsZero() {
		if currentTime.Before(budget.LastClaimedAt) {
			return sdk.Coin{}, fmt.Errorf("distribution has not started yet: start time: %s", budget.LastClaimedAt.String())
		}
	}

	if budget.TranchesLeft != 0 && budget.ClaimedAmount == nil {
		zeroCoin := sdk.NewCoin(budget.BudgetPerTranche.Denom, math.ZeroInt())
		budget.ClaimedAmount = &zeroCoin
	}

	return k.calculateClaimableFunds(ctx, recipient, budget)
}

func (k Keeper) calculateClaimableFunds(ctx sdk.Context, recipient sdk.AccAddress, budget types.Budget) (amount sdk.Coin, err error) {
	// Calculate the time elapsed since the last claim time
	timeElapsed := ctx.BlockTime().Sub(budget.LastClaimedAt)

	// Check the time elapsed has passed period length
	if timeElapsed < budget.Period {
		return sdk.Coin{}, fmt.Errorf("budget period of %f hours has not passed yet", budget.Period.Hours())
	}

	// Calculate how many periods have passed
	periodsPassed := int64(timeElapsed) / int64(budget.Period)

	if periodsPassed > int64(budget.TranchesLeft) {
		periodsPassed = int64(budget.TranchesLeft)
	}

	// Calculate the amount to distribute for all passed periods
	coinsToDistribute := math.NewInt(periodsPassed).Mul(budget.BudgetPerTranche.Amount)
	amount = sdk.NewCoin(budget.BudgetPerTranche.Denom, coinsToDistribute)

	// update the budget's remaining tranches
	if budget.TranchesLeft > uint64(periodsPassed) {
		budget.TranchesLeft -= uint64(periodsPassed)
	} else {
		budget.TranchesLeft = 0
	}

	// update the ClaimedAmount
	claimedAmount := budget.ClaimedAmount.Add(amount)
	budget.ClaimedAmount = &claimedAmount

	// Update the last claim time for the budget
	nextClaimFrom := budget.LastClaimedAt.Add(budget.Period * time.Duration(periodsPassed))
	budget.LastClaimedAt = nextClaimFrom

	ctx.Logger().Debug(fmt.Sprintf("Processing budget for recipient: %s. Amount: %s", budget.RecipientAddress, coinsToDistribute.String()))

	// Save the updated budget in the state
	if err := k.Budgets.Set(ctx, recipient, budget); err != nil {
		return sdk.Coin{}, fmt.Errorf("error while updating the budget for recipient %s", budget.RecipientAddress)
	}

	return amount, nil
}

// validateAndUpdateBudgetProposal validates the Budget included in a MsgSubmitBudgetProposal as follows:
// - BudgetPerTranche must be nonzero
// - the budget amount must be a valid sdk.Coin
// - the startTime must be valid (after current blocktime)
// - - if the startTime was nil, set it to the current blocktime
// - number of tranches must be nonzero
// - period duration must be nonzero
func (k Keeper) validateAndUpdateBudgetProposal(ctx sdk.Context, bp types.MsgSubmitBudgetProposal) (types.Budget, error) {
	if bp.BudgetPerTranche.IsZero() {
		return types.Budget{}, errors.New("invalid budget proposal: budget per tranche cannot be zero")
	}

	if err := validateAmount(sdk.NewCoins(bp.BudgetPerTranche)); err != nil {
		return types.Budget{}, fmt.Errorf("invalid budget proposal: %w", err)
	}

	currentTime := ctx.BlockTime()
	if bp.StartTime.IsZero() || bp.StartTime == nil {
		bp.StartTime = &currentTime
	}

	if currentTime.After(*bp.StartTime) {
		return types.Budget{}, errors.New("invalid budget proposal: start time cannot be less than the current block time")
	}

	if bp.Tranches == 0 {
		return types.Budget{}, errors.New("invalid budget proposal: tranches must be greater than zero")
	}

	if bp.Period == 0 {
		return types.Budget{}, errors.New("invalid budget proposal: period length should be greater than zero")
	}

	// Create and return an updated budget proposal
	updatedBudget := types.Budget{
		RecipientAddress: bp.RecipientAddress,
		BudgetPerTranche: bp.BudgetPerTranche,
		LastClaimedAt:    *bp.StartTime,
		TranchesLeft:     bp.Tranches,
		Period:           bp.Period,
	}

	return updatedBudget, nil
}

// validateContinuousFund validates the fields of the CreateContinuousFund message.
func (k Keeper) validateContinuousFund(ctx sdk.Context, msg types.MsgCreateContinuousFund) error {
	// Validate percentage
	if msg.Percentage.IsZero() || msg.Percentage.IsNil() {
		return errors.New("percentage cannot be zero or empty")
	}
	if msg.Percentage.IsNegative() {
		return errors.New("percentage cannot be negative")
	}
	if msg.Percentage.GTE(math.LegacyOneDec()) {
		return errors.New("percentage cannot be greater than or equal to one")
	}

	// Validate expiry
	currentTime := ctx.BlockTime()
	if msg.Expiry != nil && msg.Expiry.Compare(currentTime) == -1 {
		return errors.New("expiry time cannot be less than the current block time")
	}

	return nil
}
