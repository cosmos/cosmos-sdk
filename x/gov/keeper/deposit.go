package keeper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// SetDeposit sets a Deposit to the gov store
func (keeper Keeper) SetDeposit(ctx context.Context, deposit v1.Deposit) error {
	depositor, err := keeper.authKeeper.AddressCodec().StringToBytes(deposit.Depositor)
	if err != nil {
		return err
	}
	return keeper.Deposits.Set(ctx, collections.Join(deposit.ProposalId, sdk.AccAddress(depositor)), deposit)
}

// GetDeposits returns all the deposits of a proposal
func (keeper Keeper) GetDeposits(ctx context.Context, proposalID uint64) (deposits v1.Deposits, err error) {
	err = keeper.IterateDeposits(ctx, proposalID, func(_ collections.Pair[uint64, sdk.AccAddress], deposit v1.Deposit) (bool, error) {
		deposits = append(deposits, &deposit)
		return false, nil
	})
	return deposits, err
}

// DeleteAndBurnDeposits deletes and burns all the deposits on a specific proposal.
func (keeper Keeper) DeleteAndBurnDeposits(ctx context.Context, proposalID uint64) error {
	coinsToBurn := sdk.NewCoins()
	err := keeper.IterateDeposits(ctx, proposalID, func(key collections.Pair[uint64, sdk.AccAddress], deposit v1.Deposit) (stop bool, err error) {
		coinsToBurn = coinsToBurn.Add(deposit.Amount...)
		return false, keeper.Deposits.Remove(ctx, key)
	})
	if err != nil {
		return err
	}

	return keeper.bankKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn)
}

// IterateDeposits iterates over all the proposals deposits and performs a callback function
func (keeper Keeper) IterateDeposits(ctx context.Context, proposalID uint64, cb func(key collections.Pair[uint64, sdk.AccAddress], value v1.Deposit) (bool, error)) error {
	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposalID)
	err := keeper.Deposits.Walk(ctx, rng, cb)
	if err != nil {
		return err
	}
	return nil
}

