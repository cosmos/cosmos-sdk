package simapp_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
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
	priv, _, _ := testdata.KeyTestPubAddr()
	app := simapp.NewSimApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 0, encCfg, msgServiceOpt)

	msg := testdata.NewServiceMsgCreateDog(&testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}})

	txBuilder := encCfg.TxConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	txBuilder.SetMsgs(msg)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encCfg.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: 0,
	}

	err := txBuilder.SetSignatures(sigV2)
	require.NoError(t, err)

	// Second round: all signer infos are set, so each signer can sign.
	signerData := authsigning.SignerData{
		ChainID:       "test",
		AccountNumber: 0,
		Sequence:      0,
	}
	sigV2, err = tx.SignWithPrivKey(
		encCfg.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encCfg.TxConfig, 0)
	require.NoError(t, err)
	err = txBuilder.SetSignatures(sigV2)
	require.NoError(t, err)

	// Send the tx to the app
	txBytes, err := encCfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	res := app.BaseApp.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	fmt.Println("res=", res)
}
