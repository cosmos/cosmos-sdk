package runtime

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"
)

// ValidatorUpdateService is a type injected by the runtime module that can be used by modules to send validator
// updates back to tendermint.
type ValidatorUpdateService interface {

	// SetValidatorUpdates sends the provided validator updates back to tendermint. It should only be called during
	// an end blocker. The current behavior makes it an error for SetValidatorUpdates to be called twice in one block.
	SetValidatorUpdates(context.Context, []abci.ValidatorUpdate)
}
