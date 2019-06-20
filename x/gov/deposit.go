package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GetDeposit gets the deposit of a specific depositor on a specific proposal
func (keeper Keeper) GetDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) (deposit Deposit, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.DepositKey(proposalID, depositorAddr))
	if bz == nil {
		return deposit, false
	}

	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &deposit)
	return deposit, true
}

func (keeper Keeper) setDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress, deposit Deposit) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(deposit)
	store.Set(types.DepositKey(proposalID, depositorAddr), bz)
}

// AddDeposit adds or updates a deposit of a specific depositor on a specific proposal
// Activates voting period when appropriate
func (keeper Keeper) AddDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress, depositAmount sdk.Coins) (sdk.Error, bool) {
	// Checks to see if proposal exists
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return ErrUnknownProposal(keeper.codespace, proposalID), false
	}

	// Check if proposal is still depositable
	if (proposal.Status != StatusDepositPeriod) && (proposal.Status != StatusVotingPeriod) {
		return ErrAlreadyFinishedProposal(keeper.codespace, proposalID), false
	}

	// Send coins from depositor's account to DepositedCoinsAccAddr account
	// TODO: Don't use an account for this purpose; it's clumsy and prone to misuse.
	err := keeper.ck.SendCoins(ctx, depositorAddr, DepositedCoinsAccAddr, depositAmount)
	if err != nil {
		return err, false
	}

	// Update proposal
	proposal.TotalDeposit = proposal.TotalDeposit.Add(depositAmount)
	keeper.SetProposal(ctx, proposal)

	// Check if deposit has provided sufficient total funds to transition the proposal into the voting period
	activatedVotingPeriod := false
	if proposal.Status == StatusDepositPeriod && proposal.TotalDeposit.IsAllGTE(keeper.GetDepositParams(ctx).MinDeposit) {
		keeper.activateVotingPeriod(ctx, proposal)
		activatedVotingPeriod = true
	}

	// Add or update deposit object
	deposit, found := keeper.GetDeposit(ctx, proposalID, depositorAddr)
	if found {
		deposit.Amount = deposit.Amount.Add(depositAmount)
	} else {
		deposit = NewDeposit(proposalID, depositorAddr, depositAmount)
	}

	keeper.setDeposit(ctx, proposalID, depositorAddr, deposit)

	return nil, activatedVotingPeriod
}

// GetAllDeposits returns all the deposits from the store
func (keeper Keeper) GetAllDeposits(ctx sdk.Context) (deposits Deposits) {
	keeper.IterateAllDeposits(ctx, func(deposit Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	return
}

// GetDeposits returns all the deposits from a proposal
func (keeper Keeper) GetDeposits(ctx sdk.Context, proposalID uint64) (deposits Deposits) {
	keeper.IterateDeposits(ctx, proposalID, func(deposit Deposit) bool {
		deposits = append(deposits, deposit)
		return false
	})
	return
}

// GetDepositsIterator gets all the deposits on a specific proposal as an sdk.Iterator
func (keeper Keeper) GetDepositsIterator(ctx sdk.Context, proposalID uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.DepositsKey(proposalID))
}

// RefundDeposits refunds and deletes all the deposits on a specific proposal
func (keeper Keeper) RefundDeposits(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	keeper.IterateDeposits(ctx, proposalID, func(deposit Deposit) bool {
		err := keeper.ck.SendCoins(ctx, DepositedCoinsAccAddr, deposit.Depositor, deposit.Amount)
		if err != nil {
			panic(err)
		}
		store.Delete(DepositKey(proposalID, deposit.Depositor))
		return false
	})
}

// DeleteDeposits deletes all the deposits on a specific proposal without refunding them
func (keeper Keeper) DeleteDeposits(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	keeper.IterateDeposits(ctx, proposalID, func(deposit Deposit) bool {
		err := keeper.ck.SendCoins(ctx, DepositedCoinsAccAddr, BurnedDepositCoinsAccAddr, deposit.Amount)
		if err != nil {
			panic(err)
		}

		store.Delete(DepositKey(proposalID, deposit.Depositor))
		return false
	})
}
