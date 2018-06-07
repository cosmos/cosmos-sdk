package mock

import (
	"io/ioutil"
	"os"

	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
	"path/filepath"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// setupMockApp returns an application as well as a clean-up function
// to be used to quickly setup a test case with an app
// TODO Fix name ambiguity
func setupMockApp() (*MockApp, func(), error) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "mock")
	rootDir, err := ioutil.TempDir("", "mock-sdk")
	if err != nil {
		return &MockApp{}, nil, err
	}

	cleanup := func() {
		os.RemoveAll(rootDir)
	}

	app, err := newApp(rootDir, logger)
	return app.(*MockApp), cleanup, err
}

// NewApp sets up everything it can without knowing what modules will be loaded.
// TODO: Fix this name ambiguity
func newApp(rootDir string, logger log.Logger) (abci.Application, error) {
	db, err := dbm.NewGoLevelDB("mock", filepath.Join(rootDir, "data"))
	if err != nil {
		return nil, err
	}

	// Create MockApp.
	mApp := &MockApp{
		BaseApp:         bam.NewBaseApp("mockApp", nil, logger, db),
		Cdc:             wire.NewCodec(),
		KeyMain:         sdk.NewKVStoreKey("mock"),
		KeyAccountStore: sdk.NewKVStoreKey("acc"),
	}

	// Define the accountMapper.
	mApp.AccountMapper = auth.NewAccountMapper(
		mApp.Cdc,
		mApp.KeyAccountStore, // target store
		&auth.BaseAccount{},  // prototype
	)
	return mApp, nil
}
