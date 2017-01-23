package main

import (
	"flag"

	"github.com/tendermint/abci/server"
	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/plugins/counter"
	"github.com/tendermint/basecoin/plugins/paytovote"
	cmn "github.com/tendermint/go-common"
	eyes "github.com/tendermint/merkleeyes/client"
)

func main() {
	addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
	eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
	genFilePath := flag.String("genesis", "", "Genesis file, if any")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr, "socket")
	if err != nil {
		cmn.Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin app
	app := app.NewBasecoin(eyesCli)

	// create/add plugins
	counter := counter.New("counter")
	paytovote := paytovote.New("p2v")
	app.RegisterPlugin(counter)
	app.RegisterPlugin(paytovote)

	// If genesis file was specified, set key-value options
	if *genFilePath != "" {
		err := app.LoadGenesis(*genFilePath)
		if err != nil {
			cmn.Exit(cmn.Fmt("%+v", err))
		}
	}

	// Start the listener
	svr, err := server.NewServer(*addrPtr, "socket", app)
	if err != nil {
		cmn.Exit("create listener: " + err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})

}
