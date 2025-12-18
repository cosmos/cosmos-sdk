package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// UpdateParticipationEMA updates the governance participation EMA
func (keeper Keeper) UpdateParticipationEMA(ctx context.Context, proposal v1.Proposal, participation math.LegacyDec) {
	formula := func(oldValue, newValue math.LegacyDec) math.LegacyDec {
		return oldValue.Mul(math.LegacyNewDecWithPrec(8, 1)).Add(newValue.Mul(math.LegacyNewDecWithPrec(2, 1)))
	}

	kinds := keeper.ProposalKinds(proposal)
	if kinds.HasKindConstitutionAmendment() {
		current, err := keeper.ConstitutionAmendmentParticipationEMA.Get(ctx)
		if err != nil {
			panic(err)
		}

		if err := keeper.ConstitutionAmendmentParticipationEMA.Set(ctx, formula(current, participation)); err != nil {
			panic(err)
		}
	}
	if kinds.HasKindLaw() {
		current, err := keeper.LawParticipationEMA.Get(ctx)
		if err != nil {
			panic(err)
		}

		if err := keeper.LawParticipationEMA.Set(ctx, formula(current, participation)); err != nil {
			panic(err)
		}
	}
	if kinds.HasKindAny() {
		current, err := keeper.ParticipationEMA.Get(ctx)
		if err != nil {
			panic(err)
		}

		if err := keeper.ParticipationEMA.Set(ctx, formula(current, participation)); err != nil {
			panic(err)
		}
	}
}

// GetQuorum returns the dynamic quorum for governance proposals calculated
// based on the participation EMA
func (keeper Keeper) GetQuorum(ctx context.Context) math.LegacyDec {
	params, err := keeper.Params.Get(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to get params: %w", err))
	}

	participation, err := keeper.ParticipationEMA.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}

	minQuorum := math.LegacyMustNewDecFromStr(params.QuorumRange.Min)
	maxQuorum := math.LegacyMustNewDecFromStr(params.QuorumRange.Max)
	return computeQuorum(participation, minQuorum, maxQuorum)
}

// GetConstitutionAmendmentQuorum returns the dynamic quorum for constitution
// amendment governance proposals calculated based on the participation EMA
func (keeper Keeper) GetConstitutionAmendmentQuorum(ctx context.Context) math.LegacyDec {
	params, err := keeper.Params.Get(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to get params: %w", err))
	}

	participation, err := keeper.ConstitutionAmendmentParticipationEMA.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}

	minQuorum := math.LegacyMustNewDecFromStr(params.ConstitutionAmendmentQuorumRange.Min)
	maxQuorum := math.LegacyMustNewDecFromStr(params.ConstitutionAmendmentQuorumRange.Max)
	return computeQuorum(participation, minQuorum, maxQuorum)
}

// GetLawQuorum returns the dynamic quorum for law governance proposals
// calculated based on the participation EMA
func (keeper Keeper) GetLawQuorum(ctx context.Context) math.LegacyDec {
	params, err := keeper.Params.Get(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to get params: %w", err))
	}

	participation, err := keeper.LawParticipationEMA.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}

	minQuorum := math.LegacyMustNewDecFromStr(params.LawQuorumRange.Min)
	maxQuorum := math.LegacyMustNewDecFromStr(params.LawQuorumRange.Max)
	return computeQuorum(participation, minQuorum, maxQuorum)
}

// computeQuorum returns the dynamic quorum for governance proposals calculated
// based on the participation EMA, min and max quorum.
func computeQuorum(participationEma, minQuorum, maxQuorum math.LegacyDec) math.LegacyDec {
	// quorum = min_quorum + (max_quorum - min_quorum) * participationEma
	return minQuorum.Add(maxQuorum.Sub(minQuorum).Mul(participationEma))
}
