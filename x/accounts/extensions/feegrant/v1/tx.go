package v1

import (
	"errors"

	"github.com/cosmos/gogoproto/proto"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"

	errorsmod "cosmossdk.io/errors"
	xfeegrant "cosmossdk.io/x/feegrant"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgGrantAllowance construtor
func NewMsgGrantAllowance(feeAllowance xfeegrant.FeeAllowanceI, grantee string) (*MsgGrantAllowance, error) {
	msg, ok := feeAllowance.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", msg)
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &MsgGrantAllowance{
		Grantee:   grantee,
		Allowance: any,
	}, nil
}

// GetFeeAllowanceI returns unpacked FeeAllowance
func (msg MsgGrantAllowance) GetFeeAllowanceI() (xfeegrant.FeeAllowanceI, error) {
	allowance, ok := msg.Allowance.GetCachedValue().(xfeegrant.FeeAllowanceI)
	if !ok {
		return nil, errors.New("failed to get allowance")
	}

	return allowance, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgGrantAllowance) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	var allowance xfeegrant.FeeAllowanceI
	return unpacker.UnpackAny(msg.Allowance, &allowance)
}

// NewMsgUseGrantedFees constructor
func NewMsgUseGrantedFees(grantee string, msgs ...sdk.Msg) (*MsgUseGrantedFees, error) {
	msgsAny := make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		any, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}

		msgsAny[i] = any
	}

	return &MsgUseGrantedFees{
		Grantee: grantee,
		Msgs:    msgsAny,
	}, nil
}

// GetMessages returns the cache values from the MsgUseGrantedFees.Msgs if present.
func (msg MsgUseGrantedFees) GetMessages() ([]sdk.Msg, error) {
	msgs := make([]sdk.Msg, len(msg.Msgs))
	for i, msgAny := range msg.Msgs {
		msg, ok := msgAny.GetCachedValue().(sdk.Msg)
		if !ok {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("messages contains %T which is not a sdk.Msg", msgAny)
		}
		msgs[i] = msg
	}

	return msgs, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgUseGrantedFees) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	for _, x := range msg.Msgs {
		var msgExecAuthorized sdk.Msg
		err := unpacker.UnpackAny(x, &msgExecAuthorized)
		if err != nil {
			return err
		}
	}

	return nil
}
