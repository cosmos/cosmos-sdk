package types

import (
	"reflect"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	_ Authorization = &DelegateAuthorization{}
)

// NewDelegateAuthorization creates a new DelegateAuthorization object.
func NewDelegateAuthorization(validatorsAddr []sdk.ValAddress, amount *sdk.Coin) *DelegateAuthorization {
	validators := make([]string, len(validatorsAddr))
	for i, validator := range validatorsAddr {
		validators[i] = validator.String()
	}
	authorization := &DelegateAuthorization{
		ValidatorAddress: validators,
	}

	if amount != nil {
		authorization.MaxTokens = amount
	}

	return authorization
}

// MethodName implements Authorization.MethodName.
func (authorization DelegateAuthorization) MethodName() string {
	return "/cosmos.staking.v1beta1.Msg/Delegate"
}

// Accept implements Authorization.Accept.
func (authorization DelegateAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&staking.MsgDelegate{}) {
		msg, ok := msg.Request.(*staking.MsgDelegate)
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
				return &DelegateAuthorization{ValidatorAddress: authorization.ValidatorAddress}, false, nil
			}

			limitLeft := authorization.MaxTokens.Sub(msg.Amount)
			if limitLeft.IsZero() {
				return nil, true, nil
			}

			return &DelegateAuthorization{ValidatorAddress: authorization.ValidatorAddress, MaxTokens: &limitLeft}, false, nil
		}
	}

	return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type mismatch")
}
