package client

import (
	tmrpc "github.com/tendermint/tendermint/rpc/client"
	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/client/reflection/tx"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

type TxBuilder interface {
	// AddMsg adds a message to the builder
	AddMsg(msg proto.Message)
	// SetFees sets the fees
	SetFees(fees sdktypes.Coins)
	// SetMemo sets the memo
	SetMemo(memo string)
}

type Tx struct {
	rpc tmrpc.Client
	tx.UnsignedBuilder
}

func NewTx() *Tx {
	return &Tx{}
}
