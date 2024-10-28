package authz

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/authz/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var (
	AUTHZ_ACCOUNT = "authz_account"
	purge_limit   = 75
)

var (
	GranteePrefix = collections.NewPrefix(0)
	GranterPrefix = collections.NewPrefix(1)
)

type Account struct {
	Grantees      collections.Map[collections.Pair[[]byte, string], types.Grant]
	Granter       collections.Item[[]byte]
	addressCodec  address.Codec
	headerService header.Service
}

// NewAccount creates a new Account object.
func NewAccount(d accountstd.Dependencies) (*Account, error) {
	account := &Account{
		Granter:       collections.NewItem(d.SchemaBuilder, GranterPrefix, "granter", collections.BytesValue),
		Grantees:      collections.NewMap(d.SchemaBuilder, GranteePrefix, "grantees", collections.PairKeyCodec(collections.BytesKey, collections.StringKey), codec.CollValue[types.Grant](d.LegacyStateCodec)),
		addressCodec:  d.AddressCodec,
		headerService: d.Environment.HeaderService,
	}

	return account, nil
}

func (a *Account) Init(ctx context.Context, msg *types.MsgInitAuthzAccount) (
	*types.MsgInitAuthzAccountResponse, error,
) {
	granterAddr, err := a.addressCodec.StringToBytes(msg.Granter)
	if err != nil {
		return nil, err
	}
	err = a.Granter.Set(ctx, granterAddr)
	if err != nil {
		return nil, err
	}
	return &types.MsgInitAuthzAccountResponse{}, nil
}

func (a *Account) Grant(ctx context.Context, msg *types.MsgGrant) (
	*types.MsgGrantResponse, error,
) {
	granterAddr, err := a.addressCodec.StringToBytes(msg.Granter)
	if err != nil {
		return nil, err
	}

	accountGranter, err := a.Granter.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(granterAddr, accountGranter) {
		return nil, errorsmod.Wrapf(types.ErrUnauthorizedAction,
			"invalid granter address")
	}

	currentTime := a.headerService.HeaderInfo(ctx).Time
	if currentTime.After(*msg.Grant.Expiration) {
		return nil, types.ErrInvalidExpirationTime
	}

	granteeAddr, err := a.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}

	authorizationAny := msg.Grant.GetAuthorization()
	if authorizationAny == nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAuthorization, "authorization cannot be nil")
	}

	authorization, ok := authorizationAny.GetCachedValue().(types.Authorization)
	if !ok {
		return nil, errorsmod.Wrapf(types.ErrInvalidAuthorization, "authorization not impelementing interface")
	}

	msgTypeUrl := authorization.MsgTypeURL()

	err = a.Grantees.Set(ctx, collections.Join(granteeAddr, msgTypeUrl), msg.Grant)
	if err != nil {
		return nil, err
	}

	grant, err := a.Grantees.Get(ctx, collections.Join(granteeAddr, msgTypeUrl))
	if err != nil {
		return nil, err
	}

	fmt.Println("*", grant.Authorization.GetCachedValue())

	return &types.MsgGrantResponse{}, nil
}

func (a *Account) Exec(ctx context.Context, msg *types.MsgExec) (
	*types.MsgExecResponse, error,
) {
	senderAddr, err := a.addressCodec.StringToBytes(msg.Sender)
	if err != nil {
		return nil, err
	}

	granter, err := a.Granter.Get(ctx)
	if err != nil {
		return nil, err
	}

	// validate messages
	sdkmsgs, err := msg.GetMessages()
	if err != nil {
		return nil, err
	}
	err = validateMsgs(sdkmsgs)
	if err != nil {
		return nil, err
	}

	currentTime := a.headerService.HeaderInfo(ctx).Time

	executeMsgs := make([]*codectypes.Any, len(msg.Msgs))
	// If sender is not the granter then we check for the sender grant for each msg type urls
	// we dont need to check signer field for the messages here since that will be check in the
	// accountstd.ExecModuleAnys
	if !bytes.Equal(granter, senderAddr) {
		for _, msgany := range msg.Msgs {
			sdkMsg, ok := msgany.GetCachedValue().(transaction.Msg)
			if !ok {
				return nil, errors.New("failed to extract transaction message, got %v")
			}
			msgTypeUrl := sdk.MsgTypeURL(sdkMsg)
			grant, err := a.Grantees.Get(ctx, collections.Join(senderAddr, msgTypeUrl))
			if err != nil {
				return nil, errorsmod.Wrapf(types.ErrInvalidAuthorization,
					"failed to get grant with given granter: %s, grantee: %s & msgType: %s ", granter, msg.Sender, msgTypeUrl)
			}

			if grant.Expiration != nil && grant.Expiration.Before(currentTime) {
				return nil, types.ErrAuthorizationExpired
			}

			fmt.Println(grant.Authorization.GetCachedValue())

			authorization, err := grant.GetGrantAuthorization()
			if err != nil {
				return nil, err
			}

			resp, err := authorization.Accept(ctx, msg)
			if err != nil {
				return nil, err
			}

			if resp.Delete {
				err = a.Grantees.Remove(ctx, collections.Join(senderAddr, msgTypeUrl))
				if err != nil {
					return nil, err
				}
			} else if resp.Updated != nil {
				updated, ok := resp.Updated.(types.Authorization)
				if !ok {
					return nil, fmt.Errorf("should implement Authorization interface but got %T", resp.Updated)
				}

				updatedAny, err := accountstd.PackAny(updated)
				if err != nil {
					return nil, err
				}
				grant.Authorization = updatedAny

				err = a.Grantees.Set(ctx, collections.Join(senderAddr, msgTypeUrl), grant)
				if err != nil {
					return nil, err
				}
			}

			if !resp.Accept {
				return nil, sdkerrors.ErrUnauthorized
			}

			executeMsgs = append(executeMsgs, msgany)
		}
	} else {
		executeMsgs = append(executeMsgs, msg.Msgs...)
	}

	// execute messages
	resps, err := accountstd.ExecModuleAnys(ctx, executeMsgs)
	if err != nil {
		return nil, fmt.Errorf("failed to execute message; message %v: %w", msg, err)
	}

	return &types.MsgExecResponse{Responds: resps}, nil
}

