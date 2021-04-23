package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

// TODO: Revisit this once we have propoer gas fee framework.
// Tracking issues https://github.com/cosmos/cosmos-sdk/issues/9054, https://github.com/cosmos/cosmos-sdk/discussions/9072
const (
	gasCostPerIteration = uint64(10)
)

var _ FeeAllowanceI = (*AllowedMsgFeeAllowance)(nil)
var _ types.UnpackInterfacesMessage = (*AllowedMsgFeeAllowance)(nil)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (a *AllowedMsgFeeAllowance) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var allowance FeeAllowanceI
	return unpacker.UnpackAny(a.Allowance, &allowance)
}

// NewAllowedMsgFeeAllowance creates new filtered fee allowance.
func NewAllowedMsgFeeAllowance(allowance FeeAllowanceI, allowedMsgs []string) (*AllowedMsgFeeAllowance, error) {
	msg, ok := allowance.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return &AllowedMsgFeeAllowance{
		Allowance:       any,
		AllowedMessages: allowedMsgs,
	}, nil
}

// GetAllowance returns allowed fee allowance.
func (a *AllowedMsgFeeAllowance) GetAllowance() (FeeAllowanceI, error) {
	allowance, ok := a.Allowance.GetCachedValue().(FeeAllowanceI)
	if !ok {
		return nil, sdkerrors.Wrap(ErrNoAllowance, "failed to get allowance")
	}

	return allowance, nil
}

// Accept method checks for the filtered messages has valid expiry
func (a *AllowedMsgFeeAllowance) Accept(ctx sdk.Context, fee sdk.Coins, msgs []sdk.Msg) (bool, error) {
	if !a.allMsgTypesAllowed(ctx, msgs) {
		return false, sdkerrors.Wrap(ErrMessageNotAllowed, "message does not exist in allowed messages")
	}

	allowance, err := a.GetAllowance()
	if err != nil {
		return false, err
	}

	return allowance.Accept(ctx, fee, msgs)
}

func (a *AllowedMsgFeeAllowance) allowedMsgsToMap(ctx sdk.Context) map[string]bool {
	msgsMap := make(map[string]bool, len(a.AllowedMessages))
	for _, msg := range a.AllowedMessages {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "check msg")
		msgsMap[msg] = true
	}

	return msgsMap
}

func (a *AllowedMsgFeeAllowance) allMsgTypesAllowed(ctx sdk.Context, msgs []sdk.Msg) bool {
	msgsMap := a.allowedMsgsToMap(ctx)

	for _, msg := range msgs {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "check msg")
		if !msgsMap[msg.Type()] {
			return false
		}
	}

	return true
}

// PrepareForExport will adjust the expiration based on export time. In particular,
// it will subtract the dumpHeight from any height-based expiration to ensure that
// the elapsed number of blocks this allowance is valid for is fixed.
func (a *AllowedMsgFeeAllowance) PrepareForExport(dumpTime time.Time, dumpHeight int64) FeeAllowanceI {
	allowance, err := a.GetAllowance()
	if err != nil {
		panic("failed to get allowance")
	}

	f, err := NewAllowedMsgFeeAllowance(allowance.PrepareForExport(dumpTime, dumpHeight), a.AllowedMessages)
	if err != nil {
		panic("failed to export filtered fee allowance")
	}

	return f
}

// ValidateBasic implements FeeAllowance and enforces basic sanity checks
func (a *AllowedMsgFeeAllowance) ValidateBasic() error {
	if a.Allowance == nil {
		return sdkerrors.Wrap(ErrNoAllowance, "allowance should not be empty")
	}
	if len(a.AllowedMessages) == 0 {
		return sdkerrors.Wrap(ErrNoMessages, "allowed messages shouldn't be empty")
	}

	allowance, err := a.GetAllowance()
	if err != nil {
		return err
	}

	return allowance.ValidateBasic()
}
