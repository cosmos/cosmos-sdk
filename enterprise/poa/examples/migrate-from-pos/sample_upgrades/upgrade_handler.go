// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

// Package sample_upgrades provides a reference POS → POA upgrade handler.
// Copy into your chain's upgrade package and adapt.
package sample_upgrades

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const StandaloneUpgradeName = "pos-to-poa"

// StandaloneStoreUpgrades returns store upgrades for the final clean POA binary.
// The transitional binary must keep staking/distribution/slashing stores mounted
// so the upgrade handler can read them; apply these deletions in a subsequent upgrade.
func StandaloneStoreUpgrades() storetypes.StoreUpgrades {
	return storetypes.StoreUpgrades{
		Added: []string{poatypes.StoreKey},
		Deleted: []string{
			stakingtypes.StoreKey,
			distrtypes.StoreKey,
			slashingtypes.StoreKey,
		},
	}
}

// NewPOSToPOAUpgradeHandler returns the POS → POA upgrade handler.
//
// This loads all delegations into memory. For large chains (>100K delegations),
// swap GetAllDelegations/GetValidatorDelegations for their iterator equivalents
// and increase CometBFT timeouts for the upgrade height.
func NewPOSToPOAUpgradeHandler(
	cdc codec.Codec,
	accountKeeper authkeeper.AccountKeeper,
	stakingKeeper stakingkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	distributionKeeper distrkeeper.Keeper,
	govKeeper *govkeeper.Keeper,
	poaKeeper *poakeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// 1. Snapshot staking validators BEFORE any teardown.
		poaValidators, err := snapshotStakingValidators(ctx, stakingKeeper)
		if err != nil {
			return nil, err
		}

		// 2. Withdraw all distribution rewards (must precede unbond).
		if err := withdrawAllRewards(ctx, accountKeeper, stakingKeeper, distributionKeeper); err != nil {
			return nil, err
		}

		// 3. Force-unbond all delegations — returns tokens to delegators.
		if err := forceUnbondAll(ctx, accountKeeper, stakingKeeper, bankKeeper); err != nil {
			return nil, err
		}

		// 4. Complete in-flight unbonding delegations immediately.
		if err := completeAllUnbondings(ctx, accountKeeper, stakingKeeper, bankKeeper); err != nil {
			return nil, err
		}

		// 5. Remove all in-flight redelegations.
		if err := removeAllRedelegations(ctx, stakingKeeper); err != nil {
			return nil, err
		}

		// 6. Drain rounding dust from staking pools to community pool.
		if err := drainPools(ctx, accountKeeper, bankKeeper, distributionKeeper); err != nil {
			return nil, err
		}

		// 7. Fail all active governance proposals and refund deposits.
		if err := failActiveProposals(ctx, govKeeper); err != nil {
			return nil, err
		}

		// 8. Initialize POA module with snapshotted validators.
		if err := initializePOA(ctx, cdc, poaKeeper, defaultAdminAddress(), poaValidators); err != nil {
			return nil, err
		}

		return fromVM, nil
	}
}

// snapshotStakingValidators converts the bonded validator set to POA validators.
func snapshotStakingValidators(
	ctx context.Context,
	stakingKeeper stakingkeeper.Keeper,
) ([]poatypes.Validator, error) {
	bondedVals, err := stakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return nil, err
	}

	validators := make([]poatypes.Validator, 0, len(bondedVals))
	for _, val := range bondedVals {
		pubKey, err := val.ConsPubKey()
		if err != nil {
			return nil, fmt.Errorf("get consensus pubkey for validator %s: %w", val.GetOperator(), err)
		}
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return nil, fmt.Errorf("wrap pubkey as Any for validator %s: %w", val.GetOperator(), err)
		}

		// Convert valoper address to account address for POA OperatorAddress.
		valBz, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			return nil, fmt.Errorf("decode operator address %s: %w", val.GetOperator(), err)
		}
		operatorAccAddr := sdk.AccAddress(valBz).String()

		validators = append(validators, poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  val.GetConsensusPower(stakingKeeper.PowerReduction(ctx)),
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         val.GetMoniker(),
				OperatorAddress: operatorAccAddr,
			},
		})
	}
	return validators, nil
}

// withdrawAllRewards withdraws all delegation rewards and validator commissions.
// Must run before forceUnbondAll — Unbond triggers distribution hooks.
func withdrawAllRewards(
	ctx context.Context,
	accountKeeper authkeeper.AccountKeeper,
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distrkeeper.Keeper,
) error {
	validators, err := stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}

	for _, val := range validators {
		valBz, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			return err
		}
		valAddr := sdk.ValAddress(valBz)

		// Zero commission is not an error.
		if _, err := distributionKeeper.WithdrawValidatorCommission(ctx, valAddr); err != nil {
			if !errors.Is(err, distrtypes.ErrNoValidatorCommission) {
				return err
			}
		}

		delegations, err := stakingKeeper.GetValidatorDelegations(ctx, valAddr)
		if err != nil {
			return err
		}
		for _, del := range delegations {
			delBz, err := accountKeeper.AddressCodec().StringToBytes(del.GetDelegatorAddr())
			if err != nil {
				return err
			}
			if _, err := distributionKeeper.WithdrawDelegationRewards(ctx, sdk.AccAddress(delBz), valAddr); err != nil {
				if !errors.Is(err, distrtypes.ErrEmptyDelegationDistInfo) {
					return err
				}
			}
		}
	}
	return nil
}

