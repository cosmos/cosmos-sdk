package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/x/feemarket/types"
)

// Keeper is the x/feemarket keeper.
type Keeper struct {
	appmodule.Environment

	cdc      codec.BinaryCodec
	ak       types.AccountKeeper
	resolver types.DenomResolver

	enabledHeight collections.Item[int64]
	state         collections.Item[types.State]
	params        collections.Item[types.Params]

	// The address that is capable of executing a MsgParams message.
	// Typically, this will be the governance module's address.
	authority string
}

// NewKeeper constructs a new feemarket keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	env appmodule.Environment,
	authKeeper types.AccountKeeper,
	resolver types.DenomResolver,
	authority string,
) *Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	sb := collections.NewSchemaBuilder(env.KVStoreService)

	k := &Keeper{
		cdc:           cdc,
		Environment:   env,
		ak:            authKeeper,
		resolver:      resolver,
		authority:     authority,
		enabledHeight: collections.NewItem(sb, types.KeyEnabledHeight, "enabled_height", collections.Int64Value),
		state:         collections.NewItem(sb, types.KeyState, "state", codec.CollValue[types.State](cdc)),
		params:        collections.NewItem(sb, types.KeyParams, "params", codec.CollValue[types.Params](cdc)),
	}

	return k
}

// GetAuthority returns the address that is capable of executing a MsgUpdateParams message.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetEnabledHeight returns the height at which the feemarket was enabled.
func (k *Keeper) GetEnabledHeight(ctx context.Context) (int64, error) {
	bz, err := k.enabledHeight.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return -1, nil
		} else {
			return -1, err
		}
	}
	return bz, err
}

// SetEnabledHeight sets the height at which the feemarket was enabled.
func (k *Keeper) SetEnabledHeight(ctx context.Context, height int64) error {
	if err := k.enabledHeight.Set(ctx, height); err != nil {
		return err
	}

	return nil
}

// ResolveToDenom converts the given coin to the given denomination.
func (k *Keeper) ResolveToDenom(ctx context.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error) {
	if k.resolver == nil {
		return sdk.DecCoin{}, types.ErrResolverNotSet
	}

	return k.resolver.ConvertToDenom(ctx, coin, denom)
}

// SetDenomResolver sets the keeper's denom resolver.
func (k *Keeper) SetDenomResolver(resolver types.DenomResolver) {
	k.resolver = resolver
}

// GetState returns the feemarket module's state.
func (k *Keeper) GetState(ctx context.Context) (types.State, error) {
	state, err := k.state.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.State{}, nil
		} else {
			return types.State{}, status.Error(codes.Internal, err.Error())
		}
	}

	return state, nil
}

// SetState sets the feemarket module's state.
func (k *Keeper) SetState(ctx context.Context, state types.State) error {
	return k.state.Set(ctx, state)
}

// GetParams returns the feemarket module's parameters.
func (k *Keeper) GetParams(ctx context.Context) (types.Params, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.Params{}, nil
		} else {
			return types.Params{}, status.Error(codes.Internal, err.Error())
		}
	}

	return params, nil
}

// SetParams sets the feemarket module's parameters.
func (k *Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}
