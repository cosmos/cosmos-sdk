package app

import (
	"io"
	"os"

	"github.com/cometbft/cometbft/libs/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

const (
	// TJR token parameters
	TotalSupply = 1_000_000_000 // 1 billion TJR
	InitialMint = 500_000_000   // 500 million TJR
	StakeReward = 500_000_000   // 500 million TJR reserved for staking rewards

	// TJR token denom
	TokenDenom = "tjr"
)

var (
	DefaultNodeHome = os.ExpandEnv("$USERPROFILE/.tajeor")
)

type TajeorApp struct {
	*baseapp.BaseApp
	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry codectypes.InterfaceRegistry
}

func NewTajeorApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *TajeorApp {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	// Create the base app
	bApp := baseapp.NewBaseApp(
		"tajeor",
		nil, // simplified logger for now
		db,
		nil, // TxDecoder - will add later
		baseAppOptions...,
	)

	app := &TajeorApp{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(err)
		}
	}

	return app
}

func (app *TajeorApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *TajeorApp) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs []string) (servertypes.ExportedApp, error) {
	return servertypes.ExportedApp{}, nil
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *TajeorApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	// Simple implementation for now
}

// RegisterNodeService registers the node gRPC service for the provided client.
func (app *TajeorApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	// Simple implementation for now
}

// RegisterTendermintService registers the Tendermint gRPC service for the provided client.
func (app *TajeorApp) RegisterTendermintService(clientCtx client.Context) {
	// Simple implementation for now
}

// RegisterTxService registers the Tx gRPC service for the provided client.
func (app *TajeorApp) RegisterTxService(clientCtx client.Context) {
	// Simple implementation for now
}
