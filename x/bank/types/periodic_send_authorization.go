package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

var (
	_ authz.Authorization = &PeriodicSendAuthorization{}
)

// NewPeriodicSendAuthorization creates a new PeriodicSendAuthorization object.
func NewPeriodicSendAuthorization(periodicAllowance feegrant.PeriodicAllowance, spendLimit sdk.Coins) *PeriodicSendAuthorization {
	return &PeriodicSendAuthorization{
		PeriodicAllowance: periodicAllowance,
		SpendLimit:        spendLimit,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a PeriodicSendAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgSend{})
}

// Accept implements Authorization.Accept.
func (a PeriodicSendAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	mSend, ok := msg.(*MsgSend)
	if !ok {
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
	}
	remove, err := a.PeriodicAllowance.Accept(ctx, mSend.Amount, nil)
	fmt.Println(remove)
	if err != nil {
		fmt.Println(err)
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("some error")
	}

	return authz.AcceptResponse{Accept: true, Delete: false, Updated: &a}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a PeriodicSendAuthorization) ValidateBasic() error {
	err := a.PeriodicAllowance.ValidateBasic()
	if err != nil {
		fmt.Println(err)
		return sdkerrors.ErrInvalidType.Wrap("some error")
	}
	return nil
}
