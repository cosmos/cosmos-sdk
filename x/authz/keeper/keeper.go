package keeper

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	router     *middleware.MsgServiceRouter
	authKeeper authkeeper.AccountKeeper
}

// NewKeeper constructs a message authorization Keeper
func NewKeeper(storeKey storetypes.StoreKey, cdc codec.BinaryCodec, router *middleware.MsgServiceRouter, ak authkeeper.AccountKeeper) Keeper {
	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		router:     router,
		authKeeper: ak,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", authz.ModuleName))
}

// getGrant returns grant stored at skey.
func (k Keeper) getGrant(ctx sdk.Context, skey []byte) (grant authz.Grant, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(skey)
	if bz == nil {
		return grant, false
	}
	k.cdc.MustUnmarshal(bz, &grant)
	return grant, true
}

func (k Keeper) update(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, updated authz.Authorization) error {
	skey := grantStoreKey(grantee, granter, updated.MsgTypeURL())
	grant, found := k.getGrant(ctx, skey)
	if !found {
		return sdkerrors.ErrNotFound.Wrap("authorization not found")
	}

	msg, ok := updated.(proto.Message)
	if !ok {
		sdkerrors.ErrPackAny.Wrapf("cannot proto marshal %T", updated)
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}

	grant.Authorization = any
	store := ctx.KVStore(k.storeKey)
	store.Set(skey, k.cdc.MustMarshal(&grant))
	return nil
}

// DispatchActions attempts to execute the provided messages via authorization
// grants from the message signer to the grantee.
func (k Keeper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) ([][]byte, error) {
	var results = make([][]byte, len(msgs))
	for i, msg := range msgs {
		signers := msg.GetSigners()
		if len(signers) != 1 {
			return nil, sdkerrors.ErrInvalidRequest.Wrap("authorization can be given to msg with only one signer")
		}
		granter := signers[0]

		// if granter != grantee then check authorization.Accept, otherwise we implicitly accept.
		if !granter.Equals(grantee) {
			authorization, _ := k.GetCleanAuthorization(ctx, grantee, granter, sdk.MsgTypeURL(msg))
			if authorization == nil {
				return nil, sdkerrors.ErrUnauthorized.Wrap("authorization not found")
			}
			resp, err := authorization.Accept(ctx, msg)
			if err != nil {
				return nil, err
			}
			if resp.Delete {
				err = k.DeleteGrant(ctx, grantee, granter, sdk.MsgTypeURL(msg))
			} else if resp.Updated != nil {
				err = k.update(ctx, grantee, granter, resp.Updated)
			}
			if err != nil {
				return nil, err
			}
			if !resp.Accept {
				return nil, sdkerrors.ErrUnauthorized
			}
		}

		handler := k.router.Handler(msg)
		if handler == nil {
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized message route: %s", sdk.MsgTypeURL(msg))
		}

		msgResp, err := handler(ctx, msg)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to execute message; message %v", msg)
		}
		results[i] = msgResp.Data

		// emit the events from the dispatched actions
		events := msgResp.Events
		sdkEvents := make([]sdk.Event, 0, len(events))
		for i := 0; i < len(events); i++ {
			sdkEvents = append(sdkEvents, sdk.Event(events[i]))
		}
		ctx.EventManager().EmitEvents(sdkEvents)
	}

	return results, nil
}

// SaveGrant method grants the provided authorization to the grantee on the granter's account
// with the provided expiration time and insert authorization key into the grants queue. If there is an existing authorization grant for the
// same `sdk.Msg` type, this grant overwrites that.
func (k Keeper) SaveGrant(ctx sdk.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration time.Time) error {
	store := ctx.KVStore(k.storeKey)
	skey := grantStoreKey(grantee, granter, authorization.MsgTypeURL())

	grant, found := k.getGrant(ctx, skey)
	// remove old grant key from the grant queue
	if found {
		k.removeFromGrantQueue(ctx, skey, grant.Expiration)
	}

	grant, err := authz.NewGrant(authorization, expiration)
	if err != nil {
		return err
	}

	bz := k.cdc.MustMarshal(&grant)
	store.Set(skey, bz)
	k.insertIntoGrantQueue(ctx, granter, grantee, authorization.MsgTypeURL(), expiration)

	return ctx.EventManager().EmitTypedEvent(&authz.EventGrant{
		MsgTypeUrl: authorization.MsgTypeURL(),
		Granter:    granter.String(),
		Grantee:    grantee.String(),
	})
}

