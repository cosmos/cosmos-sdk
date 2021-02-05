package exported

import (
	"github.com/gogo/protobuf/proto"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Authorization interface {
	proto.Message

	// MethodName returns the fully-qualified Msg service method name as described in ADR 031.
	MethodName() string

	// Accept determines whether this grant permits the provided sdk.ServiceMsg to be performed, and if
	// so provides an upgraded authorization instance.
	Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated Authorization, delete bool, err error)
}
