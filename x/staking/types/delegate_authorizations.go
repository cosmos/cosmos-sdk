package types

import (
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
func NewDelegateAuthorization(allowed []sdk.ValAddress, denied []sdk.ValAddress, amount *sdk.Coin) *DelegateAuthorization {
	allowedValidators := make([]string, len(allowed))
	authorization := DelegateAuthorization{}
	for i, validator := range allowed {
		allowedValidators[i] = validator.String()
	}
	authorization.AllowList = allowedValidators

	deniedValidators := make([]string, len(denied))
	for i, validator := range denied {
		deniedValidators[i] = validator.String()
	}
	authorization.DenyList = deniedValidators

	if amount != nil {
		authorization.MaxTokens = amount
	}

	return &authorization
}

// MethodName implements Authorization.MethodName.
func (authorization DelegateAuthorization) MethodName() string {
	return "/cosmos.staking.v1beta1.Msg/Delegate"
}

// Accept implements Authorization.Accept.
func (authorization DelegateAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated authz.Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&MsgDelegate{}) {
		msg, ok := msg.Request.(*MsgDelegate)
		if ok {

			for _, validator := range authorization.DenyList {
				if validator == msg.ValidatorAddress {
					return nil, false, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, " cannot delegate to %s validator", validator)
				}
			}

			isValidatorExists := false
			for _, validator := range authorization.AllowList {
				if validator == msg.ValidatorAddress {
					isValidatorExists = true
					break
				}
			}

			if !isValidatorExists {
				return nil, false, sdkerrors.Wrapf(sdkerrors.ErrNotFound, " validator not found")
			}

			if authorization.MaxTokens == nil {
				return &DelegateAuthorization{AllowList: authorization.AllowList, DenyList: authorization.DenyList}, false, nil
			}

			limitLeft := authorization.MaxTokens.Sub(msg.Amount)
			if limitLeft.IsZero() {
				return nil, true, nil
			}

			return &DelegateAuthorization{AllowList: authorization.AllowList, DenyList: authorization.DenyList, MaxTokens: &limitLeft}, false, nil
		}
	}

	return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type mismatch")
}
