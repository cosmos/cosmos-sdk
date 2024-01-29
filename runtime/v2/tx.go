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
	msgs := make([]sdk.Msg, len(w.Tx.GetMessages()))
	for i, msg := range w.Tx.GetMessages() {
		msg := dynamicpb.NewMessage(msg.ProtoReflect().Descriptor())
		proto.Merge(msg, msg.ProtoReflect().Interface())
		msgs[i] = msg
	}

	return msgs
}

func (w wrapperServerTx) GetMsgsV2() ([]proto.Message, error) {
	return w.Tx.GetMessages(), nil
}
