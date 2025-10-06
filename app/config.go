package app

import (
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
	}
}