// AddDeposit adds or updates a deposit of a specific depositor on a specific proposal.
// Activates voting period when appropriate and returns true in that case, else returns false.
func (keeper Keeper) AddDeposit(ctx context.Context, proposalID uint64, depositorAddr sdk.AccAddress, depositAmount sdk.Coins) (bool, error) {
	// Checks to see if proposal exists
	proposal, err := keeper.Proposals.Get(ctx, proposalID)
	if err != nil {
		return false, err
	}

	// Check if proposal is still depositable
	if (proposal.Status != v1.StatusDepositPeriod) && (proposal.Status != v1.StatusVotingPeriod) {
		return false, errors.Wrapf(types.ErrInactiveProposal, "%d", proposalID)
	}

	// Check coins to be deposited match the proposal's deposit params
	params, err := keeper.Params.Get(ctx)
	if err != nil {
		return false, err
	}

	minDepositAmount := proposal.GetMinDepositFromParams(params)
	minDepositRatio, err := sdkmath.LegacyNewDecFromStr(params.GetMinDepositRatio())
	if err != nil {
		return false, err
	}

	// the deposit must only contain valid denoms (listed in the min deposit param)
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
			return false, errors.Wrapf(types.ErrMinDepositTooSmall, "received %s but need at least one of the following: %s", depositAmount, strings.Join(thresholds, ","))
		}
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
	if proposal.Status == v1.StatusDepositPeriod && sdk.NewCoins(proposal.TotalDeposit...).IsAllGTE(minDepositAmount) {
		err = keeper.ActivateVotingPeriod(ctx, proposal)
		if err != nil {
			return false, err
		}

		activatedVotingPeriod = true
	}

	// Add or update deposit object
	deposit, err := keeper.Deposits.Get(ctx, collections.Join(proposalID, depositorAddr))
	switch {
	case err == nil:
		// deposit exists
		deposit.Amount = sdk.NewCoins(deposit.Amount...).Add(depositAmount...)
	case errors.IsOf(err, collections.ErrNotFound):
		// deposit doesn't exist
		deposit = v1.NewDeposit(proposalID, depositorAddr, depositAmount)
	default:
		// failed to get deposit
		return false, err
	}

	// called when deposit has been added to a proposal, however the proposal may not be active
	err = keeper.Hooks().AfterProposalDeposit(ctx, proposalID, depositorAddr)
	if err != nil {
		return false, err
	}

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
	rate := sdkmath.LegacyMustNewDecFromStr(proposalCancelRate)
	var cancellationCharges sdk.Coins

	deposits, err := keeper.GetDeposits(ctx, proposalID)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		depositerAddress, err := keeper.authKeeper.AddressCodec().StringToBytes(deposit.Depositor)
		if err != nil {
			return err
		}

		var remainingAmount sdk.Coins

		for _, coin := range deposit.Amount {
			burnAmount := sdkmath.LegacyNewDecFromInt(coin.Amount).Mul(rate).TruncateInt()
			// remaining amount = deposits amount - burn amount
			remainingAmount = remainingAmount.Add(
				sdk.NewCoin(
					coin.Denom,
					coin.Amount.Sub(burnAmount),
				),
			)
			cancellationCharges = cancellationCharges.Add(
				sdk.NewCoin(
					coin.Denom,
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
		err = keeper.Deposits.Remove(ctx, collections.Join(deposit.ProposalId, sdk.AccAddress(depositerAddress)))
		if err != nil {
			return err
		}
	}

	// burn the cancellation fee or sent the cancellation charges to destination address.
	if !cancellationCharges.IsZero() {
		// get the distribution module account address
		distributionAddress := keeper.authKeeper.GetModuleAddress(disttypes.ModuleName)
		switch {
		case destAddress == "":
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
			destAccAddress, err := keeper.authKeeper.AddressCodec().StringToBytes(destAddress)
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
	return keeper.IterateDeposits(ctx, proposalID, func(key collections.Pair[uint64, sdk.AccAddress], deposit v1.Deposit) (bool, error) {
		depositor := key.K2()
		err := keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, deposit.Amount)
		if err != nil {
			return false, err
		}
		err = keeper.Deposits.Remove(ctx, key)
		return false, err
	})
}

// validateInitialDeposit validates if initial deposit is greater than or equal to the minimum
// required at the time of proposal submission. This threshold amount is determined by
// the deposit parameters. Returns nil on success, error otherwise.
func (keeper Keeper) validateInitialDeposit(ctx context.Context, params v1.Params, initialDeposit sdk.Coins, expedited bool) error {
	if !initialDeposit.IsValid() || initialDeposit.IsAnyNegative() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, initialDeposit.String())
	}

	minInitialDepositRatio, err := sdkmath.LegacyNewDecFromStr(params.MinInitialDepositRatio)
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
		minDepositCoins[i].Amount = sdkmath.LegacyNewDecFromInt(minDepositCoins[i].Amount).Mul(minInitialDepositRatio).RoundInt()
	}
	if !initialDeposit.IsAllGTE(minDepositCoins) {
		return errors.Wrapf(types.ErrMinDepositTooSmall, "was (%s), need (%s)", initialDeposit, minDepositCoins)
	}
	return nil
}

// validateDepositDenom validates if the deposit denom is accepted by the governance module.
func (keeper Keeper) validateDepositDenom(ctx context.Context, params v1.Params, depositAmount sdk.Coins) error {
	denoms := []string{}
	acceptedDenoms := make(map[string]bool, len(params.MinDeposit))
	for _, coin := range params.MinDeposit {
		acceptedDenoms[coin.Denom] = true
		denoms = append(denoms, coin.Denom)
	}

	for _, coin := range depositAmount {
		if _, ok := acceptedDenoms[coin.Denom]; !ok {
			return errors.Wrapf(types.ErrInvalidDepositDenom, "deposited %s, but gov accepts only the following denom(s): %v", depositAmount, denoms)
		}
	}

	return nil
}
