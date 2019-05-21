package querier

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/tendermint/tendermint/crypto"
)

type testInput struct {
	mApp     *mock.App
	keeper   keeper.Keeper
	addrs    []sdk.AccAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

//func getMockApp(t *testing.T, numGenAccs int, genAccs []auth.Account) testInput {
//	mApp := mock.NewApp()
//
//	types.RegisterCodec(mApp.Cdc)
//
//	keyNFT := sdk.NewKVStoreKey(types.StoreKey)
//	keeperInstance := keeper.NewKeeper(keyNFT, types.ModuleCdc)
//
//	mApp.Router().AddRoute(types.RouterKey, nft.NewHandler(keeperInstance))
//	mApp.QueryRouter().AddRoute(types.QuerierRoute, NewQuerier(keeperInstance))
//
//	require.NoError(t, mApp.CompleteSetup(keyNFT))
//
//	valTokens := sdk.TokensFromTendermintPower(42)
//
//	var (
//		addrs    []sdk.AccAddress
//		pubKeys  []crypto.PubKey
//		privKeys []crypto.PrivKey
//	)
//
//	if genAccs == nil || len(genAccs) == 0 {
//		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
//			sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})
//	}
//
//	mock.SetGenesis(mApp, genAccs)
//
//	return testInput{mApp, keeper, addrs, pubKeys, privKeys}
//}
