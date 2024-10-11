package feegrant

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/event"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/feegrant/v1"
	"cosmossdk.io/x/accounts/internal/implementation"
	xfeegrant "cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Dependencies = implementation.Dependencies

var (
	// todo: prefix conflicts when embedded in other account type. bump +100
	PrefixGrants       = collections.NewPrefix(100)
	PrefixGrantsExpiry = collections.NewPrefix(101)
)
var (
	_ accountstd.Interface = (*Feegrant)(nil)
)

type Feegrant struct {
	env          appmodule.Environment
	addressCodec address.Codec
	feeAllowance collections.Map[sdk.AccAddress, v1.Grant]
	// FeeAllowanceQueue key: expiration time+grantee | value: bool
	feeAllowanceQueue collections.Map[collections.Pair[time.Time, sdk.AccAddress], bool]
}

// RegisterInitHandler implements implementation.Account.
func (f Feegrant) RegisterInitHandler(builder *implementation.InitBuilder) {
	accountstd.RegisterInitHandler(builder, f.Init)
}

func (f Feegrant) RegisterQueryHandlers(builder *implementation.QueryBuilder) {
}

func NewAccount(d Dependencies) (*Feegrant, error) {
	f := &Feegrant{
		env:               d.Environment,
		addressCodec:      d.AddressCodec,
		feeAllowance:      collections.NewMap(d.SchemaBuilder, PrefixGrants, "fee_allowance", sdk.AccAddressKey, codec.CollValue[v1.Grant](d.LegacyStateCodec)),
		feeAllowanceQueue: collections.NewMap(d.SchemaBuilder, PrefixGrantsExpiry, "fee_allowance_expiry", collections.PairKeyCodec(sdk.TimeKey, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), collections.BoolValue),
	}

	return f, nil
}

func (f Feegrant) Init(ctx context.Context, msg *v1.MsgInit) (*v1.MsgInitResponse, error) {
	return &v1.MsgInitResponse{}, nil
}

func (f *Feegrant) GrantAllowance(ctx context.Context, msg *v1.MsgGrantAllowance) (*v1.MsgGrantAllowanceResponse, error) {
	grantee, err := f.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}

	if f, _ := f.GetAllowance(ctx, grantee); f != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "fee allowance already exists")
	}

	allowance, err := msg.GetFeeAllowanceI()
	if err != nil {
		return nil, err
	}
	exp, err := allowance.ExpiresAt()
	if err != nil {
		return nil, err
	}

	// expiration shouldn't be in the past.

	now := f.env.HeaderService.HeaderInfo(ctx).Time
	if exp != nil && exp.Before(now) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "expiration is before current block time")
	}

	// if expiry is not nil, add the new key to pruning queue.
	if exp != nil {
		err = f.feeAllowanceQueue.Set(ctx, collections.Join(*exp, sdk.AccAddress(grantee)), true)
		if err != nil {
			return nil, err
		}
	}
	granter := accountstd.Whoami(ctx)
	granterStr, err := f.addressCodec.BytesToString(granter)
	if err != nil {
		return nil, err
	}
	granteeStr := msg.Grantee

	// if block time is not zero, update the period reset
	// if it is zero, it could be genesis initialization, so we don't need to update the period reset
	if !now.IsZero() {
		if err = allowance.UpdatePeriodReset(now); err != nil {
			return nil, err
		}
	}

	grant, err := v1.NewGrant(allowance)
	if err != nil {
		return nil, err
	}
	if err := f.feeAllowance.Set(ctx, grantee, *grant); err != nil {
		return nil, err
	}

	return &v1.MsgGrantAllowanceResponse{}, f.env.EventService.EventManager(ctx).EmitKV(
		xfeegrant.EventTypeSetFeeGrant,
		event.NewAttribute(xfeegrant.AttributeKeyGranter, granterStr),
		event.NewAttribute(xfeegrant.AttributeKeyGrantee, granteeStr),
	)
}

func (f Feegrant) GetAllowance(ctx context.Context, grantee sdk.AccAddress) (xfeegrant.FeeAllowanceI, error) {
	grant, err := f.feeAllowance.Get(ctx, grantee)
	if err != nil {
		return nil, err
	}

	return grant.GetFeeAllowanceI()
}

