package simulation

import (
	"math/rand"
	"time"
)

// Simulation parameter constants
const (
	UnbondingTime = "unbonding_time"
	MaxValidators = "max_validators"
)

// enParams generates random staking parameters
func GenParams(paramSims map[string]func(r *rand.Rand) interface{}) {
	paramSims[UnbondingTime] = func(r *rand.Rand) interface{} {
		return time.Duration(RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second
	}

	paramSims[MaxValidators] = func(r *rand.Rand) interface{} {
		return uint16(r.Intn(250) + 1)
	}
}
