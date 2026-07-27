package keeper_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
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
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
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
		authtypes.StoreKey, banktypes.StoreKey, consensusparamtypes.StoreKey, evidencetypes.StoreKey, stakingtypes.StoreKey, slashingtypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, evidence.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{}, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		minttypes.ModuleName:                {authtypes.Minter},
		stakingtypes.BondedPoolName:         {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:      {authtypes.Burner, authtypes.Staking},
		stakingtypes.KeyRotationFeePoolName: {authtypes.Burner},
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

	// wire slashing hooks into staking so consensus key rotations migrate
	// signing info / missed-block state to the new key.
	stakingKeeper.SetHooks(stakingtypes.NewMultiStakingHooks(slashingKeeper.Hooks()))

	evidenceKeeper := keeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[evidencetypes.StoreKey]), stakingKeeper, slashingKeeper, addresscodec.NewBech32Codec("cosmos"), runtime.ProvideCometInfoService())
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

	assert.NilError(tb, slashingKeeper.SetParams(sdkCtx, testutil.TestParams()))

	// set default staking params
	assert.NilError(tb, stakingKeeper.SetParams(sdkCtx, stakingtypes.DefaultParams()))

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
	stakingParams, err := f.stakingKeeper.GetParams(ctx)
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
	assert.DeepEqual(t, selfDelegation, val.GetValidatorPower())

	assert.NilError(t, f.slashingKeeper.AddPubkey(f.sdkCtx, valpubkey))

	info := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(valpubkey.Address()), f.sdkCtx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	assert.NilError(t, f.slashingKeeper.SetValidatorSigningInfo(f.sdkCtx, sdk.ConsAddress(valpubkey.Address()), info))

	// handle a signature to set signing info
	assert.NilError(t, f.slashingKeeper.HandleValidatorSignature(ctx, valpubkey.Address(), selfDelegation.Int64(), comet.BlockIDFlagCommit))

	// double sign less than max age
	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	oldTokens := val.GetTokens()

	nci := NewCometInfo(abci.RequestFinalizeBlock{
		Misbehavior: []abci.Misbehavior{{
			Validator: abci.Validator{Address: valpubkey.Address(), Power: power},
			Type:      abci.MisbehaviorType_DUPLICATE_VOTE,
			Time:      time.Now().UTC(),
			Height:    1,
		}},
	})

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
	ctx = ctx.WithBlockTime(time.Unix(1, 0).Add(stakingParams.UnbondingTime))

	// require we cannot unjail
	assert.Error(t, f.slashingKeeper.Unjail(ctx, operatorAddr), slashingtypes.ErrValidatorJailed.Error())

	// require we be able to unbond now
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	del, _ := f.stakingKeeper.GetDelegation(ctx, sdk.AccAddress(operatorAddr), operatorAddr)
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

	ctx := f.sdkCtx.WithIsCheckTx(false).WithBlockHeight(1).WithBlockTime(time.Now())
	populateValidators(t, f)

	power := int64(100)
	stakingParams, err := f.stakingKeeper.GetParams(ctx)
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
	assert.DeepEqual(t, amt, val.GetValidatorPower())

	nci := NewCometInfo(abci.RequestFinalizeBlock{
		Misbehavior: []abci.Misbehavior{{
			Validator: abci.Validator{Address: valpubkey.Address(), Power: power},
			Type:      abci.MisbehaviorType_DUPLICATE_VOTE,
			Time:      ctx.BlockTime(),
			Height:    0,
		}},
	})

	assert.NilError(t, f.app.StoreConsensusParams(ctx, *simtestutil.DefaultConsensusParams))
	cp := f.app.GetConsensusParams(ctx)

	ctx = ctx.WithCometInfo(nci)
	ctx = ctx.WithConsensusParams(cp)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(cp.Evidence.MaxAgeDuration + 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + cp.Evidence.MaxAgeNumBlocks + 1)

	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx))

	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.IsJailed() == false)
	assert.Assert(t, f.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(valpubkey.Address())) == false)
}

