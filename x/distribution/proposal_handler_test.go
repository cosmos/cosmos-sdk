package distribution_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())

	amount = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
)

func testProposal(recipient sdk.AccAddress, amount sdk.Coins) *types.CommunityPoolSpendProposal {
	return types.NewCommunityPoolSpendProposal("Test", "description", recipient, amount)
}

func TestProposalHandlerPassed(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	recipient := delAddr1

	// add coins to the module account
	macc := distrKeeper.GetDistributionAccount(ctx)
	balances := bankKeeper.GetAllBalances(ctx, macc.GetAddress())
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, macc.GetName(), amount))

	accountKeeper.SetModuleAccount(ctx, macc)

	account := accountKeeper.NewAccountWithAddress(ctx, recipient)
	accountKeeper.SetAccount(ctx, account)
	require.True(t, bankKeeper.GetAllBalances(ctx, account.GetAddress()).IsZero())

	feePool := distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(amount...)
	distrKeeper.SetFeePool(ctx, feePool)

	tp := testProposal(recipient, amount)
	hdlr := distribution.NewCommunityPoolSpendProposalHandler(distrKeeper)
	require.NoError(t, hdlr(ctx, tp))

	balances = bankKeeper.GetAllBalances(ctx, recipient)
	require.Equal(t, balances, amount)
}

func TestProposalHandlerFailed(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// reset fee pool
	distrKeeper.SetFeePool(ctx, types.InitialFeePool())

	recipient := delAddr1

	account := accountKeeper.NewAccountWithAddress(ctx, recipient)
	accountKeeper.SetAccount(ctx, account)
	require.True(t, bankKeeper.GetAllBalances(ctx, account.GetAddress()).IsZero())

	tp := testProposal(recipient, amount)
	hdlr := distribution.NewCommunityPoolSpendProposalHandler(distrKeeper)
	require.Error(t, hdlr(ctx, tp))

	balances := bankKeeper.GetAllBalances(ctx, recipient)
	require.True(t, balances.IsZero())
}
