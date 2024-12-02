package evidence_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/evidence/exported"
	"cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	minttypes "cosmossdk.io/x/mint/types"
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	slashingtypes "cosmossdk.io/x/slashing/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtestutil "cosmossdk.io/x/staking/testutil"
	stakingtypes "cosmossdk.io/x/staking/types"

	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/runtime/v2/services"
	_ "cosmossdk.io/x/accounts"  // import as blank for app wiring
	_ "cosmossdk.io/x/bank"      // import as blank for app wiring
	_ "cosmossdk.io/x/consensus" // import as blank for app wiring
	_ "cosmossdk.io/x/evidence"  // import as blank for app wiring
	_ "cosmossdk.io/x/slashing"  // import as blank for app wiring
	_ "cosmossdk.io/x/staking"   // import as blank for app wiring
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
	"github.com/stretchr/testify/require"
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
	initAmt          = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins        = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))
	cometInfoService = &services.ContextAwareCometInfoService{}
)

type fixture struct {
	app *integration.App

	ctx context.Context
	cdc codec.Codec

	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	evidenceKeeper keeper.Keeper
	slashingKeeper slashingkeeper.Keeper
	stakingKeeper  *stakingkeeper.Keeper
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	res := fixture{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.StakingModule(),
		configurator.SlashingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.EvidenceModule(),
		configurator.GenutilModule(),
	}

	var err error
	startupCfg := integration.DefaultStartUpConfig(t)

	msgRouterService := integration.NewRouterService()
	res.registerMsgRouterService(msgRouterService)

	var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
		return msgRouterService
	}

	queryRouterService := integration.NewRouterService()
	res.registerQueryRouterService(queryRouterService)

	serviceBuilder := runtime.NewRouterBuilder(routerFactory, queryRouterService)

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = &integration.HeaderService{}

	res.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.bankKeeper, &res.accountKeeper, &res.stakingKeeper, &res.slashingKeeper, &res.evidenceKeeper, &res.cdc)
	require.NoError(t, err)

	router := evidencetypes.NewRouter()
	router = router.AddRoute(evidencetypes.RouteEquivocation, testEquivocationHandler(res.evidenceKeeper))
	res.evidenceKeeper.SetRouter(router)

	res.ctx = res.app.StateLatestContext(t)

	return &res
}

func (s *fixture) registerMsgRouterService(router *integration.RouterService) {
	// register custom router service
	bankSendHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*banktypes.MsgSend)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := bankkeeper.NewMsgServerImpl(s.bankKeeper)
		resp, err := msgServer.Send(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(bankSendHandler, "cosmos.bank.v1beta1.MsgSend")
}

func (s *fixture) registerQueryRouterService(router *integration.RouterService) {
	// register custom router service
}

