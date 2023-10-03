package keeper_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/evidence/exported"
	"cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramtypes "cosmossdk.io/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	pubkeys = []cryptotypes.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	}

	valAddresses = []sdk.ValAddress{
		sdk.ValAddress(pubkeys[0].Address()),
		sdk.ValAddress(pubkeys[1].Address()),
		sdk.ValAddress(pubkeys[2].Address()),
	}

	// The default power validators are initialized to have within tests
	initAmt   = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))
)

type fixture struct {
	app *integration.App

	sdkCtx sdk.Context
	cdc    codec.Codec

	bankKeeper     bankkeeper.Keeper
	evidenceKeeper *keeper.Keeper
	slashingKeeper slashingkeeper.Keeper
	stakingKeeper  *stakingkeeper.Keeper
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, paramtypes.StoreKey, consensusparamtypes.StoreKey, evidencetypes.StoreKey, stakingtypes.StoreKey, slashingtypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, evidence.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr), addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr))

	slashingKeeper := slashingkeeper.NewKeeper(cdc, codec.NewLegacyAmino(), runtime.NewKVStoreService(keys[slashingtypes.StoreKey]), stakingKeeper, authority.String())

	evidenceKeeper := keeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[evidencetypes.StoreKey]), stakingKeeper, slashingKeeper, addresscodec.NewBech32Codec("cosmos"))
	router := evidencetypes.NewRouter()
	router = router.AddRoute(evidencetypes.RouteEquivocation, testEquivocationHandler(evidenceKeeper))
	evidenceKeeper.SetRouter(router)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper)
	slashingModule := slashing.NewAppModule(cdc, slashingKeeper, accountKeeper, bankKeeper, stakingKeeper, cdc.InterfaceRegistry())
	evidenceModule := evidence.NewAppModule(*evidenceKeeper)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:     authModule,
		banktypes.ModuleName:     bankModule,
		stakingtypes.ModuleName:  stakingModule,
		slashingtypes.ModuleName: slashingModule,
		evidencetypes.ModuleName: evidenceModule,
	})

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// Register MsgServer and QueryServer
	evidencetypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), keeper.NewMsgServerImpl(*evidenceKeeper))
	evidencetypes.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewQuerier(evidenceKeeper))

	assert.NilError(tb, slashingKeeper.Params.Set(sdkCtx, testutil.TestParams()))

	// set default staking params
	assert.NilError(tb, stakingKeeper.Params.Set(sdkCtx, stakingtypes.DefaultParams()))

	return &fixture{
		app:            integrationApp,
		sdkCtx:         sdkCtx,
		cdc:            cdc,
		bankKeeper:     bankKeeper,
		evidenceKeeper: evidenceKeeper,
		slashingKeeper: slashingKeeper,
		stakingKeeper:  stakingKeeper,
	}
}

func TestHandleDoubleSign(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx.WithIsCheckTx(false).WithBlockHeight(1)
	populateValidators(t, f)

	power := int64(100)
	stakingParams, err := f.stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	operatorAddr, valpubkey := valAddresses[0], pubkeys[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)

	selfDelegation := tstaking.CreateValidatorWithValPower(operatorAddr, valpubkey, power, true)

	// execute end-blocker and verify validator attributes
	_, err = f.stakingKeeper.EndBlocker(f.sdkCtx)
	assert.NilError(t, err)
	assert.DeepEqual(t,
		f.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)).String(),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(selfDelegation))).String(),
	)
	val, err := f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.DeepEqual(t, selfDelegation, val.GetBondedTokens())

	assert.NilError(t, f.slashingKeeper.AddrPubkeyRelation.Set(f.sdkCtx, valpubkey.Address(), valpubkey))

	info := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(valpubkey.Address()), f.sdkCtx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	err = f.slashingKeeper.ValidatorSigningInfo.Set(f.sdkCtx, sdk.ConsAddress(valpubkey.Address()), info)
	assert.NilError(t, err)
	// handle a signature to set signing info
	err = f.slashingKeeper.HandleValidatorSignature(ctx, valpubkey.Address(), selfDelegation.Int64(), comet.BlockIDFlagCommit)
	assert.NilError(t, err)
	// double sign less than max age
	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	oldTokens := val.GetTokens()

	nci := comet.Info{
		Evidence: []comet.Evidence{{
			Validator: comet.Validator{Address: valpubkey.Address(), Power: power},
			Type:      comet.DuplicateVote,
			Time:      time.Now().UTC(),
			Height:    1,
		}},
	}

	ctx = ctx.WithCometInfo(nci)
	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx.WithCometInfo(nci)))

	// should be jailed and tombstoned
	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.IsJailed())
	assert.Assert(t, f.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(valpubkey.Address())))

	// tokens should be decreased
	newTokens := val.GetTokens()
	assert.Assert(t, newTokens.LT(oldTokens))

	// submit duplicate evidence
	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx))

	// tokens should be the same (capped slash)
	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.GetTokens().Equal(newTokens))

	// jump to past the unbonding period
	ctx = ctx.WithHeaderInfo(header.Info{Time: time.Unix(1, 0).Add(stakingParams.UnbondingTime)})

	// require we cannot unjail
	assert.Error(t, f.slashingKeeper.Unjail(ctx, operatorAddr), slashingtypes.ErrValidatorJailed.Error())

	// require we be able to unbond now
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	del, _ := f.stakingKeeper.Delegations.Get(ctx, collections.Join(sdk.AccAddress(operatorAddr), operatorAddr))
	validator, _ := f.stakingKeeper.GetValidator(ctx, operatorAddr)
	totalBond := validator.TokensFromShares(del.GetShares()).TruncateInt()
	tstaking.Ctx = ctx
	tstaking.Denom = stakingParams.BondDenom
	tstaking.Undelegate(sdk.AccAddress(operatorAddr), operatorAddr, totalBond, true)

	// query evidence from store
	iter, err := f.evidenceKeeper.Evidences.Iterate(ctx, nil)
	assert.NilError(t, err)
	values, err := iter.Values()
	assert.NilError(t, err)
	assert.Assert(t, len(values) == 1)
}

