package keeper

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/testutil/x/staking/types"
)

type Keeper struct {
	appmodule.Environment

	cdc                   codec.BinaryCodec
	authKeeper            types.AccountKeeper
	bankKeeper            types.BankKeeper
	validatorAddressCodec addresscodec.Codec
	// ValidatorByConsensusAddress key: consAddr | value: valAddr
	ValidatorByConsensusAddress collections.Map[sdk.ConsAddress, sdk.ValAddress]
	// Delegations key: AccAddr+valAddr | value: Delegation
	Delegations collections.Map[collections.Pair[sdk.AccAddress, sdk.ValAddress], types.Delegation]
	// Validators key: valAddr | value: Validator
	Validators collections.Map[[]byte, types.Validator]
	// Params key: ParamsKeyPrefix | value: Params
	Params collections.Item[types.Params]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	env appmodule.Environment,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	validatorAddressCodec addresscodec.Codec,
) *Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)
	return &Keeper{
		Environment:           env,
		cdc:                   cdc,
		authKeeper:            ak,
		bankKeeper:            bk,
		validatorAddressCodec: validatorAddressCodec,
		Delegations: collections.NewMap(
			sb, types.DelegationKey, "delegations",
			collections.PairKeyCodec(
				sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
				sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			),
			codec.CollValue[types.Delegation](cdc),
		),
		Validators: collections.NewMap(sb, types.ValidatorsKey, "validators", sdk.LengthPrefixedBytesKey, codec.CollValue[types.Validator](cdc)), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
		// key is: 113 (it's a direct prefix)
		Params: collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}
}

// BondDenom - Bondable coin denomination
func (k Keeper) BondDenom(ctx context.Context) (string, error) {
	params, err := k.Params.Get(ctx)
	return params.BondDenom, err
}

// MsgServer

var _ types.MsgServer = Keeper{}

// CreateValidator defines a method for creating a new validator
func (k Keeper) CreateValidator(ctx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, k.HeaderService.HeaderInfo(ctx).Time,
	)

	validator := types.Validator{
		OperatorAddress:         msg.ValidatorAddress,
		ConsensusPubkey:         msg.Pubkey,
		Jailed:                  false,
		Status:                  types.Bonded,
		Tokens:                  msg.Value.Amount,
		DelegatorShares:         msg.Value.Amount.ToLegacyDec(),
		Description:             msg.Description,
		UnbondingHeight:         int64(0),
		UnbondingTime:           time.Unix(0, 0).UTC(),
		Commission:              commission,
		MinSelfDelegation:       msg.MinSelfDelegation,
		UnbondingOnHoldRefCount: 0,
	}

	delAddr := sdk.AccAddress(valAddr)

	delAddrStr, err := k.authKeeper.AddressCodec().BytesToString(delAddr)
	if err != nil {
		return nil, err
	}

	delegation := types.NewDelegation(delAddrStr, validator.OperatorAddress, msg.Value.Amount.ToLegacyDec())

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, msg.Value.Amount))
	if err := k.bankKeeper.DelegateCoinsFromAccountToModule(ctx, delAddr, types.BondedPoolName, coins); err != nil {
		return nil, err
	}

	_ = k.Validators.Set(ctx, sdk.ValAddress(valAddr), validator)

	_ = k.Delegations.Set(ctx, collections.Join(delAddr, sdk.ValAddress(valAddr)), delegation)

	if err := k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeCreateValidator,
		event.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
		event.NewAttribute(sdk.AttributeKeyAmount, msg.Value.String()),
	); err != nil {
		return nil, err
	}

	return &types.MsgCreateValidatorResponse{}, nil
}
