package app

import (
	"encoding/json"
	"io"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

const (
	appName = "GaiaApp"
)

// default home directories for expected binaries
var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.gaiacli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.gaiad")
)

// Extended ABCI application
type GaiaApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keyMain          *sdk.KVStoreKey
	keyAccount       *sdk.KVStoreKey
	keyStaking       *sdk.KVStoreKey
	tkeyStaking      *sdk.TransientStoreKey
	keySlashing      *sdk.KVStoreKey
	keyMint          *sdk.KVStoreKey
	keyDistr         *sdk.KVStoreKey
	tkeyDistr        *sdk.TransientStoreKey
	keyGov           *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	keyParams        *sdk.KVStoreKey
	tkeyParams       *sdk.TransientStoreKey

	// keepers
	accountKeeper       auth.AccountKeeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	bankKeeper          bank.Keeper
	stakingKeeper       staking.Keeper
	slashingKeeper      slashing.Keeper
	mintKeeper          mint.Keeper
	distrKeeper         distr.Keeper
	govKeeper           gov.Keeper
	crisisKeeper        crisis.Keeper
	paramsKeeper        params.Keeper

	// the module manager
	mm *sdk.ModuleManager
}

// custom tx codec
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	bank.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	distr.RegisterCodec(cdc)
	slashing.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	crisis.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// NewGaiaApp returns a reference to an initialized GaiaApp.
func NewGaiaApp(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint,
	baseAppOptions ...func(*bam.BaseApp)) *GaiaApp {

	cdc := MakeCodec()

	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)

	var app = &GaiaApp{
		BaseApp:          bApp,
		cdc:              cdc,
		invCheckPeriod:   invCheckPeriod,
		keyMain:          sdk.NewKVStoreKey(bam.MainStoreKey),
		keyAccount:       sdk.NewKVStoreKey(auth.StoreKey),
		keyStaking:       sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking:      sdk.NewTransientStoreKey(staking.TStoreKey),
		keyMint:          sdk.NewKVStoreKey(mint.StoreKey),
		keyDistr:         sdk.NewKVStoreKey(distr.StoreKey),
		tkeyDistr:        sdk.NewTransientStoreKey(distr.TStoreKey),
		keySlashing:      sdk.NewKVStoreKey(slashing.StoreKey),
		keyGov:           sdk.NewKVStoreKey(gov.StoreKey),
		keyFeeCollection: sdk.NewKVStoreKey(auth.FeeStoreKey),
		keyParams:        sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams:       sdk.NewTransientStoreKey(params.TStoreKey),
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams)
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := app.paramsKeeper.Subspace(mint.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	govSubspace := app.paramsKeeper.Subspace(gov.DefaultParamspace)
	crisisSubspace := app.paramsKeeper.Subspace(crisis.DefaultParamspace)

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(app.cdc, app.keyAccount, authSubspace, auth.ProtoBaseAccount)
	app.bankKeeper = bank.NewBaseKeeper(app.accountKeeper, bankSubspace, bank.DefaultCodespace)
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.keyFeeCollection)
	stakingKeeper := staking.NewKeeper(app.cdc, app.keyStaking, app.tkeyStaking, app.bankKeeper,
		stakingSubspace, staking.DefaultCodespace)
	app.mintKeeper = mint.NewKeeper(app.cdc, app.keyMint, mintSubspace, &stakingKeeper, app.feeCollectionKeeper)
	app.distrKeeper = distr.NewKeeper(app.cdc, app.keyDistr, distrSubspace, app.bankKeeper, &stakingKeeper,
		app.feeCollectionKeeper, distr.DefaultCodespace)
	app.slashingKeeper = slashing.NewKeeper(app.cdc, app.keySlashing, &stakingKeeper,
		slashingSubspace, slashing.DefaultCodespace)
	app.govKeeper = gov.NewKeeper(app.cdc, app.keyGov, app.paramsKeeper, govSubspace,
		app.bankKeeper, &stakingKeeper, gov.DefaultCodespace)
	app.crisisKeeper = crisis.NewKeeper(crisisSubspace, app.distrKeeper, app.bankKeeper, app.feeCollectionKeeper)

	// register the staking hooks
	// NOTE: The stakingKeeper above is passed by reference, so that it can be
	// modified like below:
	app.stakingKeeper = *stakingKeeper.SetHooks(
		sdk.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()))

	app.mm = sdk.NewModuleManager(
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		crisis.NewAppModule(app.crisisKeeper),
		distr.NewAppModule(app.distrKeeper),
		gov.NewAppModule(app.govKeeper),
		mint.NewAppModule(app.mintKeeper),
		slashing.NewAppModule(app.slashingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.feeCollectionKeeper, app.distrKeeper, app.accountKeeper),
	)

	// NOTE: during begin block slashing happens after distr.BeginBlocker so
	// that there is nothing left over in the validator fee pool, so as to keep
	// the CanWithdrawInvariant invariant.
	// TODO: slashing should really happen at EndBlocker.
	app.mm.SetOrderBeginBlockers(mint.ModuleName, distr.ModuleName, slashing.ModuleName)
	app.mm.SetOrderEndBlockers(gov.ModuleName, staking.ModuleName)
	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// initialize BaseApp
	app.MountStores(app.keyMain, app.keyAccount, app.keyStaking, app.keyMint,
		app.keyDistr, app.keySlashing, app.keyGov, app.keyFeeCollection,
		app.keyParams, app.tkeyParams, app.tkeyStaking, app.tkeyDistr)
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper))
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		err := app.LoadLatestVersion(app.keyMain)
		if err != nil {
			cmn.Exit(err.Error())
		}
	}

	return app
}

