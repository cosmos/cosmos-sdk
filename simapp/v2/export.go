package simapp

import (
	"context"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *SimApp[T]) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (servertypes.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := context.Background()

	// We export at last height + 1, because that's the height at which
	// CometBFT will start InitChain.
	latestHeight, err := app.LoadLatestHeight()

	if err != nil {
		return servertypes.ExportedApp{}, err
	}
	height := latestHeight + 1
	// if forZeroHeight {
	// 	height = 0
	// 	app.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
	// }

	genesis, err := app.ExportGenesis(ctx, height)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        genesis,
		Validators:      nil,
		Height:          int64(height),
		ConsensusParams: cmtproto.ConsensusParams{}, // TODO: CometBFT consensus params
	}, err
}

// prepare for fresh start at zero height
// NOTE zero height genesis is a temporary feature which will be deprecated
//
//	in favor of export at a block height
// func (app *SimApp[T]) prepForZeroHeightGenesis(ctx context.Context, jailAllowedAddrs []string) {
// 	applyAllowedAddrs := false

// 	// check if there is a allowed address list
// 	if len(jailAllowedAddrs) > 0 {
// 		applyAllowedAddrs = true
// 	}

// 	allowedAddrsMap := make(map[string]bool)

// 	for _, addr := range jailAllowedAddrs {
// 		_, err := sdk.ValAddressFromBech32(addr)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		allowedAddrsMap[addr] = true
// 	}

// 	/* Handle fee distribution state. */

// 	// withdraw all validator commission
// 	err := app.StakingKeeper.IterateValidators(ctx, func(_ int64, val sdk.ValidatorI) (stop bool) {
// 		valBz, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
// 		if err != nil {
// 			panic(err)
// 		}
// 		_, _ = app.DistrKeeper.WithdrawValidatorCommission(ctx, valBz)
// 		return false
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	// withdraw all delegator rewards
// 	dels, err := app.StakingKeeper.GetAllDelegations(ctx)
// 	if err != nil {
// 		panic(err)
// 	}

// 	for _, delegation := range dels {
// 		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
// 		if err != nil {
// 			panic(err)
// 		}

// 		delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)

// 		_, _ = app.DistrKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
// 	}

// 	// clear validator slash events
// 	err = app.DistrKeeper.ValidatorSlashEvents.Clear(ctx, nil)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// clear validator historical rewards
// 	err = app.DistrKeeper.ValidatorHistoricalRewards.Clear(ctx, nil)
// 	if err != nil {
// 		panic(err)
// 	}

// 	//TODO: set height to 0

// 	// reinitialize all validators
// 	err = app.StakingKeeper.IterateValidators(ctx, func(_ int64, val sdk.ValidatorI) (stop bool) {
// 		valBz, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
// 		if err != nil {
// 			panic(err)
// 		}
// 		// donate any unwithdrawn outstanding reward tokens to the community pool
// 		rewards, err := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, valBz)
// 		if err != nil {
// 			panic(err)
// 		}
// 		feePool, err := app.DistrKeeper.FeePool.Get(ctx)
// 		if err != nil {
// 			panic(err)
// 		}
// 		feePool.DecimalPool = feePool.DecimalPool.Add(rewards...) // distribution will allocate this to the protocolpool eventually
// 		if err := app.DistrKeeper.FeePool.Set(ctx, feePool); err != nil {
// 			panic(err)
// 		}

// 		if err := app.DistrKeeper.Hooks().AfterValidatorCreated(ctx, valBz); err != nil {
// 			panic(err)
// 		}
// 		return false
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	// reinitialize all delegations
// 	for _, del := range dels {
// 		valAddr, err := sdk.ValAddressFromBech32(del.ValidatorAddress)
// 		if err != nil {
// 			panic(err)
// 		}
// 		delAddr := sdk.MustAccAddressFromBech32(del.DelegatorAddress)

// 		if err := app.DistrKeeper.Hooks().BeforeDelegationCreated(ctx, delAddr, valAddr); err != nil {
// 			// never called as BeforeDelegationCreated always returns nil
// 			panic(fmt.Errorf("error while incrementing period: %w", err))
// 		}

// 		if err := app.DistrKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr); err != nil {
// 			// never called as AfterDelegationModified always returns nil
// 			panic(fmt.Errorf("error while creating a new delegation period record: %w", err))
// 		}
// 	}

// 	/* Handle staking state. */

// 	// iterate through redelegations, reset creation height
// 	err = app.StakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) (stop bool) {
// 		for i := range red.Entries {
// 			red.Entries[i].CreationHeight = 0
// 		}
// 		err = app.StakingKeeper.SetRedelegation(ctx, red)
// 		if err != nil {
// 			panic(err)
// 		}
// 		return false
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	// iterate through unbonding delegations, reset creation height
// 	err = app.StakingKeeper.UnbondingDelegations.Walk(
// 		ctx,
// 		nil,
// 		func(key collections.Pair[[]byte, []byte], ubd stakingtypes.UnbondingDelegation) (stop bool, err error) {
// 			for i := range ubd.Entries {
// 				ubd.Entries[i].CreationHeight = 0
// 			}
// 			err = app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
// 			if err != nil {
// 				return true, err
// 			}
// 			return false, err
// 		},
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	counter := 0
// 	iter, err := app.StakingKeeper.Validators.IterateRaw(ctx, []byte{}, []byte{}, collections.OrderDescending)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for ; iter.Valid(); iter.Next() {
// 		key, err := iter.KeyValue()
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		addr := sdk.ValAddress(stakingtypes.AddressFromValidatorsKey(key.Key))
// 		validator, err := app.StakingKeeper.GetValidator(ctx, addr)
// 		if err != nil {
// 			panic("expected validator, not found")
// 		}

// 		validator.UnbondingHeight = 0
// 		if applyAllowedAddrs && !allowedAddrsMap[addr.String()] {
// 			validator.Jailed = true
// 		}

// 		if err = app.StakingKeeper.SetValidator(ctx, validator); err != nil {
// 			panic(err)
// 		}
// 		counter++
// 	}

// 	if err := iter.Close(); err != nil {
// 		app.Logger().Error("error while closing the key-value store reverse prefix iterator: ", err)
// 		return
// 	}

// 	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	/* Handle slashing state. */

// 	// reset start height on signing infos
// 	err = app.SlashingKeeper.ValidatorSigningInfo.Walk(ctx, nil, func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool, err error) {
// 		info.StartHeight = 0
// 		err = app.SlashingKeeper.ValidatorSigningInfo.Set(ctx, addr, info)
// 		if err != nil {
// 			return true, err
// 		}
// 		return false, nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// }
