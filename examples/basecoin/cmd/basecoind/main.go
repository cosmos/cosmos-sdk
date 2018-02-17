package main

import (
	"fmt"
	"os"

	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
)

func main() {
	// TODO CREATE CLI

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")

	db, err := dbm.NewGoLevelDB("basecoind", "data")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bapp := app.NewBasecoinApp(logger, db)
	baseapp.RunForever(bapp)
}
