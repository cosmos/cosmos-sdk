package staking

import (
	"fmt"
	"log"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// InitGenesis sets the pool and parameters for the provided keeper.  For each
// validator in data, it sets that validator in the keeper along with manually
// setting the indexes. In addition, it also sets any delegations found in
// data. Finally, it updates the bonded validators.
// Returns final validator set after applying all declaration and delegations
func InitGenesis(
	ctx sdk.Context, keeper keeper.Keeper, accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper, data *types.GenesisState,
) (res []abci.ValidatorUpdate) {
	bondedTokens := sdk.ZeroInt()
	notBondedTokens := sdk.ZeroInt()

	// We need to pretend to be "n blocks before genesis", where "n" is the
	// validator update delay, so that e.g. slashing periods are correctly
	// initialized for the validator set e.g. with a one-block offset - the
	// first TM block is at height 1, so state updates applied from
	// genesis.json are in block 0.
	ctx = ctx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)

	keeper.SetParams(ctx, data.Params)
	keeper.SetLastTotalPower(ctx, data.LastTotalPower)

	for _, validator := range data.Validators {
		keeper.SetValidator(ctx, validator)

		// Manually set indices for the first time
		keeper.SetValidatorByConsAddr(ctx, validator)
		keeper.SetValidatorByPowerIndex(ctx, validator)

		// Call the creation hook if not exported
		if !data.Exported {
			keeper.AfterValidatorCreated(ctx, validator.GetOperator())
		}

		// update timeslice if necessary
		if validator.IsUnbonding() {
			keeper.InsertUnbondingValidatorQueue(ctx, validator)
		}

		switch validator.GetStatus() {
		case types.Bonded:
			bondedTokens = bondedTokens.Add(validator.GetTokens())
		case types.Unbonding, types.Unbonded:
			notBondedTokens = notBondedTokens.Add(validator.GetTokens())
		default:
			panic("invalid validator status")
		}
	}

	for _, delegation := range data.Delegations {
		delegatorAddress := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)

		// Call the before-creation hook if not exported
		if !data.Exported {
			keeper.BeforeDelegationCreated(ctx, delegatorAddress, delegation.GetValidatorAddr())
		}

		keeper.SetDelegation(ctx, delegation)
		// Call the after-modification hook if not exported
		if !data.Exported {
			keeper.AfterDelegationModified(ctx, delegatorAddress, delegation.GetValidatorAddr())
		}
	}

	for _, ubd := range data.UnbondingDelegations {
		keeper.SetUnbondingDelegation(ctx, ubd)

		for _, entry := range ubd.Entries {
			keeper.InsertUBDQueue(ctx, ubd, entry.CompletionTime)
			notBondedTokens = notBondedTokens.Add(entry.Balance)
		}
	}

	for _, red := range data.Redelegations {
		keeper.SetRedelegation(ctx, red)

		for _, entry := range red.Entries {
			keeper.InsertRedelegationQueue(ctx, red, entry.CompletionTime)
		}
	}

	bondedCoins := sdk.NewCoins(sdk.NewCoin(data.Params.BondDenom, bondedTokens))
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(data.Params.BondDenom, notBondedTokens))

	// check if the unbonded and bonded pools accounts exists
	bondedPool := keeper.GetBondedPool(ctx)
	if bondedPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.BondedPoolName))
	}
	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	bondedBalance := bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	if bondedBalance.IsZero() {
		accountKeeper.SetModuleAccount(ctx, bondedPool)
	}
	// if balance is different from bonded coins panic because genesis is most likely malformed
	if !bondedBalance.IsEqual(bondedCoins) {
		panic(fmt.Sprintf("bonded pool balance is different from bonded coins: %s <-> %s", bondedBalance, bondedCoins))
	}
	notBondedPool := keeper.GetNotBondedPool(ctx)
	if notBondedPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.NotBondedPoolName))
	}

	notBondedBalance := bankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	if notBondedBalance.IsZero() {
		accountKeeper.SetModuleAccount(ctx, notBondedPool)
	}
	// if balance is different from non bonded coins panic because genesis is most likely malformed
	if !notBondedBalance.IsEqual(notBondedCoins) {
		panic(fmt.Sprintf("not bonded pool balance is different from not bonded coins: %s <-> %s", notBondedBalance, notBondedCoins))
	}
	// don't need to run Tendermint updates if we exported
	if data.Exported {
		for _, lv := range data.LastValidatorPowers {
			valAddr, err := sdk.ValAddressFromBech32(lv.Address)
			if err != nil {
				panic(err)
			}
			keeper.SetLastValidatorPower(ctx, valAddr, lv.Power)
			validator, found := keeper.GetValidator(ctx, valAddr)

			if !found {
				panic(fmt.Sprintf("validator %s not found", lv.Address))
			}

			update := validator.ABCIValidatorUpdate(keeper.PowerReduction(ctx))
			update.Power = lv.Power // keep the next-val-set offset, use the last power for the first block
			res = append(res, update)
		}
	} else {
		var err error
		res, err = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Set the total liquid staked tokens
	keeper.SetTotalLiquidStakedTokens(ctx, data.TotalLiquidStakedTokens)

	// Set each tokenize share record, as well as the last tokenize share record ID
	latestId := uint64(0)
	for _, tokenizeShareRecord := range data.TokenizeShareRecords {
		if err := keeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord); err != nil {
			panic(err)
		}
		if tokenizeShareRecord.Id > latestId {
			latestId = tokenizeShareRecord.Id
		}
	}
	if data.LastTokenizeShareRecordId < latestId {
		panic("Tokenize share record specified with ID greater than the latest ID")
	}
	keeper.SetLastTokenizeShareRecordID(ctx, data.LastTokenizeShareRecordId)

	// Set the tokenize shares locks for accounts that have disabled tokenizing shares
	// The lock can either be in status LOCKED or LOCK_EXPIRING
	// If it is in status LOCK_EXPIRING, a the unlocking must also be queued
	for _, tokenizeShareLock := range data.TokenizeShareLocks {
		address := sdk.MustAccAddressFromBech32(tokenizeShareLock.Address)

		switch tokenizeShareLock.Status {
		case types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED.String():
			keeper.AddTokenizeSharesLock(ctx, address)

		case types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING.String():
			completionTime := tokenizeShareLock.CompletionTime

			authorizations := keeper.GetPendingTokenizeShareAuthorizations(ctx, completionTime)
			authorizations.Addresses = append(authorizations.Addresses, address.String())

			keeper.SetPendingTokenizeShareAuthorizations(ctx, completionTime, authorizations)
			keeper.SetTokenizeSharesUnlockTime(ctx, address, completionTime)

		default:
			panic(fmt.Sprintf("Unsupported tokenize share lock status %s", tokenizeShareLock.Status))
		}
	}

	return res
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, params, validators, and bonds found in
// the keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	var unbondingDelegations []types.UnbondingDelegation

	keeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd types.UnbondingDelegation) (stop bool) {
		unbondingDelegations = append(unbondingDelegations, ubd)
		return false
	})

	var redelegations []types.Redelegation

	keeper.IterateRedelegations(ctx, func(_ int64, red types.Redelegation) (stop bool) {
		redelegations = append(redelegations, red)
		return false
	})

	var lastValidatorPowers []types.LastValidatorPower

	keeper.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, power int64) (stop bool) {
		lastValidatorPowers = append(lastValidatorPowers, types.LastValidatorPower{Address: addr.String(), Power: power})
		return false
	})

	return &types.GenesisState{
		Params:                    keeper.GetParams(ctx),
		LastTotalPower:            keeper.GetLastTotalPower(ctx),
		LastValidatorPowers:       lastValidatorPowers,
		Validators:                keeper.GetAllValidators(ctx),
		Delegations:               keeper.GetAllDelegations(ctx),
		UnbondingDelegations:      unbondingDelegations,
		Redelegations:             redelegations,
		Exported:                  true,
		TokenizeShareRecords:      keeper.GetAllTokenizeShareRecords(ctx),
		LastTokenizeShareRecordId: keeper.GetLastTokenizeShareRecordID(ctx),
		TotalLiquidStakedTokens:   keeper.GetTotalLiquidStakedTokens(ctx),
		TokenizeShareLocks:        keeper.GetAllTokenizeSharesLocks(ctx),
	}
}

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx sdk.Context, keeper keeper.Keeper) (vals []tmtypes.GenesisValidator, err error) {
	keeper.IterateLastValidators(ctx, func(_ int64, validator types.ValidatorI) (stop bool) {
		pk, err := validator.ConsPubKey()
		if err != nil {
			return true
		}
		tmPk, err := cryptocodec.ToTmPubKeyInterface(pk)
		if err != nil {
			return true
		}

		vals = append(vals, tmtypes.GenesisValidator{
			Address: sdk.ConsAddress(tmPk.Address()).Bytes(),
			PubKey:  tmPk,
			Power:   validator.GetConsensusPower(keeper.PowerReduction(ctx)),
			Name:    validator.GetMoniker(),
		})

		return false
	})

	return
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate validators)
func ValidateGenesis(data *types.GenesisState) error {
	if err := validateGenesisStateValidators(data.Validators); err != nil {
		return err
	}

	return data.Params.Validate()
}

func validateGenesisStateValidators(validators []types.Validator) error {
	addrMap := make(map[string]bool, len(validators))

	for i := 0; i < len(validators); i++ {
		val := validators[i]
		consPk, err := val.ConsPubKey()
		if err != nil {
			return err
		}

		strKey := string(consPk.Bytes())

		if _, ok := addrMap[strKey]; ok {
			consAddr, err := val.GetConsAddr()
			if err != nil {
				return err
			}
			return fmt.Errorf("duplicate validator in genesis state: moniker %v, address %v", val.Description.Moniker, consAddr)
		}

		if val.Jailed && val.IsBonded() {
			consAddr, err := val.GetConsAddr()
			if err != nil {
				return err
			}
			return fmt.Errorf("validator is bonded and jailed in genesis state: moniker %v, address %v", val.Description.Moniker, consAddr)
		}

		if val.DelegatorShares.IsZero() && !val.IsUnbonding() {
			return fmt.Errorf("bonded/unbonded genesis validator cannot have zero delegator shares, validator: %v", val)
		}

		addrMap[strKey] = true
	}

	return nil
}
