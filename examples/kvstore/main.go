package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/abci/server"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
	config := cfg.DefaultConfig()
	ctx := sdk.NewServerContext(config, logger)

	rootDir := viper.GetString(cli.HomeFlag)
	db, err := dbm.NewGoLevelDB("basecoind", filepath.Join(rootDir, "data"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Capabilities key to access the main KVStore.
	var capKeyMainStore = sdk.NewKVStoreKey("main")

	// Create BaseApp.
	var baseApp = bam.NewBaseApp("kvstore", nil, ctx, db)

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStoresIAVL(capKeyMainStore)

	// Set Tx decoder
	baseApp.SetTxDecoder(decodeTx)

	// Set a handler Route.
	baseApp.Router().AddRoute("kvstore", Handler(capKeyMainStore))

	// Load latest version.
	if err := baseApp.LoadLatestVersion(capKeyMainStore); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:26658", "socket", baseApp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = srv.Start()
	if err != nil {
		cmn.Exit(err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		err = srv.Stop()
		if err != nil {
			cmn.Exit(err.Error())
		}
	})
	return
}

// KVStore Handler
func Handler(storeKey sdk.StoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		dTx, ok := msg.(kvstoreTx)
		if !ok {
			panic("Handler should only receive kvstoreTx")
		}

		// tx is already unmarshalled
		key := dTx.key
		value := dTx.value

		store := ctx.KVStore(storeKey)
		store.Set(key, value)

		return sdk.Result{
			Code: 0,
			Log:  fmt.Sprintf("set %s=%s", key, value),
		}
	}
}
