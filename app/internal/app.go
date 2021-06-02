package internal

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/spf13/cast"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/baseapp"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// NewApp is an AppCreator
func (ap *AppProvider) AppCreator(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	return ap.newApp(
		logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		appOpts,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshotStore(snapshotStore),
		baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval))),
		baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent))),
	)
}

func (ap *AppProvider) newApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *theApp {
	return &theApp{}
}

// AppExport creates a new app (optionally at a given height) and exports state.
func (ap *AppProvider) AppExportor(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string,
	appOpts servertypes.AppOptions) (servertypes.ExportedApp, error) {

	var a *theApp
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	if height != -1 {
		a = ap.newApp(logger, db, traceStore, false, map[int64]bool{}, homePath, uint(1), appOpts)

		if err := a.LoadVersion(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		a = ap.newApp(logger, db, traceStore, true, map[int64]bool{}, homePath, uint(1), appOpts)
	}

	return a.exportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
}

type theApp struct {
	*baseapp.BaseApp
	appProvider *AppProvider
	mm          module.Manager
}

var _ servertypes.Application = &theApp{}

func (a *theApp) RegisterAPIRoutes(server *api.Server, config config.APIConfig) {
	panic("implement me")
}

func (a *theApp) RegisterTxService(clientCtx client.Context) {
	panic("implement me")
}

func (a *theApp) RegisterTendermintService(clientCtx client.Context) {
	panic("implement me")
}

func (a *theApp) exportAppStateAndValidators(
	forZeroHeight bool, jailAllowedAddrs []string,
) (servertypes.ExportedApp, error) {
	//// as if they could withdraw from the start of the next block
	//ctx := a.NewContext(true, tmproto.Header{Height: a.LastBlockHeight()})
	//
	//// We export at last height + 1, because that's the height at which
	//// Tendermint will start InitChain.
	//height := a.LastBlockHeight() + 1
	//if forZeroHeight {
	//	height = 0
	//	a.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
	//}
	//
	//genState := a.mm.ExportGenesis(ctx, a.appProvider.codec)
	//appState, err := json.MarshalIndent(genState, "", "  ")
	//if err != nil {
	//	return servertypes.ExportedApp{}, err
	//}

	panic("TODO")
	//validators, err := staking.WriteValidators(ctx, app.stakingKeeper)
	//return servertypes.ExportedApp{
	//	AppState:        appState,
	//	Validators:      validators,
	//	Height:          height,
	//	ConsensusParams: a.GetConsensusParams(ctx),
	//}, err
}

func (a *theApp) prepForZeroHeightGenesis(ctx sdk.Context, jailAllowedAddrs []string) {
	panic("TODO")
	//applyAllowedAddrs := false
	//
	//// check if there is a allowed address list
	//if len(jailAllowedAddrs) > 0 {
	//	applyAllowedAddrs = true
	//}
	//
	//allowedAddrsMap := make(map[string]bool)
	//
	//for _, addr := range jailAllowedAddrs {
	//	_, err := sdk.ValAddressFromBech32(addr)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	allowedAddrsMap[addr] = true
	//}
	//
	///* Just to be safe, assert the invariants on current state. */
	//app.CrisisKeeper.AssertInvariants(ctx)
	//
	///* Handle fee distribution state. */
	//
	//// withdraw all validator commission
	//app.StakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
	//	_, _ = app.DistrKeeper.WithdrawValidatorCommission(ctx, val.GetOperator())
	//	return false
	//})
	//
	//// withdraw all delegator rewards
	//dels := app.StakingKeeper.GetAllDelegations(ctx)
	//for _, delegation := range dels {
	//	valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	delAddr, err := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
	//	if err != nil {
	//		panic(err)
	//	}
	//	_, _ = app.DistrKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
	//}
	//
	//// clear validator slash events
	//app.DistrKeeper.DeleteAllValidatorSlashEvents(ctx)
	//
	//// clear validator historical rewards
	//app.DistrKeeper.DeleteAllValidatorHistoricalRewards(ctx)
	//
	//// set context height to zero
	//height := ctx.BlockHeight()
	//ctx = ctx.WithBlockHeight(0)
	//
	//// reinitialize all validators
	//app.StakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
	//	// donate any unwithdrawn outstanding reward fraction tokens to the community pool
	//	scraps := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, val.GetOperator())
	//	feePool := app.DistrKeeper.GetFeePool(ctx)
	//	feePool.CommunityPool = feePool.CommunityPool.Add(scraps...)
	//	app.DistrKeeper.SetFeePool(ctx, feePool)
	//
	//	app.DistrKeeper.Hooks().AfterValidatorCreated(ctx, val.GetOperator())
	//	return false
	//})
	//
	//// reinitialize all delegations
	//for _, del := range dels {
	//	valAddr, err := sdk.ValAddressFromBech32(del.ValidatorAddress)
	//	if err != nil {
	//		panic(err)
	//	}
	//	delAddr, err := sdk.AccAddressFromBech32(del.DelegatorAddress)
	//	if err != nil {
	//		panic(err)
	//	}
	//	app.DistrKeeper.Hooks().BeforeDelegationCreated(ctx, delAddr, valAddr)
	//	app.DistrKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr)
	//}
	//
	//// reset context height
	//ctx = ctx.WithBlockHeight(height)
	//
	///* Handle staking state. */
	//
	//// iterate through redelegations, reset creation height
	//app.StakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) (stop bool) {
	//	for i := range red.Entries {
	//		red.Entries[i].CreationHeight = 0
	//	}
	//	app.StakingKeeper.SetRedelegation(ctx, red)
	//	return false
	//})
	//
	//// iterate through unbonding delegations, reset creation height
	//app.StakingKeeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd stakingtypes.UnbondingDelegation) (stop bool) {
	//	for i := range ubd.Entries {
	//		ubd.Entries[i].CreationHeight = 0
	//	}
	//	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	//	return false
	//})
	//
	//// Iterate through validators by power descending, reset bond heights, and
	//// update bond intra-tx counters.
	//store := ctx.KVStore(app.keys[stakingtypes.StoreKey])
	//iter := sdk.KVStoreReversePrefixIterator(store, stakingtypes.ValidatorsKey)
	//counter := int16(0)
	//
	//for ; iter.Valid(); iter.Next() {
	//	addr := sdk.ValAddress(stakingtypes.AddressFromValidatorsKey(iter.Key()))
	//	validator, found := app.StakingKeeper.GetValidator(ctx, addr)
	//	if !found {
	//		panic("expected validator, not found")
	//	}
	//
	//	validator.UnbondingHeight = 0
	//	if applyAllowedAddrs && !allowedAddrsMap[addr.String()] {
	//		validator.Jailed = true
	//	}
	//
	//	app.StakingKeeper.SetValidator(ctx, validator)
	//	counter++
	//}
	//
	//iter.Close()
	//
	//_, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	///* Handle slashing state. */
	//
	//// reset start height on signing infos
	//app.SlashingKeeper.IterateValidatorSigningInfos(
	//	ctx,
	//	func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
	//		info.StartHeight = 0
	//		app.SlashingKeeper.SetValidatorSigningInfo(ctx, addr, info)
	//		return false
	//	},
	//)
}
