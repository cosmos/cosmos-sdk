package simapp_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestMsgService(t *testing.T) {
	db := dbm.NewMemDB()
	encCfg := simapp.MakeEncodingConfig()
	msgServiceOpt := func(bapp *baseapp.BaseApp) {
		testdata.RegisterMsgServer(
			bapp.MsgServiceRouter(),
			testdata.MsgImpl{},
		)
	}
	app := simapp.NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 0, encCfg, msgServiceOpt)

	msg, err := testdata.NewServiceMsgCreateDog(&testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}})
	require.NoError(t, err)

	txf := tx.Factory{}.
		WithTxConfig(encCfg.TxConfig).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")
}
