package baseapp

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
)

// ValidatorUpdateService is the service that runtime will provide to the module that sets validator updates.
type ValidatorUpdateService interface {
	SetValidatorUpdates(context.Context, []abci.ValidatorUpdate)
}

var _ = (ValidatorUpdateService)(&ValidatorSetUpdate{})

type ValidatorSetUpdate struct {
	updatedValidators []abci.ValidatorUpdate
}

func (v *ValidatorSetUpdate) SetValidatorUpdates(ctx context.Context, updates []abci.ValidatorUpdate) {
	v.updatedValidators = updates
}
