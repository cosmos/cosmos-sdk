package keeper

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

type TestSuite struct {
	ctx         sdk.Context
	authKeeper  authkeeper.AccountKeeper
	bankKeeper  keeper.BaseKeeper
	queryClient banktypes.QueryClient
	msgServer   banktypes.MsgServer
}

func initDeterministicFixture(t *testing.T) *TestSuite {
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, vesting.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	header := cmtproto.Header{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	newCtx := sdk.NewContext(cms, header, false, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		minttypes.ModuleName: {authtypes.Minter},
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
	bankKeeper := keeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	vestingModule := vesting.NewAppModule(accountKeeper, bankKeeper)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:    authModule,
		banktypes.ModuleName:    bankModule,
		vestingtypes.ModuleName: vestingModule,
	})

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	bankSrv := keeper.NewMsgServerImpl(bankKeeper)
	banktypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), bankSrv)
	banktypes.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewQuerier(&bankKeeper))

	qr := integrationApp.QueryHelper()
	queryClient := banktypes.NewQueryClient(qr)

	f := TestSuite{
		ctx:         sdkCtx,
		authKeeper:  accountKeeper,
		bankKeeper:  bankKeeper,
		queryClient: queryClient,
		msgServer:   bankSrv,
	}

	err := banktestutil.FundModuleAccount(f.ctx, f.bankKeeper, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000000000))))
	require.NoError(t, err)

	return &f
}

func fundAccount(t *testing.T, f *TestSuite, addr sdk.AccAddress, coins sdk.Coins) {
	err := banktestutil.FundAccount(f.ctx, f.bankKeeper, addr, coins)
	require.NoError(t, err)
}

func TestIntegration_ContinuousVesting_UpdateScheduleAndSend(t *testing.T) {
	f := initDeterministicFixture(t)
	ctx := f.ctx
	require.NotNil(t, f.msgServer)

	stakeDenom := "stake"
	initialAmount := math.NewInt(1000)
	vestingDuration := time.Hour * 24
	startTime := ctx.BlockTime()
	endTime := startTime.Add(vestingDuration)

	updateTime := startTime.Add(vestingDuration / 4)
	checkTime := startTime.Add(vestingDuration / 2)

	senderAddr := sdk.AccAddress("sender_______________")
	recipientAddr := sdk.AccAddress("recipient____________")

	baseAcc := authtypes.NewBaseAccount(senderAddr, nil, 100, 4)
	initialVestingCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, initialAmount))

	cva, err := vestingtypes.NewContinuousVestingAccount(baseAcc, initialVestingCoins, startTime.Unix(), endTime.Unix())
	require.NoError(t, err)

	f.authKeeper.SetAccount(ctx, cva)
	fundAccount(t, f, senderAddr, initialVestingCoins)

	rewardCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(500)))
	accBeforeUpdate := f.authKeeper.GetAccount(ctx, senderAddr)
	cvaBeforeUpdate, ok := accBeforeUpdate.(*vestingtypes.ContinuousVestingAccount)
	require.True(t, ok, "Account should be a vesting account")
	err = cvaBeforeUpdate.UpdateSchedule(updateTime, rewardCoins)
	require.NoError(t, err)
	fundAccount(t, f, senderAddr, sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(500))))

	f.authKeeper.SetAccount(ctx, cvaBeforeUpdate)

	accAfterUpdate := f.authKeeper.GetAccount(ctx, senderAddr)
	cvaAfterUpdate, ok := accAfterUpdate.(*vestingtypes.ContinuousVestingAccount)
	require.True(t, ok)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(1500))), cvaAfterUpdate.GetOriginalVesting(), "OriginalVesting in store mismatch after update")
	ctx = ctx.WithBlockTime(checkTime)

	spendableCoins := f.bankKeeper.SpendableCoins(ctx, senderAddr)
	spendableAmount := spendableCoins.AmountOf(stakeDenom)

	require.True(t, spendableAmount.Equal(math.NewInt(750)), "Spendable amount mismatch. Expected: %s, Got: %s", math.NewInt(750), spendableAmount)

	amountToSendFail := math.NewInt(751)
	msgSendFail := banktypes.NewMsgSend(senderAddr, recipientAddr, sdk.NewCoins(sdk.NewCoin(stakeDenom, amountToSendFail)))

	_, err = f.msgServer.Send(ctx, msgSendFail)
	require.Error(t, err, "Expected error when sending more than spendable")
	require.ErrorContains(t, err, "insufficient funds", "Expected insufficient funds error")

	senderBalFail := f.bankKeeper.GetBalance(ctx, senderAddr, stakeDenom)
	require.True(t, senderBalFail.Amount.Equal(math.NewInt(1500)), "Sender balance should not change after failed send")
	recipientBalFail := f.bankKeeper.GetBalance(ctx, recipientAddr, stakeDenom)
	require.True(t, recipientBalFail.Amount.IsZero(), "Recipient balance should be zero after failed send")

	amountToSendOK := math.NewInt(375)
	require.True(t, amountToSendOK.IsPositive(), "Amount to send OK must be positive")
	msgSendOK := banktypes.NewMsgSend(senderAddr, recipientAddr, sdk.NewCoins(sdk.NewCoin(stakeDenom, amountToSendOK)))

	_, err = f.msgServer.Send(ctx, msgSendOK)
	require.NoError(t, err, "Expected no error when sending spendable amount")

	senderBalOK := f.bankKeeper.GetBalance(ctx, senderAddr, stakeDenom)
	require.True(t, senderBalOK.Amount.Equal(math.NewInt(1125)), "Sender balance mismatch after successful send. Expected: %s, Got: %s", math.NewInt(1125), senderBalOK.Amount)

	recipientBalOK := f.bankKeeper.GetBalance(ctx, recipientAddr, stakeDenom)
	require.True(t, recipientBalOK.Amount.Equal(math.NewInt(375)), "Recipient balance mismatch after successful send. Expected: %s, Got: %s", math.NewInt(375), recipientBalOK.Amount)

	currentAcc := f.authKeeper.GetAccount(ctx, senderAddr)
	currentCva, ok := currentAcc.(*vestingtypes.ContinuousVestingAccount)
	require.True(t, ok)
	lockedCoins := currentCva.LockedCoins(checkTime)
	require.True(t, lockedCoins.AmountOf(stakeDenom).Equal(math.NewInt(750)), "Locked coins mismatch after send. Expected: %s, Got: %s", math.NewInt(750), lockedCoins.AmountOf(stakeDenom))
}
