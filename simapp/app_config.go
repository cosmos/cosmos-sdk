package simapp

import (
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"cosmossdk.io/core/appconfig"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	authzmodulev1 "cosmossdk.io/api/cosmos/authz/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	capabilitymodulev1 "cosmossdk.io/api/cosmos/capability/module/v1"
	distrmodulev1 "cosmossdk.io/api/cosmos/distribution/module/v1"
	evidencemodulev1 "cosmossdk.io/api/cosmos/evidence/module/v1"
	feegrantmodulev1 "cosmossdk.io/api/cosmos/feegrant/module/v1"
	genutilmodulev1 "cosmossdk.io/api/cosmos/genutil/module/v1"
	groupmodulev1 "cosmossdk.io/api/cosmos/group/module/v1"
	mintmodulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	nftmodulev1 "cosmossdk.io/api/cosmos/nft/module/v1"
	paramsmodulev1 "cosmossdk.io/api/cosmos/params/module/v1"
	slashingmodulev1 "cosmossdk.io/api/cosmos/slashing/module/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	txmodulev1 "cosmossdk.io/api/cosmos/tx/module/v1"
	upgrademodulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"
	vestingmodulev1 "cosmossdk.io/api/cosmos/vesting/module/v1"
)

var AppConfig = appconfig.Compose(&appv1alpha1.Config{
	Modules: []*appv1alpha1.ModuleConfig{
		{
			Name: "runtime",
			Config: MakeConfig(&runtimev1alpha1.Module{
				AppName: "SimApp",
				// During begin block slashing happens after distr.BeginBlocker so that
				// there is nothing left over in the validator fee pool, so as to keep the
				// CanWithdrawInvariant invariant.
				// NOTE: staking module is required if HistoricalEntries param > 0
				// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
				BeginBlockers: []string{
					upgradetypes.ModuleName,
					capabilitytypes.ModuleName,
					minttypes.ModuleName,
					distrtypes.ModuleName,
					slashingtypes.ModuleName,
					evidencetypes.ModuleName,
					stakingtypes.ModuleName,
					authtypes.ModuleName,
					banktypes.ModuleName,
					govtypes.ModuleName,
					crisistypes.ModuleName,
					genutiltypes.ModuleName,
					authz.ModuleName,
					feegrant.ModuleName,
					nft.ModuleName,
					group.ModuleName,
					paramstypes.ModuleName,
					vestingtypes.ModuleName,
				},
				EndBlockers: []string{
					crisistypes.ModuleName,
					govtypes.ModuleName,
					stakingtypes.ModuleName,
					capabilitytypes.ModuleName,
					authtypes.ModuleName,
					banktypes.ModuleName,
					distrtypes.ModuleName,
					slashingtypes.ModuleName,
					minttypes.ModuleName,
					genutiltypes.ModuleName,
					evidencetypes.ModuleName,
					authz.ModuleName,
					feegrant.ModuleName,
					nft.ModuleName,
					group.ModuleName,
					paramstypes.ModuleName,
					upgradetypes.ModuleName,
					vestingtypes.ModuleName,
				},
				OverrideStoreKeys: []*runtimev1alpha1.StoreKeyConfig{
					{
						ModuleName: authtypes.ModuleName,
						KvStoreKey: "acc",
					},
				},
			}),
		},
		{
			Name: authtypes.ModuleName,
			Config: MakeConfig(&authmodulev1.Module{
				Bech32Prefix: "cosmos",
				ModuleAccountPermissions: []*authmodulev1.ModuleAccountPermission{
					{Account: "fee_collector"},
					{Account: distrtypes.ModuleName},
					{Account: minttypes.ModuleName, Permissions: []string{"minter"}},
					{Account: "bonded_tokens_pool", Permissions: []string{"burner", stakingtypes.ModuleName}},
					{Account: "not_bonded_tokens_pool", Permissions: []string{"burner", stakingtypes.ModuleName}},
					{Account: govtypes.ModuleName, Permissions: []string{"burner"}},
					{Account: nft.ModuleName},
				},
			}),
		},

		{
			Name:   vestingtypes.ModuleName,
			Config: MakeConfig(&vestingmodulev1.Module{}),
		},
		{
			Name:   banktypes.ModuleName,
			Config: MakeConfig(&bankmodulev1.Module{}),
		},
		{
			Name:   stakingtypes.ModuleName,
			Config: MakeConfig(&stakingmodulev1.Module{}),
		},
		{
			Name:   slashingtypes.ModuleName,
			Config: MakeConfig(&slashingmodulev1.Module{}),
		},
		{
			Name:   paramstypes.ModuleName,
			Config: MakeConfig(&paramsmodulev1.Module{}),
		},
		{
			Name:   "tx",
			Config: MakeConfig(&txmodulev1.Module{}),
		},
		{
			Name:   genutiltypes.ModuleName,
			Config: MakeConfig(&genutilmodulev1.Module{}),
		},
		{
			Name:   authz.ModuleName,
			Config: MakeConfig(&authzmodulev1.Module{}),
		},
		{
			Name:   upgradetypes.ModuleName,
			Config: MakeConfig(&upgrademodulev1.Module{}),
		},
		{
			Name:   distrtypes.ModuleName,
			Config: MakeConfig(&distrmodulev1.Module{}),
		},
		{
			Name: capabilitytypes.ModuleName,
			Config: MakeConfig(&capabilitymodulev1.Module{
				SealKeeper: true,
			}),
		},
		{
			Name:   evidencetypes.ModuleName,
			Config: MakeConfig(&evidencemodulev1.Module{}),
		},
		{
			Name:   minttypes.ModuleName,
			Config: MakeConfig(&mintmodulev1.Module{}),
		},
		{
			Name: group.ModuleName,
			Config: MakeConfig(&groupmodulev1.Module{
				MaxExecutionPeriod: durationpb.New(time.Second * 1209600),
				MaxMetadataLen:     255,
			}),
		},
		{
			Name:   nft.ModuleName,
			Config: MakeConfig(&nftmodulev1.Module{}),
		},
		{
			Name:   feegrant.ModuleName,
			Config: MakeConfig(&feegrantmodulev1.Module{}),
		},
	},
})

func MakeConfig(config protoreflect.ProtoMessage) *anypb.Any {
	m, err := anypb.New(config)
	if err != nil {
		panic(err)
	}

	return m
}
