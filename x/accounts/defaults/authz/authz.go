package authz

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/authz/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
func NewAccount(d accountstd.Dependencies) *Account {
	account := &Account{
		Granter:       collections.NewItem(d.SchemaBuilder, GranterPrefix, "granter", collections.BytesValue),
		Grantees:      collections.NewMap(d.SchemaBuilder, GranteePrefix, "grantees", collections.PairKeyCodec(collections.BytesKey, collections.StringKey), codec.CollValue[types.Grant](d.LegacyStateCodec)),
		addressCodec:  d.AddressCodec,
		headerService: d.Environment.HeaderService,
	}

	return account
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

	granteeAddr, err := a.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}

	authorization := msg.Grant.GetAuthorization()
	if authorization == nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAuthorization, "authorization cannot be nil")
	}

	authz, ok := authorization.GetCachedValue().(types.Authorization)
	if !ok {
		return nil, errorsmod.Wrapf(types.ErrInvalidAuthorization, "authorization not impelementing interface")
	}

	err = a.Grantees.Set(ctx, collections.Join(granteeAddr, authz.MsgTypeURL()), msg.Grant)
	if err != nil {
		return nil, err
	}
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
			msgTypeUrl := sdk.MsgTypeURL(msg)
			grant, err := a.Grantees.Get(ctx, collections.Join(senderAddr, msgTypeUrl))
			if err != nil {
				return nil, errorsmod.Wrapf(types.ErrInvalidAuthorization,
					"failed to get grant with given granter: %s, grantee: %s & msgType: %s ", sdk.AccAddress(granter), msg.Sender, sdk.MsgTypeURL(msg))
			}

			if grant.Expiration != nil && grant.Expiration.Before(currentTime) {
				return nil, types.ErrAuthorizationExpired
			}

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
}

func (a Account) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
}
