package keeper_test

import (
	"fmt"
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var (
	_, _, addr   = testdata.KeyTestPubAddr()
	govAcct      = authtypes.NewModuleAddress(types.ModuleName)
	distAcct     = authtypes.NewModuleAddress(disttypes.ModuleName)
	TestProposal = getTestProposal()
)

// getTestProposal creates and returns a test proposal message.
func getTestProposal() []sdk.Msg {
	legacyProposalMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		panic(err)
	}

	return []sdk.Msg{
		banktypes.NewMsgSend(govAcct, addr, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000)))),
		legacyProposalMsg,
	}
}

// setupGovKeeper creates a govKeeper as well as all its dependencies.
func setupGovKeeper(t *testing.T) (
	*keeper.Keeper,
	*govtestutil.MockAccountKeeper,
	*govtestutil.MockBankKeeper,
	*govtestutil.MockStakingKeeper,
	*govtestutil.MockDistributionKeeper,
	moduletestutil.TestEncodingConfig,
	sdk.Context,
) {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()
	v1.RegisterInterfaces(encCfg.InterfaceRegistry)
	v1beta1.RegisterInterfaces(encCfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	// Create MsgServiceRouter, but don't populate it before creating the gov
	// keeper.
	msr := baseapp.NewMsgServiceRouter()

	// gomock initializations
	ctrl := gomock.NewController(t)
	acctKeeper := govtestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := govtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := govtestutil.NewMockStakingKeeper(ctrl)
	distributionKeeper := govtestutil.NewMockDistributionKeeper(ctrl)

	acctKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(govAcct).AnyTimes()
	acctKeeper.EXPECT().GetModuleAddress(disttypes.ModuleName).Return(distAcct).AnyTimes()
	acctKeeper.EXPECT().GetModuleAccount(gomock.Any(), types.ModuleName).Return(authtypes.NewEmptyModuleAccount(types.ModuleName)).AnyTimes()
	acctKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	trackMockBalances(bankKeeper, distributionKeeper)
	stakingKeeper.EXPECT().TokensFromConsensusPower(ctx, gomock.Any()).DoAndReturn(func(ctx sdk.Context, power int64) math.Int {
		return sdk.TokensFromConsensusPower(power, math.NewIntFromUint64(1000000))
	}).AnyTimes()

	stakingKeeper.EXPECT().BondDenom(ctx).Return("stake", nil).AnyTimes()
	stakingKeeper.EXPECT().IterateBondedValidatorsByPower(gomock.Any(), gomock.Any()).AnyTimes()
	stakingKeeper.EXPECT().IterateDelegations(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stakingKeeper.EXPECT().TotalBondedTokens(gomock.Any()).Return(math.NewInt(10000000), nil).AnyTimes()
	distributionKeeper.EXPECT().FundCommunityPool(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Gov keeper initializations

	govKeeper := keeper.NewKeeper(encCfg.Codec, storeService, acctKeeper, bankKeeper, stakingKeeper, distributionKeeper, msr, types.DefaultConfig(), govAcct.String())
	require.NoError(t, govKeeper.ProposalID.Set(ctx, 1))
	govRouter := v1beta1.NewRouter() // Also register legacy gov handlers to test them too.
	govRouter.AddRoute(types.RouterKey, v1beta1.ProposalHandler)
	govKeeper.SetLegacyRouter(govRouter)
	err := govKeeper.Params.Set(ctx, v1.DefaultParams())
	require.NoError(t, err)
	err = govKeeper.Constitution.Set(ctx, "constitution")
	require.NoError(t, err)

	// Register all handlers for the MegServiceRouter.
	msr.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	v1.RegisterMsgServer(msr, keeper.NewMsgServerImpl(govKeeper))
	banktypes.RegisterMsgServer(msr, nil) // Nil is fine here as long as we never execute the proposal's Msgs.

	return govKeeper, acctKeeper, bankKeeper, stakingKeeper, distributionKeeper, encCfg, ctx
}

// trackMockBalances sets up expected calls on the Mock BankKeeper, and also
// locally tracks accounts balances (not modules balances).
func trackMockBalances(bankKeeper *govtestutil.MockBankKeeper, distributionKeeper *govtestutil.MockDistributionKeeper) {
	balances := make(map[string]sdk.Coins)
	balances[distAcct.String()] = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)))

	// We don't track module account balances.
	bankKeeper.EXPECT().MintCoins(gomock.Any(), minttypes.ModuleName, gomock.Any()).AnyTimes()
	bankKeeper.EXPECT().BurnCoins(gomock.Any(), types.ModuleName, gomock.Any()).AnyTimes()
	bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), minttypes.ModuleName, types.ModuleName, gomock.Any()).AnyTimes()

	// But we do track normal account balances.
	bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), types.ModuleName, gomock.Any()).DoAndReturn(func(_ sdk.Context, sender sdk.AccAddress, _ string, coins sdk.Coins) error {
		newBalance, negative := balances[sender.String()].SafeSub(coins...)
		if negative {
			return fmt.Errorf("not enough balance")
		}
		balances[sender.String()] = newBalance
		return nil
	}).AnyTimes()
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ sdk.Context, module string, rcpt sdk.AccAddress, coins sdk.Coins) error {
		balances[rcpt.String()] = balances[rcpt.String()].Add(coins...)
		return nil
	}).AnyTimes()
	bankKeeper.EXPECT().GetAllBalances(gomock.Any(), gomock.Any()).DoAndReturn(func(_ sdk.Context, addr sdk.AccAddress) sdk.Coins {
		return balances[addr.String()]
	}).AnyTimes()
	bankKeeper.EXPECT().GetBalance(gomock.Any(), gomock.Any(), sdk.DefaultBondDenom).DoAndReturn(func(_ sdk.Context, addr sdk.AccAddress, _ string) sdk.Coin {
		balances := balances[addr.String()]
		for _, balance := range balances {
			if balance.Denom == sdk.DefaultBondDenom {
				return balance
			}
		}
		return sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0))
	}).AnyTimes()

	distributionKeeper.EXPECT().FundCommunityPool(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ sdk.Context, coins sdk.Coins, sender sdk.AccAddress) error {
		// sender balance
		newBalance, negative := balances[sender.String()].SafeSub(coins...)
		if negative {
			return fmt.Errorf("not enough balance")
		}
		balances[sender.String()] = newBalance
		// receiver balance
		balances[distAcct.String()] = balances[distAcct.String()].Add(coins...)
		return nil
	}).AnyTimes()
}
