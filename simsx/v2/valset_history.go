package v2

import (
	"math/rand"
	"slices"
	"time"

	"cosmossdk.io/core/comet"

	"github.com/cosmos/cosmos-sdk/simsx"
)

type historicValSet struct {
	blockTime time.Time
	vals      WeightedValidators
}
type ValSetHistory struct {
	maxElements int
	blockOffset uint64
	vals        []historicValSet
}

// NewValSetHistory constructor. The maximum of historic valsets must not exceed the block or time limit for
// valid evidence.
func NewValSetHistory(initialHeight uint64) *ValSetHistory {
	return &ValSetHistory{
		maxElements: 1,
		blockOffset: initialHeight,
		vals:        make([]historicValSet, 0, 1),
	}
}

func (h *ValSetHistory) Add(blockTime time.Time, vals WeightedValidators) {
	vals = slices.DeleteFunc(vals, func(validator WeightedValidator) bool {
		return validator.Power == 0
	})
	slices.SortFunc(vals, func(a, b WeightedValidator) int {
		return b.Compare(a)
	})
	newEntry := historicValSet{blockTime: blockTime, vals: vals}
	if len(h.vals) >= h.maxElements {
		h.vals = append(h.vals[1:], newEntry)
		h.blockOffset++
		return
	}
	h.vals = append(h.vals, newEntry)
}

// MissBehaviour determines if a random validator misbehaves, creating and returning evidence for duplicate voting.
// Returns a slice of comet.Evidence if misbehavior is detected; otherwise, returns nil.
// Has a 1% chance of generating evidence for a validator's misbehavior.
// Recursively checks for other misbehavior instances and combines their evidence if any.
// Utilizes a random generator to select a validator and evidence-related attributes.
func (h *ValSetHistory) MissBehaviour(r *rand.Rand) []comet.Evidence {
	if r.Intn(100) != 0 { // 1% chance
		return nil
	}
	n := r.Intn(len(h.vals))
	badVal := simsx.OneOf(r, h.vals[n].vals)
	evidence := comet.Evidence{
		Type:             comet.DuplicateVote,
		Validator:        comet.Validator{Address: badVal.Address, Power: badVal.Power},
		Height:           int64(h.blockOffset) + int64(n),
		Time:             h.vals[n].blockTime,
		TotalVotingPower: h.vals[n].vals.TotalPower(),
	}
	if otherEvidence := h.MissBehaviour(r); otherEvidence != nil {
		return append([]comet.Evidence{evidence}, otherEvidence...)
	}
	return []comet.Evidence{evidence}
}

// SetMaxHistory sets the maximum number of historical validator sets to retain. Reduces retained history if it exceeds the limit.
func (h *ValSetHistory) SetMaxHistory(v int) {
	h.maxElements = v
	if len(h.vals) > h.maxElements {
		diff := len(h.vals) - h.maxElements
		h.vals = h.vals[diff:]
		h.blockOffset += uint64(diff)
	}
}