// DeleteGrant revokes any authorization for the provided message type granted to the grantee
// by the granter.
func (k Keeper) DeleteGrant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error {
	store := ctx.KVStore(k.storeKey)
	skey := grantStoreKey(grantee, granter, msgType)
	grant, found := k.getGrant(ctx, skey)
	if !found {
		return sdkerrors.ErrNotFound.Wrap("authorization not found")
	}

	store.Delete(skey)
	k.removeFromGrantQueue(ctx, skey, grant.Expiration)

	return ctx.EventManager().EmitTypedEvent(&authz.EventRevoke{
		MsgTypeUrl: msgType,
		Granter:    granter.String(),
		Grantee:    grantee.String(),
	})
}

// GetAuthorizations Returns list of `Authorizations` granted to the grantee by the granter.
func (k Keeper) GetAuthorizations(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress) (authorizations []authz.Authorization) {
	store := ctx.KVStore(k.storeKey)
	key := grantStoreKey(grantee, granter, "")
	iter := sdk.KVStorePrefixIterator(store, key)
	defer iter.Close()
	var authorization authz.Grant
	for ; iter.Valid(); iter.Next() {
		k.cdc.MustUnmarshal(iter.Value(), &authorization)
		authorizations = append(authorizations, authorization.GetAuthorization())
	}
	return authorizations
}

// GetCleanAuthorization returns an `Authorization` and it's expiration time for
// (grantee, granter, message name) grant. If there is no grant `nil` is returned.
// If the grant is expired, the grant is revoked, removed from the storage, and `nil` is returned.
func (k Keeper) GetCleanAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (cap authz.Authorization, expiration time.Time) {
	grant, found := k.getGrant(ctx, grantStoreKey(grantee, granter, msgType))
	if !found {
		return nil, time.Time{}
	}
	if grant.Expiration.Before(ctx.BlockHeader().Time) {
		k.DeleteGrant(ctx, grantee, granter, msgType)
		return nil, time.Time{}
	}

	return grant.GetAuthorization(), grant.Expiration
}

// IterateGrants iterates over all authorization grants
// This function should be used with caution because it can involve significant IO operations.
// It should not be used in query or msg services without charging additional gas.
func (k Keeper) IterateGrants(ctx sdk.Context,
	handler func(granterAddr sdk.AccAddress, granteeAddr sdk.AccAddress, grant authz.Grant) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, GrantKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var grant authz.Grant
		granterAddr, granteeAddr, _ := parseGrantStoreKey(iter.Key())
		k.cdc.MustUnmarshal(iter.Value(), &grant)
		if handler(granterAddr, granteeAddr, grant) {
			break
		}
	}
}

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *authz.GenesisState {
	var entries []authz.GrantAuthorization
	k.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant authz.Grant) bool {
		exp := grant.Expiration
		entries = append(entries, authz.GrantAuthorization{
			Granter:       granter.String(),
			Grantee:       grantee.String(),
			Expiration:    exp,
			Authorization: grant.Authorization,
		})
		return false
	})

	return authz.NewGenesisState(entries)
}

// InitGenesis new authz genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *authz.GenesisState) {
	for _, entry := range data.Authorization {
		grantee, err := sdk.AccAddressFromBech32(entry.Grantee)
		if err != nil {
			panic(err)
		}
		granter, err := sdk.AccAddressFromBech32(entry.Granter)
		if err != nil {
			panic(err)
		}
		a, ok := entry.Authorization.GetCachedValue().(authz.Authorization)
		if !ok {
			panic("expected authorization")
		}

		err = k.SaveGrant(ctx, grantee, granter, a, entry.Expiration)
		if err != nil {
			panic(err)
		}
	}
}

