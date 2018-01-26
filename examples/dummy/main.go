package main

import (
	"fmt"
	"os"

	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {

	// Capabilities key to access the main KVStore.
	var capKeyMainStore = sdk.NewKVStoreKey("main")

	// Create BaseApp.
	var baseApp = bam.NewBaseApp("dummy")

	// Set mounts for BaseApp's MultiStore.
	baseApp.MountStore(capKeyMainStore, sdk.StoreTypeIAVL)

	// Set Tx decoder
	baseApp.SetTxDecoder(decodeTx)

	// Set a handler Route.
	baseApp.Router().AddRoute("dummy", DummyHandler(capKeyMainStore))

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

func DummyHandler(storeKey sdk.StoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// tx is already unmarshalled
		key := msg.Get("key").([]byte)
		value := msg.Get("value").([]byte)

		store := ctx.KVStore(storeKey)
		store.Set(key, value)

		return sdk.Result{
			Code: 0,
			Log:  fmt.Sprintf("set %s=%s", key, value),
		}
	}
}