// TestHandleDoubleSign_RotatedConsKey exercises the full historical-evidence
// path end to end: a bonded validator rotates its consensus key, and
// equivocation evidence that references the OLD (rotated-away) key is then
// submitted. The validator must still be slashed, jailed, and tombstoned under
// its CURRENT key.
func TestHandleDoubleSign_RotatedConsKey(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx.WithIsCheckTx(false).WithBlockHeight(1)
	populateValidators(t, f)

	power := int64(100)
	operatorAddr, oldPubKey := valAddresses[0], pubkeys[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)

	tstaking.CreateValidatorWithValPower(operatorAddr, oldPubKey, power, true)

	// bond the validator
	_, err := f.stakingKeeper.EndBlocker(f.sdkCtx)
	assert.NilError(t, err)

	oldConsAddr := sdk.ConsAddress(oldPubKey.Address())

	// establish signing info / signature for the original key
	assert.NilError(t, f.slashingKeeper.AddPubkey(f.sdkCtx, oldPubKey))
	info := slashingtypes.NewValidatorSigningInfo(oldConsAddr, f.sdkCtx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	assert.NilError(t, f.slashingKeeper.SetValidatorSigningInfo(f.sdkCtx, oldConsAddr, info))
	assert.NilError(t, f.slashingKeeper.HandleValidatorSignature(ctx, oldPubKey.Address(), power, comet.BlockIDFlagCommit))

	// rotate the consensus key oldPubKey -> newPubKey and apply it
	newPubKey := ed25519.GenPrivKey().PubKey()
	newConsAddr := sdk.ConsAddress(newPubKey.Address())
	assert.NilError(t, f.stakingKeeper.SetConsKeyRotation(f.sdkCtx, operatorAddr, oldPubKey, newPubKey))
	assert.NilError(t, f.stakingKeeper.ApplyConsKeyRotation(f.sdkCtx, operatorAddr, newPubKey))

	// signing info migrated to the new key via the slashing hook
	assert.Assert(t, !f.slashingKeeper.HasValidatorSigningInfo(f.sdkCtx, oldConsAddr))
	assert.Assert(t, f.slashingKeeper.HasValidatorSigningInfo(f.sdkCtx, newConsAddr))

	// the old key resolves to the validator via the historical lookup, and
	// canonicalizes to the current consensus address
	histVal, err := f.stakingKeeper.ValidatorByHistoricalConsAddr(f.sdkCtx, oldConsAddr)
	assert.NilError(t, err)
	histConsAddr, err := histVal.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, newConsAddr.Bytes(), histConsAddr)

	val, err := f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	oldTokens := val.GetTokens()

	// submit equivocation evidence that references the OLD (rotated-away) key
	nci := NewCometInfo(abci.RequestFinalizeBlock{
		Misbehavior: []abci.Misbehavior{{
			Validator: abci.Validator{Address: oldPubKey.Address(), Power: power},
			Type:      abci.MisbehaviorType_DUPLICATE_VOTE,
			Time:      time.Now().UTC(),
			Height:    1,
		}},
	})
	ctx = ctx.WithCometInfo(nci)
	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx))

	// the validator is jailed and tombstoned under its CURRENT key
	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.IsJailed())
	assert.Assert(t, f.slashingKeeper.IsTombstoned(ctx, newConsAddr))
	// the old key was never given its own signing info, so it is not tombstoned
	assert.Assert(t, !f.slashingKeeper.IsTombstoned(ctx, oldConsAddr))

	// tokens were slashed
	assert.Assert(t, val.GetTokens().LT(oldTokens))

	// the evidence was recorded
	iter, err := f.evidenceKeeper.Evidences.Iterate(ctx, nil)
	assert.NilError(t, err)
	values, err := iter.Values()
	assert.NilError(t, err)
	assert.Assert(t, len(values) == 1)
}

