package feegrant

import (
	"context"
	"errors"
	"time"

	"cosmossdk.io/collections"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/event"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/extensions/feegrant/v1"
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

var _ implementation.MigrateableLegacyDataExtension = &Feegrant{}

type Feegrant struct {
	env          appmodule.Environment
	addressCodec addresscodec.Codec
	feeAllowance collections.Map[sdk.AccAddress, v1.Grant]
	// FeeAllowanceQueue key: expiration time+grantee | value: bool
	feeAllowanceQueue collections.Map[collections.Pair[time.Time, sdk.AccAddress], bool]
}

func NewAccountExtension(d Dependencies, reg implementation.ProtoMsgHandlerRegistry) (*Feegrant, error) {
	f := &Feegrant{
		env:               d.Environment,
		addressCodec:      d.AddressCodec,
		feeAllowance:      collections.NewMap(d.SchemaBuilder, PrefixGrants, "fee_allowance", sdk.AccAddressKey, codec.CollValue[v1.Grant](d.LegacyStateCodec)),
		feeAllowanceQueue: collections.NewMap(d.SchemaBuilder, PrefixGrantsExpiry, "fee_allowance_expiry", collections.PairKeyCodec(sdk.TimeKey, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), collections.BoolValue),
	}
	implementation.RegisterHandler(reg, f.Init)
	implementation.RegisterHandler(reg, f.GrantAllowance)
	implementation.RegisterHandler(reg, f.UseGrantedFees)
	implementation.RegisterHandler(reg, f.QueryAllowance)
	return f, nil
}

func (f Feegrant) Init(ctx context.Context, msg *v1.MsgInit) (*v1.MsgInitResponse, error) {
	return &v1.MsgInitResponse{}, nil
}

func (f *Feegrant) GrantAllowance(ctx context.Context, msg *v1.MsgGrantAllowance) (*v1.MsgGrantAllowanceResponse, error) {
	if !accountstd.SenderIsSelf(ctx) {
		return nil, errors.New("unauthorized sender")
	}
	grantee, err := f.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}

	if f, _ := f.GetAllowance(ctx, grantee); f != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "fee allowance already exists")
	}
	granterStr, err := f.addressCodec.BytesToString(accountstd.Whoami(ctx))
	if err != nil {
		return nil, err
	}

	allowance, err := msg.GetFeeAllowanceI()
	if err != nil {
		return nil, err
	}

	if err := f.storeGrant(ctx, grantee, allowance); err != nil {
		return nil, err
	}

	granteeStr := msg.Grantee
	return &v1.MsgGrantAllowanceResponse{}, f.env.EventService.EventManager(ctx).EmitKV(
		xfeegrant.EventTypeSetFeeGrant,
		event.NewAttribute(xfeegrant.AttributeKeyGranter, granterStr),
		event.NewAttribute(xfeegrant.AttributeKeyGrantee, granteeStr),
	)
}

func (f *Feegrant) storeGrant(ctx context.Context, grantee []byte, allowance xfeegrant.FeeAllowanceI) error {
	exp, err := allowance.ExpiresAt()
	if err != nil {
		return err
	}

	// expiration shouldn't be in the past.
	now := f.env.HeaderService.HeaderInfo(ctx).Time
	if exp != nil && exp.Before(now) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "expiration is before current block time")
	}

	// if expiry is not nil, add the new key to pruning queue.
	if exp != nil {
		err = f.feeAllowanceQueue.Set(ctx, collections.Join(*exp, sdk.AccAddress(grantee)), true)
		if err != nil {
			return err
		}
	}

	// if block time is not zero, update the period reset
	// if it is zero, it could be genesis initialization, so we don't need to update the period reset
	if !now.IsZero() {
		if err = allowance.UpdatePeriodReset(now); err != nil {
			return err
		}
	}

	grant, err := v1.NewGrant(allowance)
	if err != nil {
		return err
	}
	if err := f.feeAllowance.Set(ctx, grantee, *grant); err != nil {
		return err
	}
	return nil
}

func (f Feegrant) GetAllowance(ctx context.Context, grantee sdk.AccAddress) (xfeegrant.FeeAllowanceI, error) {
	grant, err := f.feeAllowance.Get(ctx, grantee)
	if err != nil {
		return nil, err
	}

	return grant.GetFeeAllowanceI()
}

func (f Feegrant) UseGrantedFees(ctx context.Context, msg *v1.MsgUseGrantedFees) (*v1.MsgUseGrantedFeesResponse, error) {
	// todo: who should be authorized? x/feegrant module? or make this a setup param?
	//if !accountstd.SenderIsAccountsModule(ctx) {
	//	return nil, errors.New("unauthorized: only accounts module is allowed to call this")
	//}

	msgs, err := msg.GetMessages()
	if err != nil {
		return nil, err
	}
	grantee, err := f.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}
	err = f.doUseGrantedFees(ctx, grantee, msg.GetFees(), msgs)
	if err != nil {
		return nil, err
	}
	return &v1.MsgUseGrantedFeesResponse{}, nil
}

func (f Feegrant) doUseGrantedFees(ctx context.Context, grantee sdk.AccAddress, fees sdk.Coins, msgs []sdk.Msg) error {
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

func (f Feegrant) QueryAllowance(ctx context.Context, msg *v1.QueryAllowanceRequest) (*v1.QueryAllowanceResponse, error) {
	grantee, err := f.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}
	allowance, err := f.feeAllowance.Get(ctx, grantee)
	if err != nil {
		return nil, err
	}
	return &v1.QueryAllowanceResponse{
		Allowance: allowance.GetAllowance(),
	}, nil
}

// MigrateFromLegacy migrate data for the account from x/feegrant module
func (f *Feegrant) MigrateFromLegacy(ctx context.Context) error {
	if !accountstd.SenderIsAccountsModule(ctx) { // not really needed but better safe than sorry
		return errors.New("unauthorized: only accounts module is allowed to call this")
	}
	granter, err := f.addressCodec.BytesToString(accountstd.Whoami(ctx))
	if err != nil {
		return err
	}

	sender, err := f.addressCodec.BytesToString(accountstd.Sender(ctx))
	if err != nil {
		return err
	}
	_ = sender
	resp, err := accountstd.ExecModule[*xfeegrant.MsgMigrateAllowancesResponse](ctx, &xfeegrant.MsgMigrateAllowances{
		Granter: granter,
	})
	if err != nil {
		return err
	}
	for _, g := range resp.Grants {
		grantee, err := f.addressCodec.StringToBytes(g.Grantee)
		if err != nil {
			return err
		}
		allowance, err := g.GetGrant()
		if err != nil {
			return err
		}

		if err := f.storeGrant(ctx, grantee, allowance); err != nil {
			return err
		}
	}
	// todo: any events to emit?
	return nil
}
