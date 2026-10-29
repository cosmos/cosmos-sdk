package app

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
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
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
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

	WithAuthz    bool
	WithEpochs   bool
	WithFeeGrant bool
	WithMint     bool

	WithUnorderedTx bool

	TransientStoreKeys []string
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

	OptimisticExecutionEnabled bool

	// BlockSTM enables parallel execution when configured; nil means serial execution.
	BlockSTM *BlockSTMConfig

	Upgrades []Upgrade[AppI]

	ModuleAuthority string
	GovConfig       *govtypes.Config
	GovHooks        []govtypes.GovHooks
	GovVoteCalcFn   govkeeper.CalculateVoteResultsAndVotingPowerFn
}

type BlockSTMConfig struct {
	Workers  int
	Estimate bool
}

// DefaultSDKAppConfig returns the single canonical app configuration baseline.
// It always includes server.DefaultBaseappOptions and applies a chain-id
// fallback (AppName) when app opts do not set one explicitly.
func DefaultSDKAppConfig(
	name string,
	opts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) SDKAppConfig {
	if opts == nil {
		panic("app options must not be nil")
	}
	if cast.ToString(opts.Get(flags.FlagHome)) == "" {
		panic("app options must include --home")
	}

	// Resolve chain-id with priority: explicit flag > genesis file > app name.
	// Injecting the resolved value into wrappedOpts means DefaultBaseappOptions
	// always receives a concrete chain-id without needing to open the genesis
	// file itself, preserving correct chain-id for both running nodes (which
	// have a genesis) and pre-genesis unit-test construction (which do not).
	chainID := cast.ToString(opts.Get(flags.FlagChainID))
	if chainID == "" {
		homeDir := cast.ToString(opts.Get(flags.FlagHome))
		genesisPathCfg, _ := opts.Get("genesis_file").(string)
		if genesisPathCfg == "" {
			genesisPathCfg = filepath.Join("config", "genesis.json")
		}
		if reader, err := os.Open(filepath.Join(homeDir, genesisPathCfg)); err == nil {
			if id, err := genutiltypes.ParseChainIDFromGenesis(reader); err == nil && id != "" {
				chainID = id
			}
			reader.Close()
		}
	}
	if chainID == "" {
		chainID = name
	}

	wrappedOpts := appOptionsWithDefaults{
		base: opts,
		defaults: map[string]any{
			flags.FlagChainID: chainID,
		},
	}

	baseOpts := append(server.DefaultBaseappOptions(wrappedOpts), slices.Clone(baseAppOptions)...)

	return SDKAppConfig{
		AppName: name,

		InterfaceRegistryOptions: defaultInterfaceRegistryOptions,

		AppOpts:        wrappedOpts,
		BaseAppOptions: baseOpts,
		WithAuthz:      true,
		WithEpochs:     true,
		WithFeeGrant:   true,
		WithMint:       true,

		WithUnorderedTx: true,

		ModuleAccountPerms: cloneModuleAccountPerms(defaultMaccPerms),

		OrderPreBlockers:   slices.Clone(defaultOrderPreBlockers),
		OrderBeginBlockers: slices.Clone(defaultOrderBeginBlockers),
		OrderEndBlockers:   slices.Clone(defaultOrderEndBlockers),
		OrderInitGenesis:   slices.Clone(defaultOrderInitGenesis),
		OrderExportGenesis: slices.Clone(defaultOrderExportGenesis),

		Mempool:                    mempool.NoOpMempool{},
		VerifyVoteExtensionHandler: baseapp.NoOpVerifyVoteExtensionHandler(),
		ExtendVoteHandler:          baseapp.NoOpExtendVote(),
		// leave these as nil for construction later in baseapp by default
		PrepareProposalHandler: nil,
		ProcessProposalHandler: nil,

		BlockSTM: nil,

		Upgrades: nil,

		ModuleAuthority: defaultModuleAuthority,
		GovConfig: func() *govtypes.Config {
			cfg := govtypes.DefaultConfig()
			return &cfg
		}(),
		GovHooks:      nil,
		GovVoteCalcFn: nil,
	}
}

func (appConfig SDKAppConfig) Validate() error {
	if appConfig.AppName == "" {
		return fmt.Errorf("app name must not be empty")
	}
	if appConfig.AppOpts == nil {
		return fmt.Errorf("app opts must not be nil")
	}
	if cast.ToString(appConfig.AppOpts.Get(flags.FlagHome)) == "" {
		return fmt.Errorf("app opts must include --home")
	}
	if appConfig.ModuleAccountPerms == nil {
		return fmt.Errorf("module account perms must not be nil; use DefaultSDKAppConfig or provide an explicit map")
	}
	if appConfig.BlockSTM != nil && appConfig.BlockSTM.Workers < 1 {
		return fmt.Errorf("blockstm workers must be >= 1")
	}
	if appConfig.OptimisticExecutionEnabled && appConfig.BlockSTM != nil {
		return fmt.Errorf("optimistic execution and blockstm cannot both be enabled")
	}

	return nil
}

func (appConfig *SDKAppConfig) processOptionalModules() {
	deleteModuleFromOrdering := func(moduleName string) {
		appConfig.OrderPreBlockers = slices.DeleteFunc(appConfig.OrderPreBlockers, func(s string) bool { return s == moduleName })
		appConfig.OrderBeginBlockers = slices.DeleteFunc(appConfig.OrderBeginBlockers, func(s string) bool { return s == moduleName })
		appConfig.OrderEndBlockers = slices.DeleteFunc(appConfig.OrderEndBlockers, func(s string) bool { return s == moduleName })
		appConfig.OrderInitGenesis = slices.DeleteFunc(appConfig.OrderInitGenesis, func(s string) bool { return s == moduleName })
		appConfig.OrderExportGenesis = slices.DeleteFunc(appConfig.OrderExportGenesis, func(s string) bool { return s == moduleName })
	}

	removeOptionalModule := func(enabled bool, moduleName string) {
		if enabled {
			return
		}
		delete(appConfig.ModuleAccountPerms, moduleName)
		deleteModuleFromOrdering(moduleName)
	}

	removeOptionalModule(appConfig.WithAuthz, authz.ModuleName)
	removeOptionalModule(appConfig.WithFeeGrant, feegrant.ModuleName)
	removeOptionalModule(appConfig.WithMint, minttypes.ModuleName)
	removeOptionalModule(appConfig.WithEpochs, epochstypes.ModuleName)
}

func cloneModuleAccountPerms(src map[string][]string) map[string][]string {
	cloned := make(map[string][]string, len(src))
	for moduleName, perms := range src {
		cloned[moduleName] = slices.Clone(perms)
	}
	return cloned
}

type appOptionsWithDefaults struct {
	base     servertypes.AppOptions
	defaults map[string]any
}

func (a appOptionsWithDefaults) Get(key string) any {
	v := a.base.Get(key)
	if v == nil {
		return a.defaults[key]
	}
	return v
}
