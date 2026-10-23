package keeper

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeService store.KVStoreService

	CountStore collections.Item[int64]

	hooks types.CounterHooks
}

func NewKeeper(storeService store.KVStoreService) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return &Keeper{
		storeService: storeService,
		CountStore:   collections.NewItem(sb, collections.NewPrefix(0), "count", collections.Int64Value),
	}
}

// Querier

var _ types.QueryServer = Keeper{}

// GetCount queries the x/counter count
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

type msgServer struct {
	*Keeper
}

var _ types.MsgServer = msgServer{}

// newMsgServerImpl returns an implementation of the x/counter MsgServer
// interface for the provided Keeper.
func newMsgServerImpl(k *Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

// RegisterMsgServer registers the x/counter Msg service on r. The msg server
// is unexported: code outside the module runs a msg through the
// MsgServiceRouter with msgservice.Dispatch.
func RegisterMsgServer(r grpc.ServiceRegistrar, k *Keeper) {
	types.RegisterMsgServer(r, newMsgServerImpl(k))
}

func (k msgServer) IncreaseCount(ctx context.Context, msg *types.MsgIncreaseCounter) (*types.MsgIncreaseCountResponse, error) {
	var num int64
	num, err := k.CountStore.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			num = 0
		} else {
			return nil, err
		}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.CountStore.Set(sdkCtx, num+msg.Count); err != nil {
		return nil, err
	}

	if err := k.Hooks().AfterIncreaseCount(sdkCtx, num+msg.Count); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"increase_counter",
		sdk.NewAttribute("signer", msg.Signer),
		sdk.NewAttribute("new count", fmt.Sprint(num+msg.Count)),
	),
	)

	return &types.MsgIncreaseCountResponse{
		NewCount: num + msg.Count,
	}, nil
}

// Hooks gets the hooks for counter Keeper
func (k *Keeper) Hooks() types.CounterHooks {
	if k.hooks == nil {
		// return a no-op implementation if no hooks are set
		return types.MultiCounterHooks{}
	}

	return k.hooks
}

// SetHooks sets the hooks for counter
func (k *Keeper) SetHooks(gh types.CounterHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set governance hooks twice")
	}

	k.hooks = gh
	return k
}

// IncreaseCount runs MsgIncreaseCounter on the module's msg server.
//
// Deprecated: run MsgIncreaseCounter with msgservice.Dispatch.
// Keeper.IncreaseCount will be removed in a future release.
func (k Keeper) IncreaseCount(ctx context.Context, msg *types.MsgIncreaseCounter) (*types.MsgIncreaseCountResponse, error) {
	return newMsgServerImpl(&k).IncreaseCount(ctx, msg)
}
