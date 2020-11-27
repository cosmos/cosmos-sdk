package keeper

import (
	"bytes"
	"fmt"
	"time"

	proto "github.com/gogo/protobuf/proto"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler
	router   *baseapp.MsgServiceRouter
}

// NewKeeper constructs a message authorization Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc codec.BinaryMarshaler, router *baseapp.MsgServiceRouter) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		router:   router,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// getAuthorizationGrant returns grant between granter and grantee for the given msg type
func (k Keeper) getAuthorizationGrant(ctx sdk.Context, actor []byte) (grant types.AuthorizationGrant, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(actor)
	if bz == nil {
		return grant, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &grant)
	return grant, true
}

func (k Keeper) update(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, updated types.Authorization) {
	actor := types.GetActorAuthorizationKey(grantee, granter, updated.MethodName())
	grant, found := k.getAuthorizationGrant(ctx, actor)
	if !found {
		return
	}

	msg, ok := updated.(proto.Message)
	if !ok {
		panic(fmt.Errorf("cannot proto marshal %T", updated))
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}

	grant.Authorization = any
	store := ctx.KVStore(k.storeKey)
	store.Set(actor, k.cdc.MustMarshalBinaryBare(&grant))
}

// DispatchActions attempts to execute the provided messages via authorization
// grants from the message signer to the grantee.
func (k Keeper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, serviceMsgs []sdk.ServiceMsg) (*sdk.Result, error) {
	var msgResult *sdk.Result
	var err error
	for _, serviceMsg := range serviceMsgs {
		signers := serviceMsg.GetSigners()
		if len(signers) != 1 {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "authorization can be given to msg with only one signer")
		}
		granter := signers[0]
		if !bytes.Equal(granter, grantee) {
			authorization, _ := k.GetAuthorization(ctx, grantee, granter, serviceMsg.MethodName)
			if authorization == nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "authorization not found")
			}
			allow, updated, del := authorization.Accept(serviceMsg, ctx.BlockHeader())
			if !allow {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "Requested amount is more than spent limit")
			}
			if del {
				k.Revoke(ctx, grantee, granter, serviceMsg.Type())
			} else if updated != nil {
				k.update(ctx, grantee, granter, updated)
			}
		}
		handler := k.router.Handler(serviceMsg.Route())

		if handler == nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s", serviceMsg.Route())
		}

		msgResult, err = handler(ctx, serviceMsg.Request)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to execute message; message %s", serviceMsg.MethodName)
		}
	}

	return msgResult, nil
}

// Grant method grants the provided authorization to the grantee on the granter's account with the provided expiration
// time. If there is an existing authorization grant for the same `sdk.Msg` type, this grant
// overwrites that.
func (k Keeper) Grant(ctx sdk.Context, grantee, granter sdk.AccAddress, authorization types.Authorization, expiration time.Time) {
	store := ctx.KVStore(k.storeKey)

	grant, err := types.NewAuthorizationGrant(authorization, expiration.Unix())
	if err != nil {
		panic(err)
	}

	bz := k.cdc.MustMarshalBinaryBare(&grant)
	actor := types.GetActorAuthorizationKey(grantee, granter, authorization.MethodName())
	store.Set(actor, bz)
}

// Revoke method revokes any authorization for the provided message type granted to the grantee by the granter.
func (k Keeper) Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error {
	store := ctx.KVStore(k.storeKey)
	actor := types.GetActorAuthorizationKey(grantee, granter, msgType)
	_, found := k.getAuthorizationGrant(ctx, actor)
	if !found {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "authorization not found")
	}
	store.Delete(actor)

	return nil
}

// GetAuthorizations Returns list of `Authorizations` granted to the grantee by the granter.
func (k Keeper) GetAuthorizations(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress) (authorizations []types.Authorization) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetActorAuthorizationKey(grantee, granter, "")
	iter := sdk.KVStorePrefixIterator(store, key)
	defer iter.Close()
	var authorization types.AuthorizationGrant
	for ; iter.Valid(); iter.Next() {
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &authorization)
		authorizations = append(authorizations, authorization.GetAuthorization())
	}
	return authorizations
}

// GetAuthorization Returns any `Authorization` (or `nil`), with the expiration time,
// granted to the grantee by the granter for the provided msg type.
func (k Keeper) GetAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (cap types.Authorization, expiration int64) {
	grant, found := k.getAuthorizationGrant(ctx, types.GetActorAuthorizationKey(grantee, granter, msgType))
	if !found {
		return nil, 0
	}
	if grant.Expiration != 0 && grant.Expiration < (ctx.BlockHeader().Time.Unix()) {
		k.Revoke(ctx, grantee, granter, msgType)
		return nil, 0
	}

	return grant.GetAuthorization(), grant.Expiration
}

// IterateGrants iterates over all authorization grants
func (k Keeper) IterateGrants(ctx sdk.Context,
	handler func(granterAddr sdk.AccAddress, granteeAddr sdk.AccAddress, grant types.AuthorizationGrant) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GrantKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var grant types.AuthorizationGrant
		granterAddr, granteeAddr := types.ExtractAddressesFromGrantKey(iter.Key())
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &grant)
		if handler(granterAddr, granteeAddr, grant) {
			break
		}
	}
}
