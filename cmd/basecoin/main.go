package main

import (
	"flag"

	"github.com/tendermint/basecoin/app"
	. "github.com/tendermint/go-common"
	eyes "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmsp/server"
)

func main() {

	addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
	eyesPtr := flag.String("eyes", "tcp://0.0.0.0:46659", "MerkleEyes address")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr)
	if err != nil {
		Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin app
	app := app.NewBasecoin(eyesCli)

	// Start the listener
	svr, err := server.NewServer(*addrPtr, app)
	if err != nil {
		Exit("create listener: " + err.Error())
	}

	// Wait forever
	TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})

}