// TestHandleDoubleSign_RotatedConsKeyLingersForBlockWindow reproduces the
// slashing escape this fix closes: after a rotation, block time advances past
// the unbonding window while the block height stays inside the still-open
// evidence block window. Late equivocation for the old key must still slash
// and tombstone the fully-bonded validator, because the lock is retired on the
// evidence horizon, not unbonding.
func TestHandleDoubleSign_RotatedConsKeyLingersForBlockWindow(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	unbonding, err := f.stakingKeeper.UnbondingTime(f.sdkCtx)
	assert.NilError(t, err)
	rotationTime := f.sdkCtx.BlockHeader().Time

	// evidence stays admissible for a large block window that outlasts unbonding.
	const maxAgeNumBlocks = int64(1_000_000)
	evidenceParams := cmtproto.ConsensusParams{
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeDuration:  unbonding,
			MaxAgeNumBlocks: maxAgeNumBlocks,
		},
	}

	ctx := f.sdkCtx.WithIsCheckTx(false).WithBlockHeight(1).WithConsensusParams(evidenceParams)
	populateValidators(t, f)

	power := int64(100)
	operatorAddr, oldPubKey := valAddresses[0], pubkeys[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)
	tstaking.CreateValidatorWithValPower(operatorAddr, oldPubKey, power, true)

	_, err = f.stakingKeeper.EndBlocker(f.sdkCtx.WithConsensusParams(evidenceParams))
	assert.NilError(t, err)

	oldConsAddr := sdk.ConsAddress(oldPubKey.Address())
	assert.NilError(t, f.slashingKeeper.AddPubkey(f.sdkCtx, oldPubKey))
	info := slashingtypes.NewValidatorSigningInfo(oldConsAddr, ctx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	assert.NilError(t, f.slashingKeeper.SetValidatorSigningInfo(f.sdkCtx, oldConsAddr, info))
	assert.NilError(t, f.slashingKeeper.HandleValidatorSignature(ctx, oldPubKey.Address(), power, comet.BlockIDFlagCommit))

	// rotate under the evidence params so the evidence-lock horizon is computed.
	rotateCtx := f.sdkCtx.WithConsensusParams(evidenceParams)
	newPubKey := ed25519.GenPrivKey().PubKey()
	newConsAddr := sdk.ConsAddress(newPubKey.Address())
	assert.NilError(t, f.stakingKeeper.SetConsKeyRotation(rotateCtx, operatorAddr, oldPubKey, newPubKey))
	assert.NilError(t, f.stakingKeeper.ApplyConsKeyRotation(rotateCtx, operatorAddr, newPubKey))

	// advance time past the unbonding window (the old prune horizon) while
	// keeping the height inside the evidence block window, then prune.
	prunedCtx := f.sdkCtx.WithConsensusParams(evidenceParams).
		WithBlockTime(rotationTime.Add(unbonding + time.Hour)).
		WithBlockHeight(50)
	assert.NilError(t, f.stakingKeeper.PruneMaturedConsKeyRotations(prunedCtx))

	// the RotatedFrom lock survived: the historical lookup still resolves.
	histVal, err := f.stakingKeeper.ValidatorByHistoricalConsAddr(prunedCtx, oldConsAddr)
	assert.NilError(t, err)
	histConsAddr, err := histVal.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, newConsAddr.Bytes(), histConsAddr)

	val, err := f.stakingKeeper.Validator(prunedCtx, operatorAddr)
	assert.NilError(t, err)
	oldTokens := val.GetTokens()

	// equivocation for the old key: the time window is exceeded but the block
	// window is still open, so the evidence is admissible.
	evCtx := prunedCtx.WithCometInfo(NewCometInfo(abci.RequestFinalizeBlock{
		Misbehavior: []abci.Misbehavior{{
			Validator: abci.Validator{Address: oldPubKey.Address(), Power: power},
			Type:      abci.MisbehaviorType_DUPLICATE_VOTE,
			Time:      rotationTime,
			Height:    1,
		}},
	}))
	assert.NilError(t, f.evidenceKeeper.BeginBlocker(evCtx))

	// the still-bonded rotated validator is jailed, tombstoned, and slashed.
	val, err = f.stakingKeeper.Validator(evCtx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.IsJailed())
	assert.Assert(t, f.slashingKeeper.IsTombstoned(evCtx, newConsAddr))
	assert.Assert(t, val.GetTokens().LT(oldTokens))
}

