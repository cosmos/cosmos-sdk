package app

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// export the state of gaia for a genesis file
func (app *GaiaApp) ExportAppStateAndValidators(forZeroHeight bool) (
	appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {

	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})

	if forZeroHeight {
		app.prepForZeroHeightGenesis(ctx)
	}

	// iterate to get the accounts
	accounts := []GenesisAccount{}
	appendAccount := func(acc auth.Account) (stop bool) {
		account := NewGenesisAccountI(acc)
		accounts = append(accounts, account)
		return false
	}
	app.accountKeeper.IterateAccounts(ctx, appendAccount)

	genState := NewGenesisState(
		accounts,
		auth.ExportGenesis(ctx, app.feeCollectionKeeper),
		stake.ExportGenesis(ctx, app.stakeKeeper),
		mint.ExportGenesis(ctx, app.mintKeeper),
		distr.ExportGenesis(ctx, app.distrKeeper),
		gov.ExportGenesis(ctx, app.govKeeper),
		slashing.ExportGenesis(ctx, app.slashingKeeper),
	)
	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}
	validators = stake.WriteValidators(ctx, app.stakeKeeper)
	return appState, validators, nil
}

// prepare for fresh start at zero height
func (app *GaiaApp) prepForZeroHeightGenesis(ctx sdk.Context) {

	/* TODO XXX check some invariants */

	height := ctx.BlockHeight()

	valAccum := sdk.ZeroDec()
	vdiIter := func(_ int64, vdi distr.ValidatorDistInfo) bool {
		lastValPower := app.stakeKeeper.GetLastValidatorPower(ctx, vdi.OperatorAddr)
		valAccum = valAccum.Add(vdi.GetValAccum(height, sdk.NewDecFromInt(lastValPower)))
		return false
	}
	app.distrKeeper.IterateValidatorDistInfos(ctx, vdiIter)

	lastTotalPower := sdk.NewDecFromInt(app.stakeKeeper.GetLastTotalPower(ctx))
	totalAccum := app.distrKeeper.GetFeePool(ctx).GetTotalValAccum(height, lastTotalPower)

	if !totalAccum.Equal(valAccum) {
		panic(fmt.Errorf("validator accum invariance: \n\tfee pool totalAccum: %v"+
			"\n\tvalidator accum \t%v\n", totalAccum.String(), valAccum.String()))
	}

	fmt.Printf("accum invariant ok!\n")

	/* END TODO XXX */

	/* Handle fee distribution state. */

	// withdraw all delegator & validator rewards
	vdiIter = func(_ int64, valInfo distr.ValidatorDistInfo) (stop bool) {
		err := app.distrKeeper.WithdrawValidatorRewardsAll(ctx, valInfo.OperatorAddr)
		if err != nil {
			panic(err)
		}
		return false
	}
	app.distrKeeper.IterateValidatorDistInfos(ctx, vdiIter)

	ddiIter := func(_ int64, distInfo distr.DelegationDistInfo) (stop bool) {
		err := app.distrKeeper.WithdrawDelegationReward(
			ctx, distInfo.DelegatorAddr, distInfo.ValOperatorAddr)
		if err != nil {
			panic(err)
		}
		return false
	}
	app.distrKeeper.IterateDelegationDistInfos(ctx, ddiIter)

	// delete all distribution infos
	// these will be recreated in InitGenesis
	app.distrKeeper.RemoveValidatorDistInfos(ctx)
	app.distrKeeper.RemoveDelegationDistInfos(ctx)

	// assert that the fee pool is empty
	feePool := app.distrKeeper.GetFeePool(ctx)
	if !feePool.TotalValAccum.Accum.IsZero() {
		panic("unexpected leftover validator accum")
	}
	bondDenom := app.stakeKeeper.GetParams(ctx).BondDenom
	if !feePool.ValPool.AmountOf(bondDenom).IsZero() {
		panic(fmt.Sprintf("unexpected leftover validator pool coins: %v",
			feePool.ValPool.AmountOf(bondDenom).String()))
	}

	// reset fee pool height, save fee pool
	feePool.TotalValAccum.UpdateHeight = 0
	app.distrKeeper.SetFeePool(ctx, feePool)

	/* Handle stake state. */

	// iterate through validators by power descending, reset bond height, update bond intra-tx counter
	store := ctx.KVStore(app.keyStake)
	iter := sdk.KVStoreReversePrefixIterator(store, stake.ValidatorsByPowerIndexKey)
	counter := int16(0)
	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(iter.Value())
		validator, found := app.stakeKeeper.GetValidator(ctx, addr)
		if !found {
			panic("expected validator, not found")
		}
		validator.BondHeight = 0
		validator.BondIntraTxCounter = counter
		validator.UnbondingHeight = 0
		app.stakeKeeper.SetValidator(ctx, validator)
		counter++
	}
	iter.Close()

	/* Handle slashing state. */

	// we have to clear the slashing periods, since they reference heights
	app.slashingKeeper.DeleteValidatorSlashingPeriods(ctx)

	// reset start height on signing infos
	app.slashingKeeper.IterateValidatorSigningInfos(ctx, func(addr sdk.ConsAddress, info slashing.ValidatorSigningInfo) (stop bool) {
		info.StartHeight = 0
		app.slashingKeeper.SetValidatorSigningInfo(ctx, addr, info)
		return false
	})
}
