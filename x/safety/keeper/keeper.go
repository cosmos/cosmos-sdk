package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

var _ baseapp.CircuitBreaker = Keeper{}

// Keeper defines the x/safety keeper.
type Keeper struct {
	addressCodec address.Codec
	registry     codectypes.InterfaceRegistry

	// defaultAuthority is written into the default genesis params. The live
	// authority is params.Authority.
	defaultAuthority string

	Schema collections.Schema
	// Params stores the authority, operator and operator scope.
	Params collections.Item[types.Params]
	// DisabledList stores the disabled patterns.
	DisabledList collections.KeySet[string]
}

// NewKeeper creates a new x/safety keeper. registry is used to check that a
// tripped pattern matches at least one registered msg type. defaultAuthority
// is the authority written into the default genesis; it must be a valid
// address in canonical form, or NewKeeper panics.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	addressCodec address.Codec,
	registry codectypes.InterfaceRegistry,
	defaultAuthority string,
) Keeper {
	if err := types.DefaultParams(defaultAuthority).Validate(addressCodec); err != nil {
		panic(err)
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		addressCodec:     addressCodec,
		registry:         registry,
		defaultAuthority: defaultAuthority,
		Params:           collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		DisabledList:     collections.NewKeySet(sb, types.DisabledListKey, "disabled_list", collections.StringKey),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// AddressCodec returns the account address codec.
func (k Keeper) AddressCodec() address.Codec { return k.addressCodec }

// DefaultAuthority returns the authority written into the default genesis.
func (k Keeper) DefaultAuthority() string { return k.defaultAuthority }

// IsAllowed reports whether a msg type URL may execute. It implements
// baseapp.CircuitBreaker. Protected msg types are always allowed. Otherwise
// the msg is blocked when the disabled list contains its exact type URL, or,
// unless the msg is a wrapper such as authz MsgExec, "*" or any of its
// segment-boundary prefix patterns. The check is a bounded number of point
// lookups and never walks the disabled list.
//
// The check charges no gas. The lookups run on an infinite gas meter that
// replaces ctx's meter for the check only, so the tx, the block gas meter
// and the caller's ctx see none of them, and a msg costs the same gas as with
// no circuit breaker set. The router, the SafetyDecorator and
// msgservice.Dispatch all reach the store through here. The work stays
// bounded without gas: per msg, at most two lookups plus one per package
// segment of its type URL, and a type URL is either one the interface
// registry knows or, through the IsAllowed query, at most MaxPatternLength
// bytes.
//
// A blocked msg returns false with ErrMsgDisabled, so the router, which
// passes the error on as is, rejects it under the safety codespace. A store
// error also returns false: the breaker fails closed.
func (k Keeper) IsAllowed(ctx context.Context, typeURL string) (bool, error) {
	if types.IsProtected(typeURL) {
		return true, nil
	}

	free := sdk.UnwrapSDKContext(ctx).WithGasMeter(storetypes.NewInfiniteGasMeter())
	by, err := k.disabledCover(free, typeURL)
	if err != nil {
		return false, err
	}
	if by != "" {
		return false, types.ErrMsgDisabled.Wrapf("%s by %s", typeURL, by)
	}
	return true, nil
}

// disabledCover returns the disabled entry that blocks the msg type URL
// typeURL, or "" when none does. It runs one point lookup per blocking
// pattern.
func (k Keeper) disabledCover(ctx context.Context, typeURL string) (string, error) {
	for c := range types.BlockingPatterns(typeURL) {
		has, err := k.DisabledList.Has(ctx, c)
		if err != nil {
			return "", err
		}
		if has {
			return c, nil
		}
	}
	return "", nil
}

// matchesUnprotectedMsg reports whether pattern would block at least one msg
// type URL registered in the interface registry that is not protected. A
// wrapper counts only for its exact URL.
func (k Keeper) matchesUnprotectedMsg(pattern string) bool {
	for _, typeURL := range k.registry.ListImplementations(sdk.MsgInterfaceProtoName) {
		if !types.IsProtected(typeURL) && types.Blocks(pattern, typeURL) {
			return true
		}
	}
	return false
}
