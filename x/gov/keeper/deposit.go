package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// GetDeposit gets the deposit of a specific depositor on a specific proposal
func (keeper Keeper) GetDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) (deposit v1.Deposit, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.DepositKey(proposalID, depositorAddr))
	if bz == nil {
		return deposit, false
	}

	keeper.cdc.MustUnmarshal(bz, &deposit)

	return deposit, true
}

// SetDeposit sets a Deposit to the gov store
func (keeper Keeper) SetDeposit(ctx sdk.Context, deposit v1.Deposit) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshal(&deposit)
	depositor := sdk.MustAccAddressFromBech32(deposit.Depositor)

	store.Set(types.DepositKey(deposit.ProposalId, depositor), bz)
}

// GetAllDeposits returns all the deposits from the store
func (keeper Keeper) GetAllDeposits(ctx sdk.Context) (deposits v1.Deposits) {
	keeper.IterateAllDeposits(ctx, func(deposit v1.Deposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})

	return
}

// GetDeposits returns all the deposits of a proposal
func (keeper Keeper) GetDeposits(ctx sdk.Context, proposalID uint64) (deposits v1.Deposits) {
	keeper.IterateDeposits(ctx, proposalID, func(deposit v1.Deposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})

	return
}

// DeleteAndBurnDeposits deletes and burns all the deposits on a specific proposal.
func (keeper Keeper) DeleteAndBurnDeposits(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	keeper.IterateDeposits(ctx, proposalID, func(deposit v1.Deposit) bool {
		err := keeper.bankKeeper.BurnCoins(ctx, types.ModuleName, deposit.Amount)
		if err != nil {
			panic(err)
		}

		depositor := sdk.MustAccAddressFromBech32(deposit.Depositor)

		store.Delete(types.DepositKey(proposalID, depositor))
		return false
	})
}

// IterateAllDeposits iterates over all the stored deposits and performs a callback function.
func (keeper Keeper) IterateAllDeposits(ctx sdk.Context, cb func(deposit v1.Deposit) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DepositsKeyPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit v1.Deposit

		keeper.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

// IterateDeposits iterates over all the proposals deposits and performs a callback function
func (keeper Keeper) IterateDeposits(ctx sdk.Context, proposalID uint64, cb func(deposit v1.Deposit) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DepositsKey(proposalID))

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit v1.Deposit

		keeper.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

// AddDeposit adds or updates a deposit of a specific depositor on a specific proposal.
// Activates voting period when appropriate and returns true in that case, else returns false.
func (keeper Keeper) AddDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress, depositAmount sdk.Coins) (bool, error) {
	// Checks to see if proposal exists
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return false, sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
	}

	// Check if proposal is still depositable
	if (proposal.Status != v1.StatusDepositPeriod) && (proposal.Status != v1.StatusVotingPeriod) {
		return false, sdkerrors.Wrapf(types.ErrInactiveProposal, "%d", proposalID)
	}

	// Check coins to be deposited match the proposal's deposit params
	params := keeper.GetParams(ctx)

	// NOTE: backported from v50
	// v47 does not have expedited proposals so we always use params.MinDeposit
	minDepositAmount := params.MinDeposit
	minDepositRatio, err := sdk.NewDecFromStr(params.GetMinDepositRatio())
	if err != nil {
		return false, err
	}

	if err := keeper.validateDepositDenom(ctx, params, depositAmount); err != nil {
		return false, err
	}

	// If minDepositRatio is set, the deposit must be equal or greater than minDepositAmount*minDepositRatio
	// for at least one denom. If minDepositRatio is zero we skip this check.
	if !minDepositRatio.IsZero() {
		var (
			depositThresholdMet bool
			thresholds          []string
		)
		for _, minDep := range minDepositAmount {
			// calculate the threshold for this denom, and hold a list to later return a useful error message
			threshold := sdk.NewCoin(minDep.GetDenom(), minDep.Amount.ToLegacyDec().Mul(minDepositRatio).TruncateInt())
			thresholds = append(thresholds, threshold.String())

			found, deposit := depositAmount.Find(minDep.Denom)
			if !found { // if not found, continue, as we know the deposit contains at least 1 valid denom
				continue
			}

			// Once we know at least one threshold has been met, we can break. The deposit
			// might contain other denoms but we don't care.
			if deposit.IsGTE(threshold) {
				depositThresholdMet = true
				break
			}
		}

		// the threshold must be met with at least one denom, if not, return the list of minimum deposits
		if !depositThresholdMet {
			return false, sdkerrors.Wrapf(types.ErrMinDepositTooSmall, "received %s but need at least one of the following: %s", depositAmount, strings.Join(thresholds, ","))
		}
	}

	// update the governance module's account coins pool
	err = keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, depositAmount)
	if err != nil {
		return false, err
	}

	// Update proposal
	proposal.TotalDeposit = sdk.NewCoins(proposal.TotalDeposit...).Add(depositAmount...)
	keeper.SetProposal(ctx, proposal)

	// Check if deposit has provided sufficient total funds to transition the proposal into the voting period
	activatedVotingPeriod := false

	if proposal.Status == v1.StatusDepositPeriod && sdk.NewCoins(proposal.TotalDeposit...).IsAllGTE(params.MinDeposit) {
		keeper.ActivateVotingPeriod(ctx, proposal)

		activatedVotingPeriod = true
	}

	// Add or update deposit object
	deposit, found := keeper.GetDeposit(ctx, proposalID, depositorAddr)

	if found {
		deposit.Amount = sdk.NewCoins(deposit.Amount...).Add(depositAmount...)
	} else {
		deposit = v1.NewDeposit(proposalID, depositorAddr, depositAmount)
	}

	// called when deposit has been added to a proposal, however the proposal may not be active
	err = keeper.Hooks().AfterProposalDeposit(ctx, proposalID, depositorAddr)
	if err != nil {
		return false, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	keeper.SetDeposit(ctx, deposit)

	return activatedVotingPeriod, nil
}

