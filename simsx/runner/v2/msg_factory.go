package v2

import (
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"math/rand"
)

// NextFactoryFn shuffles and processes a list of weighted factories, returning a selection function for factory objects.
func NextFactoryFn(factories []WeightedFactory, r *rand.Rand) func() common.SimMsgFactoryX {
	factCount := len(factories)
	r.Shuffle(factCount, func(i, j int) {
		factories[i], factories[j] = factories[j], factories[i]
	})
	var totalWeight int
	for k := range factories {
		totalWeight += k
	}
	return func() common.SimMsgFactoryX {
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
