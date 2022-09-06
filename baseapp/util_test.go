package baseapp_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	consensusmodulev1 "cosmossdk.io/api/cosmos/consensus/module/v1"
	mintmodulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	paramsmodulev1 "cosmossdk.io/api/cosmos/params/module/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	txmodulev1 "cosmossdk.io/api/cosmos/tx/module/v1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GenesisStateWithSingleValidator initializes GenesisState with a single validator and genesis accounts
// that also act as delegators.
func GenesisStateWithSingleValidator(t *testing.T, codec codec.Codec, builder *runtime.AppBuilder) map[string]json.RawMessage {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balances := []banktypes.Balance{
		{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
		},
	}

	genesisState := builder.DefaultGenesis()
	// sus
	genesisState, err = simtestutil.GenesisStateWithValSet(codec, genesisState, valSet, []authtypes.GenesisAccount{acc}, balances...)
	require.NoError(t, err)

	return genesisState
}

func makeTestConfig() depinject.Config {
	return appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: "runtime",
				Config: appconfig.WrapAny(&runtimev1alpha1.Module{
					AppName: "BaseAppApp",
					BeginBlockers: []string{
						"mint",
						"staking",
						"auth",
						"bank",
						"params",
						"consensus",
					},
					EndBlockers: []string{
						"staking",
						"auth",
						"bank",
						"mint",
						"params",
						"consensus",
					},
					OverrideStoreKeys: []*runtimev1alpha1.StoreKeyConfig{
						{
							ModuleName: "auth",
							KvStoreKey: "acc",
						},
					},
					InitGenesis: []string{
						"auth",
						"bank",
						"staking",
						"mint",
						"params",
						"consensus",
					},
				}),
			},
			{
				Name: "auth",
				Config: appconfig.WrapAny(&authmodulev1.Module{
					Bech32Prefix: "cosmos",
					ModuleAccountPermissions: []*authmodulev1.ModuleAccountPermission{
						{Account: authtypes.FeeCollectorName},
						{Account: minttypes.ModuleName, Permissions: []string{authtypes.Minter}},
						{Account: stakingtypes.BondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
						{Account: stakingtypes.NotBondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
					},
				}),
			},
			{
				Name:   "bank",
				Config: appconfig.WrapAny(&bankmodulev1.Module{}),
			},
			{
				Name:   "params",
				Config: appconfig.WrapAny(&paramsmodulev1.Module{}),
			},
			{
				Name:   "staking",
				Config: appconfig.WrapAny(&stakingmodulev1.Module{}),
			},
			{
				Name:   "mint",
				Config: appconfig.WrapAny(&mintmodulev1.Module{}),
			},
			{
				Name:   "consensus",
				Config: appconfig.WrapAny(&consensusmodulev1.Module{}),
			},
			{
				Name:   "tx",
				Config: appconfig.WrapAny(&txmodulev1.Module{}),
			},
		},
	})
}

func makeMinimalConfig() depinject.Config {
	return appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: "runtime",
				Config: appconfig.WrapAny(&runtimev1alpha1.Module{
					AppName: "BaseAppApp",
				}),
			},
		},
	})
}
