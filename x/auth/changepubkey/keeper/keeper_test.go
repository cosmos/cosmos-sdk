package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context

	queryClient types.QueryClient
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{Time: time.Now()})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(true)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.AccountHistoryKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func TestPubKeyHistory_StoreGet(t *testing.T) {
	app, ctx := createTestApp(true)
	cdc := app.AppCodec()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))

	currentTime := ctx.BlockTime()
	// create accounts
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc)

	// store pubkey when pubkey does not exist
	app.AccountHistoryKeeper.StoreLastPubKey(ctx, addr1, currentTime.Truncate(time.Hour), acc.GetPubKey())

	_, pub1, _ := testdata.KeyTestPubAddr()
	acc.SetPubKey(pub1)
	app.AccountKeeper.SetAccount(ctx, acc)

	// Note: store pubkey when pubkey does exist, it will overwrite previous one
	// since previous one's endtime = current one's endtime
	app.AccountHistoryKeeper.StoreLastPubKey(ctx, addr1, currentTime.Truncate(time.Hour), acc.GetPubKey())

	// store pubkey after pubkey is set
	_, pub2, _ := testdata.KeyTestPubAddr()
	acc.SetPubKey(pub2)
	app.AccountKeeper.SetAccount(ctx, acc)

	app.AccountHistoryKeeper.StoreLastPubKey(ctx, addr1, currentTime.Truncate(time.Minute), acc.GetPubKey())
	_, pub3, _ := testdata.KeyTestPubAddr()
	acc.SetPubKey(pub3)
	app.AccountKeeper.SetAccount(ctx, acc)

	// try iteration after changing pubkey 3 times
	history, err := app.AccountHistoryKeeper.GetPubKeyHistory(ctx, addr1)
	require.NoError(t, err)
	require.Equal(t, len(history), 3)

	require.Equal(t, history[0].StartTime, time.Time{})
	require.Equal(t, history[0].EndTime, currentTime.Truncate(time.Hour))
	require.Equal(t, types.DecodePubKey(cdc, history[0].GetPubKey()).Address(), pub1.Address())

	require.Equal(t, history[1].StartTime, currentTime.Truncate(time.Hour))
	require.Equal(t, history[1].EndTime, currentTime.Truncate(time.Minute))
	require.Equal(t, types.DecodePubKey(cdc, history[1].GetPubKey()).Address(), pub2.Address())

	require.Equal(t, history[2].StartTime, currentTime.Truncate(time.Minute))
	require.Equal(t, history[2].EndTime, currentTime)
	require.Equal(t, types.DecodePubKey(cdc, history[2].GetPubKey()).Address(), pub3.Address())

	history1 := app.AccountHistoryKeeper.GetPubKeyHistoricalEntry(ctx, addr1, currentTime.Truncate(2*time.Hour))
	require.Equal(t, history1, history[0])
	history2 := app.AccountHistoryKeeper.GetPubKeyHistoricalEntry(ctx, addr1, currentTime.Truncate(time.Hour))
	require.Equal(t, history2, history[0])
	history3 := app.AccountHistoryKeeper.GetPubKeyHistoricalEntry(ctx, addr1, currentTime.Truncate(2*time.Minute))
	require.Equal(t, history3, history[1])
	history4 := app.AccountHistoryKeeper.GetPubKeyHistoricalEntry(ctx, addr1, currentTime.Truncate(time.Minute))
	require.Equal(t, history4, history[1])
	history5 := app.AccountHistoryKeeper.GetPubKeyHistoricalEntry(ctx, addr1, currentTime.Truncate(time.Second))
	require.Equal(t, history5, history[2])
	history6 := app.AccountHistoryKeeper.GetPubKeyHistoricalEntry(ctx, addr1, currentTime)
	require.Equal(t, history6, history[2])
	history7 := app.AccountHistoryKeeper.GetLastPubKeyHistoricalEntry(ctx, addr1)
	require.Equal(t, history7, history[1])
	history8 := app.AccountHistoryKeeper.GetCurrentPubKeyEntry(ctx, addr1)
	require.Equal(t, history8, history[2])
}
