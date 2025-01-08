package v2

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/simsx"
)

// NextFactoryFn shuffles and processes a list of weighted factories, returning a selection function for factory objects.
func NextFactoryFn(factories []simsx.WeightedFactory, r *rand.Rand) func() simsx.SimMsgFactoryX {
	factCount := len(factories)
	r.Shuffle(factCount, func(i, j int) {
		factories[i], factories[j] = factories[j], factories[i]
	})
	var totalWeight int
	for _, f := range factories {
		totalWeight += int(f.Weight)
	}
	return func() simsx.SimMsgFactoryX {
		// this is copied from old sims WeightedOperations.getSelectOpFn
		x := r.Intn(totalWeight)
		for i := 0; i < factCount; i++ {
			if x <= int(factories[i].Weight) {
				return factories[i].Factory
			}
			x -= int(factories[i].Weight)
		}
		// shouldn't happen
		return factories[0].Factory
	}
}
