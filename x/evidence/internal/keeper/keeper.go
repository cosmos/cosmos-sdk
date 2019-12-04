package keeper

import (
	"fmt"

	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper defines the evidence module's keeper. The keeper is responsible for
// managing persistence, state transitions and query handling for the evidence
// module.
type Keeper struct {
	cdc            *codec.Codec
	storeKey       sdk.StoreKey
	paramSpace     params.Subspace
	router         types.Router
	stakingKeeper  types.StakingKeeper
	slashingKeeper types.SlashingKeeper
	codespace      sdk.CodespaceType
}

func NewKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace, codespace sdk.CodespaceType,
	stakingKeeper types.StakingKeeper, slashingKeeper types.SlashingKeeper,
) *Keeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		paramSpace:     paramSpace,
		stakingKeeper:  stakingKeeper,
		slashingKeeper: slashingKeeper,
		codespace:      codespace,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetRouter sets the Evidence Handler router for the x/evidence module. Note,
// we allow the ability to set the router after the Keeper is constructed as a
// given Handler may need access the Keeper before being constructed. The router
// may only be set once and will be sealed if it's not already sealed.
func (k *Keeper) SetRouter(rtr types.Router) {
	// It is vital to seal the Evidence Handler router as to not allow further
	// handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	if !rtr.Sealed() {
		rtr.Seal()
	}
	if k.router != nil {
		panic(fmt.Sprintf("attempting to reset router on x/%s", types.ModuleName))
	}

	k.router = rtr
}

// GetEvidenceHandler returns a registered Handler for a given Evidence type. If
// no handler exists, an error is returned.
func (k Keeper) GetEvidenceHandler(evidenceRoute string) (types.Handler, error) {
	if !k.router.HasRoute(evidenceRoute) {
		return nil, types.ErrNoEvidenceHandlerExists(k.codespace, evidenceRoute)
	}

	return k.router.GetRoute(evidenceRoute), nil
}

// SubmitEvidence attempts to match evidence against the keepers router and execute
// the corresponding registered Evidence Handler. An error is returned if no
// registered Handler exists or if the Handler fails. Otherwise, the evidence is
// persisted.
func (k Keeper) SubmitEvidence(ctx sdk.Context, evidence exported.Evidence) error {
	if _, ok := k.GetEvidence(ctx, evidence.Hash()); ok {
		return types.ErrEvidenceExists(k.codespace, evidence.Hash().String())
	}
	if !k.router.HasRoute(evidence.Route()) {
		return types.ErrNoEvidenceHandlerExists(k.codespace, evidence.Route())
	}

	handler := k.router.GetRoute(evidence.Route())
	if err := handler(ctx, evidence); err != nil {
		return types.ErrInvalidEvidence(k.codespace, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitEvidence,
			sdk.NewAttribute(types.AttributeKeyEvidenceHash, evidence.Hash().String()),
		),
	)

	k.SetEvidence(ctx, evidence)
	return nil
}

// SetEvidence sets Evidence by hash in the module's KVStore.
func (k Keeper) SetEvidence(ctx sdk.Context, evidence exported.Evidence) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEvidence)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(evidence)
	store.Set(evidence.Hash(), bz)
}

// GetEvidence retrieves Evidence by hash if it exists. If no Evidence exists for
// the given hash, (nil, false) is returned.
func (k Keeper) GetEvidence(ctx sdk.Context, hash cmn.HexBytes) (evidence exported.Evidence, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEvidence)

	bz := store.Get(hash)
	if len(bz) == 0 {
		return nil, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &evidence)
	return evidence, true
}

// IterateEvidence provides an interator over all stored Evidence objects. For
// each Evidence object, cb will be called. If the cb returns true, the iterator
// will close and stop.
func (k Keeper) IterateEvidence(ctx sdk.Context, cb func(exported.Evidence) bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEvidence)
	iterator := sdk.KVStorePrefixIterator(store, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var evidence exported.Evidence
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &evidence)

		if cb(evidence) {
			break
		}
	}
}

// GetAllEvidence returns all stored Evidence objects.
func (k Keeper) GetAllEvidence(ctx sdk.Context) (evidence []exported.Evidence) {
	k.IterateEvidence(ctx, func(e exported.Evidence) bool {
		evidence = append(evidence, e)
		return false
	})
	return evidence
}
