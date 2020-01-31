package distribution

import (
	"testing"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())

	amount = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
)

func testProposal(recipient sdk.AccAddress, amount sdk.Coins) types.CommunityPoolSpendProposal {
	return types.NewCommunityPoolSpendProposal("Test", "description", recipient, amount)
}

func TestProposalHandlerPassed(t *testing.T) {
	ctx, ak, bk, keeper, _, supplyKeeper := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	// add coins to the module account
	macc := keeper.GetDistributionAccount(ctx)
	balances := bk.GetAllBalances(ctx, macc.GetAddress())
	err := bk.SetBalances(ctx, macc.GetAddress(), balances.Add(amount...))
	require.NoError(t, err)

	supplyKeeper.SetModuleAccount(ctx, macc)

	account := ak.NewAccountWithAddress(ctx, recipient)
	ak.SetAccount(ctx, account)
	require.True(t, bk.GetAllBalances(ctx, account.GetAddress()).IsZero())

	feePool := keeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(amount...)
	keeper.SetFeePool(ctx, feePool)

	tp := testProposal(recipient, amount)
	hdlr := NewCommunityPoolSpendProposalHandler(keeper)
	require.NoError(t, hdlr(ctx, tp))

	balances = bk.GetAllBalances(ctx, recipient)
	require.Equal(t, balances, amount)
}

func TestProposalHandlerFailed(t *testing.T) {
	ctx, ak, bk, keeper, _, _ := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1

	account := ak.NewAccountWithAddress(ctx, recipient)
	ak.SetAccount(ctx, account)
	require.True(t, bk.GetAllBalances(ctx, account.GetAddress()).IsZero())

	tp := testProposal(recipient, amount)
	hdlr := NewCommunityPoolSpendProposalHandler(keeper)
	require.Error(t, hdlr(ctx, tp))

	balances := bk.GetAllBalances(ctx, recipient)
	require.True(t, balances.IsZero())
}
