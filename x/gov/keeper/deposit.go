package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// GetDeposit gets the deposit of a specific depositor on a specific proposal
func (keeper Keeper) GetDeposit(ctx context.Context, proposalID uint64, depositorAddr sdk.AccAddress) (deposit v1.Deposit, err error) {
	store := keeper.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.DepositKey(proposalID, depositorAddr))
	if err != nil {
		return deposit, err
	}

	if bz == nil {
		return deposit, types.ErrDepositNotFound
	}

	err = keeper.cdc.Unmarshal(bz, &deposit)
	if err != nil {
		return deposit, err
	}

	return deposit, nil
}

// SetDeposit sets a Deposit to the gov store
func (keeper Keeper) SetDeposit(ctx context.Context, deposit v1.Deposit) error {
	store := keeper.storeService.OpenKVStore(ctx)
	bz, err := keeper.cdc.Marshal(&deposit)
	if err != nil {
		return err
	}

	depositor, err := keeper.authKeeper.StringToBytes(deposit.Depositor)
	if err != nil {
		return err
	}

	return store.Set(types.DepositKey(deposit.ProposalId, depositor), bz)
}

// GetAllDeposits returns all the deposits from the store
func (keeper Keeper) GetAllDeposits(ctx context.Context) (deposits v1.Deposits, err error) {
	err = keeper.IterateAllDeposits(ctx, func(deposit v1.Deposit) error {
		deposits = append(deposits, &deposit)
		return nil
	})

	return
}

// GetDeposits returns all the deposits of a proposal
func (keeper Keeper) GetDeposits(ctx context.Context, proposalID uint64) (deposits v1.Deposits, err error) {
	err = keeper.IterateDeposits(ctx, proposalID, func(deposit v1.Deposit) error {
		deposits = append(deposits, &deposit)
		return nil
	})

	return
}

