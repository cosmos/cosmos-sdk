package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var oneUatom = sdk.NewCoins(sdk.NewCoin("uatom", sdk.OneInt()))

func TestSupply(t *testing.T) {
	input := getMockApp(t, 2, nil)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})
	input.sk.SetParams(ctx, staking.DefaultParams())

	balance, _, _ := input.keeper.AccountsSupply(ctx)
	require.Equal(t, int64(200e6), balance.AmountOf(input.sk.BondDenom(ctx)).Int64())

	total := input.keeper.TotalSupply(ctx)
	require.Equal(t, total, balance)
}

type testInput struct {
	mApp   *mock.App
	keeper Keeper
	dk     distribution.Keeper
	sk     staking.Keeper

	addrs    []sdk.AccAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

func getMockApp(t *testing.T, numGenAccs int, genAccs []auth.Account) testInput {
	mApp := mock.NewApp()

	staking.RegisterCodec(mApp.Cdc)
	distribution.RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tKeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyDistr := sdk.NewKVStoreKey(distribution.StoreKey)
	tkeyDistr := sdk.NewTransientStoreKey(distribution.TStoreKey)

	pk := mApp.ParamsKeeper
	bk := bank.NewBaseKeeper(mApp.AccountKeeper, mApp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk := staking.NewKeeper(mApp.Cdc, keyStaking, tKeyStaking, bk, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)

	dk := distribution.NewKeeper(mApp.Cdc, keyDistr, pk.Subspace(distribution.DefaultParamspace), bk, sk, mApp.FeeCollectionKeeper, distribution.DefaultCodespace)

	keeper := NewKeeper(mApp.Cdc, mApp.AccountKeeper, dk, mApp.FeeCollectionKeeper, sk)

	require.NoError(t, mApp.CompleteSetup(keyStaking, tKeyStaking, keyDistr, tkeyDistr))

	valTokens := sdk.TokensFromTendermintPower(100)

	var (
		addrs    []sdk.AccAddress
		pubKeys  []crypto.PubKey
		privKeys []crypto.PrivKey
	)

	if numGenAccs > 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens)))
	}

	mock.SetGenesis(mApp, genAccs)

	return testInput{mApp, keeper, dk, sk, addrs, pubKeys, privKeys}
}
