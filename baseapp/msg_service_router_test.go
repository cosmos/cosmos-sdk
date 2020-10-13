package baseapp

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func TestMsgService(t *testing.T) {
	msgServiceOpt := func(bapp *BaseApp) {
		testdata.RegisterMsgServer(
			bapp.MsgServiceRouter(),
			testdata.MsgImpl{},
		)
	}

	app := setupBaseApp(t, msgServiceOpt)

	app.InitChain(abci.RequestInitChain{})
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()

	msg, err := testdata.NewServiceMsgCreateDog(&testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}})
	require.NoError(t, err)
	tx := &txtypes.Tx{
		Body: &txtypes.TxBody{
			Messages: []*codectypes.Any{msg},
		},
	}
	txBytes, err := proto.Marshal(tx)
	require.NoError(t, err)

	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	fmt.Println(res.Data)
	require.Empty(t, res.Events)
	require.False(t, true)
}
