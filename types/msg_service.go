package types

import (
	"github.com/gogo/protobuf/proto"
)

// MsgRequest is the interface a transaction message, defined as a proto
// service method, must fulfill.
type MsgRequest interface {
	proto.Message
	ValidateBasic() error
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
