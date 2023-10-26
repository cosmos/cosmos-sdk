package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/event"
	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/counter/types"
)

var StoreKey = "Counter"

type Keeper struct {
	storeService storetypes.KVStoreService
	event        event.Service

	authority  string
	CountStore collections.Item[int64]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, authority string, em event.Service) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return Keeper{
		storeService: storeService,
		authority:    authority,
		event:        em,
		CountStore:   collections.NewItem(sb, collections.NewPrefix("Count"), "count", collections.Int64Value),
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Querier

var _ types.QueryServer = Keeper{}

// Params queries params of consensus module
func (k Keeper) GetCount(ctx context.Context, _ *types.QueryCountRequest) (*types.QueryCountResponse, error) {
	count, err := k.CountStore.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryCountResponse{TotalCount: count}, nil
}

// MsgServer

var _ types.MsgServer = Keeper{}

func (k Keeper) IncreaseCount(ctx context.Context, msg *types.MsgCounter) (*types.MsgCountResponse, error) {
	value, err := k.CountStore.Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := k.CountStore.Set(ctx, value+msg.Count); err != nil {
		return nil, err
	}

	if err := k.event.EventManager(ctx).EmitKV(
		ctx,
		"increase_counter",
		event.Attribute{Key: "signer", Value: msg.Signer},
		event.Attribute{Key: "new count", Value: fmt.Sprint(value + msg.Count)}); err != nil {
		return nil, err
	}

	return &types.MsgCountResponse{}, nil
}
