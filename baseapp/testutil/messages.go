package testutil

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers the test message types and their service descriptors
// with the provided interface registry. This is used for testing purposes to ensure
// proper message serialization and deserialization.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Register test message implementations with the interface registry
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCounter{},
		&MsgCounter2{},
		&MsgKeyValue{},
	)
	// Register message service descriptors for gRPC services
	msgservice.RegisterMsgServiceDesc(registry, &_Counter_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_Counter2_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_KeyValue_serviceDesc)

	// Register crypto codec interfaces
	codec.RegisterInterfaces(registry)
}

// Ensure MsgCounter implements the sdk.Msg interface
var _ sdk.Msg = &MsgCounter{}

// GetSigners returns the list of signers for MsgCounter.
// This test message has no signers.
func (msg *MsgCounter) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }

// ValidateBasic performs basic validation on MsgCounter.
// It ensures the counter is a non-negative integer.
func (msg *MsgCounter) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return errorsmod.Wrap(sdkerrors.ErrInvalidSequence, "counter should be a non-negative integer")
}

// Ensure MsgCounter2 implements the sdk.Msg interface
var _ sdk.Msg = &MsgCounter2{}

// GetSigners returns the list of signers for MsgCounter2.
// This test message has no signers.
func (msg *MsgCounter2) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }

// ValidateBasic performs basic validation on MsgCounter2.
// It ensures the counter is a non-negative integer.
func (msg *MsgCounter2) ValidateBasic() error {
	if msg.Counter >= 0 {
		return nil
	}
	return errorsmod.Wrap(sdkerrors.ErrInvalidSequence, "counter should be a non-negative integer")
}

// Ensure MsgKeyValue implements the sdk.Msg interface
var _ sdk.Msg = &MsgKeyValue{}

// GetSigners returns the list of signers for MsgKeyValue.
// It parses the signer address from the message if provided.
func (msg *MsgKeyValue) GetSigners() []sdk.AccAddress {
	if len(msg.Signer) == 0 {
		return []sdk.AccAddress{}
	}

	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Signer)}
}

// ValidateBasic performs basic validation on MsgKeyValue.
// It ensures both key and value are not nil.
func (msg *MsgKeyValue) ValidateBasic() error {
	if msg.Key == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "key cannot be nil")
	}
	if msg.Value == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "value cannot be nil")
	}
	return nil
}
