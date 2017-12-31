package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
)

func main() {

	app := app.NewApp("dummy")

	db, err := dbm.NewGoLevelDB("dummy", "dummy-data")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create CommitStoreLoader
	cacheSize := 10000
	numHistory := int64(100)
	loader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)

	// Create MultiStore
	multiStore := store.NewMultiStore(db)
	multiStore.SetSubstoreLoader("main", loader)

	// Create Handler
	handler := types.Decorate(unmarshalDecorator, dummyHandler)

	// Set everything on the app and load latest
	app.SetCommitMultiStore(multiStore)
	app.SetHandler(handler)
	if err := app.LoadLatestVersion(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:46658", "socket", app)
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

type dummyTx struct {
	key   []byte
	value []byte

	bytes []byte
}

func (tx dummyTx) Get(key interface{}) (value interface{}) {
	switch k := key.(type) {
	case string:
		switch k {
		case "key":
			return tx.key
		case "value":
			return tx.value
		}
	}
	return nil
}

func (tx dummyTx) SignBytes() []byte {
	return tx.bytes
}

// Should the app be calling this? Or only handlers?
func (tx dummyTx) ValidateBasic() error {
	return nil
}

func (tx dummyTx) Signers() [][]byte {
	return nil
}

func (tx dummyTx) TxBytes() []byte {
	return tx.bytes
}

func (tx dummyTx) Signatures() []types.StdSignature {
	return nil
}

func unmarshalDecorator(ctx types.Context, ms types.MultiStore, tx types.Tx, next types.Handler) types.Result {
	txBytes := ctx.TxBytes()

	split := bytes.Split(txBytes, []byte("="))
	if len(split) == 1 {
		k := split[0]
		tx = dummyTx{k, k, txBytes}
	} else if len(split) == 2 {
		k, v := split[0], split[1]
		tx = dummyTx{k, v, txBytes}
	} else {
		return types.Result{
			Code: 1,
			Log:  "too many =",
		}
	}

	return next(ctx, ms, tx)
}

func dummyHandler(ctx types.Context, ms types.MultiStore, tx types.Tx) types.Result {
	// tx is already unmarshalled
	key := tx.Get("key").([]byte)
	value := tx.Get("value").([]byte)

	main := ms.GetKVStore("main")
	main.Set(key, value)

	return types.Result{
		Code: 0,
		Log:  fmt.Sprintf("set %s=%s", key, value),
	}
}
