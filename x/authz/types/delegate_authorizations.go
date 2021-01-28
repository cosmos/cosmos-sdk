package types

import (
	"reflect"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	_ Authorization = &DelegateAuthorization{}
)

// NewDelegateAuthorization creates a new DelegateAuthorization object.
func NewDelegateAuthorization(validatorAddress sdk.ValAddress, amount sdk.Coins) *DelegateAuthorization {
	return &DelegateAuthorization{
		ValidatorAddress: validatorAddress.String(),
		Amount:           amount,
	}
}

// MethodName implements Authorization.MethodName.
func (authorization DelegateAuthorization) MethodName() string {
	return "/cosmos.staking.v1beta1.Msg/Delegate"
}

// Accept implements Authorization.Accept.
func (authorization DelegateAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (allow bool, updated Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&bank.MsgSend{}) {
		msg, ok := msg.Request.(*bank.MsgSend)
		if ok {
			limitLeft, isNegative := authorization.Amount.SafeSub(msg.Amount)
			if isNegative {
				return false, nil, false, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "requested amount is more than spend limit")
			}
			if limitLeft.IsZero() {
				return true, nil, true, nil
			}

			return true, &DelegateAuthorization{Amount: limitLeft}, false, nil
		}
	}
	return false, nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type mismatch")
}
