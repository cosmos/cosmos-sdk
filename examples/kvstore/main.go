package main

import (
	"fmt"
	"os"

	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")

	db, err := dbm.NewGoLevelDB("basecoind", "data")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Capabilities key to access the main KVStore.
	var capKeyMainStore = sdk.NewKVStoreKey("main")

	// Create BaseApp.
	var baseApp = bam.NewBaseApp("kvstore", logger, db)

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStore(capKeyMainStore, sdk.StoreTypeIAVL)

	// Set Tx decoder
	baseApp.SetTxDecoder(decodeTx)

	// Set a handler Route.
	baseApp.Router().AddRoute("kvstore", KVStoreHandler(capKeyMainStore))

	// Load latest version.
	if err := baseApp.LoadLatestVersion(capKeyMainStore); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:46658", "socket", baseApp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	srv.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
	return
}

func KVStoreHandler(storeKey sdk.StoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		dTx, ok := msg.(kvstoreTx)
		if !ok {
			panic("KVStoreHandler should only receive kvstoreTx")
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
