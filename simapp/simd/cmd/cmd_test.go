package cmd_test

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func TestInitCmd(t *testing.T) {
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig())

}
