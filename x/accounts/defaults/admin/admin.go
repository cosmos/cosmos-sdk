package admin

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/accounts/accountstd"
	admintypes "cosmossdk.io/x/accounts/defaults/admin/v1"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	OwnerPrefix = collections.NewPrefix(0)
	DenomPrefix = collections.NewPrefix(1)
	Type        = "admin"
)

// NewAdmin creates a new Admin object.
func NewAdmin(d accountstd.Dependencies) (*Admin, error) {
	Admin := &Admin{
		Owner:         collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		Denom:         collections.NewItem(d.SchemaBuilder, DenomPrefix, "denom", collections.StringValue),
		addressCodec:  d.AddressCodec,
		headerService: d.Environment.HeaderService,
	}

	return Admin, nil
}

type Admin struct {
	// Owner is the address of the account owner.
	Owner         collections.Item[[]byte]
	Denom         collections.Item[string]
	addressCodec  address.Codec
	headerService header.Service
}

func (ad *Admin) Init(ctx context.Context, msg *admintypes.MsgInit) (
	*admintypes.MsgInitResponse, error,
) {
	owner, err := ad.addressCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}
	err = ad.Owner.Set(ctx, owner)
	if err != nil {
		return nil, err
	}

	if msg.Denom == "" {
		return nil, errors.New("empty denom")
	}
	err = ad.Denom.Set(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}

	return &admintypes.MsgInitResponse{}, nil
}

func (ad *Admin) HaveMintPerm(ctx context.Context, msg *admintypes.QueryMintPerm) (
	*admintypes.QueryMintPermResponse, error,
) {
	denom, err := ad.Denom.Get(ctx)
	if err != nil {
		return nil, err
	}
	if denom != msg.Denom {
		return nil, fmt.Errorf("invalid denom, got %s, expected %s", msg.Denom, denom)
	}
	sender, err := ad.addressCodec.StringToBytes(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}
	owner, err := ad.Owner.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(sender, owner) {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("%s does not have mint permission", string(sender))
	}

	return &admintypes.QueryMintPermResponse{ShouldMint: true}, nil
}

func (ad *Admin) HaveBurnPerm(ctx context.Context, msg *admintypes.QueryBurnPerm) (
	*admintypes.QueryBurnPermResponse, error,
) {
	denom, err := ad.Denom.Get(ctx)
	if err != nil {
		return nil, err
	}
	if denom != msg.Denom {
		return nil, fmt.Errorf("invalid denom, got %s, expected %s", msg.Denom, denom)
	}
	sender, err := ad.addressCodec.StringToBytes(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}
	owner, err := ad.Owner.Get(ctx)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(sender, owner) {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("%s does not have burn permission", string(sender))
	}

	return &admintypes.QueryBurnPermResponse{ShouldBurn: true}, nil
}

func (ad *Admin) QueryOwner(ctx context.Context, msg *admintypes.QueryOwner) (
	*admintypes.QueryOwnerResponse, error,
) {
	owner, err := ad.Owner.Get(ctx)
	if err != nil {
		return nil, err
	}

	ownerStr, err := ad.addressCodec.BytesToString(owner)
	if err != nil {
		return nil, err
	}

	return &admintypes.QueryOwnerResponse{Owner: ownerStr}, nil
}

// RegisterExecuteHandlers implements implementation.Account.
func (a *Admin) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
}

// RegisterInitHandler implements implementation.Account.
func (a *Admin) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

// RegisterQueryHandlers implements implementation.Account.
func (a *Admin) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.HaveMintPerm)
	accountstd.RegisterQueryHandler(builder, a.HaveBurnPerm)
}
