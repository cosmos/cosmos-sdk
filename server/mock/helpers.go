package mock

import (
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
)

// SetupApp returns an application as well as a clean-up function to be used to
// quickly setup a test case with an app.
func SetupApp() (abci.Application, func(), error) {
	var logger log.Logger
	if testing.Verbose() {
		logger = log.NewLoggerWithKV("module", "mock")
	} else {
		logger = log.NewNopLogger()
	}

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