// forceUnbondAll unbonds every delegation and sends tokens back to delegators.
// Unbond() does not move coins — explicit SendCoinsFromModuleToAccount is required.
func forceUnbondAll(
	ctx context.Context,
	accountKeeper authkeeper.AccountKeeper,
	stakingKeeper stakingkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) error {
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	delegations, err := stakingKeeper.GetAllDelegations(ctx)
	if err != nil {
		return err
	}

	for _, del := range delegations {
		valBz, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(del.GetValidatorAddr())
		if err != nil {
			return err
		}
		valAddr := sdk.ValAddress(valBz)

		val, err := stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			continue // validator may have been removed by a previous unbond
		}

		delBz, err := accountKeeper.AddressCodec().StringToBytes(del.GetDelegatorAddr())
		if err != nil {
			return err
		}
		delAddr := sdk.AccAddress(delBz)

		returnAmount, err := stakingKeeper.Unbond(ctx, delAddr, valAddr, del.GetShares())
		if err != nil {
			return err
		}

		pool := stakingtypes.NotBondedPoolName
		if val.IsBonded() {
			pool = stakingtypes.BondedPoolName
		}
		coins := sdk.NewCoins(sdk.NewCoin(bondDenom, returnAmount))
		if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, pool, delAddr, coins); err != nil {
			return err
		}
	}
	return nil
}

// completeAllUnbondings completes all in-flight unbonding delegations immediately.
func completeAllUnbondings(
	ctx context.Context,
	accountKeeper authkeeper.AccountKeeper,
	stakingKeeper stakingkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) error {
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	var iterErr error
	if err := stakingKeeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd stakingtypes.UnbondingDelegation) bool {
		delBz, err := accountKeeper.AddressCodec().StringToBytes(ubd.DelegatorAddress)
		if err != nil {
			iterErr = fmt.Errorf("decode delegator address %s: %w", ubd.DelegatorAddress, err)
			return true
		}
		delAddr := sdk.AccAddress(delBz)

		for _, entry := range ubd.Entries {
			coins := sdk.NewCoins(sdk.NewCoin(bondDenom, entry.Balance))
			if err := bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, delAddr, coins); err != nil {
				iterErr = fmt.Errorf("undelegate coins to %s: %w", delAddr, err)
				return true
			}
		}

		if err := stakingKeeper.RemoveUnbondingDelegation(ctx, ubd); err != nil {
			iterErr = fmt.Errorf("remove unbonding delegation for %s: %w", ubd.DelegatorAddress, err)
			return true
		}
		return false
	}); err != nil {
		return err
	}
	return iterErr
}

// removeAllRedelegations removes all in-flight redelegations. No token movement
// needed — tokens moved to the destination validator when the redelegation started.
func removeAllRedelegations(ctx context.Context, stakingKeeper stakingkeeper.Keeper) error {
	var iterErr error
	if err := stakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) bool {
		if err := stakingKeeper.RemoveRedelegation(ctx, red); err != nil {
			iterErr = fmt.Errorf("remove redelegation %s -> %s: %w",
				red.DelegatorAddress, red.ValidatorDstAddress, err)
			return true
		}
		return false
	}); err != nil {
		return err
	}
	return iterErr
}

// drainPools sends rounding dust from staking pools to the community pool.
func drainPools(
	ctx context.Context,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	distributionKeeper distrkeeper.Keeper,
) error {
	for _, poolName := range []string{stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName} {
		poolAddr := accountKeeper.GetModuleAddress(poolName)
		balance := bankKeeper.GetAllBalances(ctx, poolAddr)
		if !balance.IsZero() {
			if err := distributionKeeper.FundCommunityPool(ctx, balance, poolAddr); err != nil {
				return err
			}
		}
	}
	return nil
}

// failActiveProposals fails all proposals in voting/deposit period, removes them
// from the timed queues, and refunds deposits.
func failActiveProposals(ctx context.Context, govKeeper *govkeeper.Keeper) error {
	return govKeeper.Proposals.Walk(ctx, nil, func(proposalID uint64, proposal govv1.Proposal) (bool, error) {
		switch proposal.Status {
		case govv1.StatusVotingPeriod:
			if proposal.VotingEndTime != nil {
				if err := govKeeper.ActiveProposalsQueue.Remove(ctx, collections.Join(*proposal.VotingEndTime, proposalID)); err != nil {
					return true, err
				}
			}
		case govv1.StatusDepositPeriod:
			if proposal.DepositEndTime != nil {
				if err := govKeeper.InactiveProposalsQueue.Remove(ctx, collections.Join(*proposal.DepositEndTime, proposalID)); err != nil {
					return true, err
				}
			}
		default:
			return false, nil
		}

		proposal.Status = govv1.StatusFailed
		if err := govKeeper.Proposals.Set(ctx, proposalID, proposal); err != nil {
			return true, err
		}
		if err := govKeeper.RefundAndDeleteDeposits(ctx, proposalID); err != nil {
			return true, err
		}
		return false, nil
	})
}

// initializePOA runs POA InitGenesis with the given admin and validators.
// WithBlockHeight(0) is required because InitGenesis expects height 0.
func initializePOA(
	ctx context.Context,
	cdc codec.Codec,
	poaKeeper *poakeeper.Keeper,
	adminAddress string,
	validators []poatypes.Validator,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx).WithBlockHeight(0)

	genesis := &poatypes.GenesisState{
		Params:     poatypes.Params{Admin: adminAddress},
		Validators: validators,
	}
	_, err := poaKeeper.InitGenesis(sdkCtx, cdc, genesis)
	return err
}

// defaultAdminAddress returns the governance module address as the POA admin.
func defaultAdminAddress() string {
	return authtypes.NewModuleAddress(govtypes.ModuleName).String()
}
