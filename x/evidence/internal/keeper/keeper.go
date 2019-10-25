package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper defines the evidence module's keeper. The keeper is responsible for
// managing persistence, state transitions and query handling for the evidence
// module.
type Keeper struct {
	cdc        *codec.Codec
	storeKey   sdk.StoreKey
	paramSpace params.Subspace
	router     types.Router
	codespace  sdk.CodespaceType
}

func NewKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace, rtr types.Router,
	codespace sdk.CodespaceType,
) Keeper {

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace,
		router:     rtr,
		codespace:  codespace,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SubmitEvidence attempts to match evidence against the keepers router and execute
// the corresponding registered Evidence Handler. An error is returned if no
// registered Handler exists or if the Handler fails. Otherwise, the evidence is
// persisted.
func (k Keeper) SubmitEvidence(ctx sdk.Context, evidence types.Evidence) error {
	if !k.router.HasRoute(evidence.Route()) {
		return types.ErrNoEvidenceHandlerExists(k.codespace, evidence.Route())
	}

	handler := k.router.GetRoute(evidence.Route())
	if err := handler(ctx, evidence); err != nil {
		return types.ErrInvalidEvidence(k.codespace, err)
	}

	return nil
}
