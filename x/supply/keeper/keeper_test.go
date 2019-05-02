package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var oneUatom = sdk.NewCoins(sdk.NewCoin("uatom", sdk.OneInt()))

func TestSupply(t *testing.T) {
	// TODO:
}

type testInput struct {
	mApp     *mock.App
	keeper   Keeper
	router   Router
	sk       staking.Keeper
	addrs    []sdk.AccAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

func getMockApp(t *testing.T, numGenAccs int, genState GenesisState, genAccs []auth.Account) testInput {
	mApp := mock.NewApp()

	staking.RegisterCodec(mApp.Cdc)
	RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tKeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyGov := sdk.NewKVStoreKey(StoreKey)

	rtr := NewRouter().AddRoute(RouterKey, ProposalHandler)

	pk := mApp.ParamsKeeper
	ck := bank.NewBaseKeeper(mApp.AccountKeeper, mApp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk := staking.NewKeeper(mApp.Cdc, keyStaking, tKeyStaking, ck, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	keeper := NewKeeper(mApp.Cdc, keyGov, pk, pk.Subspace("testgov"), ck, sk, DefaultCodespace, rtr)

	mApp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mApp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(keeper))

	mApp.SetEndBlocker(getEndBlocker(keeper))
	mApp.SetInitChainer(getInitChainer(mApp, keeper, sk, genState))

	require.NoError(t, mApp.CompleteSetup(keyStaking, tKeyStaking, keyGov))

	valTokens := sdk.TokensFromTendermintPower(42)

	var (
		addrs    []sdk.AccAddress
		pubKeys  []crypto.PubKey
		privKeys []crypto.PrivKey
	)

	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})
	}

	mock.SetGenesis(mApp, genAccs)

	return testInput{mApp, keeper, rtr, sk, addrs, pubKeys, privKeys}
}
