package app

import (
	"maps"
	"slices"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/authz"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             *codec.ProtoCodec
	LegacyAmino       *codec.LegacyAmino
	TxConfig          client.TxConfig
}

func NewEncodingConfigFromOptions(opts types.InterfaceRegistryOptions) EncodingConfig {
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(opts)
	if err != nil {
		panic(err)
	}

	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	if err := interfaceRegistry.SigningContext().Validate(); err != nil {
		panic(err)
	}

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             appCodec,
		LegacyAmino:       legacyAmino,
		TxConfig:          txConfig,
	}
}

type SDKAppConfig struct {
	AppName string

	AppOpts        servertypes.AppOptions
	BaseAppOptions []func(*baseapp.BaseApp)

	InterfaceRegistryOptions types.InterfaceRegistryOptions

	WithProtocolPool bool
	WithAuthz        bool
	WithEpochs       bool
	WithFeeGrant     bool
	WithMint         bool
	// TODO gov optional?
	// TODO any other optional modules?

	WithUnorderedTx bool

	Keys               []string
	OrderPreBlockers   []string
	OrderBeginBlockers []string
	OrderEndBlockers   []string
	OrderInitGenesis   []string
	OrderExportGenesis []string

	ModuleAccountPerms map[string][]string

	Mempool mempool.Mempool

	VerifyVoteExtensionHandler sdk.VerifyVoteExtensionHandler
	PrepareProposalHandler     sdk.PrepareProposalHandler
	ProcessProposalHandler     sdk.ProcessProposalHandler
	ExtendVoteHandler          sdk.ExtendVoteHandler

	Upgrades []Upgrade[AppI]

	ModuleAuthority string
}

func DefaultSDKAppConfig(
	name string,
	opts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) SDKAppConfig {
	defaultOptions := server.DefaultBaseappOptions(opts)

	// TODO - populate if nil to fix any issues

	baseAppOptions = append(defaultOptions, baseAppOptions...)

	return SDKAppConfig{
		AppName: name,

		InterfaceRegistryOptions: defaultInterfaceRegistryOptions,

		AppOpts:          opts,
		BaseAppOptions:   baseAppOptions,
		WithProtocolPool: true,
		WithAuthz:        true,
		WithEpochs:       true,
		WithFeeGrant:     true,
		WithMint:         true,

		WithUnorderedTx: true,

		ModuleAccountPerms: defaultMaccPerms,

		OrderPreBlockers:   defaultOrderPreBlockers,
		OrderBeginBlockers: defaultOrderBeginBlockers,
		OrderEndBlockers:   defaultOrderEndBlockers,
		OrderInitGenesis:   defaultOrderInitGenesis,
		OrderExportGenesis: defaultOrderExportGenesis,

		Mempool:                    mempool.NoOpMempool{},
		VerifyVoteExtensionHandler: baseapp.NoOpVerifyVoteExtensionHandler(),
		ExtendVoteHandler:          baseapp.NoOpExtendVote(),
		// leave these as nil for construction later in baseapp by default
		PrepareProposalHandler: nil,
		ProcessProposalHandler: nil,

		Upgrades: nil,

		ModuleAuthority: defaultModuleAuthority,
	}
}

// TODO test thoroughly
func (appConfig *SDKAppConfig) processOptionalModules() {
	checkForModuleInclusion := func(moduleName string) func(string) bool {
		return func(s string) bool {
			return moduleName == s
		}
	}

	deleteModuleFromOrdering := func(moduleName string) {
		defaultOrderPreBlockers = slices.DeleteFunc(defaultOrderPreBlockers, checkForModuleInclusion(moduleName))
		defaultOrderBeginBlockers = slices.DeleteFunc(defaultOrderBeginBlockers, checkForModuleInclusion(moduleName))
		defaultOrderEndBlockers = slices.DeleteFunc(defaultOrderEndBlockers, checkForModuleInclusion(moduleName))
		defaultOrderInitGenesis = slices.DeleteFunc(defaultOrderInitGenesis, checkForModuleInclusion(moduleName))
		defaultOrderExportGenesis = slices.DeleteFunc(defaultOrderExportGenesis, checkForModuleInclusion(moduleName))
	}

	if !appConfig.WithProtocolPool {
		// remove from macc permissions
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case protocolpooltypes.ModuleName:
				return true
			case protocolpooltypes.ProtocolPoolEscrowAccount:
				return true
			default:
				return false
			}
		})

		deleteModuleFromOrdering(protocolpooltypes.ModuleName)
	}

	if !appConfig.WithAuthz {
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case authz.ModuleName:
				return true
			default:
				return false
			}
		})

		deleteModuleFromOrdering(authz.ModuleName)
	}

	if !appConfig.WithFeeGrant {
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case feegrant.ModuleName:
				return true
			default:
				return false
			}
		})

		deleteModuleFromOrdering(feegrant.ModuleName)
	}

	if !appConfig.WithMint {
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case minttypes.ModuleName:
				return true
			default:
				return false
			}
		})
	}

	if !appConfig.WithEpochs {
		maps.DeleteFunc(appConfig.ModuleAccountPerms, func(s string, _ []string) bool {
			switch s {
			case epochstypes.ModuleName:
				return true
			default:
				return false
			}
		})

		deleteModuleFromOrdering(epochstypes.ModuleName)
	}
}
