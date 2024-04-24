package types

import (
	"context"
	time "time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/runtime/protoiface"
)

type ProtoMsg = protoiface.MessageV1

type GetLockedCoinsFunc = func(ctx context.Context, time time.Time, denoms ...string) (sdk.Coins, error)

type BaseAccount interface {
	Init(ctx context.Context, msg *MsgInitLockupAccount, amount sdk.Coins) (*MsgInitLockupAccountResponse, error)
	SendCoins(ctx context.Context, msg *MsgSend, getLockedCoinsFunc GetLockedCoinsFunc) (*MsgExecuteMessagesResponse, error)
	WithdrawUnlockedCoins(ctx context.Context, msg *MsgWithdraw, getLockedCoinsFunc GetLockedCoinsFunc) (*MsgWithdrawResponse, error)
	QueryAccountBaseInfo(ctx context.Context, req *QueryLockupAccountInfoRequest) (*QueryLockupAccountInfoResponse, error)

	GetEndTime() collections.Item[time.Time]
	GetHeaderService() header.Service
	GetOriginalFunds() collections.Map[string, math.Int]
}
