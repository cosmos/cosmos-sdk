package simsx

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

type WeightSource interface {
	Get(name string, defaultValue uint32) uint32
}
type WeightSourceFn func(name string, defaultValue uint32) uint32

func (f WeightSourceFn) Get(name string, defaultValue uint32) uint32 {
	return f(name, defaultValue)
}

func ParamWeightSource(p simtypes.AppParams) WeightSource {
	return WeightSourceFn(func(name string, defaultValue uint32) uint32 {
		var result uint32
		p.GetOrGenerate("op_weight_"+name, &result, nil, func(_ *rand.Rand) {
			result = defaultValue
		})
		return result
	})
}