func (a *Account) Revoke(ctx context.Context, msg *types.MsgRevoke) (*types.MsgRevokeResponse, error) {
	granterAddr, err := a.addressCodec.StringToBytes(msg.Granter)
	if err != nil {
		return nil, err
	}

	granteeAddr, err := a.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}

	accountGranter, err := a.Granter.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(granterAddr, accountGranter) {
		return nil, errorsmod.Wrapf(types.ErrUnauthorizedAction,
			"invalid granter address")
	}

	_, err = a.Grantees.Get(ctx, collections.Join(granteeAddr, msg.MsgTypeUrl))
	if err != nil {
		return nil, err
	}

	err = a.Grantees.Remove(ctx, collections.Join(granteeAddr, msg.MsgTypeUrl))
	return &types.MsgRevokeResponse{}, err
}

func (a *Account) RevokeAll(ctx context.Context, msg *types.MsgRevokeAll) (*types.MsgRevokeAllResponse, error) {
	granterAddr, err := a.addressCodec.StringToBytes(msg.Granter)
	if err != nil {
		return nil, err
	}

	accountGranter, err := a.Granter.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(granterAddr, accountGranter) {
		return nil, errorsmod.Wrapf(types.ErrUnauthorizedAction,
			"invalid granter address")
	}

	// if grantee address is provided then only revoke all grants for that specific address
	var granteeAddr []byte
	var rng *collections.PairRange[[]byte, string]
	if msg.Grantee != "" {
		granteeAddr, err = a.addressCodec.StringToBytes(msg.Grantee)
		if err != nil {
			return nil, err
		}

		rng = collections.NewPrefixedPairRange[[]byte, string](granteeAddr)
	}

	err = a.Grantees.Clear(ctx, rng)
	return &types.MsgRevokeAllResponse{}, err
}

func (a *Account) PurgeExpiredGrants(ctx context.Context, msg *types.MsgPurgeExpiredGrants) (*types.MsgPurgeExpiredGrantsResponse, error) {
	currentTime := a.headerService.HeaderInfo(ctx).Time
	count := 0
	err := a.Grantees.Walk(ctx, nil, func(key collections.Pair[[]byte, string], value types.Grant) (stop bool, err error) {
		if value.Expiration.Before(currentTime) {
			err = a.Grantees.Remove(ctx, key)
			if err != nil {
				return true, fmt.Errorf("faild to remove key %v: %w", key, err)
			}
		}
		count++
		if count == purge_limit {
			return true, nil
		}

		return false, nil
	})
	return &types.MsgPurgeExpiredGrantsResponse{}, err
}

func (a Account) QueryAllGrants(ctx context.Context, req *types.QueryGrantsRequest) (*types.QueryGrantsResponse, error) {
	grants, pageResp, err := query.CollectionPaginate(
		ctx,
		a.Grantees,
		req.Pagination,
		func(key collections.Pair[[]byte, string], value types.Grant) (*types.Grant, error) {
			return &value, nil
		},
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryGrantsResponse{
		Grants:     grants,
		Pagination: pageResp,
	}, nil
}

func (a Account) QueryGranteeGrants(ctx context.Context, req *types.QueryGranteeGrantsRequest) (*types.QueryGranteeGrantsResponse, error) {
	granteeAddr, err := a.addressCodec.StringToBytes(req.Grantee)
	if err != nil {
		return nil, err
	}

	grants, pageResp, err := query.CollectionPaginate(
		ctx,
		a.Grantees,
		req.Pagination,
		func(key collections.Pair[[]byte, string], value types.Grant) (*types.Grant, error) {
			return &value, nil
		},
		query.WithCollectionPaginationPairPrefix[[]byte, string](granteeAddr),
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryGranteeGrantsResponse{
		Grants:     grants,
		Pagination: pageResp,
	}, nil
}

func validateMsgs(msgs []sdk.Msg) error {
	for i, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return errorsmod.Wrapf(err, "msg %d", i)
		}
	}

	return nil
}

func (a Account) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

func (a Account) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.Grant)
	accountstd.RegisterExecuteHandler(builder, a.Exec)
	accountstd.RegisterExecuteHandler(builder, a.Revoke)
	accountstd.RegisterExecuteHandler(builder, a.RevokeAll)
	accountstd.RegisterExecuteHandler(builder, a.PurgeExpiredGrants)
}

func (a Account) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QueryAllGrants)
	accountstd.RegisterQueryHandler(builder, a.QueryGranteeGrants)
}
