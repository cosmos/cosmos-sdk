package distribution_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	distrAcc = authtypes.NewEmptyModuleAccount(types.ModuleName)

	amount     = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
	zeroAmount = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))
)

func testProposal(recipient sdk.AccAddress, amount sdk.Coins) *types.CommunityPoolSpendProposal {
	return types.NewCommunityPoolSpendProposal("Test", "description", recipient, amount)
}

func TestProposalHandlerPassed(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	recipient := delAddr1

	feePool := types.InitialFeePool()
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(amount...)
	distrKeeper.SetFeePool(ctx, feePool)

	bankKeeper.EXPECT().BlockedAddr(recipient).Return(false)
	bankKeeper.EXPECT().SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, amount).Return(nil)

	tp := testProposal(recipient, amount)
	hdlr := distribution.NewCommunityPoolSpendProposalHandler(distrKeeper)
	require.NoError(t, hdlr(ctx, tp))
}

func TestProposalHandlerFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: time.Now()})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	recipient := delAddr1

	feePool := types.InitialFeePool()
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(amount...)
	distrKeeper.SetFeePool(ctx, feePool)

	bankKeeper.EXPECT().BlockedAddr(recipient).Return(false)
	bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, amount).
		Return(sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "%s is smaller than %s", zeroAmount, amount))

	tp := testProposal(recipient, amount)
	hdlr := distribution.NewCommunityPoolSpendProposalHandler(distrKeeper)
	require.ErrorIs(t, hdlr(ctx, tp), sdkerrors.ErrInsufficientFunds)
}