func (f Feegrant) UseGrantedFees(ctx context.Context, msg *v1.MsgUseGrantedFees) (*v1.MsgUseGrantedFeesResponse, error) {
	msgs, err := msg.GetMessages()
	if err != nil {
		return nil, err
	}
	grantee, err := f.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}
	err = f.DoUseGrantedFees(ctx, grantee, msg.GetFees(), msgs)
	if err != nil {
		return nil, err
	}
	return &v1.MsgUseGrantedFeesResponse{}, nil
}

func (f Feegrant) DoUseGrantedFees(ctx context.Context, grantee sdk.AccAddress, fees sdk.Coins, msgs []sdk.Msg) error {
	grant, err := f.GetAllowance(ctx, grantee)
	if err != nil {
		return err
	}

	granter := accountstd.Whoami(ctx)
	granterStr, err := f.addressCodec.BytesToString(granter)
	if err != nil {
		return err
	}
	granteeStr, err := f.addressCodec.BytesToString(grantee)
	if err != nil {
		return err
	}

	remove, err := grant.Accept(context.WithValue(ctx, corecontext.EnvironmentContextKey, f.env), fees, msgs)
	if remove && err == nil {
		// Ignoring the `revokeFeeAllowance` error, because the user has enough feeAllowance to perform this transaction.
		_ = f.revokeAllowance(ctx, granter, grantee)

		return f.emitUseGrantEvent(ctx, granterStr, granteeStr)
	}
	if err != nil {
		return err
	}
	if err := f.emitUseGrantEvent(ctx, granterStr, granteeStr); err != nil {
		return err
	}

	// if fee allowance is accepted, store the updated state of the allowance
	return f.UpdateAllowance(ctx, granter, grantee, grant)
}

func (f Feegrant) emitUseGrantEvent(ctx context.Context, granter, grantee string) error {
	return f.env.EventService.EventManager(ctx).EmitKV(
		xfeegrant.EventTypeUseFeeGrant,
		event.NewAttribute(xfeegrant.AttributeKeyGranter, granter),
		event.NewAttribute(xfeegrant.AttributeKeyGrantee, grantee),
	)
}

func (f Feegrant) UpdateAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance xfeegrant.FeeAllowanceI) error {
	_, err := f.GetAllowance(ctx, grantee)
	if err != nil {
		return err
	}

	granterStr, err := f.addressCodec.BytesToString(granter)
	if err != nil {
		return err
	}
	granteeStr, err := f.addressCodec.BytesToString(grantee)
	if err != nil {
		return err
	}

	grant, err := v1.NewGrant(feeAllowance)
	if err != nil {
		return err
	}

	if err := f.feeAllowance.Set(ctx, grantee, *grant); err != nil {
		return err
	}

	return f.env.EventService.EventManager(ctx).EmitKV(
		xfeegrant.EventTypeUpdateFeeGrant,
		event.NewAttribute(xfeegrant.AttributeKeyGranter, granterStr),
		event.NewAttribute(xfeegrant.AttributeKeyGrantee, granteeStr),
	)
}

// revokeAllowance removes an existing grant
func (f Feegrant) revokeAllowance(ctx context.Context, granter, grantee sdk.AccAddress) error {
	grant, err := f.GetAllowance(ctx, grantee)
	if err != nil {
		return err
	}

	if err := f.feeAllowance.Remove(ctx, grantee); err != nil {
		return err
	}

	exp, err := grant.ExpiresAt()
	if err != nil {
		return err
	}

	if exp != nil {
		if err := f.feeAllowanceQueue.Remove(ctx, collections.Join(*exp, grantee)); err != nil {
			return err
		}
	}

	granterStr, err := f.addressCodec.BytesToString(granter)
	if err != nil {
		return err
	}
	granteeStr, err := f.addressCodec.BytesToString(grantee)
	if err != nil {
		return err
	}

	return f.env.EventService.EventManager(ctx).EmitKV(
		xfeegrant.EventTypeRevokeFeeGrant,
		event.NewAttribute(xfeegrant.AttributeKeyGranter, granterStr),
		event.NewAttribute(xfeegrant.AttributeKeyGrantee, granteeStr),
	)
}

func (f *Feegrant) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, f.Init)
	accountstd.RegisterExecuteHandler(builder, f.GrantAllowance)
	accountstd.RegisterExecuteHandler(builder, f.UseGrantedFees)
}
