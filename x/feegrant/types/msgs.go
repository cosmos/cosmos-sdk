package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/feegrant/exported"
)

func (msg MsgGrantFeeAllowance) NewMsgGrantFeeAllowance(feeAllowanceI exported.FeeAllowance, granter, grantee sdk.AccAddress) (MsgGrantFeeAllowance, error) {
	feeallowance := &FeeAllowance{}

	if err := feeallowance.SetFeeAllowance(feeAllowanceI); err != nil {
		return MsgGrantFeeAllowance{}, err
	}

	return MsgGrantFeeAllowance{
		Granter:   granter,
		Grantee:   grantee,
		Allowance: feeallowance,
	}, nil
}

func (msg MsgGrantFeeAllowance) GetFeeGrant() exported.FeeAllowance {
	return msg.Allowance.GetFeeAllowance()
}

func (msg MsgGrantFeeAllowance) GetGrantee() sdk.AccAddress {
	return msg.Grantee
}

func (msg MsgGrantFeeAllowance) GetGranter() sdk.AccAddress {
	return msg.Granter
}

// PrepareForExport will make all needed changes to the allowance to prepare to be
// re-imported at height 0, and return a copy of this grant.
func (a MsgGrantFeeAllowance) PrepareForExport(dumpTime time.Time, dumpHeight int64) FeeAllowanceGrant {
	err := a.GetFeeGrant().PrepareForExport(dumpTime, dumpHeight)
	if err != nil {
		//TODO handle this error
	}

	feegrant := FeeAllowanceGrant{Granter: a.Granter, Grantee: a.Grantee, Allowance: a.Allowance}
	return feegrant
}

func (msg MsgGrantFeeAllowance) Route() string {
	return RouterKey
}

func (msg MsgGrantFeeAllowance) Type() string {
	return "grant-fee-allowance"
}

func (msg MsgGrantFeeAllowance) ValidateBasic() error {
	if msg.Granter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}

	return nil
}

func (msg MsgGrantFeeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgGrantFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}

func NewMsgRevokeFeeAllowance(granter sdk.AccAddress, grantee sdk.AccAddress) MsgRevokeFeeAllowance {
	return MsgRevokeFeeAllowance{Granter: granter, Grantee: grantee}
}

func (msg MsgRevokeFeeAllowance) Route() string {
	return RouterKey
}

func (msg MsgRevokeFeeAllowance) Type() string {
	return "revoke-fee-allowance"
}

func (msg MsgRevokeFeeAllowance) ValidateBasic() error {
	if msg.Granter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing granter address")
	}
	if msg.Grantee.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing grantee address")
	}

	return nil
}

func (msg MsgRevokeFeeAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgRevokeFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}
