package runtime

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"

	"cosmossdk.io/core/transaction"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ServerTxToSDKTx(tx transaction.Tx) sdk.Tx {
	return wrapperServerTx{tx}
}

var _ sdk.Tx = wrapperServerTx{}

type wrapperServerTx struct {
	transaction.Tx
}

func (w wrapperServerTx) GetMsgs() []sdk.Msg {
	txMsgs, err := w.Tx.GetMessages()
	if err != nil {
		panic(err)
	}
	msgs := make([]sdk.Msg, len(txMsgs))
	for i, msg := range txMsgs {
		msg := dynamicpb.NewMessage(msg.ProtoReflect().Descriptor())
		proto.Merge(msg, msg.ProtoReflect().Interface())
		msgs[i] = msg
	}

	return msgs
}

func (w wrapperServerTx) GetMsgsV2() ([]proto.Message, error) {
	txMsgs, err := w.Tx.GetMessages()
	if err != nil {
		return nil, err
	}
	return txMsgs, nil
}