// DeleteAndBurnDeposits deletes and burns all the deposits on a specific proposal.
func (keeper Keeper) DeleteAndBurnDeposits(ctx context.Context, proposalID uint64) error {
	store := keeper.storeService.OpenKVStore(ctx)

	err := keeper.IterateDeposits(ctx, proposalID, func(deposit v1.Deposit) error {
		err := keeper.bankKeeper.BurnCoins(ctx, types.ModuleName, deposit.Amount)
		if err != nil {
			return err
		}

		depositor, err := keeper.authKeeper.StringToBytes(deposit.Depositor)
		if err != nil {
			return err
		}

		err = store.Delete(types.DepositKey(proposalID, depositor))
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

// IterateAllDeposits iterates over all the stored deposits and performs a callback function.
func (keeper Keeper) IterateAllDeposits(ctx context.Context, cb func(deposit v1.Deposit) error) error {
	store := keeper.storeService.OpenKVStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.DepositsKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit v1.Deposit

		err := keeper.cdc.Unmarshal(iterator.Value(), &deposit)
		if err != nil {
			return err
		}

		err = cb(deposit)
		// exit early without error if cb returns ErrStopIterating
		if errors.IsOf(err, errors.ErrStopIterating) {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

// IterateDeposits iterates over all the proposals deposits and performs a callback function
func (keeper Keeper) IterateDeposits(ctx context.Context, proposalID uint64, cb func(deposit v1.Deposit) error) error {
	store := keeper.storeService.OpenKVStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.DepositsKey(proposalID))

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit v1.Deposit
		err := keeper.cdc.Unmarshal(iterator.Value(), &deposit)
		if err != nil {
			return err
		}

		err = cb(deposit)
		// exit early without error if cb returns ErrStopIterating
		if errors.IsOf(err, errors.ErrStopIterating) {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

// AddDeposit adds or updates a deposit of a specific depositor on a specific proposal.
// Activates voting period when appropriate and returns true in that case, else returns false.
func (keeper Keeper) AddDeposit(ctx context.Context, proposalID uint64, depositorAddr sdk.AccAddress, depositAmount sdk.Coins) (bool, error) {
	// Checks to see if proposal exists
	proposal, err := keeper.GetProposal(ctx, proposalID)
	if err != nil {
		if errors.IsOf(err, types.ErrProposalNotFound) {
			return false, errors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
		}
		return false, err
	}

	// Check if proposal is still depositable
	if (proposal.Status != v1.StatusDepositPeriod) && (proposal.Status != v1.StatusVotingPeriod) {
		return false, errors.Wrapf(types.ErrInactiveProposal, "%d", proposalID)
	}

	// update the governance module's account coins pool
	err = keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, depositAmount)
	if err != nil {
		return false, err
	}

	// Update proposal
	proposal.TotalDeposit = sdk.NewCoins(proposal.TotalDeposit...).Add(depositAmount...)
	err = keeper.SetProposal(ctx, proposal)
	if err != nil {
		return false, err
	}

	// Check if deposit has provided sufficient total funds to transition the proposal into the voting period
	activatedVotingPeriod := false
	params, err := keeper.GetParams(ctx)
	if err != nil {
		return false, err
	}
	minDepositAmount := proposal.GetMinDepositFromParams(params)

	if proposal.Status == v1.StatusDepositPeriod && sdk.NewCoins(proposal.TotalDeposit...).IsAllGTE(minDepositAmount) {
		err = keeper.ActivateVotingPeriod(ctx, proposal)
		if err != nil {
			return false, err
		}

		activatedVotingPeriod = true
	}

	// Add or update deposit object
	deposit, err := keeper.GetDeposit(ctx, proposalID, depositorAddr)
	switch {
	case err == nil:
		// deposit exists
		deposit.Amount = sdk.NewCoins(deposit.Amount...).Add(depositAmount...)
	case errors.IsOf(err, types.ErrDepositNotFound):
		// deposit doesn't exist
		deposit = v1.NewDeposit(proposalID, depositorAddr, depositAmount)
	default:
		// failed to get deposit
		return false, err
	}

	// called when deposit has been added to a proposal, however the proposal may not be active
	keeper.Hooks().AfterProposalDeposit(ctx, proposalID, depositorAddr)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	err = keeper.SetDeposit(ctx, deposit)
	if err != nil {
		return false, err
	}

	return activatedVotingPeriod, nil
}

// ChargeDeposit will charge proposal cancellation fee (deposits * proposal_cancel_burn_rate)  and
// send to a destAddress if defined or burn otherwise.
// Remaining funds are send back to the depositor.
func (keeper Keeper) ChargeDeposit(ctx context.Context, proposalID uint64, destAddress, proposalCancelRate string) error {
	store := keeper.storeService.OpenKVStore(ctx)
	rate := sdkmath.LegacyMustNewDecFromStr(proposalCancelRate)
	var cancellationCharges sdk.Coins

	deposits, err := keeper.GetDeposits(ctx, proposalID)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		depositerAddress, err := keeper.authKeeper.StringToBytes(deposit.Depositor)
		if err != nil {
			return err
		}

		var remainingAmount sdk.Coins

		for _, coins := range deposit.Amount {
			burnAmount := sdk.NewDecFromInt(coins.Amount).Mul(rate).TruncateInt()
			// remaining amount = deposits amount - burn amount
			remainingAmount = remainingAmount.Add(
				sdk.NewCoin(
					coins.Denom,
					coins.Amount.Sub(burnAmount),
				),
			)
			cancellationCharges = cancellationCharges.Add(
				sdk.NewCoin(
					coins.Denom,
					burnAmount,
				),
			)
		}

		if !remainingAmount.IsZero() {
			err := keeper.bankKeeper.SendCoinsFromModuleToAccount(
				ctx, types.ModuleName, depositerAddress, remainingAmount,
			)
			if err != nil {
				return err
			}
		}

		err = store.Delete(types.DepositKey(deposit.ProposalId, depositerAddress))
		if err != nil {
			return err
		}
	}

	// burn the cancellation fee or sent the cancellation charges to destination address.
	if !cancellationCharges.IsZero() {
		// get the distribution module account address
		distributionAddress := keeper.authKeeper.GetModuleAddress(disttypes.ModuleName)
		switch {
		case len(destAddress) == 0:
			// burn the cancellation charges from deposits
			err := keeper.bankKeeper.BurnCoins(ctx, types.ModuleName, cancellationCharges)
			if err != nil {
				return err
			}
		case distributionAddress.String() == destAddress:
			err := keeper.distrKeeper.FundCommunityPool(ctx, cancellationCharges, keeper.ModuleAccountAddress())
			if err != nil {
				return err
			}
		default:
			destAccAddress, err := keeper.authKeeper.StringToBytes(destAddress)
			if err != nil {
				return err
			}
			err = keeper.bankKeeper.SendCoinsFromModuleToAccount(
				ctx, types.ModuleName, destAccAddress, cancellationCharges,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// RefundAndDeleteDeposits refunds and deletes all the deposits on a specific proposal.
func (keeper Keeper) RefundAndDeleteDeposits(ctx context.Context, proposalID uint64) error {
	store := keeper.storeService.OpenKVStore(ctx)

	err := keeper.IterateDeposits(ctx, proposalID, func(deposit v1.Deposit) error {
		depositor, err := keeper.authKeeper.StringToBytes(deposit.Depositor)
		if err != nil {
			return err
		}

		err = keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, deposit.Amount)
		if err != nil {
			return err
		}

		err = store.Delete(types.DepositKey(proposalID, depositor))
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

// validateInitialDeposit validates if initial deposit is greater than or equal to the minimum
// required at the time of proposal submission. This threshold amount is determined by
// the deposit parameters. Returns nil on success, error otherwise.
func (keeper Keeper) validateInitialDeposit(ctx context.Context, initialDeposit sdk.Coins, expedited bool) error {
	params, err := keeper.GetParams(ctx)
	if err != nil {
		return err
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
		return errors.Wrapf(types.ErrMinDepositTooSmall, "was (%s), need (%s)", initialDeposit, minDepositCoins)
	}
	return nil
}
