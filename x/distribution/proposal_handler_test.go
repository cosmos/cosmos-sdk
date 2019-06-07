package distribution

import (
	"testing"

	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/require"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
)

func testProposal(recipient sdk.AccAddress, amount sdk.Coins) types.CommunityPoolSpendProposal {
	return types.NewCommunityPoolSpendProposal(
		"Test",
		"description",
		recipient,
		amount,
	)
}

func TestProposalHandlerPassed(t *testing.T) {
	ctx, accountKeeper, keeper, _, _ := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1
	amount := sdk.NewCoin("stake", sdk.NewInt(1))

	account := accountKeeper.NewAccountWithAddress(ctx, recipient)
	require.True(t, account.GetCoins().IsZero())
	accountKeeper.SetAccount(ctx, account)

	feePool := keeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.DecCoins{sdk.NewDecCoinFromCoin(amount)}
	keeper.SetFeePool(ctx, feePool)

	tp := testProposal(recipient, sdk.NewCoins(amount))
	hdlr := NewCommunityPoolSpendProposalHandler(keeper)
	require.NoError(t, hdlr(ctx, tp))
	require.Equal(t, accountKeeper.GetAccount(ctx, recipient).GetCoins(), sdk.NewCoins(amount))
}

func TestProposalHandlerFailed(t *testing.T) {
	ctx, accountKeeper, keeper, _, _ := CreateTestInputDefault(t, false, 10)
	recipient := delAddr1
	amount := sdk.NewCoin("stake", sdk.NewInt(1))

	account := accountKeeper.NewAccountWithAddress(ctx, recipient)
	require.True(t, account.GetCoins().IsZero())
	accountKeeper.SetAccount(ctx, account)

	tp := testProposal(recipient, sdk.NewCoins(amount))
	hdlr := NewCommunityPoolSpendProposalHandler(keeper)
	require.Error(t, hdlr(ctx, tp))
	require.True(t, accountKeeper.GetAccount(ctx, recipient).GetCoins().IsZero())
}
