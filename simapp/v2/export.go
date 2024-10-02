package simapp

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	slashingtypes "cosmossdk.io/x/slashing/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *SimApp[T]) ExportAppStateAndValidators(jailAllowedAddrs []string) (v2.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := context.Background()

	latestHeight, err := app.LoadLatestHeight()
	if err != nil {
		return v2.ExportedApp{}, err
	}

	genesis, err := app.ExportGenesis(ctx, latestHeight)
	if err != nil {
		return v2.ExportedApp{}, err
	}

	// get the current bonded validators
	resp, err := app.Query(ctx, 0, latestHeight, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
	})

	vals, ok := resp.(*stakingtypes.QueryValidatorsResponse)
	if !ok {
		return v2.ExportedApp{}, fmt.Errorf("invalid response, expected QueryValidatorsResponse")
	}

	// convert to genesis validator
	var genesisVals []sdk.GenesisValidator
	for _, val := range vals.Validators {
		pk, err := val.ConsPubKey()
		if err != nil {
			return v2.ExportedApp{}, err
		}
		jsonPk, err := cryptocodec.PubKeyFromProto(pk)
		if err != nil {
			return v2.ExportedApp{}, err
		}

		genesisVals = append(genesisVals, sdk.GenesisValidator{
			Address: sdk.ConsAddress(pk.Address()).Bytes(),
			PubKey:  jsonPk,
			Power:   val.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx)),
			Name:    val.Description.Moniker,
		})
	}

	return v2.ExportedApp{
		AppState:   genesis,
		Height:     int64(latestHeight),
		Validators: genesisVals,
	}, err
}

// prepForZeroHeightGenesis prepares for fresh start at zero height
func (app *SimApp[T]) prepForZeroHeightGenesis(ctx context.Context, jailAllowedAddrs []string) {
	applyAllowedAddrs := false

	// check if there is a allowed address list
	if len(jailAllowedAddrs) > 0 {
		applyAllowedAddrs = true
	}

	allowedAddrsMap := make(map[string]bool)

	for _, addr := range jailAllowedAddrs {
		_, err := app.TxConfig().SigningContext().ValidatorAddressCodec().StringToBytes(addr)
		if err != nil {
			panic(err)
		}
		allowedAddrsMap[addr] = true
	}

	/* Handle fee distribution state. */

	// withdraw all validator commission
	err := app.StakingKeeper.IterateValidators(ctx, func(_ int64, val sdk.ValidatorI) (stop bool) {
		valBz, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			panic(err)
		}
		_, _ = app.DistrKeeper.WithdrawValidatorCommission(ctx, valBz)
		return false
	})
	if err != nil {
		panic(err)
	}

	// withdraw all delegator rewards
	dels, err := app.StakingKeeper.GetAllDelegations(ctx)
	if err != nil {
		panic(err)
	}

	for _, delegation := range dels {
		valAddr, err := app.TxConfig().SigningContext().ValidatorAddressCodec().StringToBytes(delegation.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		delAddr, err := app.TxConfig().SigningContext().AddressCodec().StringToBytes(delegation.DelegatorAddress)
		if err != nil {
			panic(err)
		}

		_, _ = app.DistrKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
	}

	// clear validator slash events
	err = app.DistrKeeper.ValidatorSlashEvents.Clear(ctx, nil)
	if err != nil {
		panic(err)
	}

	// clear validator historical rewards
	err = app.DistrKeeper.ValidatorHistoricalRewards.Clear(ctx, nil)
	if err != nil {
		panic(err)
	}

	// reinitialize all validators
	err = app.StakingKeeper.IterateValidators(ctx, func(_ int64, val sdk.ValidatorI) (stop bool) {
		valBz, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			panic(err)
		}
		// donate any unwithdrawn outstanding reward tokens to the community pool
		rewards, err := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, valBz)
		if err != nil {
			panic(err)
		}
		feePool, err := app.DistrKeeper.FeePool.Get(ctx)
		if err != nil {
			panic(err)
		}
		feePool.DecimalPool = feePool.DecimalPool.Add(rewards...) // distribution will allocate this to the protocolpool eventually
		if err := app.DistrKeeper.FeePool.Set(ctx, feePool); err != nil {
			panic(err)
		}

		if err := app.DistrKeeper.Hooks().AfterValidatorCreated(ctx, valBz); err != nil {
			panic(err)
		}
		return false
	})
	if err != nil {
		panic(err)
	}

	// reinitialize all delegations
	for _, del := range dels {
		valAddr, err := app.TxConfig().SigningContext().ValidatorAddressCodec().StringToBytes(del.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delAddr, err := app.TxConfig().SigningContext().AddressCodec().StringToBytes(del.DelegatorAddress)
		if err != nil {
			panic(err)
		}

		if err := app.DistrKeeper.Hooks().BeforeDelegationCreated(ctx, delAddr, valAddr); err != nil {
			// never called as BeforeDelegationCreated always returns nil
			panic(fmt.Errorf("error while incrementing period: %w", err))
		}

		if err := app.DistrKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr); err != nil {
			// never called as AfterDelegationModified always returns nil
			panic(fmt.Errorf("error while creating a new delegation period record: %w", err))
		}
	}

	/* Handle staking state. */

	// iterate through redelegations, reset creation height
	err = app.StakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		err = app.StakingKeeper.SetRedelegation(ctx, red)
		if err != nil {
			panic(err)
		}
		return false
	})
	if err != nil {
		panic(err)
	}

	// iterate through unbonding delegations, reset creation height
	err = app.StakingKeeper.UnbondingDelegations.Walk(
		ctx,
		nil,
		func(_ collections.Pair[[]byte, []byte], ubd stakingtypes.UnbondingDelegation) (stop bool, err error) {
			for i := range ubd.Entries {
				ubd.Entries[i].CreationHeight = 0
			}
			err = app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
			if err != nil {
				return true, err
			}
			return false, err
		},
	)
	if err != nil {
		panic(err)
	}
	// Iterate through validators by power descending, reset bond heights, and
	// update bond intra-tx counters.
	store := ctx.KVStore(app.UnsafeFindStoreKey(stakingtypes.StoreKey))
	iter := storetypes.KVStoreReversePrefixIterator(store, stakingtypes.ValidatorsKey)
	counter := int16(0)

	app.StakingKeeper.Validators.Walk(ctx)

	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(stakingtypes.AddressFromValidatorsKey(iter.Key()))
		validator, err := app.StakingKeeper.GetValidator(ctx, addr)
		if err != nil {
			panic("expected validator, not found")
		}

		valAddr, err := app.StakingKeeper.ValidatorAddressCodec().BytesToString(addr)
		if err != nil {
			panic(err)
		}

		validator.UnbondingHeight = 0
		if applyAllowedAddrs && !allowedAddrsMap[valAddr] {
			validator.Jailed = true
		}

		if err = app.StakingKeeper.SetValidator(ctx, validator); err != nil {
			panic(err)
		}
		counter++
	}

	if err := iter.Close(); err != nil {
		app.Logger().Error("error while closing the key-value store reverse prefix iterator: ", err)
		return
	}

	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		panic(err)
	}

	/* Handle slashing state. */

	// reset start height on signing infos
	err = app.SlashingKeeper.ValidatorSigningInfo.Walk(ctx, nil, func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool, err error) {
		info.StartHeight = 0
		err = app.SlashingKeeper.ValidatorSigningInfo.Set(ctx, addr, info)
		if err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		panic(err)
	}
}
