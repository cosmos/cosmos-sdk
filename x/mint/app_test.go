package mint

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type MockApp struct {
	*mock.App

	keySupply   *sdk.KVStoreKey
	keyStaking  *sdk.KVStoreKey
	tkeyStaking *sdk.KVStoreKey
	keyMint     *sdk.KVStoreKey

	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	stakingKeeper staking.Keeper
	mintKeeper    Keeper
}

func registerCodec(cdc *codec.Codec) {
	supply.RegisterCodec(cdc)
}

func getMockApp(t *testing.T, numGenAccs int, balance int64, mintParams Params) (mockApp *MockApp, addrKeysSlice mock.AddrKeysSlice) {
	mapp := mock.NewApp()
	registerCodec(mapp.Cdc)

	mockApp = &MockApp{
		App: mapp,

		keySupply:   sdk.NewKVStoreKey(supply.StoreKey),
		keyStaking:  sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking: sdk.NewKVStoreKey(staking.TStoreKey),
		keyMint:     sdk.NewKVStoreKey(types.StoreKey),
	}

	mockApp.bankKeeper = bank.NewBaseKeeper(mockApp.AccountKeeper,
		mockApp.ParamsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace, nil)

	maccPerms := map[string][]string{
		auth.FeeCollectorName:     nil,
		types.ModuleName:          []string{supply.Minter},
		staking.NotBondedPoolName: []string{supply.Burner, supply.Staking},
		staking.BondedPoolName:    []string{supply.Burner, supply.Staking},
	}

	mockApp.supplyKeeper = supply.NewKeeper(mockApp.Cdc, mockApp.keySupply, mockApp.AccountKeeper,
		mockApp.bankKeeper, maccPerms)

	mockApp.stakingKeeper = staking.NewKeeper(
		mockApp.Cdc, mockApp.keyStaking, mockApp.tkeyStaking, mockApp.supplyKeeper,
		mockApp.ParamsKeeper.Subspace(staking.DefaultParamspace), staking.DefaultCodespace,
	)

	mockApp.mintKeeper = NewKeeper(mockApp.Cdc, mockApp.keyMint,
		mockApp.ParamsKeeper.Subspace(types.DefaultParamspace), &mockApp.stakingKeeper,
		mockApp.supplyKeeper, auth.FeeCollectorName)

	//mockApp.Router().AddRoute("", nil)
	mockApp.QueryRouter().AddRoute(QuerierRoute, NewQuerier(mockApp.mintKeeper))

	decCoins, _ := sdk.ParseDecCoins(fmt.Sprintf("%d%s",
		balance, sdk.DefaultBondDenom))
	coins := decCoins

	keysSlice, genAccs := CreateGenAccounts(numGenAccs, coins)
	addrKeysSlice = keysSlice

	mockApp.SetBeginBlocker(getBeginBlocker(mockApp.mintKeeper))
	mockApp.SetInitChainer(getInitChainer(mockApp.App, mockApp.supplyKeeper, mockApp.mintKeeper, mockApp.stakingKeeper,
		genAccs, mintParams))

	// todo: checkTx in mock app
	mockApp.SetAnteHandler(nil)

	app := mockApp
	require.NoError(t, app.CompleteSetup(
		app.keyStaking,
		app.tkeyStaking,
		app.keyMint,
		app.keySupply,
	))
	mock.SetGenesis(mockApp.App, genAccs)

	for i := 0; i < numGenAccs; i++ {
		mock.CheckBalance(t, app.App, keysSlice[i].Address, coins)
	}

	return
}

func getBeginBlocker(keeper Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		BeginBlocker(ctx, keeper)
		return abci.ResponseBeginBlock{}
	}
}

func getInitChainer(mapp *mock.App, supplyKeeper supply.Keeper, mintKeeper Keeper, stakingkeeper staking.Keeper,
	genAccs []auth.Account, mintParams Params) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		for _, acc := range genAccs {
			mapp.TotalCoinsSupply = mapp.TotalCoinsSupply.Add(acc.GetCoins())
		}
		supplyKeeper.SetSupply(ctx, supply.NewSupply(mapp.TotalCoinsSupply))

		mintKeeper.SetParams(ctx, mintParams)
		mintKeeper.SetMinterCustom(ctx, types.MinterCustom{})

		stakingkeeper.SetParams(ctx, staking.DefaultParams())

		return abci.ResponseInitChain{}
	}
}

func CreateGenAccounts(numAccs int, genCoins sdk.Coins) (addrKeysSlice mock.AddrKeysSlice,
	genAccs []auth.Account) {
	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		addrKeys := mock.NewAddrKeys(addr, pubKey, privKey)
		account := &auth.BaseAccount{
			Address: addr,
			Coins:   genCoins,
		}
		genAccs = append(genAccs, account)
		addrKeysSlice = append(addrKeysSlice, addrKeys)
	}
	return
}