// TestHandleDoubleSign_RotatedThenRemoved verifies that equivocation evidence
// referencing a rotated-away key whose validator has since been removed from
// the store is ignored rather than surfacing a consensus-level error from
// BeginBlocker.
func TestHandleDoubleSign_RotatedThenRemoved(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx.WithIsCheckTx(false).WithBlockHeight(1)
	populateValidators(t, f)

	power := int64(100)
	operatorAddr, oldPubKey := valAddresses[0], pubkeys[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)
	tstaking.CreateValidatorWithValPower(operatorAddr, oldPubKey, power, true)

	_, err := f.stakingKeeper.EndBlocker(f.sdkCtx)
	assert.NilError(t, err)

	oldConsAddr := sdk.ConsAddress(oldPubKey.Address())

	// rotate the consensus key and apply it, then remove the validator record
	// while keeping the RotatedFrom historical lock in place.
	newPubKey := ed25519.GenPrivKey().PubKey()
	assert.NilError(t, f.stakingKeeper.SetConsKeyRotation(f.sdkCtx, operatorAddr, oldPubKey, newPubKey))
	assert.NilError(t, f.stakingKeeper.ApplyConsKeyRotation(f.sdkCtx, operatorAddr, newPubKey))

	// drain the validator and remove its record, leaving only the RotatedFrom
	// historical lock. ValidatorByHistoricalConsAddr(oldConsAddr) now resolves
	// the lock but its GetValidator lookup fails with ErrNoValidatorFound.
	rotated, err := f.stakingKeeper.GetValidator(f.sdkCtx, operatorAddr)
	assert.NilError(t, err)
	rotated.Status = stakingtypes.Unbonded
	rotated.Tokens = math.ZeroInt()
	rotated.DelegatorShares = math.LegacyZeroDec()
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, rotated))
	assert.NilError(t, f.stakingKeeper.RemoveValidator(f.sdkCtx, operatorAddr))

	_, err = f.stakingKeeper.ValidatorByHistoricalConsAddr(f.sdkCtx, oldConsAddr)
	assert.ErrorContains(t, err, stakingtypes.ErrNoValidatorFound.Error())

	// evidence referencing the old key must be ignored without error
	nci := NewCometInfo(abci.RequestFinalizeBlock{
		Misbehavior: []abci.Misbehavior{{
			Validator: abci.Validator{Address: oldPubKey.Address(), Power: power},
			Type:      abci.MisbehaviorType_DUPLICATE_VOTE,
			Time:      time.Now().UTC(),
			Height:    1,
		}},
	})
	ctx = ctx.WithCometInfo(nci)
	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx))

	// no tombstone, no evidence recorded
	assert.Assert(t, !f.slashingKeeper.IsTombstoned(ctx, oldConsAddr))
	iter, err := f.evidenceKeeper.Evidences.Iterate(ctx, nil)
	assert.NilError(t, err)
	values, err := iter.Values()
	assert.NilError(t, err)
	assert.Assert(t, len(values) == 0)
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

type CometService struct {
	Evidence []abci.Misbehavior
}

func NewCometInfo(bg abci.RequestFinalizeBlock) comet.BlockInfo {
	return CometService{
		Evidence: bg.Misbehavior,
	}
}

func (r CometService) GetEvidence() comet.EvidenceList {
	return evidenceWrapper{evidence: r.Evidence}
}

func (CometService) GetValidatorsHash() []byte {
	return []byte{}
}

func (CometService) GetProposerAddress() []byte {
	return []byte{}
}

func (CometService) GetLastCommit() comet.CommitInfo {
	return nil
}

type evidenceWrapper struct {
	evidence []abci.Misbehavior
}

func (e evidenceWrapper) Len() int {
	return len(e.evidence)
}

func (e evidenceWrapper) Get(i int) comet.Evidence {
	return misbehaviorWrapper{e.evidence[i]}
}

type misbehaviorWrapper struct {
	abci.Misbehavior
}

func (m misbehaviorWrapper) Type() comet.MisbehaviorType {
	return comet.MisbehaviorType(m.Misbehavior.Type)
}

func (m misbehaviorWrapper) Height() int64 {
	return m.Misbehavior.Height
}

func (m misbehaviorWrapper) Validator() comet.Validator {
	return validatorWrapper{m.Misbehavior.Validator}
}

func (m misbehaviorWrapper) Time() time.Time {
	return m.Misbehavior.Time
}

func (m misbehaviorWrapper) TotalVotingPower() int64 {
	return m.Misbehavior.TotalVotingPower
}

// validatorWrapper is a wrapper around abci.Validator that implements Validator interface
type validatorWrapper struct {
	abci.Validator
}

var _ comet.Validator = (*validatorWrapper)(nil)

func (v validatorWrapper) Address() []byte {
	return v.Validator.Address
}

func (v validatorWrapper) Power() int64 {
	return v.Validator.Power
}
