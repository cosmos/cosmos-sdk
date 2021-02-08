package types

import (
	"reflect"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz/exported"
)

var (
	_ authz.Authorization = &UndelegateAuthorization{}
)

func validateAndBech32fy(allowed []sdk.ValAddress, denied []sdk.ValAddress) ([]string, []string, error) {
	if len(allowed) == 0 && len(denied) == 0 {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "both allowed & deny list cannot be empty")
	}

	if len(allowed) > 0 && len(denied) > 0 {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "cannot set both allowed & deny list")
	}

	allowedValidators := make([]string, len(allowed))
	if len(allowed) > 0 {
		for i, validator := range allowed {
			allowedValidators[i] = validator.String()
		}
		return allowedValidators, nil, nil
	}

	deniedValidators := make([]string, len(denied))
	for i, validator := range denied {
		deniedValidators[i] = validator.String()
	}

	return nil, deniedValidators, nil
}

// NewUndelegateAuthorization creates a new UndlegateAuthorization object.
func NewUndelegateAuthorization(allowed []sdk.ValAddress, denied []sdk.ValAddress, amount *sdk.Coin) (*UndelegateAuthorization, error) {
	authorization := UndelegateAuthorization{}

	allowedValidators, deniedValidators, err := validateAndBech32fy(allowed, denied)
	if err != nil {
		return nil, err
	}

	if allowedValidators != nil {
		authorization.Validators = &UndelegateAuthorization_AllowList{AllowList: &UndelegateAuthorization_Validators{Address: allowedValidators}}
	} else {
		authorization.Validators = &UndelegateAuthorization_DenyList{DenyList: &UndelegateAuthorization_Validators{Address: deniedValidators}}
	}

	if amount != nil {
		authorization.MaxTokens = amount
	}

	return &authorization, nil
}

// MethodName implements Authorization.MethodName.
func (authorization UndelegateAuthorization) MethodName() string {
	return "/cosmos.staking.v1beta1.Msg/Undelegate"
}

// Accept implements Authorization.Accept.
func (authorization UndelegateAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated authz.Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&MsgUndelegate{}) {
		msg, ok := msg.Request.(*MsgUndelegate)
		if !ok {
			return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", MsgUndelegate{}, msg)
		}

		isValidatorExists := false
		switch x := authorization.Validators.(type) {
		case *UndelegateAuthorization_AllowList:
			allowedList := x.AllowList.GetAddress()
			for _, validator := range allowedList {
				if validator == msg.ValidatorAddress {
					isValidatorExists = true
					break
				}
			}
		case *UndelegateAuthorization_DenyList:
			denyList := x.DenyList.GetAddress()
			for _, validator := range denyList {
				if validator == msg.ValidatorAddress {
					return nil, false, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, " cannot undelegate from %s validator", validator)
				}
			}
		default:
			return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "authorization has unexpected type %T", x)
		}

		if !isValidatorExists {
			return nil, false, sdkerrors.Wrapf(sdkerrors.ErrNotFound, " validator not found")
		}

		if authorization.MaxTokens == nil {
			return &UndelegateAuthorization{Validators: authorization.Validators}, false, nil
		}

		limitLeft := authorization.MaxTokens.Sub(msg.Amount)
		if limitLeft.IsZero() {
			return nil, true, nil
		}

		return &UndelegateAuthorization{Validators: authorization.Validators, MaxTokens: &limitLeft}, false, nil
	}

	return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", MsgUndelegate{}, msg)
}
