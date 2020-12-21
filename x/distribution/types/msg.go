//nolint
package types

import (
	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"strings"
)

// distribution message types
const (
	TypeMsgSetWithdrawAddress          = "set_withdraw_address"
	TypeMsgWithdrawDelegatorReward     = "withdraw_delegator_reward"
	TypeMsgWithdrawValidatorCommission = "withdraw_validator_commission"
	TypeMsgFundCommunityPool           = "fund_community_pool"
)

// Verify interface at compile time
var _, _, _ sdk.Msg = &MsgSetWithdrawAddress{}, &MsgWithdrawDelegatorReward{}, &MsgWithdrawValidatorCommission{}

func NewMsgSetWithdrawAddress(delAddr, withdrawAddr sdk.AccAddress) *MsgSetWithdrawAddress {
	return &MsgSetWithdrawAddress{
		DelegatorAddress: delAddr.String(),
		WithdrawAddress:  withdrawAddr.String(),
	}
}

func (msg MsgSetWithdrawAddress) Route() string { return ModuleName }
func (msg MsgSetWithdrawAddress) Type() string  { return TypeMsgSetWithdrawAddress }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgSetWithdrawAddress) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgSetWithdrawAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgSetWithdrawAddress) ValidateBasic() error {
	if msg.DelegatorAddress == "" {
		return ErrEmptyDelegatorAddr
	}
	if msg.WithdrawAddress == "" {
		return ErrEmptyWithdrawAddr
	}

	return nil
}

func NewMsgWithdrawDelegatorReward(delAddr sdk.AccAddress, valAddr sdk.ValAddress) *MsgWithdrawDelegatorReward {
	return &MsgWithdrawDelegatorReward{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
	}
}

func (msg MsgWithdrawDelegatorReward) Route() string { return ModuleName }
func (msg MsgWithdrawDelegatorReward) Type() string  { return TypeMsgWithdrawDelegatorReward }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgWithdrawDelegatorReward) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgWithdrawDelegatorReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawDelegatorReward) ValidateBasic() error {
	if msg.DelegatorAddress == "" {
		return ErrEmptyDelegatorAddr
	}
	if msg.ValidatorAddress == "" {
		return ErrEmptyValidatorAddr
	}
	return nil
}

// Rosetta Msg interface.
func (msg *MsgWithdrawDelegatorReward) ToOperations(withStatus bool, hasError bool) []*rosettatypes.Operation {
	var operations []*rosettatypes.Operation
	delAddr := msg.DelegatorAddress
	valAddr := msg.ValidatorAddress
	delOp := func(account string, index int) *rosettatypes.Operation {
		var status string
		if withStatus {
			status = "Success"
			if hasError {
				status = "Reverted"
			}
		}
		return &rosettatypes.Operation{
			OperationIdentifier: &rosettatypes.OperationIdentifier{
				Index: int64(index),
			},
			Type:   proto.MessageName(msg),
			Status: status,
			Account: &rosettatypes.AccountIdentifier{
				Address: account,
			},
		}
	}
	operations = append(operations,
		delOp(delAddr, 0),
		delOp(valAddr, 1),
	)
	return operations
}

func (msg MsgWithdrawDelegatorReward) FromOperations(ops []*rosettatypes.Operation) (sdk.Msg, error) {
	var (
		delAddr sdk.AccAddress
		valAddr sdk.ValAddress
		err     error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, sdk.GetConfig().GetBech32ValidatorAddrPrefix()) {
			valAddr, err = sdk.ValAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, err
			}
			continue
		}

		delAddr, err = sdk.AccAddressFromBech32(op.Account.Address)
		if err != nil {
			return nil, err
		}
	}

	return NewMsgWithdrawDelegatorReward(delAddr, valAddr), nil
}

func NewMsgWithdrawValidatorCommission(valAddr sdk.ValAddress) *MsgWithdrawValidatorCommission {
	return &MsgWithdrawValidatorCommission{
		ValidatorAddress: valAddr.String(),
	}
}

func (msg MsgWithdrawValidatorCommission) Route() string { return ModuleName }
func (msg MsgWithdrawValidatorCommission) Type() string  { return TypeMsgWithdrawValidatorCommission }

// Return address that must sign over msg.GetSignBytes()
func (msg MsgWithdrawValidatorCommission) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{valAddr.Bytes()}
}

// get the bytes for the message signer to sign on
func (msg MsgWithdrawValidatorCommission) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgWithdrawValidatorCommission) ValidateBasic() error {
	if msg.ValidatorAddress == "" {
		return ErrEmptyValidatorAddr
	}
	return nil
}

// NewMsgFundCommunityPool returns a new MsgFundCommunityPool with a sender and
// a funding amount.
func NewMsgFundCommunityPool(amount sdk.Coins, depositor sdk.AccAddress) *MsgFundCommunityPool {
	return &MsgFundCommunityPool{
		Amount:    amount,
		Depositor: depositor.String(),
	}
}

// Route returns the MsgFundCommunityPool message route.
func (msg MsgFundCommunityPool) Route() string { return ModuleName }

// Type returns the MsgFundCommunityPool message type.
func (msg MsgFundCommunityPool) Type() string { return TypeMsgFundCommunityPool }

// GetSigners returns the signer addresses that are expected to sign the result
// of GetSignBytes.
func (msg MsgFundCommunityPool) GetSigners() []sdk.AccAddress {
	depoAddr, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depoAddr}
}

// GetSignBytes returns the raw bytes for a MsgFundCommunityPool message that
// the expected signer needs to sign.
func (msg MsgFundCommunityPool) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic MsgFundCommunityPool message validation.
func (msg MsgFundCommunityPool) ValidateBasic() error {
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}
	if msg.Depositor == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Depositor)
	}

	return nil
}
