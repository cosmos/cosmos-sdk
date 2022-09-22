package mock

import (
	"fmt"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

// SetupApp returns an application as well as a clean-up function
// to be used to quickly setup a test case with an app
func SetupApp() (abci.Application, func(), error) {
	logger := tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout)).With("module", "mock")

	rootDir, err := os.MkdirTemp("", "mock-sdk")
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		err := os.RemoveAll(rootDir)
		if err != nil {
			fmt.Printf("could not delete %s, had error %s\n", rootDir, err.Error())
		}
	}

	app, err := NewApp(rootDir, logger)
	return app, cleanup, err
}