// application updates every end block
func (app *GaiaApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	tags := app.mm.BeginBlock(ctx, req)
	return abci.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// application updates every end block
// nolint: unparam
func (app *GaiaApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	validatorUpdates, tags := app.mm.EndBlock(ctx, req)

	if app.invCheckPeriod != 0 && ctx.BlockHeight()%int64(app.invCheckPeriod) == 0 {
		app.assertRuntimeInvariants()
	}
	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Tags:             tags,
	}
}

// initialize store from a genesis state
func (app *GaiaApp) initFromGenesisState(ctx sdk.Context, genesisState GenesisState) []abci.ValidatorUpdate {
	genesisState.Sanitize()

	// load the accounts
	for _, gacc := range genesisState.Accounts {
		acc := gacc.ToAccount()
		acc = app.accountKeeper.NewAccount(ctx, acc) // set account number
		app.accountKeeper.SetAccount(ctx, acc)
	}

	// initialize distribution (must happen before staking)
	distr.InitGenesis(ctx, app.distrKeeper, genesisState.DistrData)

	// load the initial staking information
	validators, err := staking.InitGenesis(ctx, app.stakingKeeper, genesisState.StakingData)
	if err != nil {
		panic(err)
	}

	// initialize module-specific stores
	auth.InitGenesis(ctx, app.accountKeeper, app.feeCollectionKeeper, genesisState.AuthData)
	bank.InitGenesis(ctx, app.bankKeeper, genesisState.BankData)
	slashing.InitGenesis(ctx, app.slashingKeeper, genesisState.SlashingData, genesisState.StakingData.Validators.ToSDKValidators())
	gov.InitGenesis(ctx, app.govKeeper, genesisState.GovData)
	crisis.InitGenesis(ctx, app.crisisKeeper, genesisState.CrisisData)
	mint.InitGenesis(ctx, app.mintKeeper, genesisState.MintData)

	// validate genesis state
	if err := GaiaValidateGenesisState(genesisState); err != nil {
		panic(err)
	}

	if len(genesisState.GenTxs) > 0 {
		validators = app.deliverGenTxs(ctx, genesisState.GenTxs, app.BaseApp.DeliverTx)
	}

	return validators
}

type deliverTxFn func([]byte) abci.ResponseDeliverTx

// TODO move this genTx functionality to staking
func (app *GaiaApp) deliverGenTxs(ctx sdk.Context, genTxs []json.RawMessage, deliverTx deliverTxFn) []abci.ValidatorUpdate {
	for _, genTx := range genTxs {
		var tx auth.StdTx
		app.cdc.MustUnmarshalJSON(genTx, &tx)
		bz := app.cdc.MustMarshalBinaryLengthPrefixed(tx)
		res := deliverTx(bz)
		if !res.IsOK() {
			panic(res.Log)
		}
	}
	return app.stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
}

// custom logic for gaia initialization
func (app *GaiaApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {

	var genesisState GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	validators := app.initFromGenesisState(ctx, genesisState)

	// XXX TODO add initChainer stuff here to the module manager
	// assert runtime invariants
	app.assertRuntimeInvariants()

	return abci.ResponseInitChain{
		Validators: validators,
	}
}

// load a particular height
func (app *GaiaApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keyMain)
}
