package server_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/module/server"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	group "github.com/cosmos/cosmos-sdk/x/group/module"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestServer(t *testing.T) {
	ff := server.NewFixtureFactory(t, 6)
	cdc := ff.Codec()
	// Setting up bank keeper
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry())
	authtypes.RegisterInterfaces(cdc.InterfaceRegistry())

	paramsKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	authKey := sdk.NewKVStoreKey(authtypes.StoreKey)
	bankKey := sdk.NewKVStoreKey(banktypes.StoreKey)
	mintKey := sdk.NewKVStoreKey(minttypes.StoreKey)
	stakingKey := sdk.NewKVStoreKey(stakingtypes.StoreKey)
	tkey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	amino := codec.NewLegacyAmino()

	authSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, authtypes.ModuleName)
	bankSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, banktypes.ModuleName)
	// stakingSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, stakingtypes.ModuleName)
	// mintSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, minttypes.ModuleName)

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc, authKey, authSubspace, authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
	)

	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc, bankKey, accountKeeper, bankSubspace, modAccAddrs,
	)

	// stakingKeeper := stakingkeeper.NewKeeper(
	// 	cdc, stakingKey, accountKeeper, bankKeeper, stakingSubspace,
	// )

	// mintKeeper := mintkeeper.NewKeeper(
	// 	cdc, mintKey, mintSubspace, stakingKeeper, accountKeeper, bankKeeper, authtypes.FeeCollectorName,
	// )

	baseApp := ff.BaseApp()
	baseApp.MsgServiceRouter().SetInterfaceRegistry(cdc.InterfaceRegistry())
	banktypes.RegisterMsgServer(baseApp.MsgServiceRouter(), bankkeeper.NewMsgServerImpl(bankKeeper))
	baseApp.MountStore(tkey, types.StoreTypeTransient)
	baseApp.MountStore(paramsKey, types.StoreTypeIAVL)
	baseApp.MountStore(authKey, types.StoreTypeIAVL)
	baseApp.MountStore(bankKey, types.StoreTypeIAVL)
	baseApp.MountStore(stakingKey, types.StoreTypeIAVL)
	baseApp.MountStore(mintKey, types.StoreTypeIAVL)

	ff.SetModules([]module.Module{
		group.Module{AccountKeeper: accountKeeper},
	})

	// s := testsuite.NewIntegrationTestSuite(ff, accountKeeper, bankKeeper, mintKeeper, ecocreditSubspace)

	// suite.Run(t, s)
}
