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

// NewUndelegateAuthorization creates a new UndlegateAuthorization object.
func NewUndelegateAuthorization(validatorsAddr []sdk.ValAddress, amount *sdk.Coin) *UndelegateAuthorization {
	validators := make([]string, len(validatorsAddr))
	for i, validator := range validatorsAddr {
		validators[i] = validator.String()
	}
	authorization := &UndelegateAuthorization{
		ValidatorAddress: validators,
	}

	if amount != nil {
		authorization.MaxTokens = amount
	}

	return authorization
}

// MethodName implements Authorization.MethodName.
func (authorization UndelegateAuthorization) MethodName() string {
	return "/cosmos.staking.v1beta1.Msg/Undelegate"
}

// Accept implements Authorization.Accept.
func (authorization UndelegateAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated authz.Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&MsgUndelegate{}) {
		msg, ok := msg.Request.(*MsgUndelegate)
		if ok {
			isValidatorExists := false

			for _, validator := range authorization.ValidatorAddress {
				if validator == msg.ValidatorAddress {
					isValidatorExists = true
					break
				}
			}

			if !isValidatorExists {
				return nil, false, sdkerrors.Wrapf(sdkerrors.ErrNotFound, " validator not found")
			}

			if authorization.MaxTokens == nil {
				return &UndelegateAuthorization{ValidatorAddress: authorization.ValidatorAddress}, false, nil
			}

			limitLeft := authorization.MaxTokens.Sub(msg.Amount)
			if limitLeft.IsZero() {
				return nil, true, nil
			}

			return &UndelegateAuthorization{ValidatorAddress: authorization.ValidatorAddress, MaxTokens: &limitLeft}, false, nil
		}
	}

	return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type mismatch")
}
