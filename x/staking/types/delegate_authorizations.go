package types

import (
	fmt "fmt"
	"reflect"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz/exported"
)

var (
	_ authz.Authorization = &DelegateAuthorization{}
)

// NewDelegateAuthorization creates a new DelegateAuthorization object.
func NewDelegateAuthorization(allowed []sdk.ValAddress, denied []sdk.ValAddress, amount *sdk.Coin) (*DelegateAuthorization, error) {
	allowedValidators, deniedValidators, err := validateAndBech32fy(allowed, denied)
	if err != nil {
		return nil, err
	}

	authorization := DelegateAuthorization{}
	if allowedValidators != nil {
		authorization.Validators = &DelegateAuthorization_AllowList{AllowList: &DelegateAuthorization_Validators{Address: allowedValidators}}
	} else {
		authorization.Validators = &DelegateAuthorization_DenyList{DenyList: &DelegateAuthorization_Validators{Address: deniedValidators}}
	}

	if amount != nil {
		authorization.MaxTokens = amount
	}

	return &authorization, nil
}

// MethodName implements Authorization.MethodName.
func (authorization DelegateAuthorization) MethodName() string {
	return "/cosmos.staking.v1beta1.Msg/Delegate"
}

// Accept implements Authorization.Accept.
func (authorization DelegateAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated authz.Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&MsgDelegate{}) {
		msg, ok := msg.Request.(*MsgDelegate)
		if !ok {
			return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", MsgDelegate{}, msg)
		}
		isValidatorExists := false
		switch x := authorization.Validators.(type) {
		case *DelegateAuthorization_AllowList:
			allowedList := x.AllowList.GetAddress()
			for _, validator := range allowedList {
				if validator == msg.ValidatorAddress {
					isValidatorExists = true
					break
				}
			}
		case *DelegateAuthorization_DenyList:
			denyList := x.DenyList.GetAddress()
			for _, validator := range denyList {
				if validator == msg.ValidatorAddress {
					return nil, false, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, " cannot delegate to %s validator", validator)
				}
			}
		default:
			return nil, false, fmt.Errorf("authorization has unexpected type %T", x)
		}

		if !isValidatorExists {
			return nil, false, sdkerrors.Wrapf(sdkerrors.ErrNotFound, " validator not found")
		}

		if authorization.MaxTokens == nil {
			return &DelegateAuthorization{Validators: authorization.Validators}, false, nil
		}

		limitLeft := authorization.MaxTokens.Sub(msg.Amount)
		if limitLeft.IsZero() {
			return nil, true, nil
		}

		return &DelegateAuthorization{Validators: authorization.Validators, MaxTokens: &limitLeft}, false, nil
	}
	return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", MsgDelegate{}, msg)
}
