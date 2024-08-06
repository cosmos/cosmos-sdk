package simapp

import (
	"context"
	app2 "cosmossdk.io/core/app"
	serverv2 "cosmossdk.io/server/v2"
	bank "cosmossdk.io/x/bank/types"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	comettypes "cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func NewTestApp(t *testing.T) *SimApp[transaction.Tx] {
	logger := log.NewTestLogger(t)
	vp := viper.New()
	vp.Set("store.app-db-backend", string(db.DBTypeGoLevelDB))
	vp.Set(serverv2.FlagHome, t.TempDir())

	app := NewSimApp[transaction.Tx](logger, vp)
	genesis := app.ModuleManager().DefaultGenesis()
	genesisBytes, err := json.Marshal(genesis)
	require.NoError(t, err)

	store := app.GetStore().(comettypes.Store)
	ci, err := store.LastCommitID()
	require.NoError(t, err)

	bz := sha256.Sum256([]byte{})

	_, _, err = app.InitGenesis(
		context.Background(),
		&app2.BlockRequest[transaction.Tx]{
			0,
			time.Now(),
			bz[:],
			"theChain",
			ci.Hash,
			nil,
			nil,
			true,
		},
		genesisBytes,
		nil,
	)
	require.NoError(t, err)

	return app
}

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	app := NewTestApp(t)

	gen, err := app.ExportAppStateAndValidators(false, nil, []string{bank.ModuleName})
	require.NoError(t, err)

	fmt.Printf("Exported genesis: %s\n", gen.AppState)
}
