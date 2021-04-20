package exported

import (
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Authorization represents the interface of various Authorization types.
type Authorization interface {
	proto.Message

	// MethodName returns the fully-qualified Msg service method name as described in ADR 031.
	MethodName() string

	// Accept determines whether this grant permits the provided sdk.ServiceMsg to be performed, and if
	// so provides an upgraded authorization instance.
	Accept(ctx sdk.Context, msg sdk.ServiceMsg) (updated Authorization, delete bool, err error)

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error
}