func (keeper Keeper) getGrantQueueItem(ctx sdk.Context, expiration time.Time) (*authz.GrantQueueItem, error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(GrantQueueKey(expiration))
	if bz == nil {
		return &authz.GrantQueueItem{}, nil
	}

	var queueItems authz.GrantQueueItem
	if err := keeper.cdc.Unmarshal(bz, &queueItems); err != nil {
		return nil, err
	}
	return &queueItems, nil
}

func (k Keeper) setGrantQueueItem(ctx sdk.Context, expiration time.Time, queueItems *authz.GrantQueueItem) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(queueItems)
	if err != nil {
		return err
	}
	store.Set(GrantQueueKey(expiration), bz)

	return nil
}

// insertIntoGrantQueue Inserts a grant key into the grant queue
func (keeper Keeper) insertIntoGrantQueue(ctx sdk.Context, granter, grantee sdk.AccAddress, msgType string,
	expiration time.Time) error {
	queueItems, err := keeper.getGrantQueueItem(ctx, expiration)
	if err != nil {
		return err
	}

	ggmTriple := authz.GrantStoreKey{
		Granter:    granter.String(),
		Grantee:    grantee.String(),
		MsgTypeUrl: msgType,
	}
	if len(queueItems.GgmTriples) == 0 {
		keeper.setGrantQueueItem(ctx, expiration, &authz.GrantQueueItem{
			GgmTriples: []*authz.GrantStoreKey{
				&ggmTriple,
			},
		})
	} else {
		queueItems.GgmTriples = append(queueItems.GgmTriples, &ggmTriple)
		keeper.setGrantQueueItem(ctx, expiration, queueItems)
	}

	return nil
}

// removeFromGrantQueue removes a grant key from the grant queue
func (keeper Keeper) removeFromGrantQueue(ctx sdk.Context, grantKey []byte, expiration time.Time) error {
	store := ctx.KVStore(keeper.storeKey)
	key := GrantQueueKey(expiration)
	bz := store.Get(key)
	if bz == nil {
		return sdkerrors.ErrLogic.Wrap("grant key not found")
	}

	var queueItem authz.GrantQueueItem
	if err := keeper.cdc.Unmarshal(bz, &queueItem); err != nil {
		return err
	}

	granter, grantee, msgType := parseGrantStoreKey(grantKey)
	queueItems := queueItem.GgmTriples

	for index, ggmTriple := range queueItems {
		if ggmTriple.Granter == granter.String() &&
			ggmTriple.Grantee == grantee.String() &&
			ggmTriple.MsgTypeUrl == msgType {
			end := len(queueItem.GgmTriples) - 1
			queueItems[index] = queueItems[end]
			queueItems = queueItems[:end]

			if err := keeper.setGrantQueueItem(ctx, expiration, &authz.GrantQueueItem{
				GgmTriples: queueItems,
			}); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// DequeueAllMatureGrants returns a concatenated list of all the queue items, and deletes them from the grant queue
func (k Keeper) DequeueAllMatureGrants(ctx sdk.Context) ([]*authz.GrantStoreKey, error) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(GrantQueuePrefix, sdk.InclusiveEndBytes(GrantQueueKey(ctx.BlockTime())))
	defer iterator.Close()

	var matureGrants []*authz.GrantStoreKey
	for ; iterator.Valid(); iterator.Next() {
		var queueItem authz.GrantQueueItem
		if err := k.cdc.Unmarshal(iterator.Value(), &queueItem); err != nil {
			return nil, err
		}

		matureGrants = append(matureGrants, queueItem.GgmTriples...)
		store.Delete(iterator.Key())
	}

	return matureGrants, nil
}

// DeleteExpiredGrants deletes expired grants from the state
func (k Keeper) DeleteExpiredGrants(ctx sdk.Context, grants []*authz.GrantStoreKey) error {
	store := ctx.KVStore(k.storeKey)

	for _, grant := range grants {
		granter, err := sdk.AccAddressFromBech32(grant.Granter)
		if err != nil {
			return err
		}

		grantee, err := sdk.AccAddressFromBech32(grant.Grantee)
		if err != nil {
			return err
		}

		store.Delete(grantStoreKey(grantee, granter, grant.MsgTypeUrl))
	}

	return nil
}
