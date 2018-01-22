package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/tendermint/abci/server"
	crypto "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {

	db, err := dbm.NewGoLevelDB("dummy", "dummy-data")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create CommitStoreLoader
	cacheSize := 10000
	numHistory := int64(100)
	loader := store.NewIAVLStoreLoader(db, cacheSize, numHistory)

	// key to access the main KVStore
	var mainStoreKey = sdk.NewKVStoreKey("main")

	// Create MultiStore
	multiStore := store.NewCommitMultiStore(db)
	multiStore.SetSubstoreLoader(mainStoreKey, loader)

	// Set everything on the baseApp and load latest
	baseApp := bam.NewBaseApp("dummy", multiStore)

	// Set Tx decoder
	baseApp.SetTxDecoder(decodeTx)

	baseApp.Router().AddRoute("dummy", DummyHandler(mainStoreKey))

	if err := baseApp.LoadLatestVersion(mainStoreKey); err != nil {
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

func (tx dummyTx) Type() string {
	return "dummy"
}

func (tx dummyTx) GetSignBytes() []byte {
	return tx.bytes
}

// Should the app be calling this? Or only handlers?
func (tx dummyTx) ValidateBasic() error {
	return nil
}

func (tx dummyTx) GetSigners() []crypto.Address {
	return nil
}

func (tx dummyTx) GetSignatures() []sdk.StdSignature {
	return nil
}

func (tx dummyTx) GetFeePayer() crypto.Address {
	return nil
}

func decodeTx(txBytes []byte) (sdk.Tx, error) {
	var tx sdk.Tx

	split := bytes.Split(txBytes, []byte("="))
	if len(split) == 1 {
		k := split[0]
		tx = dummyTx{k, k, txBytes}
	} else if len(split) == 2 {
		k, v := split[0], split[1]
		tx = dummyTx{k, v, txBytes}
	} else {
		return nil, fmt.Errorf("too many =")
	}

	return tx, nil
}

func DummyHandler(storeKey sdk.SubstoreKey) sdk.Handler {
	return func(ctx sdk.Context, tx sdk.Tx) sdk.Result {
		// tx is already unmarshalled
		key := tx.Get("key").([]byte)
		value := tx.Get("value").([]byte)

		store := ctx.KVStore(storeKey)
		store.Set(key, value)

		return sdk.Result{
			Code: 0,
			Log:  fmt.Sprintf("set %s=%s", key, value),
		}
	}
}
