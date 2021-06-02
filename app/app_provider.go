package app

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/container"

	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

type AppProvider struct {
	config            *Config
	container         *container.Container
	interfaceRegistry codecTypes.InterfaceRegistry
	codec             codec.Codec
	txConfig          client.TxConfig
	amino             *codec.LegacyAmino
}

func (ap *AppProvider) InterfaceRegistry() codecTypes.InterfaceRegistry {
	return ap.interfaceRegistry
}

func (ap *AppProvider) Codec() codec.Codec {
	return ap.codec
}

func (ap *AppProvider) TxConfig() client.TxConfig {
	return ap.txConfig
}

func (ap *AppProvider) Amino() *codec.LegacyAmino {
	return ap.amino
}

func NewApp(config *Config) (*AppProvider, error) {
	ctr := container.NewContainer()
	interfaceRegistry := codecTypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(marshaler, tx.DefaultSignModes)
	amino := codec.NewLegacyAmino()

	err := ctr.RegisterProvider(func() (codecTypes.InterfaceRegistry, codec.Codec, client.TxConfig, *codec.LegacyAmino) {
		return interfaceRegistry, marshaler, txConfig, amino
	})
	if err != nil {
		return nil, err
	}

	return &AppProvider{
		config:            config,
		container:         ctr,
		interfaceRegistry: interfaceRegistry,
		codec:             marshaler,
		txConfig:          txConfig,
		amino:             amino,
	}, nil
}

// NewApp is an AppCreator
func (ap *AppProvider) AppCreator(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	return ap.newApp(
		logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		appOpts,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshotStore(snapshotStore),
		baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval))),
		baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent))),
	)
}

func (ap *AppProvider) newApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *app {
	return &app{}
}

// AppExport creates a new app (optionally at a given height) and exports state.
func (ap *AppProvider) AppExportor(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string,
	appOpts servertypes.AppOptions) (servertypes.ExportedApp, error) {

	var a *app
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	if height != -1 {
		a = ap.newApp(logger, db, traceStore, false, map[int64]bool{}, homePath, uint(1), appOpts)

		if err := a.LoadVersion(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		a = ap.newApp(logger, db, traceStore, true, map[int64]bool{}, homePath, uint(1), appOpts)
	}

	return a.exportAppStateAndValidators(forZeroHeight, jailAllowedAddrs)
}
