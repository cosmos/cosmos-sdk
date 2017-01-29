package main

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/tendermint/abci/server"
	"github.com/tendermint/basecoin/app"
	cmn "github.com/tendermint/go-common"
	eyes "github.com/tendermint/merkleeyes/client"
)

const EyesCacheSize = 10000

func cmdStart(c *cli.Context) error {

	// Connect to MerkleEyes
	var eyesCli *eyes.Client
	if c.String("eyes") == "local" {
		eyesCli = eyes.NewLocalClient(c.String("eyes-db"), EyesCacheSize)
	} else {
		var err error
		eyesCli, err = eyes.NewClient(c.String("eyes"))
		if err != nil {
			return errors.New("connect to MerkleEyes: " + err.Error())
		}
	}

	// Create Basecoin app
	app := app.NewBasecoin(eyesCli)

	// If genesis file was specified, set key-value options
	if c.String("genesis") != "" {
		err := app.LoadGenesis(c.String("genesis"))
		if err != nil {
			return errors.New(cmn.Fmt("%+v", err))
		}
	}

	// Start the listener
	svr, err := server.NewServer(c.String("address"), "socket", app)
	if err != nil {
		return errors.New("create listener: " + err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})

	return nil
}