// RefundAndDeleteDeposits refunds and deletes all the deposits on a specific proposal.
func (keeper Keeper) RefundAndDeleteDeposits(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	keeper.IterateDeposits(ctx, proposalID, func(deposit v1.Deposit) bool {
		depositor := sdk.MustAccAddressFromBech32(deposit.Depositor)

		err := keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, deposit.Amount)
		if err != nil {
			panic(err)
		}

		store.Delete(types.DepositKey(proposalID, depositor))
		return false
	})
}

// validateInitialDeposit validates if initial deposit is greater than or equal to the minimum
// required at the time of proposal submission. This threshold amount is determined by
// the deposit parameters. Returns nil on success, error otherwise.
func (keeper Keeper) validateInitialDeposit(ctx sdk.Context, params v1.Params, initialDeposit sdk.Coins, expedited bool) error {
	if !initialDeposit.IsValid() || initialDeposit.IsAnyNegative() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, initialDeposit.String())
	}

	minInitialDepositRatio, err := sdk.NewDecFromStr(params.MinInitialDepositRatio)
	if err != nil {
		return err
	}
	if minInitialDepositRatio.IsZero() {
		return nil
	}

	var minDepositCoins sdk.Coins
	if expedited {
		minDepositCoins = params.ExpeditedMinDeposit
	} else {
		minDepositCoins = params.MinDeposit
	}

	for i := range minDepositCoins {
		minDepositCoins[i].Amount = sdk.NewDecFromInt(minDepositCoins[i].Amount).Mul(minInitialDepositRatio).RoundInt()
	}
	if !initialDeposit.IsAllGTE(minDepositCoins) {
		return sdkerrors.Wrapf(types.ErrMinDepositTooSmall, "was (%s), need (%s)", initialDeposit, minDepositCoins)
	}
	return nil
}

// validateDepositDenom validates if the deposit denom is accepted by the governance module.
func (keeper Keeper) validateDepositDenom(ctx sdk.Context, params v1.Params, depositAmount sdk.Coins) error {
	denoms := []string{}
	acceptedDenoms := make(map[string]bool, len(params.MinDeposit))
	for _, coin := range params.MinDeposit {
		acceptedDenoms[coin.Denom] = true
		denoms = append(denoms, coin.Denom)
	}

	for _, coin := range depositAmount {
		if _, ok := acceptedDenoms[coin.Denom]; !ok {
			return sdkerrors.Wrapf(types.ErrInvalidDepositDenom, "deposited %s, but gov accepts only the following denom(s): %v", depositAmount, denoms)
		}
	}

	return nil
}