func TestHandleDoubleSign(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.ctx
	populateValidators(t, f)

	power := int64(100)
	stakingParams, err := f.stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	operatorAddr, valpubkey := valAddresses[0], pubkeys[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)
	f.accountKeeper.SetAccount(f.ctx, f.accountKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(operatorAddr)))
	selfDelegation := tstaking.CreateValidatorWithValPower(operatorAddr, valpubkey, power, true)

	// execute end-blocker and verify validator attributes
	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t,
		f.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)).String(),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(selfDelegation))).String(),
	)
	val, err := f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.DeepEqual(t, selfDelegation, val.GetBondedTokens())

	assert.NilError(t, f.slashingKeeper.AddrPubkeyRelation.Set(f.ctx, valpubkey.Address(), valpubkey))

	consaddrStr, err := f.stakingKeeper.ConsensusAddressCodec().BytesToString(valpubkey.Address())
	assert.NilError(t, err)
	height := f.app.LastBlockHeight() + 1
	info := slashingtypes.NewValidatorSigningInfo(consaddrStr, int64(height), time.Unix(0, 0), false, int64(0))
	err = f.slashingKeeper.ValidatorSigningInfo.Set(f.ctx, sdk.ConsAddress(valpubkey.Address()), info)
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

	ctx = context.WithValue(ctx, corecontext.CometInfoKey, nci)
	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx, cometInfoService))

	// should be jailed and tombstoned
	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.IsJailed())
	assert.Assert(t, f.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(valpubkey.Address())))

	// tokens should be decreased
	newTokens := val.GetTokens()
	assert.Assert(t, newTokens.LT(oldTokens))

	// submit duplicate evidence
	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx, cometInfoService))

	// tokens should be the same (capped slash)
	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.GetTokens().Equal(newTokens))

	// jump to past the unbonding period
	ctx = integration.SetHeaderInfo(ctx, header.Info{Time: time.Unix(1, 0).Add(stakingParams.UnbondingTime)})

	// require we cannot unjail
	assert.Error(t, f.slashingKeeper.Unjail(ctx, operatorAddr), slashingtypes.ErrValidatorJailed.Error())

	// require we be able to unbond now
	del, _ := f.stakingKeeper.Delegations.Get(ctx, collections.Join(sdk.AccAddress(operatorAddr), operatorAddr))
	validator, _ := f.stakingKeeper.GetValidator(ctx, operatorAddr)
	totalBond := validator.TokensFromShares(del.GetShares()).TruncateInt()
	tstaking.Ctx = ctx
	tstaking.Denom = stakingParams.BondDenom
	accAddr, err := f.accountKeeper.AddressCodec().BytesToString(operatorAddr)
	assert.NilError(t, err)
	opAddr, err := f.stakingKeeper.ValidatorAddressCodec().BytesToString(operatorAddr)
	assert.NilError(t, err)
	tstaking.Undelegate(accAddr, opAddr, totalBond, true)

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

	ctx := integration.SetHeaderInfo(f.ctx, header.Info{Height: 1, Time: time.Now()})
	populateValidators(t, f)

	power := int64(100)
	stakingParams, err := f.stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	operatorAddr, valpubkey := valAddresses[0], pubkeys[0]

	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)
	f.accountKeeper.SetAccount(f.ctx, f.accountKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(operatorAddr)))
	amt := tstaking.CreateValidatorWithValPower(operatorAddr, valpubkey, power, true)

	// execute end-blocker and verify validator attributes
	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t,
		f.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(amt))),
	)
	val, err := f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.DeepEqual(t, amt, val.GetBondedTokens())

	// nci := comet.Info{Evidence: []comet.Evidence{{
	// 	Validator: comet.Validator{Address: valpubkey.Address(), Power: power},
	// 	Type:      comet.DuplicateVote, //
	// 	Time:      ctx.HeaderInfo().Time,
	// 	Height:    0,
	// }}}

	// cp := simtestutil.DefaultConsensusParams

	// ctx = ctx.WithCometInfo(nci)
	// ctx = ctx.WithConsensusParams(*simtestutil.DefaultConsensusParams)
	// ctx = ctx.WithHeaderInfo(header.Info{Height: ctx.BlockHeight() + cp.Evidence.MaxAgeNumBlocks + 1, Time: ctx.HeaderInfo().Time.Add(cp.Evidence.MaxAgeDuration + 1)})

	assert.NilError(t, f.evidenceKeeper.BeginBlocker(ctx, cometInfoService))

	val, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, val.IsJailed() == false)
	assert.Assert(t, f.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(valpubkey.Address())) == false)
}

