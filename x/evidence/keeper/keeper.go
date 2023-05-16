package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/x/evidence/exported"
	"cosmossdk.io/x/evidence/types"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines the evidence module's keeper. The keeper is responsible for
// managing persistence, state transitions and query handling for the evidence
// module.
type Keeper struct {
	cdc            codec.BinaryCodec
	storeService   store.KVStoreService
	router         types.Router
	stakingKeeper  types.StakingKeeper
	slashingKeeper types.SlashingKeeper
	addressCodec   address.Codec

	cometInfo comet.BlockInfoService
}

// NewKeeper creates a new Keeper object.
func NewKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService, stakingKeeper types.StakingKeeper,
	slashingKeeper types.SlashingKeeper, ac address.Codec, ci comet.BlockInfoService,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeService:   storeService,
		stakingKeeper:  stakingKeeper,
		slashingKeeper: slashingKeeper,
		addressCodec:   ac,
		cometInfo:      ci,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
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
		return nil, errors.Wrap(types.ErrNoEvidenceHandlerExists, evidenceRoute)
	}

	return k.router.GetRoute(evidenceRoute), nil
}

// SubmitEvidence attempts to match evidence against the keepers router and execute
// the corresponding registered Evidence Handler. An error is returned if no
// registered Handler exists or if the Handler fails. Otherwise, the evidence is
// persisted.
func (k Keeper) SubmitEvidence(ctx context.Context, evidence exported.Evidence) error {
	if _, err := k.GetEvidence(ctx, evidence.Hash()); err == nil {
		return errors.Wrap(types.ErrEvidenceExists, strings.ToUpper(hex.EncodeToString(evidence.Hash())))
	}
	if !k.router.HasRoute(evidence.Route()) {
		return errors.Wrap(types.ErrNoEvidenceHandlerExists, evidence.Route())
	}

	handler := k.router.GetRoute(evidence.Route())
	if err := handler(ctx, evidence); err != nil {
		return errors.Wrap(types.ErrInvalidEvidence, err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitEvidence,
			sdk.NewAttribute(types.AttributeKeyEvidenceHash, strings.ToUpper(hex.EncodeToString(evidence.Hash()))),
		),
	)

	return k.SetEvidence(ctx, evidence)
}

// SetEvidence sets Evidence by hash in the module's KVStore.
func (k Keeper) SetEvidence(ctx context.Context, evidence exported.Evidence) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixEvidence)
	prefixStore.Set(evidence.Hash(), k.MustMarshalEvidence(evidence))
	return nil
}

// GetEvidence retrieves Evidence by hash if it exists. If no Evidence exists for
// the given hash, (nil, types.ErrNoEvidenceExists) is returned.
func (k Keeper) GetEvidence(ctx context.Context, hash []byte) (exported.Evidence, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.EvidenceKey(hash))
	if len(bz) == 0 {
		return nil, types.ErrNoEvidenceExists
	}

	if err != nil {
		return nil, err
	}

	return k.UnmarshalEvidence(bz)
}

// IterateEvidence provides an interator over all stored Evidence objects. For
// each Evidence object, cb will be called. If the cb returns true, the iterator
// will close and stop.
func (k Keeper) IterateEvidence(ctx context.Context, cb func(exported.Evidence) error) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.KeyPrefixEvidence, storetypes.PrefixEndBytes(types.KeyPrefixEvidence))
	if err != nil {
		return err
	}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		evidence, err := k.UnmarshalEvidence(iterator.Value())
		if err != nil {
			return err
		}

		err = cb(evidence)
		if errors.IsOf(err, errors.ErrStopIterating) {
			return nil
		} else if err != nil {
			return err
		}
	}
	return nil
}

// GetAllEvidence returns all stored Evidence objects.
func (k Keeper) GetAllEvidence(ctx context.Context) (evidence []exported.Evidence) {
	k.IterateEvidence(ctx, func(e exported.Evidence) error {
		evidence = append(evidence, e)
		return nil
	})

	return evidence
}

// MustUnmarshalEvidence attempts to decode and return an Evidence object from
// raw encoded bytes. It panics on error.
func (k Keeper) MustUnmarshalEvidence(bz []byte) exported.Evidence {
	evidence, err := k.UnmarshalEvidence(bz)
	if err != nil {
		panic(fmt.Errorf("failed to decode evidence: %w", err))
	}

	return evidence
}

// MustMarshalEvidence attempts to encode an Evidence object and returns the
// raw encoded bytes. It panics on error.
func (k Keeper) MustMarshalEvidence(evidence exported.Evidence) []byte {
	bz, err := k.MarshalEvidence(evidence)
	if err != nil {
		panic(fmt.Errorf("failed to encode evidence: %w", err))
	}

	return bz
}

// MarshalEvidence protobuf serializes an Evidence interface
func (k Keeper) MarshalEvidence(evidenceI exported.Evidence) ([]byte, error) {
	return k.cdc.MarshalInterface(evidenceI)
}

// UnmarshalEvidence returns an Evidence interface from raw encoded evidence
// bytes of a Proto-based Evidence type
func (k Keeper) UnmarshalEvidence(bz []byte) (exported.Evidence, error) {
	var evi exported.Evidence
	return evi, k.cdc.UnmarshalInterface(bz, &evi)
}
