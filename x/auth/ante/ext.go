package ante

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type HasExtensionOptionsTx interface {
	GetExtensionOptions() []*codectypes.Any
	GetNonCriticalExtensionOptions() []*codectypes.Any
}

// RejectExtensionOptionsDecorator is an AnteDecorator that rejects all extension
// options which can optionally be included in protobuf transactions. Users that
// need extension options should create a custom AnteHandler chain that handles
// needed extension options properly and rejects unknown ones.
type RejectExtensionOptionsDecorator struct{}

// NewRejectExtensionOptionsDecorator creates a new RejectExtensionOptionsDecorator
func NewRejectExtensionOptionsDecorator() RejectExtensionOptionsDecorator {
	return RejectExtensionOptionsDecorator{}
}

var _ types.AnteDecorator = RejectExtensionOptionsDecorator{}

// AnteHandle implements the AnteDecorator.AnteHandle method
func (r RejectExtensionOptionsDecorator) AnteHandle(ctx types.Context, tx types.Tx, simulate bool, next types.AnteHandler) (newCtx types.Context, err error) {
	if hasExtOptsTx, ok := tx.(HasExtensionOptionsTx); ok {
		if len(hasExtOptsTx.GetExtensionOptions()) != 0 {
			return ctx, sdkerrors.ErrUnknownExtensionOptions
		}
	}

	return next(ctx, tx, simulate)
}