func TestHandleDoubleSignAfterRotation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	ctx := sdkCtx.WithIsCheckTx(false).WithBlockHeight(1).WithHeaderInfo(header.Info{Time: time.Now()})
	populateValidators(t, f)

	power := int64(100)
	stakingParams, err := f.stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)

	operatorAddr, valpubkey := valAddresses[0], pubkeys[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)
	f.accountKeeper.SetAccount(f.ctx, f.accountKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(operatorAddr)))
	selfDelegation := tstaking.CreateValidatorWithValPower(operatorAddr, valpubkey, power, true)

	// execute end-blocker and verify validator attributes
	_, err = f.stakingKeeper.EndBlocker(ctx)
	assert.NilError(t, err)

	assert.DeepEqual(t,
		f.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)).String(),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(selfDelegation))).String(),
	)

	valInfo, err := f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	consAddrBeforeRotn, err := valInfo.GetConsAddr()

	assert.NilError(t, err)
	assert.DeepEqual(t, selfDelegation, valInfo.GetBondedTokens())

	NewConsPubkey := newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53")

	msgServer := stakingkeeper.NewMsgServerImpl(f.stakingKeeper)
	msg, err := stakingtypes.NewMsgRotateConsPubKey(operatorAddr.String(), NewConsPubkey)
	assert.NilError(t, err)
	_, err = msgServer.RotateConsPubKey(ctx, msg)
	assert.NilError(t, err)

	// execute end-blocker and verify validator attributes
	_, err = f.stakingKeeper.EndBlocker(ctx)
	assert.NilError(t, err)

	valInfo, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	consAddrAfterRotn, err := valInfo.GetConsAddr()
	assert.NilError(t, err)
	assert.Equal(t, bytes.Equal(consAddrBeforeRotn, consAddrAfterRotn), false)

	// handle a signature to set signing info
	err = f.slashingKeeper.HandleValidatorSignature(ctx, NewConsPubkey.Address().Bytes(), selfDelegation.Int64(), comet.BlockIDFlagCommit)
	assert.NilError(t, err)

	// double sign less than max age
	valInfo, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	oldTokens := valInfo.GetTokens()
	nci := comet.Info{
		Evidence: []comet.Evidence{{
			Validator: comet.Validator{Address: valpubkey.Address(), Power: power},
			Type:      comet.DuplicateVote,
			Time:      time.Unix(0, 0),
			Height:    0,
		}},
	}

	err = f.evidenceKeeper.BeginBlocker(ctx.WithCometInfo(nci), cometInfoService)
	assert.NilError(t, err)

	// should be jailed and tombstoned
	valInfo, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, valInfo.IsJailed())
	assert.Assert(t, f.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(NewConsPubkey.Address())))

	// tokens should be decreased
	valInfo, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	newTokens := valInfo.GetTokens()
	assert.Assert(t, newTokens.LT(oldTokens))

	// submit duplicate evidence
	err = f.evidenceKeeper.BeginBlocker(ctx.WithCometInfo(nci), cometInfoService)
	assert.NilError(t, err)

	// tokens should be the same (capped slash)
	valInfo, err = f.stakingKeeper.Validator(ctx, operatorAddr)
	assert.NilError(t, err)
	assert.Assert(t, valInfo.GetTokens().Equal(newTokens))

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
	accAddr, err := f.accountKeeper.AddressCodec().BytesToString(operatorAddr)
	assert.NilError(t, err)
	opAddr, err := f.stakingKeeper.ValidatorAddressCodec().BytesToString(operatorAddr)
	assert.NilError(t, err)
	tstaking.Undelegate(accAddr, opAddr, totalBond, true)

	// query evidence from store
	var evidences []exported.Evidence
	assert.NilError(t, f.evidenceKeeper.Evidences.Walk(ctx, nil, func(
		key []byte,
		value exported.Evidence,
	) (stop bool, err error) {
		evidences = append(evidences, value)
		return false, nil
	}))
	// evidences, err := f.evidenceKeeper.GetAllEvidence(ctx)
	assert.NilError(t, err)
	assert.Assert(t, len(evidences) == 1)
}

func populateValidators(t assert.TestingT, f *fixture) {
	// add accounts and set total supply
	totalSupplyAmt := initAmt.MulRaw(int64(len(valAddresses)))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupplyAmt))
	assert.NilError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, totalSupply))

	for _, addr := range valAddresses {
		assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, minttypes.ModuleName, (sdk.AccAddress)(addr), initCoins))
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
