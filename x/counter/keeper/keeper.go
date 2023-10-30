package keeper

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/event"
	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/x/counter/types"
)

var StoreKey = "Counter"

type Keeper struct {
	event event.Service

	CountStore collections.Item[int64]
}

func NewKeeper(storeService storetypes.KVStoreService, em event.Service) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return Keeper{
		event:      em,
		CountStore: collections.NewItem(sb, collections.NewPrefix(0), "count", collections.Int64Value),
	}
}

// Querier

var _ types.QueryServer = Keeper{}

// Params queries params of consensus module
func (k Keeper) GetCount(ctx context.Context, _ *types.QueryGetCountRequest) (*types.QueryGetCountResponse, error) {
	count, err := k.CountStore.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &types.QueryGetCountResponse{TotalCount: 0}, nil
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &types.QueryGetCountResponse{TotalCount: count}, nil
}

// MsgServer

var _ types.MsgServer = Keeper{}

func (k Keeper) IncreaseCount(ctx context.Context, msg *types.MsgIncreaseCounter) (*types.MsgIncreaseCountResponse, error) {
	var num int64
	num, err := k.CountStore.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			num = 0
		} else {
			return nil, err
		}
	}
	if err := k.CountStore.Set(ctx, num+msg.Count); err != nil {
		return nil, err
	}

	if err := k.event.EventManager(ctx).EmitKV(
		ctx,
		"increase_counter",
		event.Attribute{Key: "signer", Value: msg.Signer},
		event.Attribute{Key: "new count", Value: fmt.Sprint(num + msg.Count)}); err != nil {
		return nil, err
	}

	return &types.MsgIncreaseCountResponse{
		NewCount: num + msg.Count,
	}, nil
}
