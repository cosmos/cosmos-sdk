package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/contrib/x/evidence/exported"
	types2 "github.com/cosmos/cosmos-sdk/contrib/x/evidence/types"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/store"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines the evidence module's keeper. The keeper is responsible for
// managing persistence, state transitions and query handling for the evidence
// module.
type Keeper struct {
	cdc            codec.BinaryCodec
	storeService   store.KVStoreService
	router         types2.Router
	stakingKeeper  types2.StakingKeeper
	slashingKeeper types2.SlashingKeeper
	addressCodec   address.Codec

	cometInfo comet.BlockInfoService

	Schema    collections.Schema
	Evidences collections.Map[[]byte, exported.Evidence]
}

// NewKeeper creates a new Keeper object.
func NewKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService, stakingKeeper types2.StakingKeeper,
	slashingKeeper types2.SlashingKeeper, ac address.Codec, ci comet.BlockInfoService,
) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := &Keeper{
		cdc:            cdc,
		storeService:   storeService,
		stakingKeeper:  stakingKeeper,
		slashingKeeper: slashingKeeper,
		addressCodec:   ac,
		cometInfo:      ci,
		Evidences:      collections.NewMap(sb, types2.KeyPrefixEvidence, "evidences", collections.BytesKey, codec.CollInterfaceValue[exported.Evidence](cdc)),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types2.ModuleName)
}

// SetRouter sets the Evidence Handler router for the x/evidence module. Note,
// we allow the ability to set the router after the Keeper is constructed as a
// given Handler may need access the Keeper before being constructed. The router
// may only be set once and will be sealed if it's not already sealed.
func (k *Keeper) SetRouter(rtr types2.Router) {
	// It is vital to seal the Evidence Handler router as to not allow further
	// handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	if !rtr.Sealed() {
		rtr.Seal()
	}
	if k.router != nil {
		panic(fmt.Sprintf("attempting to reset router on x/%s", types2.ModuleName))
	}

	k.router = rtr
}

// GetEvidenceHandler returns a registered Handler for a given Evidence type. If
// no handler exists, an error is returned.
func (k Keeper) GetEvidenceHandler(evidenceRoute string) (types2.Handler, error) {
	if !k.router.HasRoute(evidenceRoute) {
		return nil, errors.Wrap(types2.ErrNoEvidenceHandlerExists, evidenceRoute)
	}

	return k.router.GetRoute(evidenceRoute), nil
}

// SubmitEvidence attempts to match evidence against the keepers router and execute
// the corresponding registered Evidence Handler. An error is returned if no
// registered Handler exists or if the Handler fails. Otherwise, the evidence is
// persisted.
func (k Keeper) SubmitEvidence(ctx context.Context, evidence exported.Evidence) error {
	if _, err := k.Evidences.Get(ctx, evidence.Hash()); err == nil {
		return errors.Wrap(types2.ErrEvidenceExists, strings.ToUpper(hex.EncodeToString(evidence.Hash())))
	}
	if !k.router.HasRoute(evidence.Route()) {
		return errors.Wrap(types2.ErrNoEvidenceHandlerExists, evidence.Route())
	}

	handler := k.router.GetRoute(evidence.Route())
	if err := handler(ctx, evidence); err != nil {
		return errors.Wrap(types2.ErrInvalidEvidence, err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types2.EventTypeSubmitEvidence,
			sdk.NewAttribute(types2.AttributeKeyEvidenceHash, strings.ToUpper(hex.EncodeToString(evidence.Hash()))),
		),
	)

	return k.Evidences.Set(ctx, evidence.Hash(), evidence)
}
