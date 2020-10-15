package types

import (
	"github.com/gogo/protobuf/proto"
)

// MsgRequest is the interface a transaction message, defined as a proto
// service method, must fulfill.
type MsgRequest interface {
	proto.Message
	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error
	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []AccAddress
}

// ServiceMsg is the struct into which an Any whose typeUrl matches a service
// method format (ex. `/cosmos.gov.Msg/SubmitProposal`) unpacks.
type ServiceMsg struct {
	// MethodName is the fully-qualified service method name.
	MethodName string
	// Request is the request payload.
	Request MsgRequest
}

var _ Msg = ServiceMsg{}

func (msg ServiceMsg) ProtoMessage()  {}
func (msg ServiceMsg) Reset()         {}
func (msg ServiceMsg) String() string { return "ServiceMsg" }

// Route implements Msg.Route method.
func (msg ServiceMsg) Route() string {
	return msg.MethodName
}

// ValidateBasic implements Msg.ValidateBasic method.
func (msg ServiceMsg) ValidateBasic() error {
	return msg.Request.ValidateBasic()
}

// GetSignBytes implements Msg.GetSignBytes method.
func (msg ServiceMsg) GetSignBytes() []byte {
	panic("ServiceMsg does not have a GetSignBytes method")
}

// GetSigners implements Msg.GetSigners method.
func (msg ServiceMsg) GetSigners() []AccAddress {
	return msg.Request.GetSigners()
}

// Type implements Msg.Type method.
func (msg ServiceMsg) Type() string {
	return msg.MethodName
}
