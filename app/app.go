package app

import (
	"github.com/cosmos/cosmos-sdk/types/module"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/baseapp"

	grpc1 "github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

type app struct {
	*baseapp.BaseApp
	appProvider *AppProvider
	mm          module.Manager
}

var _ servertypes.Application = &app{}

func (a *app) exportAppStateAndValidators(
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

func (a *app) Info(info abci.RequestInfo) abci.ResponseInfo {
	panic("implement me")
}

func (a *app) SetOption(option abci.RequestSetOption) abci.ResponseSetOption {
	panic("implement me")
}

func (a *app) Query(query abci.RequestQuery) abci.ResponseQuery {
	panic("implement me")
}

func (a *app) CheckTx(checkTx abci.RequestCheckTx) abci.ResponseCheckTx {
	panic("implement me")
}

func (a *app) InitChain(chain abci.RequestInitChain) abci.ResponseInitChain {
	panic("implement me")
}

func (a *app) BeginBlock(block abci.RequestBeginBlock) abci.ResponseBeginBlock {
	panic("implement me")
}

func (a *app) DeliverTx(deliverTx abci.RequestDeliverTx) abci.ResponseDeliverTx {
	panic("implement me")
}

func (a *app) EndBlock(block abci.RequestEndBlock) abci.ResponseEndBlock {
	panic("implement me")
}

func (a *app) Commit() abci.ResponseCommit {
	panic("implement me")
}

func (a *app) ListSnapshots(listSnapshots abci.RequestListSnapshots) abci.ResponseListSnapshots {
	panic("implement me")
}

func (a *app) OfferSnapshot(snapshot abci.RequestOfferSnapshot) abci.ResponseOfferSnapshot {
	panic("implement me")
}

func (a *app) LoadSnapshotChunk(chunk abci.RequestLoadSnapshotChunk) abci.ResponseLoadSnapshotChunk {
	panic("implement me")
}

func (a *app) ApplySnapshotChunk(chunk abci.RequestApplySnapshotChunk) abci.ResponseApplySnapshotChunk {
	panic("implement me")
}

func (a *app) RegisterAPIRoutes(a2 *api.Server, config config.APIConfig) {
	panic("implement me")
}

func (a *app) RegisterGRPCServer(context client.Context, g grpc1.Server) {
	panic("implement me")
}

func (a *app) RegisterTxService(clientCtx client.Context) {
	panic("implement me")
}

func (a *app) RegisterTendermintService(clientCtx client.Context) {
	panic("implement me")
}

func (a *app) prepForZeroHeightGenesis(ctx sdk.Context, jailAllowedAddrs []string) {
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
