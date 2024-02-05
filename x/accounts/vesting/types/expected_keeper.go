package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/runtime/protoiface"
)

type ProtoMsg = protoiface.MessageV1

type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type AccountsKeeper interface {
	Init(ctx context.Context, accountType string, creator []byte, initRequest ProtoMsg) (ProtoMsg, []byte, error)
	Execute(ctx context.Context, accountAddr []byte, sender []byte, execRequest ProtoMsg) (ProtoMsg, error)
}
