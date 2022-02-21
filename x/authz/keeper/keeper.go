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

// TODO: Revisit this once we have propoer gas fee framework.
// Tracking issues https://github.com/cosmos/cosmos-sdk/issues/9054,
// https://github.com/cosmos/cosmos-sdk/discussions/9072
const gasCostPerIteration = uint64(20)

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
			grant, found := k.getGrant(ctx, grantStoreKey(grantee, granter, sdk.MsgTypeURL(msg)))
			if !found {
				return nil, sdkerrors.ErrUnauthorized.Wrap("authorization not found")
			}

			if grant.Expiration.Before(ctx.BlockTime()) {
				return nil, sdkerrors.ErrUnauthorized.Wrap("authorization expired")
			}

			authorization, err := grant.GetAuthorization()
			if err != nil {
				return nil, err
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
	oldGrant, found := k.getGrant(ctx, skey)

	grant, err := authz.NewGrant(ctx.BlockTime(), authorization, expiration)
	if err != nil {
		return err
	}

	bz := k.cdc.MustMarshal(&grant)
	store.Set(skey, bz)

	if found {
		// if expiration is not the same, remove old key and add the new key to queue
		if !oldGrant.Expiration.Equal(expiration) {
			if err := k.removeFromGrantQueue(ctx, skey, oldGrant.Expiration, granter, grantee); err != nil {
				return err
			}

			k.insertIntoGrantQueue(ctx, granter, grantee, authorization.MsgTypeURL(), expiration)
		}
	} else {
		k.insertIntoGrantQueue(ctx, granter, grantee, authorization.MsgTypeURL(), expiration)
	}

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
	if err := k.removeFromGrantQueue(ctx, skey, grant.Expiration, granter, grantee); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&authz.EventRevoke{
		MsgTypeUrl: msgType,
		Granter:    granter.String(),
		Grantee:    grantee.String(),
	})
}

// GetAuthorizations Returns list of `Authorizations` granted to the grantee by the granter.
func (k Keeper) GetAuthorizations(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress) ([]authz.Authorization, error) {
	store := ctx.KVStore(k.storeKey)
	key := grantStoreKey(grantee, granter, "")
	iter := sdk.KVStorePrefixIterator(store, key)
	defer iter.Close()

	var authorization authz.Grant
	var authorizations []authz.Authorization
	for ; iter.Valid(); iter.Next() {
		if err := k.cdc.Unmarshal(iter.Value(), &authorization); err != nil {
			return nil, err
		}

		a, err := authorization.GetAuthorization()
		if err != nil {
			return nil, err
		}

		authorizations = append(authorizations, a)
	}

	return authorizations, nil
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
		// ignore expired authorizations
		if entry.Expiration.Before(ctx.BlockTime()) {
			continue
		}

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

func (keeper Keeper) getGrantQueueItem(ctx sdk.Context, expiration time.Time, granter, grantee sdk.AccAddress) (*authz.GrantQueueItem, error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(GrantQueueKey(expiration, granter, grantee))
	if bz == nil {
		return &authz.GrantQueueItem{}, nil
	}

	var queueItems authz.GrantQueueItem
	if err := keeper.cdc.Unmarshal(bz, &queueItems); err != nil {
		return nil, err
	}
	return &queueItems, nil
}

func (k Keeper) setGrantQueueItem(ctx sdk.Context, expiration time.Time,
	granter sdk.AccAddress, grantee sdk.AccAddress, queueItems *authz.GrantQueueItem) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(queueItems)
	if err != nil {
		return err
	}
	store.Set(GrantQueueKey(expiration, granter, grantee), bz)

	return nil
}

// insertIntoGrantQueue inserts a grant key into the grant queue
func (keeper Keeper) insertIntoGrantQueue(ctx sdk.Context, granter, grantee sdk.AccAddress, msgType string,
	expiration time.Time) error {
	queueItems, err := keeper.getGrantQueueItem(ctx, expiration, granter, grantee)
	if err != nil {
		return err
	}

	if len(queueItems.MsgTypeUrls) == 0 {
		keeper.setGrantQueueItem(ctx, expiration, granter, grantee, &authz.GrantQueueItem{
			MsgTypeUrls: []string{msgType},
		})
	} else {
		queueItems.MsgTypeUrls = append(queueItems.MsgTypeUrls, msgType)
		keeper.setGrantQueueItem(ctx, expiration, granter, grantee, queueItems)
	}

	return nil
}

// removeFromGrantQueue removes a grant key from the grant queue
func (keeper Keeper) removeFromGrantQueue(ctx sdk.Context, grantKey []byte, expiration time.Time, granter, grantee sdk.AccAddress) error {
	store := ctx.KVStore(keeper.storeKey)
	key := GrantQueueKey(expiration, granter, grantee)
	bz := store.Get(key)
	if bz == nil {
		return sdkerrors.ErrLogic.Wrap("grant key not found")
	}

	var queueItem authz.GrantQueueItem
	if err := keeper.cdc.Unmarshal(bz, &queueItem); err != nil {
		return err
	}

	_, _, msgType := parseGrantStoreKey(grantKey)
	queueItems := queueItem.MsgTypeUrls

	for index, typeUrl := range queueItems {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "grant queue")

		if typeUrl == msgType {
			end := len(queueItem.MsgTypeUrls) - 1
			queueItems[index] = queueItems[end]
			queueItems = queueItems[:end]

			if err := keeper.setGrantQueueItem(ctx, expiration, granter, grantee, &authz.GrantQueueItem{
				MsgTypeUrls: queueItems,
			}); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// DequeueAndDeleteExpiredGrants deletes expired grants from the state and grant queue.
func (k Keeper) DequeueAndDeleteExpiredGrants(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(GrantQueuePrefix, sdk.InclusiveEndBytes(GrantQueueTimePrefix(ctx.BlockTime())))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var queueItem authz.GrantQueueItem
		if err := k.cdc.Unmarshal(iterator.Value(), &queueItem); err != nil {
			return err
		}

		_, granter, grantee, err := parseGrantQueueKey(iterator.Key())
		if err != nil {
			return err
		}

		store.Delete(iterator.Key())

		for _, typeUrl := range queueItem.MsgTypeUrls {
			store.Delete(grantStoreKey(grantee, granter, typeUrl))
		}
	}

	return nil
}