func TestHandleDoubleSign_TooOld(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx.WithIsCheckTx(false).WithBlockHeight(1).WithHeaderInfo(header.Info{Time: time.Now()})
	populateValidators(t, f)

	power := int64(100)
	stakingParams, err := f.stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	operatorAddr, valpubkey := valAddresses[0], pubkeys[0]

	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)

	amt := tstaking.CreateValidatorWithValPower(operatorAddr, valpubkey, power, true)

	// execute end-blocker and verify validator attributes
	_, err = f.stakingKeeper.EndBlocker(f.sdkCtx)
	assert.NilError(t, err)
	assert.DeepEqual(t,
		f.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(amt))),
	)
	val, err := f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.DeepEqual(t, amt, val.GetBondedTokens())

	nci := comet.Info{Evidence: []comet.Evidence{{
		Validator: comet.Validator{Address: valpubkey.Address(), Power: power},
		Type:      comet.MisbehaviorType(abci.MisbehaviorType_DUPLICATE_VOTE),
		Time:      ctx.HeaderInfo().Time,
		Height:    0,
	}}}

	assert.NilError(t, f.app.BaseApp.StoreConsensusParams(ctx, *simtestutil.DefaultConsensusParams))
	cp := f.app.BaseApp.GetConsensusParams(ctx)

	ctx = ctx.WithCometInfo(nci)
	ctx = ctx.WithConsensusParams(cp)
	ctx = ctx.WithHeaderInfo(header.Info{Time: ctx.HeaderInfo().Time.Add(cp.Evidence.MaxAgeDuration + 1)})
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + cp.Evidence.MaxAgeNumBlocks + 1)

	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx))

	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.IsJailed() == false)
	assert.Assert(t, f.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(valpubkey.Address())) == false)
}

func populateValidators(t assert.TestingT, f *fixture) {
	// add accounts and set total supply
	totalSupplyAmt := initAmt.MulRaw(int64(len(valAddresses)))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupplyAmt))
	assert.NilError(t, f.bankKeeper.MintCoins(f.sdkCtx, minttypes.ModuleName, totalSupply))

	for _, addr := range valAddresses {
		assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, minttypes.ModuleName, (sdk.AccAddress)(addr), initCoins))
	}
}

func newPubKey(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}

	pubkey := &ed25519.PubKey{Key: pkBytes}

	return pubkey
}

func testEquivocationHandler(_ interface{}) evidencetypes.Handler {
	return func(ctx context.Context, e exported.Evidence) error {
		if err := e.ValidateBasic(); err != nil {
			return err
		}

		ee, ok := e.(*evidencetypes.Equivocation)
		if !ok {
			return fmt.Errorf("unexpected evidence type: %T", e)
		}
		if ee.Height%2 == 0 {
			return fmt.Errorf("unexpected even evidence height: %d", ee.Height)
		}

		return nil
	}
}
