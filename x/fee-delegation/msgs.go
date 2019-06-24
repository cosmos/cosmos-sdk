package fee_delegation

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgDelegateFeeAllowance struct {
	Granter   sdk.AccAddress `json:"granter"`
	Grantee   sdk.AccAddress `json:"grantee"`
	Allowance FeeAllowance   `json:"allowance"`
}

func NewMsgDelegateFeeAllowance(granter sdk.AccAddress, grantee sdk.AccAddress, allowance FeeAllowance) MsgDelegateFeeAllowance {
	return MsgDelegateFeeAllowance{Granter: granter, Grantee: grantee, Allowance: allowance}
}

func (msg MsgDelegateFeeAllowance) Route() string {
	return "delegation"
}

func (msg MsgDelegateFeeAllowance) Type() string {
	return "delegate-fee-allowance"
}

func (msg MsgDelegateFeeAllowance) ValidateBasic() sdk.Error {
	return nil
}

func (msg MsgDelegateFeeAllowance) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgDelegateFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}

type MsgRevokeFeeAllowance struct {
	Granter sdk.AccAddress `json:"granter"`
	Grantee sdk.AccAddress `json:"grantee"`
}

func NewMsgRevokeFeeAllowance(granter sdk.AccAddress, grantee sdk.AccAddress) MsgRevokeFeeAllowance {
	return MsgRevokeFeeAllowance{Granter: granter, Grantee: grantee}
}

func (msg MsgRevokeFeeAllowance) Route() string {
	return "delegation"
}

func (msg MsgRevokeFeeAllowance) Type() string {
	return "revoke-fee-allowance"
}

func (msg MsgRevokeFeeAllowance) ValidateBasic() sdk.Error {
	return nil
}

func (msg MsgRevokeFeeAllowance) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgRevokeFeeAllowance) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Granter}
}
