package baseapp

import (
	"github.com/tendermint/abci/server"
	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
)

// RunForever - BasecoinApp execution and cleanup
func RunForever(app abci.Application) {

	// Start the ABCI server
	srv, err := server.NewServer("0.0.0.0:46658", "socket", app)
	if err != nil {
		cmn.Exit(err.Error())
	}
	srv.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
}
